// Package diagnostics renders credential-free local context and startup
// diagnostics behind the same loader used by the welcome screen.
package diagnostics

import (
	"context"

	tea "charm.land/bubbletea/v2"

	welcomebridge "github.com/pluralsh/plural-cli/pkg/bridge/welcome"
	"github.com/pluralsh/plural-cli/tui/navigation"
	"github.com/pluralsh/plural-cli/tui/theme"
)

type loadedMsg struct {
	snapshot welcomebridge.Snapshot
	err      error
}

type keyAction uint8

const (
	keyActionNone keyAction = iota
	keyActionBack
	keyActionRefresh
)

var keyActionKeystrokes = map[keyAction]string{
	keyActionBack:    "esc",
	keyActionRefresh: "r",
}

func actionForKeystroke(keystroke string) keyAction {
	for action, candidate := range keyActionKeystrokes {
		if keystroke == candidate {
			return action
		}
	}

	return keyActionNone
}

type Model struct {
	ctx      context.Context
	loader   welcomebridge.Loader
	theme    theme.Theme
	loading  bool
	snapshot welcomebridge.Snapshot
	err      error
}

func New(ctx context.Context, loader welcomebridge.Loader, t theme.Theme) Model {
	return Model{ctx: ctx, loader: loader, theme: t, loading: loader != nil}
}
func (m Model) Init() tea.Cmd {
	if m.loader == nil {
		return nil
	}
	return m.load
}
func (m Model) load() tea.Msg { value, err := m.loader.Load(m.ctx); return loadedMsg{value, err} }
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case loadedMsg:
		m.loading = false
		m.snapshot = msg.snapshot
		m.err = msg.err
	case tea.KeyPressMsg:
		switch actionForKeystroke(msg.Keystroke()) {
		case keyActionBack:
			return m, navigation.Navigate(navigation.Welcome)
		case keyActionRefresh:
			m.loading = true
			return m, m.load
		}
	}
	return m, nil
}
