package ai

import (
	tea "charm.land/bubbletea/v2"

	"github.com/pluralsh/plural-cli/tui/navigation"
	"github.com/pluralsh/plural-cli/tui/theme"
)

type keyAction uint8

const (
	keyActionNone keyAction = iota
	keyActionBack
	keyActionMoveUp
	keyActionMoveDown
	keyActionConfirm
)

var keyActionKeystrokes = map[keyAction][]string{
	keyActionBack:     {"esc"},
	keyActionMoveUp:   {"up", "k"},
	keyActionMoveDown: {"down", "j"},
	keyActionConfirm:  {"enter"},
}

func actionForKeystroke(keystroke string) keyAction {
	for action, keystrokes := range keyActionKeystrokes {
		for _, candidate := range keystrokes {
			if keystroke == candidate {
				return action
			}
		}
	}
	return keyActionNone
}

type item struct {
	number   string
	shortcut string
	title    string
	blurb    string
	command  string
	usage    string
}

var items = []item{
	{number: "1", shortcut: "a", title: "Agents", blurb: "list and resume runs", command: "plural agents resume [run-id]", usage: "Resume or inspect an agent run from the console-backed TUI flow."},
	{number: "2", shortcut: "w", title: "Workbenches", blurb: "PR follow-up prompts", command: "plural workbenches pr-followup --prompt ...", usage: "Send a follow-up prompt to a workbench-backed pull request."},
}

// Model owns the AI hub navigation state.
type Model struct {
	theme  theme.Theme
	cursor int
}

func New(t theme.Theme) Model { return Model{theme: t} }

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m.updateKey(msg)
	}
	return m, nil
}

func (m Model) updateKey(key tea.KeyPressMsg) (Model, tea.Cmd) {
	text := key.Text
	if text == "" && key.Code > 0 && key.Code < 128 {
		text = string(key.Code)
	}
	for i, item := range items {
		if text == item.number || text == item.shortcut {
			m.cursor = i
			return m.open()
		}
	}

	switch actionForKeystroke(key.Keystroke()) {
	case keyActionMoveUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case keyActionMoveDown:
		if m.cursor < len(items)-1 {
			m.cursor++
		}
	case keyActionConfirm:
		return m.open()
	case keyActionBack:
		return m, navigation.Navigate(navigation.Welcome)
	}
	return m, nil
}

func (m Model) open() (Model, tea.Cmd) {
	if m.cursor == 0 {
		return m, navigation.Navigate(navigation.Agents)
	}
	return m, navigation.Navigate(navigation.Workbenches)
}

func (m Model) Snapshot() item { return items[m.cursor] }
