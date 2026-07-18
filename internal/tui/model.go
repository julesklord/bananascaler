package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/julesklord/bananascaler/internal/config"
	"github.com/julesklord/bananascaler/internal/hardware"
	"github.com/julesklord/bananascaler/internal/pipeline"
)

// ── Constants ─────────────────────────────────────────────────────────────────

type stageStatus int

const (
	stagePending stageStatus = iota
	stageActive
	stageDone
	stageError
)

type tuiState int

const (
	stateSelectFile tuiState = iota
	statePipeline
)

var modelNames = []string{
	"realesr-animevideov3-x2",
	"realesrgan-x4plus",
	"realesrgan-x4plus-anime",
}

// ── Data types ────────────────────────────────────────────────────────────────

type stageState struct {
	name    string
	status  stageStatus
	current int
	total   int
	eta     time.Duration // 0 = unknown
}

type logEntry struct {
	level string
	text  string
}

// ── Bubbletea model ───────────────────────────────────────────────────────────

// Model is the Bubbletea model for the bananascaler TUI.
type Model struct {
	cfg         *config.Config
	events      chan PipelineEvent
	done        chan error
	stages      [3]stageState
	logs        []logEntry
	width       int
	height      int
	quitting    bool
	finished    bool
	pipelineErr error

	// Explorer
	state        tuiState
	currentDir   string
	files        []os.DirEntry
	cursor       int
	scrollOffset int
	explorerErr  error
}

// NewModel creates a new TUI model.
func NewModel(cfg *config.Config) Model {
	state := statePipeline
	if cfg.Input == "" {
		state = stateSelectFile
	}

	// Auto-detect profile if none was set via CLI
	if cfg.Profile == nil {
		preset := hardware.PresetBalanced
		if cfg.PresetStr != "" {
			if p, err := hardware.ParsePreset(cfg.PresetStr); err == nil {
				preset = p
			}
		}
		profile, _ := hardware.AutoProfileWithPreset(preset)
		cfg.Profile = profile
		cfg.Model = profile.Model
	}

	wd, _ := os.Getwd()
	return Model{
		cfg:        cfg,
		events:     make(chan PipelineEvent, 64),
		done:       make(chan error, 1),
		state:      state,
		currentDir: wd,
		stages: [3]stageState{
			{name: "Frame Extraction", status: stagePending},
			{name: "Neural Upscaling", status: stagePending},
			{name: "Re-encode + Mux", status: stagePending},
		},
	}
}

func (m *Model) Events() chan<- PipelineEvent { return m.events }
func (m *Model) SetDone(err error)            { m.done <- err }

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.waitForEvent(), tea.EnterAltScreen}
	if m.state == stateSelectFile {
		cmds = append(cmds, m.readDirCmd())
	}
	return tea.Batch(cmds...)
}

// ── Directory reading ─────────────────────────────────────────────────────────

type readDirMsg struct {
	dir   string
	files []os.DirEntry
	err   error
}

func (m Model) readDirCmd() tea.Cmd {
	return func() tea.Msg {
		dir := m.currentDir
		if dir == "" {
			var err error
			dir, err = os.Getwd()
			if err != nil {
				return readDirMsg{err: err}
			}
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return readDirMsg{err: err}
		}
		var dirs, files []os.DirEntry
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), ".") {
				continue
			}
			if e.IsDir() {
				dirs = append(dirs, e)
			} else {
				files = append(files, e)
			}
		}
		return readDirMsg{dir: dir, files: append(dirs, files...)}
	}
}

func isVideoFile(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".mp4", ".mkv", ".avi", ".mov", ".webm", ".flv", ".wmv", ".m4v", ".mpg", ".mpeg", ".3gp":
		return true
	}
	return false
}

// ── Event loop ────────────────────────────────────────────────────────────────

func (m Model) waitForEvent() tea.Cmd {
	return func() tea.Msg {
		select {
		case event, ok := <-m.events:
			if !ok {
				return pipelineDoneMsg{err: nil}
			}
			return pipelineEventMsg{e: event}
		case err := <-m.done:
			return pipelineDoneMsg{err: err}
		}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case readDirMsg:
		if msg.err != nil {
			m.explorerErr = msg.err
			return m, nil
		}
		m.currentDir = msg.dir
		m.files = msg.files
		m.cursor = 0
		m.scrollOffset = 0
		m.explorerErr = nil
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
		if m.state == stateSelectFile {
			return m.updateExplorer(msg)
		}
		return m.updatePipeline(msg)

	case pipelineEventMsg:
		return m.handlePipelineEvent(msg.e)

	case pipelineDoneMsg:
		m.finished = true
		m.pipelineErr = msg.err
		if msg.err != nil {
			m.addLog("err", msg.err.Error())
		}
		return m, tea.Quit
	}

	return m, nil
}

func (m Model) updateExplorer(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	visH := m.explorerVisibleHeight()
	switch msg.String() {
	case "q", "esc":
		m.quitting = true
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			if m.cursor < m.scrollOffset {
				m.scrollOffset = m.cursor
			}
		}
	case "down", "j":
		if m.cursor < len(m.files)-1 {
			m.cursor++
			if m.cursor >= m.scrollOffset+visH {
				m.scrollOffset = m.cursor - visH + 1
			}
		}
	case "backspace", "h", "left":
		m.currentDir = filepath.Dir(m.currentDir)
		return m, m.readDirCmd()
	case "s":
		m.cfg.Scale++
		if m.cfg.Scale > 4 {
			m.cfg.Scale = 2
		}
	case "g":
		if m.cfg.GPU == -1 {
			m.cfg.GPU = 0
		} else if m.cfg.GPU == 0 {
			m.cfg.GPU = 1
		} else {
			m.cfg.GPU = -1
		}
	case "m":
		idx := 0
		for i, n := range modelNames {
			if n == m.cfg.Model {
				idx = i
				break
			}
		}
		m.cfg.Model = modelNames[(idx+1)%len(modelNames)]
	case "p":
		// Cycle through presets: fast → balanced → quality → fast
		presets := hardware.AllPresets()
		idx := 0
		if m.cfg.Profile != nil {
			for i, p := range presets {
				if p == m.cfg.Profile.Preset {
					idx = i
					break
				}
			}
		}
		nextPreset := presets[(idx+1)%len(presets)]
		profile, _ := hardware.AutoProfileWithPreset(nextPreset)
		m.cfg.Profile = profile
		m.cfg.Model = profile.Model
	case "enter", "right":
		if len(m.files) == 0 {
			break
		}
		sel := m.files[m.cursor]
		if sel.IsDir() {
			m.currentDir = filepath.Join(m.currentDir, sel.Name())
			return m, m.readDirCmd()
		}
		// Launch pipeline
		m.cfg.Input = filepath.Join(m.currentDir, sel.Name())
		dir := filepath.Dir(m.cfg.Input)
		base := strings.TrimSuffix(filepath.Base(m.cfg.Input), filepath.Ext(m.cfg.Input))
		m.cfg.Output = filepath.Join(dir, base+"_upscaled.mp4")
		go func() {
			log := &tuiLogger{events: m.Events()}
			m.SetDone(pipeline.Run(m.cfg, log))
		}()
		m.state = statePipeline
	}
	return m, nil
}

func (m Model) updatePipeline(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.quitting = true
		return m, tea.Quit
	case "v":
		m.cfg.Verbose = !m.cfg.Verbose
	}
	return m, nil
}

func (m Model) handlePipelineEvent(e PipelineEvent) (tea.Model, tea.Cmd) {
	switch e.Kind {
	case EventLog:
		m.addLog(e.Level, e.Message)
	case EventStageStart:
		if e.Stage >= 1 && e.Stage <= 3 {
			m.stages[e.Stage-1].status = stageActive
			if e.Total > 0 {
				m.stages[e.Stage-1].total = e.Total
			}
		}
	case EventStageProgress:
		if e.Stage >= 1 && e.Stage <= 3 {
			m.stages[e.Stage-1].current = e.Current
			m.stages[e.Stage-1].eta = e.ETA
			if e.Total > 0 {
				m.stages[e.Stage-1].total = e.Total
			}
		}
	case EventStageDone:
		if e.Stage >= 1 && e.Stage <= 3 {
			idx := e.Stage - 1
			m.stages[idx].status = stageDone
			m.stages[idx].current = m.stages[idx].total
			m.stages[idx].eta = 0
		}
	}
	return m, m.waitForEvent()
}

func (m *Model) addLog(level, text string) {
	m.logs = append(m.logs, logEntry{level, text})
	if len(m.logs) > maxVisibleLogs {
		m.logs = m.logs[len(m.logs)-maxVisibleLogs:]
	}
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m Model) View() string {
	if m.quitting && !m.finished {
		return lipgloss.NewStyle().Foreground(colAmber).Render("\n  ⚡ Interrupted — cleaning up…\n")
	}
	if m.state == stateSelectFile {
		return m.viewExplorer()
	}
	return m.viewPipeline()
}

// ── Explorer view ─────────────────────────────────────────────────────────────

func (m Model) explorerVisibleHeight() int {
	h := m.height - 14
	if h < 4 {
		h = 4
	}
	return h
}

func (m Model) viewExplorer() string {
	w := m.width
	if w < 40 {
		w = 40
	}
	innerW := w - 4

	var b strings.Builder

	// ── Header ─────────────────────────────────────────────────────────────
	title := titleStyle.Render("  🍌 bananascaler")
	tag := lipgloss.NewStyle().
		Foreground(colGold).
		Render("  file selector")
	b.WriteString(title + tag + "\n")

	dirLine := subtitleStyle.Render("  " + truncatePath(m.currentDir, innerW-3))
	b.WriteString(dirLine + "\n")
	b.WriteString(hRule(innerW) + "\n\n")

	// ── File list ──────────────────────────────────────────────────────────
	visH := m.explorerVisibleHeight()

	if m.explorerErr != nil {
		b.WriteString("  " + errorBoxStyle.Render(fmt.Sprintf(" %v ", m.explorerErr)) + "\n")
	} else if len(m.files) == 0 {
		b.WriteString("  " + lipgloss.NewStyle().Foreground(colGray2).Italic(true).Render("empty directory") + "\n")
	} else {
		end := m.scrollOffset + visH
		if end > len(m.files) {
			end = len(m.files)
		}
		for i := m.scrollOffset; i < end; i++ {
			f := m.files[i]
			name := f.Name()

			var icon, rendered string
			if f.IsDir() {
				icon = "  "
				rendered = explorerDirStyle.Render(name + "/")
			} else if isVideoFile(name) {
				icon = "  "
				rendered = explorerVideoStyle.Render(name)
			} else {
				icon = "  "
				rendered = explorerFileStyle.Render(name)
			}

			line := icon + rendered
			if i == m.cursor {
				// Pad line to fill width so highlight stretches
				var raw string
				if f.IsDir() {
					raw = icon + name + "/"
				} else {
					raw = icon + name
				}
				padding := innerW - len(raw)
				if padding < 0 {
					padding = 0
				}
				line = explorerCursorStyle.Render(" " + icon[1:] + name + strings.Repeat(" ", padding+1))
				_ = rendered // suppress unused warning
			}
			b.WriteString(line + "\n")
		}
		// Fill remaining lines so layout stays stable
		shown := end - m.scrollOffset
		for i := shown; i < visH; i++ {
			b.WriteString("\n")
		}
	}

	// ── Settings bar ───────────────────────────────────────────────────────
	b.WriteString(hRule(innerW) + "\n")

	gpuVal := "CPU"
	if m.cfg.GPU >= 0 {
		gpuVal = fmt.Sprintf("GPU %d", m.cfg.GPU)
	}
	profileVal := "none"
	if m.cfg.Profile != nil {
		profileVal = fmt.Sprintf("%s/%s", m.cfg.Profile.Tier, m.cfg.Profile.Preset)
	}
	settings := []struct{ k, v, kb string }{
		{"Scale", fmt.Sprintf("%d×", m.cfg.Scale), "s"},
		{"GPU", gpuVal, "g"},
		{"Model", shortModel(m.cfg.Model), "m"},
		{"Profile", profileVal, "p"},
	}
	settingParts := make([]string, len(settings))
	for i, s := range settings {
		settingParts[i] = settingKeyStyle.Render(s.k+": ") +
			settingValStyle.Render(s.v) +
			settingKbStyle.Render(" ["+s.kb+"]")
	}
	b.WriteString("  " + strings.Join(settingParts, footerDivStyle.Render("   │   ")) + "\n\n")

	// ── Footer keybinds ────────────────────────────────────────────────────
	binds := [][2]string{
		{"↑↓ / jk", "navigate"},
		{"Enter", "open / select"},
		{"⌫ / h", "go up"},
		{"p", "cycle profile"},
		{"q", "quit"},
	}
	b.WriteString("  " + renderFooterBinds(binds))

	return b.String()
}

// ── Pipeline view ─────────────────────────────────────────────────────────────

func (m Model) viewPipeline() string {
	w := m.width
	if w < 40 {
		w = 40
	}
	innerW := w - 4

	var b strings.Builder

	// ── Header ─────────────────────────────────────────────────────────────
	b.WriteString(titleStyle.Render("  🍌 bananascaler") + "\n")

	// Meta row: profile | gpu | model | scale
	gpuLabel := "CPU · libx265"
	if m.cfg.GPU >= 0 {
		gpuLabel = fmt.Sprintf("GPU %d · NVDEC+NVENC", m.cfg.GPU)
	}
	profileLabel := "legacy"
	if m.cfg.Profile != nil {
		profileLabel = fmt.Sprintf("%s · %s", m.cfg.Profile.Tier, m.cfg.Profile.Preset)
	}
	meta := renderBadge("Profile", profileLabel) + "   " +
		renderBadge("GPU", gpuLabel) + "   " +
		renderBadge("Model", shortModel(m.cfg.Model)) + "   " +
		renderBadge("Scale", fmt.Sprintf("%d×", m.cfg.Scale))
	b.WriteString("  " + meta + "\n")

	// Input / output paths
	inShort := truncatePath(m.cfg.Input, (innerW/2)-8)
	outShort := truncatePath(m.cfg.Output, (innerW/2)-8)
	b.WriteString("  " + subtitleStyle.Render("in  ") +
		lipgloss.NewStyle().Foreground(colGray1).Render(inShort) + "\n")
	b.WriteString("  " + subtitleStyle.Render("out ") +
		lipgloss.NewStyle().Foreground(colGray1).Render(outShort) + "\n")
	b.WriteString(hRule(innerW) + "\n\n")

	// ── Stages ─────────────────────────────────────────────────────────────
	for i, s := range m.stages {
		num := fmt.Sprintf("%d/3", i+1)
		b.WriteString(m.renderStage(num, s, innerW))
		b.WriteString("\n")
	}

	// ── Log panel ──────────────────────────────────────────────────────────
	b.WriteString(hRule(innerW) + "\n")
	logH := maxVisibleLogs
	if m.height > 0 {
		logH = max(3, min(maxVisibleLogs, m.height-20))
	}
	visible := m.logs
	if len(visible) > logH {
		visible = visible[len(visible)-logH:]
	}
	for _, entry := range visible {
		b.WriteString("  " + m.renderLogLine(entry) + "\n")
	}
	// Pad to fixed height so layout doesn't jump
	for i := len(visible); i < logH; i++ {
		b.WriteString("\n")
	}

	// ── Footer ─────────────────────────────────────────────────────────────
	b.WriteString(hRule(innerW) + "\n")
	binds := [][2]string{{"q / Esc", "cancel"}, {"v", "verbose"}}
	b.WriteString("  " + renderFooterBinds(binds))

	// ── Completion banner ───────────────────────────────────────────────────
	if m.finished {
		b.WriteString("\n\n")
		if m.pipelineErr != nil {
			b.WriteString("  " + errorBoxStyle.Render(" ✖  "+m.pipelineErr.Error()+" "))
		} else {
			b.WriteString("  " + doneBoxStyle.Render(" ✔  Done → "+m.cfg.Output+" "))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (m Model) renderStage(num string, s stageState, innerW int) string {
	var icon string
	var labelStyle lipgloss.Style
	switch s.status {
	case stageActive:
		icon = iconActive
		labelStyle = stageLabelActiveStyle
	case stageDone:
		icon = iconDone
		labelStyle = stageLabelDoneStyle
	case stageError:
		icon = iconError
		labelStyle = stageLabelErrorStyle
	default:
		icon = iconPending
		labelStyle = stageLabelPendingStyle
	}

	numStr := lipgloss.NewStyle().Foreground(colSubtext).Render(num)
	label := labelStyle.Render(s.name)
	header := fmt.Sprintf("  %s  %s  %s", icon, numStr, label)

	var b strings.Builder
	b.WriteString(header + "\n")

	switch s.status {
	case stageActive:
		bar := renderBar(s.current, max(s.total, 1), barWidth, false)
		count := ""
		if s.total > 0 {
			count = renderPct(s.current, s.total, s.eta)
		} else {
			count = barCountStyle.Render("processing…")
		}
		b.WriteString(fmt.Sprintf("       %s  %s\n", bar, count))

	case stageDone:
		bar := renderBar(s.current, max(s.total, 1), barWidth, true)
		count := renderPct(s.current, s.total, 0)
		b.WriteString(fmt.Sprintf("       %s  %s\n", bar, count))

	case stagePending:
		empty := barEmptyStyle.Render(strings.Repeat("─", barWidth))
		waiting := lipgloss.NewStyle().Foreground(colGray2).Italic(true).Render("waiting")
		b.WriteString(fmt.Sprintf("       %s  %s\n", empty, waiting))

	case stageError:
		bar := strings.Repeat("─", barWidth)
		b.WriteString(fmt.Sprintf("       %s  %s\n",
			lipgloss.NewStyle().Foreground(colRed).Render(bar),
			lipgloss.NewStyle().Foreground(colRed).Render("failed")))
	}

	return b.String()
}

func (m Model) renderLogLine(e logEntry) string {
	var prefix string
	switch e.level {
	case LevelOK:
		prefix = logPrefixOK.Render("✔ ok   ")
	case LevelWarn:
		prefix = logPrefixWarn.Render("⚠ warn ")
	case LevelErr:
		prefix = logPrefixErr.Render("✖ err  ")
	case LevelStep:
		prefix = logPrefixStep.Render("◆ step ")
	default:
		prefix = logPrefixInfo.Render("· info ")
	}
	return prefix + logTextStyle.Render(e.text)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func renderBadge(k, v string) string {
	return badgeKeyStyle.Render(k+": ") + badgeValStyle.Render(v)
}

func shortModel(m string) string {
	// "realesr-animevideov3-x2" → "animevideo-x2"
	// "realesrgan-x4plus-anime" → "x4plus-anime"
	m = strings.TrimPrefix(m, "realesr-")
	m = strings.TrimPrefix(m, "realesrgan-")
	return m
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// fmtETA formats a duration as "Xm Ys" or "Xs".
func fmtETA(d time.Duration) string {
	d = d.Round(time.Second)
	if d <= 0 {
		return ""
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
}
