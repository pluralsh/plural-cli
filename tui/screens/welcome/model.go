package welcome

import (
	"context"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	welcomebridge "github.com/pluralsh/plural-cli/pkg/bridge/welcome"
	"github.com/pluralsh/plural-cli/tui/components/commandbar"
	pluralspinner "github.com/pluralsh/plural-cli/tui/components/spinner"
	"github.com/pluralsh/plural-cli/tui/navigation"
	"github.com/pluralsh/plural-cli/tui/theme"
)

type loadedMsg struct{ snapshot welcomebridge.Snapshot }
type failedMsg struct{ err error }

type Model struct {
	ctx      context.Context
	loader   welcomebridge.Loader
	theme    theme.Theme
	spinner  spinner.Model
	command  commandbar.Model
	loading  bool
	snapshot welcomebridge.Snapshot
	err      error
}

func New(ctx context.Context, loader welcomebridge.Loader, t theme.Theme) Model {
	commands := []string{"access", "console", "diagnostics", "help", "profiles", "services", "workspace"}
	return Model{
		ctx:     ctx,
		loader:  loader,
		theme:   t,
		spinner: pluralspinner.New(t),
		command: commandbar.New(t, commands),
		loading: loader != nil,
	}
}

func (m Model) Init() tea.Cmd {
	if !m.loading {
		return nil
	}
	return tea.Batch(m.spinner.Tick, m.loadSnapshot)
}

func (m Model) loadSnapshot() tea.Msg {
	snapshot, err := m.loader.Load(m.ctx)
	if err != nil {
		return failedMsg{err: err}
	}
	return loadedMsg{snapshot: snapshot}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case loadedMsg:
		m.loading = false
		m.snapshot = msg.snapshot
		m.err = nil
		return m, nil
	case failedMsg:
		m.loading = false
		m.err = msg.err
		return m, nil
	case spinner.TickMsg:
		if !m.loading {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case commandbar.SubmittedMsg:
		switch msg.Command {
		case "access", "console", "profiles":
			return m, navigation.Navigate(navigation.Access)
		case "diagnostics", "workspace":
			return m, navigation.Navigate(navigation.Diagnostics)
		case "services":
			return m, navigation.Navigate(navigation.Services)
		}
		return m, nil
	default:
		var cmd tea.Cmd
		m.command, cmd = m.command.Update(msg)
		return m, cmd
	}
}

func (m Model) Snapshot() welcomebridge.Snapshot { return m.snapshot }
