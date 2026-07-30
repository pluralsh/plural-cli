// Package app composes the root Bubble Tea model and runs the TUI process.
package app

import (
	"context"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	accessbridge "github.com/pluralsh/plural-cli/pkg/bridge/access"
	clustersbridge "github.com/pluralsh/plural-cli/pkg/bridge/clusters"
	notificationsbridge "github.com/pluralsh/plural-cli/pkg/bridge/notifications"
	pipelinesbridge "github.com/pluralsh/plural-cli/pkg/bridge/pipelines"
	providersbridge "github.com/pluralsh/plural-cli/pkg/bridge/providers"
	repositoriesbridge "github.com/pluralsh/plural-cli/pkg/bridge/repositories"
	servicesbridge "github.com/pluralsh/plural-cli/pkg/bridge/services"
	welcomebridge "github.com/pluralsh/plural-cli/pkg/bridge/welcome"
	"github.com/pluralsh/plural-cli/tui/navigation"
	accessscreen "github.com/pluralsh/plural-cli/tui/screens/access"
	clustersscreen "github.com/pluralsh/plural-cli/tui/screens/clusters"
	deploymentsscreen "github.com/pluralsh/plural-cli/tui/screens/deployments"
	diagnosticsscreen "github.com/pluralsh/plural-cli/tui/screens/diagnostics"
	notificationsscreen "github.com/pluralsh/plural-cli/tui/screens/notifications"
	pipelinesscreen "github.com/pluralsh/plural-cli/tui/screens/pipelines"
	providersscreen "github.com/pluralsh/plural-cli/tui/screens/providers"
	repositoriesscreen "github.com/pluralsh/plural-cli/tui/screens/repositories"
	servicesscreen "github.com/pluralsh/plural-cli/tui/screens/services"
	welcomescreen "github.com/pluralsh/plural-cli/tui/screens/welcome"
	"github.com/pluralsh/plural-cli/tui/theme"
)

// Dependencies contains the services required by TUI screens.
type Dependencies struct {
	Welcome       welcomebridge.Loader
	Access        accessbridge.Manager
	Services      servicesbridge.Loader
	Clusters      clustersbridge.Loader
	Repositories  repositoriesbridge.Loader
	Pipelines     pipelinesbridge.Loader
	Notifications notificationsbridge.Loader
	Providers     providersbridge.Loader
}

// Model is the root TUI model. It owns global input and delegates screen state
// to the active screen model.
type Model struct {
	width  int
	height int

	theme theme.Theme
	quit  key.Binding

	welcome       welcomescreen.Model
	access        accessscreen.Model
	diagnostics   diagnosticsscreen.Model
	deployments   deploymentsscreen.Model
	services      servicesscreen.Model
	clusters      clustersscreen.Model
	repositories  repositoriesscreen.Model
	pipelines     pipelinesscreen.Model
	notifications notificationsscreen.Model
	providers     providersscreen.Model
	route         navigation.Route
}

// New composes the root model with caller-provided dependencies.
func New(ctx context.Context, t theme.Theme, dependencies Dependencies) Model {
	return Model{
		theme:         t,
		welcome:       welcomescreen.New(ctx, dependencies.Welcome, t),
		access:        accessscreen.New(ctx, dependencies.Access, t),
		diagnostics:   diagnosticsscreen.New(ctx, dependencies.Welcome, t),
		deployments:   deploymentsscreen.New(ctx, t, ""),
		services:      servicesscreen.New(ctx, dependencies.Services, t),
		clusters:      clustersscreen.New(ctx, dependencies.Clusters, t),
		repositories:  repositoriesscreen.New(ctx, dependencies.Repositories, t),
		pipelines:     pipelinesscreen.New(ctx, dependencies.Pipelines, t),
		notifications: notificationsscreen.New(ctx, dependencies.Notifications, t),
		providers:     providersscreen.New(ctx, dependencies.Providers, t),
		route:         navigation.Welcome,
		quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
}

func (m Model) Init() tea.Cmd { return m.welcome.Init() }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case navigation.NavigateMsg:
		m.route = msg.Route
		switch m.route {
		case navigation.Access:
			return m, m.access.Init()
		case navigation.Diagnostics:
			return m, m.diagnostics.Init()
		case navigation.Deployments:
			m.deployments.SetConsoleURL(m.welcome.Snapshot().Console.URL)
			return m, m.deployments.Init()
		case navigation.Services:
			return m, m.services.Init()
		case navigation.Clusters:
			return m, m.clusters.Init()
		case navigation.Repositories:
			return m, m.repositories.Init()
		case navigation.Pipelines:
			return m, m.pipelines.Init()
		case navigation.Notifications:
			return m, m.notifications.Init()
		case navigation.Providers:
			return m, m.providers.Init()
		default:
			return m, m.welcome.Init()
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyPressMsg:
		if key.Matches(msg, m.quit) {
			if m.route == navigation.Access && m.access.HasCancellableOperation() {
				var cmd tea.Cmd
				m.access, cmd = m.access.Update(msg)
				return m, cmd
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	switch m.route {
	case navigation.Access:
		m.access, cmd = m.access.Update(msg)
	case navigation.Diagnostics:
		m.diagnostics, cmd = m.diagnostics.Update(msg)
	case navigation.Deployments:
		m.deployments, cmd = m.deployments.Update(msg)
	case navigation.Services:
		m.services, cmd = m.services.Update(msg)
	case navigation.Clusters:
		m.clusters, cmd = m.clusters.Update(msg)
	case navigation.Repositories:
		m.repositories, cmd = m.repositories.Update(msg)
	case navigation.Pipelines:
		m.pipelines, cmd = m.pipelines.Update(msg)
	case navigation.Notifications:
		m.notifications, cmd = m.notifications.Update(msg)
	case navigation.Providers:
		m.providers, cmd = m.providers.Update(msg)
	default:
		m.welcome, cmd = m.welcome.Update(msg)
	}
	return m, cmd
}
