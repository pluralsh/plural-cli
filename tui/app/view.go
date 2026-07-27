package app

import (
	tea "charm.land/bubbletea/v2"

	"github.com/pluralsh/plural-cli/tui/navigation"
)

const windowTitle = "Plural"

func (m Model) View() tea.View {
	content := m.welcome.View(m.width, m.height)
	switch m.route {
	case navigation.Access:
		content = m.access.View(m.width, m.height)
	case navigation.Diagnostics:
		content = m.diagnostics.View(m.width, m.height)
	}
	view := tea.NewView(content)
	view.AltScreen = true
	view.WindowTitle = windowTitle
	view.BackgroundColor = m.theme.Colors.Background
	view.ForegroundColor = m.theme.Colors.Text
	return view
}
