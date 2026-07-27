// Package app composes the root Bubble Tea model and runs the TUI process.
package app

import (
	"context"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	accessbridge "github.com/pluralsh/plural-cli/pkg/bridge/access"
	welcomebridge "github.com/pluralsh/plural-cli/pkg/bridge/welcome"
	"github.com/pluralsh/plural-cli/tui/navigation"
	accessscreen "github.com/pluralsh/plural-cli/tui/screens/access"
	diagnosticsscreen "github.com/pluralsh/plural-cli/tui/screens/diagnostics"
	welcomescreen "github.com/pluralsh/plural-cli/tui/screens/welcome"
	"github.com/pluralsh/plural-cli/tui/theme"
)

// Dependencies contains the services required by TUI screens.
type Dependencies struct {
	Welcome welcomebridge.Loader
	Access  accessbridge.Manager
}

// Model is the root TUI model. It owns global input and delegates screen state
// to the active screen model.
type Model struct {
	width  int
	height int

	theme theme.Theme
	quit  key.Binding

	welcome     welcomescreen.Model
	access      accessscreen.Model
	diagnostics diagnosticsscreen.Model
	route       navigation.Route
}

// New composes the root model with caller-provided dependencies.
func New(ctx context.Context, t theme.Theme, dependencies Dependencies) Model {
	return Model{
		theme:       t,
		welcome:     welcomescreen.New(ctx, dependencies.Welcome, t),
		access:      accessscreen.New(ctx, dependencies.Access, t),
		diagnostics: diagnosticsscreen.New(ctx, dependencies.Welcome, t),
		route:       navigation.Welcome,
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
	default:
		m.welcome, cmd = m.welcome.Update(msg)
	}
	return m, cmd
}
