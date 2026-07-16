package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/julesklord/bananascaler/internal/config"
	"github.com/julesklord/bananascaler/internal/pipeline"
)

// tuiLogger implements pipeline.Logger and sends events to the TUI via a channel.
type tuiLogger struct {
	events chan<- PipelineEvent
}

func (l *tuiLogger) Info(msg string) {
	l.events <- PipelineEvent{Kind: EventLog, Level: LevelInfo, Message: msg}
}

func (l *tuiLogger) OK(msg string) {
	l.events <- PipelineEvent{Kind: EventLog, Level: LevelOK, Message: msg}
}

func (l *tuiLogger) Warn(msg string) {
	l.events <- PipelineEvent{Kind: EventLog, Level: LevelWarn, Message: msg}
}

func (l *tuiLogger) Step(msg string) {
	// Parse step messages like "[1/3] Extracting frames..." to detect stage starts
	stage, name := parseStep(msg)
	if stage > 0 {
		l.events <- PipelineEvent{Kind: EventStageStart, Stage: stage, StageName: name}
	}
	l.events <- PipelineEvent{Kind: EventLog, Level: LevelStep, Message: msg}
}

func (l *tuiLogger) Err(msg string) {
	l.events <- PipelineEvent{Kind: EventLog, Level: LevelErr, Message: msg}
}

func (l *tuiLogger) Progress(stage, current, total int) {
	if current == 0 && total == 0 {
		return // initial call, stage already started
	}
	kind := EventStageProgress
	if current >= total && total > 0 {
		kind = EventStageDone
	}
	l.events <- PipelineEvent{
		Kind:    kind,
		Stage:   stage,
		Current: current,
		Total:   total,
	}
}

func parseStep(msg string) (int, string) {
	// Parse "[1/3] Extracting frames..." format
	if len(msg) < 6 {
		return 0, ""
	}
	if msg[0] != '[' {
		return 0, ""
	}
	var stage int
	for i := 1; i < len(msg) && msg[i] != '/'; i++ {
		if msg[i] >= '0' && msg[i] <= '9' {
			stage = stage*10 + int(msg[i]-'0')
		}
	}
	// Extract name after "] "
	name := ""
	for i := 0; i < len(msg)-1; i++ {
		if msg[i] == ']' && i+2 < len(msg) {
			name = msg[i+2:]
			break
		}
	}
	return stage, name
}

// RunTUI launches the Bubbletea TUI and runs the pipeline.
func RunTUI(cfg *config.Config) error {
	m := NewModel(cfg)

	// Start pipeline in background only if input is already selected (CLI mode)
	if cfg.Input != "" {
		go func() {
			log := &tuiLogger{events: m.Events()}
			err := pipeline.Run(cfg, log)
			m.SetDone(err)
		}()
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
