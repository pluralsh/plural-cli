package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func (this *TerminalUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case tea.KeyMsg:
		return this, tea.Quit
	case spinner.TickMsg:
		return this, this.handleSpinnerUpdate(v)
	case Event:
		return this, this.handleEventUpdate(v)
	}

	return this, nil
}

func (this *TerminalUI) handleSpinnerUpdate(tick spinner.TickMsg) tea.Cmd {
	var cmd tea.Cmd
	this.spinner, cmd = this.spinner.Update(tick)

	return cmd
}

func (this *TerminalUI) handleEventUpdate(event Event) tea.Cmd {
	this.events = append(this.events, event)

	return this.waitForEvent()
}

func (this *TerminalUI) waitForEvent() tea.Cmd {
	return func() tea.Msg {
		return <-this.in
	}
}
