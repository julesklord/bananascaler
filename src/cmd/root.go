// Package cmd defines the bananascaler CLI using Cobra.
package cmd

import (
	"fmt"
	"os"

	"github.com/julesklord/bananascaler/internal/config"
	"github.com/julesklord/bananascaler/internal/pipeline"
	"github.com/spf13/cobra"
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
  nohup bananascaler input.mp4 --output out.mp4 --scale 4 > run.log 2>&1 &`,
	Args:          cobra.ExactArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg.Input = args[0]
		return pipeline.Run(&cfg)
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
	rootCmd.Flags().StringVarP(&cfg.Output, "output", "o", "",
		"Output file path (default: <input>_upscaled.mp4)")
	rootCmd.Flags().IntVarP(&cfg.Scale, "scale", "s", 2,
		"Upscale factor: 2, 3, or 4")
	rootCmd.Flags().IntVarP(&cfg.GPU, "gpu", "g", 0,
		"GPU device index for Real-ESRGAN (-1 = CPU)")
	rootCmd.Flags().StringVarP(&cfg.Model, "model", "m", "realesr-animevideov3-x2",
		"Real-ESRGAN model name")
	rootCmd.Flags().BoolVarP(&cfg.Verbose, "verbose", "v", false,
		"Forward ffmpeg/realesrgan output to terminal")
}
