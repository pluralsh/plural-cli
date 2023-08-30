package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	events = make(chan Event)
)

// TerminalUI implements tea.Model interface
type TerminalUI struct {
	in      chan Event
	events  []Event
	spinner spinner.Model
}

type View interface {
	View()
}

type printView struct {
	in     chan Event
	events []Event
}

type progressView struct {
	in      chan Event
	events  []Event
	spinner spinner.Model
}

type Mode string

const (
	ModeProgress = "progress"
	ModePrint    = "print"
)

type EventType string

const (
	EventTypeProgress = EventType("progress")
	EventTypeMessage  = EventType("message")
)

type Event struct {
	Name    string
	Message string
	Took    time.Duration
}

func NewTerminalUI() tea.Model {
	return &TerminalUI{
		in:      events,
		events:  make([]Event, 5),
		spinner: newSpinner(),
	}
}
