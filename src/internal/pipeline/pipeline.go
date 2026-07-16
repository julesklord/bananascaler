// Package pipeline orchestrates the full bananascaler processing chain:
// frame extraction → neural upscaling (with live progress) → re-encode + mux.
package pipeline

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/julesklord/bananascaler/internal/config"
	"github.com/julesklord/bananascaler/internal/hardware"
	"github.com/schollz/progressbar/v3"
)

// ── ANSI color helpers ────────────────────────────────────────────────────────

const (
	cReset  = "\033[0m"
	cBold   = "\033[1m"
	cDim    = "\033[2m"
	cRed    = "\033[31m"
	cGreen  = "\033[32m"
	cYellow = "\033[33m"
	cCyan   = "\033[36m"
)

func logInfo(msg string)  { fmt.Printf("%s%s[INFO]%s %s\n", cBold, cCyan, cReset, msg) }
func logOK(msg string)    { fmt.Printf("%s%s[ OK ]%s %s\n", cBold, cGreen, cReset, msg) }
func logWarn(msg string)  { fmt.Printf("%s%s[WARN]%s %s\n", cBold, cYellow, cReset, msg) }
func logStep(msg string)  { fmt.Printf("\n%s%s🍌 %s%s\n", cBold, cYellow, msg, cReset) }
func logErr(msg string)   { fmt.Fprintf(os.Stderr, "%s%s[ERR ]%s %s\n", cBold, cRed, cReset, msg) }

// ── Public entry point ────────────────────────────────────────────────────────

// Run executes the full bananascaler pipeline for the given Config.
// It blocks until completion, cancellation, or a fatal error.
func Run(cfg *config.Config) error {
	// Validate scale factor
	if cfg.Scale < 2 || cfg.Scale > 4 {
		return fmt.Errorf("invalid scale factor %d: must be 2, 3, or 4", cfg.Scale)
	}

	// Validate input file
	if _, err := os.Stat(cfg.Input); err != nil {
		return fmt.Errorf("input file not found: %q", cfg.Input)
	}

	// Auto-name output
	if cfg.Output == "" {
		dir := filepath.Dir(cfg.Input)
		base := strings.TrimSuffix(filepath.Base(cfg.Input), filepath.Ext(cfg.Input))
		cfg.Output = filepath.Join(dir, base+"_upscaled.mp4")
	}

	// Upfront dependency check — fail before touching any file
	if err := hardware.CheckDeps(); err != nil {
		return err
	}

	// ── Hardware detection ────────────────────────────────────────────────
	var decFlags, encFlags []string
	if hardware.HasNVIDIA() {
		logInfo("NVIDIA GPU detected — enabling NVDEC + NVENC.")
		decFlags = []string{"-hwaccel", "cuda"}
		encFlags = []string{"-c:v", "hevc_nvenc", "-pix_fmt", "yuv420p"}
	} else {
		logWarn("No NVIDIA GPU — falling back to CPU (libx265).")
		encFlags = []string{"-c:v", "libx265", "-preset", "medium", "-crf", "22", "-pix_fmt", "yuv420p"}
	}

	// ── Media probes ──────────────────────────────────────────────────────
	hasAudio, err := hardware.HasAudio(cfg.Input)
	if err != nil {
		return err
	}
	framerate, err := hardware.Framerate(cfg.Input)
	if err != nil {
		return err
	}

	if hasAudio {
		logInfo("Audio stream found — will remux without re-encoding.")
	} else {
		logWarn("No audio stream — output will be video-only.")
	}
	logInfo(fmt.Sprintf(
		"Framerate: %s%s%s fps  |  Scale: %s%d×%s  |  GPU index: %s%d%s",
		cBold, framerate, cReset,
		cBold, cfg.Scale, cReset,
		cBold, cfg.GPU, cReset,
	))
	fmt.Printf("%sOutput → %s%s\n", cDim, cfg.Output, cReset)

	// ── Session temp dirs ─────────────────────────────────────────────────
	sessionID := fmt.Sprintf("bananascaler_%d_%d", time.Now().Unix(), os.Getpid())
	tempIn := filepath.Join(os.TempDir(), sessionID+"_in")
	tempOut := filepath.Join(os.TempDir(), sessionID+"_out")
	tmpOutput := cfg.Output + ".tmp"

	for _, d := range []string{tempIn, tempOut} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("create temp dir %q: %w", d, err)
		}
	}

	// ── Cleanup: always runs on return or signal ───────────────────────────
	cleanup := func() {
		os.RemoveAll(tempIn)
		os.RemoveAll(tempOut)
		os.Remove(tmpOutput)
	}
	defer cleanup()

	// ── Context + signal handling ─────────────────────────────────────────
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println()
		logErr("Interrupted — cancelling and cleaning up...")
		cancel()
	}()

	// ── Stage 1: Frame extraction ─────────────────────────────────────────
	logStep("[1/3] Extracting frames...")
	extractArgs := append(
		[]string{"-y", "-stats", "-loglevel", "warning"},
		decFlags...,
	)
	extractArgs = append(extractArgs,
		"-i", cfg.Input,
		"-f", "image2", "-vcodec", "mjpeg", "-q:v", "2",
		filepath.Join(tempIn, "frame_%05d.jpg"),
	)
	if err := run(ctx, cfg.Verbose, "ffmpeg", extractArgs...); err != nil {
		return fmt.Errorf("frame extraction: %w", err)
	}
	frameCount, err := countFiles(tempIn, ".jpg")
	if err != nil {
		return err
	}
	logOK(fmt.Sprintf("%d frames extracted.", frameCount))

	// ── Stage 2: Neural upscale (with live progress bar) ─────────────────
	logStep(fmt.Sprintf("[2/3] Neural upscaling (%d×) via Real-ESRGAN...", cfg.Scale))
	if err := upscaleWithProgress(ctx, cfg, tempIn, tempOut, frameCount); err != nil {
		return fmt.Errorf("upscaling: %w", err)
	}
	logOK("Upscaling complete.")

	// ── Stage 3: Re-encode + audio mux → atomic write ────────────────────
	logStep("[3/3] Re-encoding and muxing audio...")
	encArgs := []string{
		"-y", "-stats", "-loglevel", "warning",
		"-framerate", framerate,
		"-i", filepath.Join(tempOut, "frame_%05d.jpg"),
		"-i", cfg.Input,
		"-map", "0:v",
	}
	if hasAudio {
		encArgs = append(encArgs, "-map", "1:a", "-c:a", "copy")
	}
	encArgs = append(encArgs, encFlags...)
	encArgs = append(encArgs, tmpOutput)

	if err := run(ctx, cfg.Verbose, "ffmpeg", encArgs...); err != nil {
		return fmt.Errorf("video assembly: %w", err)
	}

	// Atomic rename — only reach here on exit code 0
	if err := os.Rename(tmpOutput, cfg.Output); err != nil {
		return fmt.Errorf("atomic rename: %w", err)
	}

	logOK(fmt.Sprintf("Done! → %s%s%s", cBold, cfg.Output, cReset))
	return nil
}

// ── Upscale with live progress ────────────────────────────────────────────────

// upscaleWithProgress launches realesrgan-ncnn-vulkan and polls the output
// directory every 500ms to drive a real-time progress bar.
func upscaleWithProgress(ctx context.Context, cfg *config.Config, tempIn, tempOut string, total int) error {
	bar := progressbar.NewOptions(total,
		progressbar.OptionSetDescription("  upscaling  "),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerPadding: "░",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionShowCount(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionUseANSICodes(true),
	)

	cmd := exec.CommandContext(ctx, "realesrgan-ncnn-vulkan",
		"-i", tempIn,
		"-o", tempOut,
		"-n", cfg.Model,
		"-s", fmt.Sprintf("%d", cfg.Scale),
		"-g", fmt.Sprintf("%d", cfg.GPU),
	)
	if cfg.Verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start realesrgan: %w", err)
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var prev int
	for {
		select {

		case err := <-done:
			// Drain progress bar to 100% on success
			if err == nil {
				final, _ := countFiles(tempOut, ".jpg")
				if delta := final - prev; delta > 0 {
					_ = bar.Add(delta)
				}
				_ = bar.Finish()
				fmt.Println()
			}
			return err

		case <-ticker.C:
			n, _ := countFiles(tempOut, ".jpg")
			if delta := n - prev; delta > 0 {
				_ = bar.Add(delta)
				prev = n
			}

		case <-ctx.Done():
			_ = cmd.Process.Kill()
			return ctx.Err()
		}
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// run executes an external command under the given context.
// stderr is always forwarded; stdout only if verbose is true.
func run(ctx context.Context, verbose bool, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stderr = os.Stderr
	if verbose {
		cmd.Stdout = os.Stdout
	}
	return cmd.Run()
}

// countFiles returns the number of regular files matching ext in dir.
func countFiles(dir, ext string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ext) {
			n++
		}
	}
	return n, nil
}
