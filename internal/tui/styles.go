package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ── Palette ───────────────────────────────────────────────────────────────────
var (
	colBanana  = lipgloss.Color("#F5C542") // warm gold
	colGold    = lipgloss.Color("#E8A020") // deep amber accent
	colGreen   = lipgloss.Color("#3DD68C") // mint green
	colBlue    = lipgloss.Color("#5B9CF6") // calm blue (active)
	colRed     = lipgloss.Color("#F87171") // soft red
	colAmber   = lipgloss.Color("#FBBF24") // amber warning
	colPurple  = lipgloss.Color("#A78BFA") // lavender (step)
	colGray1   = lipgloss.Color("#94A3B8") // light muted
	colGray2   = lipgloss.Color("#475569") // mid separator
	colGray3   = lipgloss.Color("#1E293B") // dark panel bg
	colWhite   = lipgloss.Color("#F1F5F9") // near-white text
	colSubtext = lipgloss.Color("#64748B") // secondary text
)

// ── Layout constants ──────────────────────────────────────────────────────────
const (
	barWidth       = 38
	maxVisibleLogs = 10
	innerPad       = 2
)

// ── Base styles ───────────────────────────────────────────────────────────────
var (
	// Title: bold banana gold, large
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colBanana)

	// Dim accent line beneath title
	subtitleStyle = lipgloss.NewStyle().
			Foreground(colSubtext)

	// Thin separator line
	sepStyle = lipgloss.NewStyle().
			Foreground(colGray2)

	// Panel with rounded border
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colGray2).
			Padding(0, 1)

	// Key=value badges in header
	badgeKeyStyle = lipgloss.NewStyle().
			Foreground(colSubtext).
			Bold(false)

	badgeValStyle = lipgloss.NewStyle().
			Foreground(colWhite).
			Bold(true)

	// Stage icons (pre-rendered)
	iconDone    = lipgloss.NewStyle().Foreground(colGreen).Bold(true).Render("✔")
	iconActive  = lipgloss.NewStyle().Foreground(colBlue).Bold(true).Render("▶")
	iconPending = lipgloss.NewStyle().Foreground(colGray2).Render("○")
	iconError   = lipgloss.NewStyle().Foreground(colRed).Bold(true).Render("✖")

	// Stage label
	stageLabelActiveStyle  = lipgloss.NewStyle().Bold(true).Foreground(colWhite)
	stageLabelPendingStyle = lipgloss.NewStyle().Foreground(colGray1)
	stageLabelDoneStyle    = lipgloss.NewStyle().Foreground(colGreen)
	stageLabelErrorStyle   = lipgloss.NewStyle().Foreground(colRed)

	// Progress bar
	barFilledStyle = lipgloss.NewStyle().Foreground(colBlue)
	barLeadStyle   = lipgloss.NewStyle().Foreground(colBanana) // leading edge glow
	barEmptyStyle  = lipgloss.NewStyle().Foreground(colGray2)
	barDoneStyle   = lipgloss.NewStyle().Foreground(colGreen)
	barCountStyle  = lipgloss.NewStyle().Foreground(colGray1)
	barPctStyle    = lipgloss.NewStyle().Foreground(colWhite).Bold(true)

	// Log prefixes
	logPrefixInfo = lipgloss.NewStyle().Foreground(colBlue).Bold(true)
	logPrefixOK   = lipgloss.NewStyle().Foreground(colGreen).Bold(true)
	logPrefixWarn = lipgloss.NewStyle().Foreground(colAmber).Bold(true)
	logPrefixErr  = lipgloss.NewStyle().Foreground(colRed).Bold(true)
	logPrefixStep = lipgloss.NewStyle().Foreground(colPurple).Bold(true)
	logTextStyle  = lipgloss.NewStyle().Foreground(colGray1)

	// Footer keybind row
	footerKeyStyle = lipgloss.NewStyle().
			Foreground(colBanana).
			Bold(true)
	footerDescStyle = lipgloss.NewStyle().
			Foreground(colSubtext)
	footerDivStyle = lipgloss.NewStyle().
			Foreground(colGray2)

	// Error box
	errorBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colRed).
			Padding(0, 2).
			Foreground(colRed).
			Bold(true)

	// Done box
	doneBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colGreen).
			Padding(0, 2).
			Foreground(colGreen).
			Bold(true)

	// File explorer
	explorerDirStyle   = lipgloss.NewStyle().Foreground(colBlue).Bold(true)
	explorerVideoStyle = lipgloss.NewStyle().Foreground(colGreen).Bold(true)
	explorerFileStyle  = lipgloss.NewStyle().Foreground(colGray1)

	explorerCursorStyle = lipgloss.NewStyle().
				Background(colGray3).
				Foreground(colBanana).
				Bold(true)

	settingKeyStyle = lipgloss.NewStyle().Foreground(colSubtext)
	settingValStyle = lipgloss.NewStyle().Foreground(colBanana).Bold(true)
	settingKbStyle  = lipgloss.NewStyle().Foreground(colGray2)
)

// ── Progress bar renderer ─────────────────────────────────────────────────────

// renderBar renders a premium segmented progress bar.
// Done stages show in green; active shows blue+amber lead.
func renderBar(current, total, width int, done bool) string {
	if total <= 0 || width <= 0 {
		return barEmptyStyle.Render(strings.Repeat("─", width))
	}

	pct := float64(current) / float64(total)
	if pct > 1 {
		pct = 1
	}
	filled := int(pct * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	if done || pct >= 1.0 {
		return barDoneStyle.Render(strings.Repeat("█", width))
	}

	// Active: filled body + amber lead char + empty
	var bar strings.Builder
	if filled > 1 {
		bar.WriteString(barFilledStyle.Render(strings.Repeat("█", filled-1)))
		bar.WriteString(barLeadStyle.Render("▓"))
	} else if filled == 1 {
		bar.WriteString(barLeadStyle.Render("▓"))
	}
	bar.WriteString(barEmptyStyle.Render(strings.Repeat("░", empty)))
	return bar.String()
}

// renderPct returns a formatted percentage + optional ETA label.
func renderPct(current, total int, eta time.Duration) string {
	if total <= 0 {
		return barCountStyle.Render("···")
	}
	pct := int(float64(current) / float64(total) * 100)
	base := barPctStyle.Render(fmt.Sprintf("%3d%%", pct)) +
		barCountStyle.Render(fmt.Sprintf("  %d/%d", current, total))
	if e := fmtETA(eta); e != "" {
		base += barCountStyle.Render("  ETA ") +
			lipgloss.NewStyle().Foreground(colAmber).Bold(true).Render(e)
	}
	return base
}

// ── Helper ────────────────────────────────────────────────────────────────────

func repeat(s string, n int) string {
	if n <= 0 {
		return ""
	}
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString(s)
	}
	return b.String()
}

// hRule returns a horizontal rule fitted to w chars.
func hRule(w int) string {
	if w < 4 {
		w = 4
	}
	return sepStyle.Render(strings.Repeat("─", w))
}

// renderFooterBinds builds the keybind hint row.
func renderFooterBinds(binds [][2]string) string {
	parts := make([]string, 0, len(binds))
	for _, b := range binds {
		key := footerKeyStyle.Render(b[0])
		desc := footerDescStyle.Render(" " + b[1])
		parts = append(parts, key+desc)
	}
	div := footerDivStyle.Render("  ·  ")
	return strings.Join(parts, div)
}

// truncatePath shortens a path to at most max runes.
func truncatePath(p string, maxLen int) string {
	if len(p) <= maxLen {
		return p
	}
	return "…" + p[len(p)-(maxLen-1):]
}
