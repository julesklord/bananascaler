package cmd

import (
	"github.com/julesklord/bananascaler/internal/tui"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Open interactive file manager TUI to select and upscale videos",
	Long: `Opens the bananascaler TUI, listing all video files in the current
working directory. Navigate, configure options, and start upscaling.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Run TUI in selection mode by setting Input to empty
		cfg.Input = ""
		return tui.RunTUI(&cfg)
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
