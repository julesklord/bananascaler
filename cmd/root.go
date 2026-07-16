// Package cmd defines the bananascaler CLI using Cobra.
package cmd

import (
	"fmt"
	"os"

	"github.com/julesklord/bananascaler/internal/config"
	"github.com/julesklord/bananascaler/internal/hardware"
	"github.com/julesklord/bananascaler/internal/pipeline"
	"github.com/julesklord/bananascaler/internal/tui"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var cfg config.Config

var rootCmd = &cobra.Command{
	Use:   "bananascaler <input>",
	Short: "🍌 GPU-accelerated neural video upscaler",
	Long: `bananascaler — GPU-accelerated neural video upscaler

Orchestrates Real-ESRGAN (Vulkan) + FFmpeg (NVDEC/NVENC) to upscale
video files up to 4× using neural super-resolution.

PROFILES
  bananascaler auto-detects your GPU tier and recommends optimal settings.
  Use --auto to enable automatic profile detection, or --profile to pick
  a preset manually.

  Tiers:   low-end (<6GB) | mid-range (6-10GB) | high-end (10GB+)
  Presets: fast | balanced | quality

EXAMPLES
  bananascaler input.mp4                          # auto-detect profile
  bananascaler input.mp4 --auto                   # explicit auto-detect
  bananascaler input.mp4 --profile fast           # fast preset for your GPU
  bananascaler input.mp4 --profile quality        # quality preset for your GPU
  bananascaler input.mp4 --scale 4
  bananascaler input.mp4 --output output_4k.mp4 --scale 4
  bananascaler input.mp4 --scale 2 --gpu 1 --verbose
  bananascaler input.mp4 --no-tui
  nohup bananascaler input.mp4 --output out.mp4 --scale 4 > run.log 2>&1 &`,
	Args:          cobra.ExactArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg.Input = args[0]

		// Resolve profile before validation
		if err := cfg.ResolveProfile(); err != nil {
			return err
		}

		if err := cfg.Validate(); err != nil {
			return err
		}

		// TUI mode if stdout is a TTY and --no-tui is not set
		if !cfg.NoTUI && term.IsTerminal(int(os.Stdout.Fd())) {
			return tui.RunTUI(&cfg)
		}

		// Plain mode: no TUI, just stdout logging
		log := &pipeline.StdoutLogger{Verbose: cfg.Verbose}
		return pipeline.Run(&cfg, log)
	},
}

// Execute is the public entry point. Called by main.go.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "\033[1m\033[31m[ERR ]\033[0m %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfg.Output, "output", "o", "",
		"Output file path (default: <input>_upscaled.mp4)")
	rootCmd.PersistentFlags().IntVarP(&cfg.Scale, "scale", "s", 2,
		"Upscale factor: 2, 3, or 4")
	rootCmd.PersistentFlags().IntVarP(&cfg.GPU, "gpu", "g", 0,
		"GPU device index for Real-ESRGAN (-1 = CPU)")
	rootCmd.PersistentFlags().StringVarP(&cfg.Model, "model", "m", config.DefaultModel,
		"Real-ESRGAN model name")
	rootCmd.PersistentFlags().BoolVarP(&cfg.Verbose, "verbose", "v", false,
		"Forward ffmpeg/realesrgan output to terminal")
	rootCmd.PersistentFlags().BoolVar(&cfg.NoTUI, "no-tui", false,
		"Disable interactive TUI (use plain text output)")

	// Profile flags
	rootCmd.PersistentFlags().StringVar(&cfg.PresetStr, "profile", "",
		"Performance preset: fast, balanced, or quality (auto-detects GPU tier)")
	rootCmd.PersistentFlags().BoolVar(&cfg.AutoDetect, "auto", false,
		"Auto-detect GPU and apply optimal profile (balanced preset)")

	// Detect subcommand
	rootCmd.AddCommand(detectCmd)
}

var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Detect hardware and show recommended profiles",
	Long: `Scans your system for GPU capabilities and displays all three
performance profiles (fast/balanced/quality) adapted to your hardware.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		info := hardware.DetectGPU()

		fmt.Printf("\033[1m🍌 Hardware Detection\033[0m\n\n")

		if info.HasNVIDIA {
			vram := fmt.Sprintf("%d MB", info.VRAMMB)
			if info.VRAMMB >= 1024 {
				vram = fmt.Sprintf("%.1f GB", float64(info.VRAMMB)/1024.0)
			}
			fmt.Printf("  GPU:     %s\n", info.Name)
			fmt.Printf("  VRAM:    %s\n", vram)
		} else {
			fmt.Printf("  GPU:     none detected (CPU-only mode)\n")
		}
		fmt.Printf("  CPU:     %d cores\n", info.CPUCores)

		fmt.Printf("\n\033[1mAvailable Profiles\033[0m\n\n")

		for _, preset := range hardware.AllPresets() {
			profile := hardware.GetProfile(
				hardware.TierLowEnd, preset,
			)
			_ = profile // just checking we can get them

			// Get the profile for detected hardware
			profile, _ = hardware.AutoProfileWithPreset(preset)
			icon := "  "
			if preset == "balanced" {
				icon = "★ "
			}
			fmt.Printf("  %s\033[1m%s\033[0m\n", icon, preset)
			fmt.Printf("     %s\n", profile.Description)
			fmt.Printf("     tile=%d  model=%s  x265=%s/CRF%d", profile.TileSize, profile.Model, profile.X265Preset, profile.X265CRF)
			if profile.NVEncPreset != "" {
				fmt.Printf("  nvenc=%s", profile.NVEncPreset)
			}
			fmt.Println()
			fmt.Println()
		}

		// Also show all tiers for reference
		fmt.Printf("\033[1mAll Tier Presets (reference)\033[0m\n\n")
		for _, tier := range []hardware.HardwareTier{hardware.TierLowEnd, hardware.TierMidRange, hardware.TierHighEnd} {
			fmt.Printf("  \033[1m%s\033[0m\n", tier)
			for _, preset := range hardware.AllPresets() {
				p := hardware.GetProfile(tier, preset)
				fmt.Printf("    %-10s tile=%-4d model=%-30s x265=%-10s CRF=%-3d",
					preset, p.TileSize, p.Model, p.X265Preset, p.X265CRF)
				if p.NVEncPreset != "" {
					fmt.Printf(" nvenc=%s", p.NVEncPreset)
				}
				fmt.Println()
			}
			fmt.Println()
		}

		return nil
	},
}
