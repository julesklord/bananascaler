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
)

// Logger abstracts output so the pipeline can be driven by a TUI or stdout.
type Logger interface {
	Info(msg string)
	OK(msg string)
	Warn(msg string)
	Step(msg string)
	Err(msg string)
	Progress(stage, current, total int)
}

// StdoutLogger writes pipeline events to the terminal with ANSI colors.
type StdoutLogger struct {
	Verbose bool
}

const (
	cReset  = "\033[0m"
	cBold   = "\033[1m"
	cDim    = "\033[2m"
	cRed    = "\033[31m"
	cGreen  = "\033[32m"
	cYellow = "\033[33m"
	cCyan   = "\033[36m"
)

func (l *StdoutLogger) Info(msg string)  { fmt.Printf("%s%s[INFO]%s %s\n", cBold, cCyan, cReset, msg) }
func (l *StdoutLogger) OK(msg string)    { fmt.Printf("%s%s[ OK ]%s %s\n", cBold, cGreen, cReset, msg) }
func (l *StdoutLogger) Warn(msg string)  { fmt.Printf("%s%s[WARN]%s %s\n", cBold, cYellow, cReset, msg) }
func (l *StdoutLogger) Step(msg string)  { fmt.Printf("\n%s%s🍌 %s%s\n", cBold, cYellow, msg, cReset) }
func (l *StdoutLogger) Err(msg string)   { fmt.Fprintf(os.Stderr, "%s%s[ERR ]%s %s\n", cBold, cRed, cReset, msg) }
func (l *StdoutLogger) Progress(stage, current, total int) {}

// pipelineParams extracts the effective parameters for this run,
// preferring profile values when available, falling back to legacy defaults.
func pipelineParams(cfg *config.Config) (tileSize int, jpegQuality int, nvencPreset string, x265Preset string, x265CRF int) {
	// Defaults (legacy mode — no profile)
	tileSize = 400
	jpegQuality = 2
	nvencPreset = ""
	x265Preset = "medium"
	x265CRF = 22

	if cfg.Profile != nil {
		p := cfg.Profile
		tileSize = p.TileSize
		jpegQuality = p.JPEGQuality
		nvencPreset = p.NVEncPreset
		x265Preset = p.X265Preset
		x265CRF = p.X265CRF
	}
	return
}

// Run executes the full bananascaler pipeline.
func Run(cfg *config.Config, log Logger) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	if err := hardware.CheckDeps(); err != nil {
		return err
	}

	// Resolve profile parameters
	tileSize, jpegQuality, nvencPreset, x265Preset, x265CRF := pipelineParams(cfg)

	// Hardware detection
	gpuInfo := hardware.DetectGPU()
	hasNVIDIA := gpuInfo.HasNVIDIA && cfg.GPU != -1

	var decFlags, encFlags []string
	if hasNVIDIA {
		log.Info("NVIDIA GPU detected — enabling NVDEC hardware-accelerated decoding and NVENC hardware-accelerated encoding.")
		decFlags = []string{"-hwaccel", "cuda"}
		encFlags = []string{"-c:v", "hevc_nvenc", "-pix_fmt", "yuv420p"}
		if nvencPreset != "" {
			encFlags = append([]string{"-preset", nvencPreset}, encFlags...)
		}
	} else {
		log.Warn("Running in CPU mode — falling back to CPU (libx265).")
		decFlags = []string{}
		encFlags = []string{"-c:v", "libx265", "-preset", x265Preset, "-crf", fmt.Sprintf("%d", x265CRF), "-pix_fmt", "yuv420p"}
	}

	// Log profile info
	if cfg.Profile != nil {
		log.Info(hardware.ProfileSummary(gpuInfo, cfg.Profile))
		if cfg.Verbose {
			log.Info(hardware.ProfileDisplay(gpuInfo, cfg.Profile))
		}
	} else {
		log.Info(fmt.Sprintf("Tile: %d | JPEG: %d | Model: %s (legacy defaults)", tileSize, jpegQuality, cfg.Model))
	}

	// Tile safety check — warn if tile size may exceed VRAM budget
	if warning := hardware.CheckTileSafety(gpuInfo, tileSize, cfg.Model); warning != "" {
		log.Warn(warning)
	}

	// Media probes
	hasAudio, err := hardware.HasAudio(cfg.Input)
	if err != nil {
		return err
	}
	framerate, err := hardware.Framerate(cfg.Input)
	if err != nil {
		return err
	}

	if hasAudio {
		log.Info("Audio stream found — will remux without re-encoding.")
	} else {
		log.Warn("No audio stream — output will be video-only.")
	}
	log.Info(fmt.Sprintf(
		"Framerate: %s fps  |  Scale: %d×  |  GPU index: %d",
		framerate, cfg.Scale, cfg.GPU,
	))
	log.Info(fmt.Sprintf("Output → %s", cfg.Output))

	// Session temp dirs
	sessionID := fmt.Sprintf("bananascaler_%d_%d", time.Now().Unix(), os.Getpid())
	tempIn := filepath.Join(os.TempDir(), sessionID+"_in")
	tempOut := filepath.Join(os.TempDir(), sessionID+"_out")
	tmpOutput := cfg.Output + ".tmp"

	for _, d := range []string{tempIn, tempOut} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("create temp dir %q: %w", d, err)
		}
	}

	// Cleanup
	cleanup := func() {
		os.RemoveAll(tempIn)
		os.RemoveAll(tempOut)
		os.Remove(tmpOutput)
	}
	defer cleanup()

	// Context + signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)
	go func() {
		<-sigCh
		fmt.Println()
		log.Err("Interrupted — cancelling and cleaning up...")
		cancel()
	}()

	// ── Stage 1: Frame extraction ─────────────────────────────────────────
	log.Step("[1/3] Extracting frames...")
	log.Progress(1, 0, 0)
	extractArgs := append(
		[]string{"-y", "-stats", "-loglevel", "warning"},
		decFlags...,
	)
	extractArgs = append(extractArgs,
		"-i", cfg.Input,
		"-f", "image2", "-vcodec", "mjpeg", "-q:v", fmt.Sprintf("%d", jpegQuality),
		filepath.Join(tempIn, "frame_%05d.jpg"),
	)
	if err := runCmd(ctx, cfg.Verbose, "ffmpeg", extractArgs...); err != nil {
		return fmt.Errorf("frame extraction: %w", err)
	}
	frameCount, err := countFiles(tempIn, ".jpg")
	if err != nil {
		return err
	}
	log.Progress(1, frameCount, frameCount)
	log.OK(fmt.Sprintf("%d frames extracted.", frameCount))

	// ── Stage 2: Neural upscale ──────────────────────────────────────────
	log.Step(fmt.Sprintf("[2/3] Neural upscaling (%d×) via Real-ESRGAN...", cfg.Scale))
	log.Progress(2, 0, frameCount)
	if err := upscale(ctx, cfg, log, tempIn, tempOut, frameCount, tileSize); err != nil {
		return fmt.Errorf("upscaling: %w", err)
	}
	log.Progress(2, frameCount, frameCount)
	log.OK("Upscaling complete.")

	// ── Stage 3: Re-encode + audio mux → atomic write ────────────────────
	log.Step("[3/3] Re-encoding and muxing audio...")
	log.Progress(3, 0, 0)
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

	if err := runCmd(ctx, cfg.Verbose, "ffmpeg", encArgs...); err != nil {
		return fmt.Errorf("video assembly: %w", err)
	}
	log.Progress(3, 1, 1)

	// Atomic rename
	if err := os.Rename(tmpOutput, cfg.Output); err != nil {
		return fmt.Errorf("atomic rename: %w", err)
	}

	log.OK(fmt.Sprintf("Done! → %s", cfg.Output))
	return nil
}

// upscale runs realesrgan-ncnn-vulkan and reports progress via the logger.
func upscale(ctx context.Context, cfg *config.Config, log Logger, tempIn, tempOut string, total, tileSize int) error {
	cmd := exec.CommandContext(ctx, "realesrgan-ncnn-vulkan",
		"-i", tempIn,
		"-o", tempOut,
		"-n", cfg.Model,
		"-s", fmt.Sprintf("%d", cfg.Scale),
		"-g", fmt.Sprintf("%d", cfg.GPU),
		"-t", fmt.Sprintf("%d", tileSize),
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
			if err == nil {
				final, _ := countFiles(tempOut, ".jpg")
				if final > prev {
					log.Progress(2, final, total)
				}
			}
			return err

		case <-ticker.C:
			n, _ := countFiles(tempOut, ".jpg")
			if n > prev {
				log.Progress(2, n, total)
				prev = n
			}

		case <-ctx.Done():
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
			return ctx.Err()
		}
	}
}

func runCmd(ctx context.Context, verbose bool, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stderr = os.Stderr
	if verbose {
		cmd.Stdout = os.Stdout
	}
	return cmd.Run()
}

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
