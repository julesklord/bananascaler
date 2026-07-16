package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/julesklord/bananascaler/internal/config"
)

const (
	barWidth       = 36
	maxVisibleLogs = 12
)

type stageStatus int

const (
	stagePending stageStatus = iota
	stageActive
	stageDone
	stageError
)

type stageState struct {
	name    string
	status  stageStatus
	current int
	total   int
}

type logEntry struct {
	level string
	text  string
}

// Model is the Bubbletea model for the bananascaler TUI.
type Model struct {
	cfg    *config.Config
	events chan PipelineEvent
	done   chan error

	stages  [3]stageState
	logs    []logEntry
	width   int
	height  int
	quitting bool
	finished bool
	pipelineErr error
}

// NewModel creates a new TUI model.
func NewModel(cfg *config.Config) Model {
	return Model{
		cfg:    cfg,
		events: make(chan PipelineEvent, 64),
		done:   make(chan error, 1),
		stages: [3]stageState{
			{name: "Frame Extraction", status: stagePending},
			{name: "Neural Upscaling", status: stagePending},
			{name: "Re-encode + Mux", status: stagePending},
		},
	}
}

// Events returns the event channel for the pipeline to send to.
func (m *Model) Events() chan<- PipelineEvent {
	return m.events
}

// SetDone signals the pipeline is finished.
func (m *Model) SetDone(err error) {
	m.done <- err
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.waitForEvent(),
		tea.EnterAltScreen,
	)
}

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

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "v":
			m.cfg.Verbose = !m.cfg.Verbose
			return m, nil
		}

	case pipelineEventMsg:
		return m.handleEvent(msg.e)

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

func (m Model) handleEvent(event PipelineEvent) (tea.Model, tea.Cmd) {
	switch event.Kind {
	case EventLog:
		m.addLog(event.Level, event.Message)

	case EventStageStart:
		if event.Stage >= 1 && event.Stage <= 3 {
			idx := event.Stage - 1
			m.stages[idx].status = stageActive
			if event.Total > 0 {
				m.stages[idx].total = event.Total
			}
		}

	case EventStageProgress:
		if event.Stage >= 1 && event.Stage <= 3 {
			idx := event.Stage - 1
			m.stages[idx].current = event.Current
			if event.Total > 0 {
				m.stages[idx].total = event.Total
			}
		}

	case EventStageDone:
		if event.Stage >= 1 && event.Stage <= 3 {
			idx := event.Stage - 1
			m.stages[idx].status = stageDone
			m.stages[idx].current = m.stages[idx].total
		}
	}

	return m, m.waitForEvent()
}

func (m *Model) addLog(level, text string) {
	m.logs = append(m.logs, logEntry{level: level, text: text})
	if len(m.logs) > maxVisibleLogs {
		m.logs = m.logs[len(m.logs)-maxVisibleLogs:]
	}
}

func (m Model) View() string {
	if m.quitting && !m.finished {
		return "\n  Interrupted. Cleaning up...\n"
	}

	var b strings.Builder

	// Header
	b.WriteString(titleStyle.Render("🍌 bananascaler"))
	b.WriteString("\n")

	// System info
	gpuInfo := "CPU (libx265)"
	nvidiaLine := m.cfg.GPU
	if nvidiaLine >= 0 {
		gpuInfo = fmt.Sprintf("GPU device %d (NVDEC+NVENC)", m.cfg.GPU)
	}
	b.WriteString(sysInfoStyle.Render(fmt.Sprintf(
		"  %s  │  Model: %s  │  Scale: %d×",
		gpuInfo, m.cfg.Model, m.cfg.Scale,
	)))
	b.WriteString("\n")

	outDisplay := m.cfg.Output
	if len(outDisplay) > 50 {
		outDisplay = "..." + outDisplay[len(outDisplay)-47:]
	}
	b.WriteString(infoStyle.Render(fmt.Sprintf(
		"  In: %s  │  Out: %s",
		m.cfg.Input, outDisplay,
	)))
	b.WriteString("\n\n")

	// Separator
	b.WriteString(separatorStyle.Render(strings.Repeat("─", min(60, max(40, m.width-4)))))
	b.WriteString("\n\n")

	// Stages
	for i, s := range m.stages {
		icon := iconPending
		switch s.status {
		case stageActive:
			icon = iconActive
		case stageDone:
			icon = iconDone
		case stageError:
			icon = iconError
		}

		stageLabel := fmt.Sprintf("Stage %d/3 — %s", i+1, s.name)
		b.WriteString(fmt.Sprintf("  %s %s\n", icon, stageNameStyle.Render(stageLabel)))

		if s.status == stageActive || s.status == stageDone {
			bar := renderProgress(s.current, max(s.total, 1), barWidth)
			countText := fmt.Sprintf("%d/%d", s.current, max(s.total, 0))
			if s.total == 0 && s.status == stageActive {
				countText = "..."
			}
			b.WriteString(fmt.Sprintf("      %s  %s\n",
				bar,
				progressTextStyle.Render(countText),
			))
		} else if s.status == stagePending {
			b.WriteString(fmt.Sprintf("      %s\n", stageInfoStyle.Render("waiting...")))
		} else if s.status == stageError {
			b.WriteString(fmt.Sprintf("      %s\n", stageInfoStyle.Render("failed")))
		}
		b.WriteString("\n")
	}

	// Separator
	b.WriteString(separatorStyle.Render(strings.Repeat("─", min(60, max(40, m.width-4)))))
	b.WriteString("\n")

	// Log area
	logHeight := max(3, min(maxVisibleLogs, m.height-18))
	visibleLogs := m.logs
	if len(visibleLogs) > logHeight {
		visibleLogs = visibleLogs[len(visibleLogs)-logHeight:]
	}
	for _, entry := range visibleLogs {
		prefix := logInfoStyle.Render("[INFO] ")
		switch entry.level {
		case LevelOK:
			prefix = logOKStyle.Render("[ OK ] ")
		case LevelWarn:
			prefix = logWarnStyle.Render("[WARN] ")
		case LevelErr:
			prefix = logErrStyle.Render("[ERR ] ")
		case LevelStep:
			prefix = logStepStyle.Render("🍌 ")
		}
		b.WriteString("  " + prefix + logTextStyle.Render(entry.text) + "\n")
	}

	// Footer
	b.WriteString("\n")
	b.WriteString(footerStyle.Render("  q: cancel  │  v: toggle verbose"))

	// Finished state
	if m.finished {
		b.WriteString("\n\n")
		if m.pipelineErr != nil {
			b.WriteString(errorBoxStyle.Render(fmt.Sprintf(" Error: %s", m.pipelineErr)))
		} else {
			b.WriteString(lipgloss.NewStyle().
				Foreground(colorGreen).
				Bold(true).
				Render(fmt.Sprintf("  ✓ Done → %s", m.cfg.Output)))
		}
		b.WriteString("\n")
	}

	return b.String()
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
