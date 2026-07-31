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
	case navigation.Deployments:
		content = m.deployments.View(m.width, m.height)
	case navigation.Services:
		content = m.services.View(m.width, m.height)
	case navigation.Clusters:
		content = m.clusters.View(m.width, m.height)
	case navigation.Repositories:
		content = m.repositories.View(m.width, m.height)
	case navigation.Pipelines:
		content = m.pipelines.View(m.width, m.height)
	case navigation.Notifications:
		content = m.notifications.View(m.width, m.height)
	case navigation.Providers:
		content = m.providers.View(m.width, m.height)
	case navigation.Stacks:
		content = m.stacks.View(m.width, m.height)
	case navigation.PullRequests:
		content = m.pullrequests.View(m.width, m.height)
	case navigation.AI:
		content = m.ai.View(m.width, m.height)
	case navigation.Agents:
		content = m.agents.View(m.width, m.height)
	case navigation.Workbenches:
		content = m.workbenches.View(m.width, m.height)
	}
	view := tea.NewView(content)
	view.AltScreen = true
	view.WindowTitle = windowTitle
	view.BackgroundColor = m.theme.Colors.Background
	view.ForegroundColor = m.theme.Colors.Text
	return view
}
