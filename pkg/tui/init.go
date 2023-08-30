package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (this *TerminalUI) Init() tea.Cmd {
	return tea.Batch(
		this.spinner.Tick,
		this.waitForEvent(),
	)
}
