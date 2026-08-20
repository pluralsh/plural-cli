package welcome

import (
	"context"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	welcomebridge "github.com/pluralsh/plural-cli/pkg/bridge/welcome"
	pluralspinner "github.com/pluralsh/plural-cli/tui/components/spinner"
	"github.com/pluralsh/plural-cli/tui/navigation"
	"github.com/pluralsh/plural-cli/tui/theme"
)

type loadedMsg struct{ snapshot welcomebridge.Snapshot }
type failedMsg struct{ err error }

type keyAction uint8

const (
	keyActionNone keyAction = iota
	keyActionUp
	keyActionDown
	keyActionConfirm
)

var keyActionKeystrokes = map[keyAction]string{
	keyActionUp:      "up",
	keyActionDown:    "down",
	keyActionConfirm: "enter",
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
	spinner  spinner.Model
	groups   []group
	cursor   int
	loading  bool
	snapshot welcomebridge.Snapshot
	err      error
	helpOpen bool
}

func New(ctx context.Context, loader welcomebridge.Loader, t theme.Theme) Model {
	return Model{
		ctx:     ctx,
		loader:  loader,
		theme:   t,
		spinner: pluralspinner.New(t),
		groups:  welcomeGroups(),
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
	case tea.KeyPressMsg:
		return m.updateKey(msg)
	default:
		return m, nil
	}
}

func (m Model) updateKey(key tea.KeyPressMsg) (Model, tea.Cmd) {
	if m.helpOpen {
		m.helpOpen = false
		if key.Keystroke() == "esc" {
			return m, nil
		}
	}

	switch actionForKeystroke(key.Keystroke()) {
	case keyActionUp:
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case keyActionDown:
		if m.cursor < len(m.groups)-1 {
			m.cursor++
		}
		return m, nil
	case keyActionConfirm:
		return m.openGroup(m.groups[m.cursor])
	}

	text := key.Text
	if text == "" && key.Code > 0 && key.Code < 128 {
		text = string(rune(key.Code))
	}
	for i, g := range m.groups {
		if text == g.number || text == g.shortcut {
			m.cursor = i
			return m.openGroup(g)
		}
	}
	return m, nil
}

func (m Model) openGroup(g group) (Model, tea.Cmd) {
	if g.route == "" {
		m.helpOpen = true
		return m, nil
	}
	return m, navigation.Navigate(g.route)
}

func (m Model) Snapshot() welcomebridge.Snapshot { return m.snapshot }
