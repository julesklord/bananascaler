package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	colorBanana = lipgloss.Color("#FFD700")
	colorGreen  = lipgloss.Color("#00CC66")
	colorRed    = lipgloss.Color("#FF4444")
	colorYellow = lipgloss.Color("#FFAA00")
	colorCyan   = lipgloss.Color("#00CCCC")
	colorDim    = lipgloss.Color("#666666")
	colorBg     = lipgloss.Color("#1A1A2E")
	colorFg     = lipgloss.Color("#E0E0E0")
	colorActive = lipgloss.Color("#4488FF")

	// Title style
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBanana).
			PaddingBottom(1)

	// Subtitle/info style
	infoStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			PaddingBottom(0)

	// System info line style
	sysInfoStyle = lipgloss.NewStyle().
			Foreground(colorCyan)

	// Separator
	separatorStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	// Stage status indicators (rendered once at init)
	iconDone    = lipgloss.NewStyle().Foreground(colorGreen).Bold(true).Render("[✓]")
	iconActive  = lipgloss.NewStyle().Foreground(colorActive).Bold(true).Render("[▶]")
	iconPending = lipgloss.NewStyle().Foreground(colorDim).Render("[·]")
	iconError   = lipgloss.NewStyle().Foreground(colorRed).Bold(true).Render("[✗]")

	// Stage name styles
	stageNameStyle = lipgloss.NewStyle().Bold(true).Foreground(colorFg)
	stageInfoStyle = lipgloss.NewStyle().Foreground(colorDim)

	// Progress bar styles
	progressFilledStyle = lipgloss.NewStyle().Foreground(colorActive)
	progressEmptyStyle  = lipgloss.NewStyle().Foreground(colorDim)
	progressTextStyle   = lipgloss.NewStyle().Foreground(colorFg)

	// Log styles
	logInfoStyle = lipgloss.NewStyle().Foreground(colorCyan).Bold(true)
	logOKStyle   = lipgloss.NewStyle().Foreground(colorGreen).Bold(true)
	logWarnStyle = lipgloss.NewStyle().Foreground(colorYellow).Bold(true)
	logErrStyle  = lipgloss.NewStyle().Foreground(colorRed).Bold(true)
	logStepStyle = lipgloss.NewStyle().Foreground(colorYellow).Bold(true)
	logTextStyle = lipgloss.NewStyle().Foreground(colorFg)

	// Footer
	footerStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			PaddingTop(1)

	// Error box
	errorBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(colorRed).
			Padding(0, 1).
			Foreground(colorRed)
)

func renderProgress(current, total, width int) string {
	if total <= 0 {
		return ""
	}
	filled := (current * width) / total
	if filled > width {
		filled = width
	}
	empty := width - filled
	bar := progressFilledStyle.Render(repeat("█", filled)) +
		progressEmptyStyle.Render(repeat("░", empty))
	return bar
}

func repeat(s string, n int) string {
	result := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		result = append(result, s...)
	}
	return string(result)
}
