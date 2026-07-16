package tui

import "github.com/julesklord/bananascaler/internal/config"

// EventKind represents the type of pipeline event.
type EventKind int

const (
	EventLog EventKind = iota
	EventStageStart
	EventStageProgress
	EventStageDone
	EventDone
)

// Level constants for log events.
const (
	LevelInfo = "info"
	LevelOK   = "ok"
	LevelWarn = "warn"
	LevelErr  = "err"
	LevelStep = "step"
)

// PipelineEvent is sent from the pipeline goroutine to the TUI.
type PipelineEvent struct {
	Kind      EventKind
	Level     string
	Message   string
	Stage     int
	StageName string
	Current   int
	Total     int
}

// --- Bubbletea message wrappers ---

type pipelineEventMsg struct{ e PipelineEvent }
type pipelineDoneMsg struct{ err error }
type pipelineStartMsg struct{ cfg *config.Config }
