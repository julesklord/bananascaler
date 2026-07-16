// Package cmd defines the bananascaler CLI using Cobra.
package cmd

import (
	"fmt"
	"os"

	"github.com/julesklord/bananascaler/internal/config"
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

EXAMPLES
  bananascaler input.mp4
  bananascaler input.mp4 --output output_4k.mp4 --scale 4
  bananascaler input.mp4 --scale 2 --gpu 1 --verbose
  bananascaler input.mp4 --no-tui
  nohup bananascaler input.mp4 --output out.mp4 --scale 4 > run.log 2>&1 &`,
	Args:          cobra.ExactArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg.Input = args[0]

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
}
