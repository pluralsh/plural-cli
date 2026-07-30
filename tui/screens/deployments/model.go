package deployments

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/pluralsh/plural-cli/tui/navigation"
	"github.com/pluralsh/plural-cli/tui/theme"
)

type resourceID uint8

const (
	resourceServices resourceID = iota
	resourceClusters
	resourceRepositories
	resourcePipelines
	resourceNotifications
	resourceProviders
)

type resource struct {
	id       resourceID
	number   string
	shortcut string
	title    string
	blurb    string
	soon     bool
	route    navigation.Route
}

func resources() []resource {
	return []resource{
		{id: resourceServices, number: "1", shortcut: "s", title: "Services", blurb: "browse · kick · create · …", route: navigation.Services},
		{id: resourceClusters, number: "2", shortcut: "c", title: "Clusters", blurb: "list · describe", route: navigation.Clusters},
		{id: resourceRepositories, number: "3", shortcut: "r", title: "Repositories", blurb: "list · describe", route: navigation.Repositories},
		{id: resourcePipelines, number: "4", shortcut: "p", title: "Pipelines", blurb: "list · describe", route: navigation.Pipelines},
		{id: resourceNotifications, number: "5", shortcut: "n", title: "Notifications", blurb: "sinks", soon: true},
		{id: resourceProviders, number: "6", shortcut: "v", title: "Providers", blurb: "list", soon: true},
	}
}

type keyAction uint8

const (
	keyActionNone keyAction = iota
	keyActionUp
	keyActionDown
	keyActionConfirm
	keyActionBack
)

var keyActionKeystrokes = map[keyAction]string{
	keyActionUp:      "up",
	keyActionDown:    "down",
	keyActionConfirm: "enter",
	keyActionBack:    "esc",
}

func actionForKeystroke(keystroke string) keyAction {
	for action, candidate := range keyActionKeystrokes {
		if keystroke == candidate {
			return action
		}
	}
	return keyActionNone
}

// Model is the CD / Deployments hub.
type Model struct {
	theme   theme.Theme
	items   []resource
	cursor  int
	console string
}

// New creates the deployments hub. consoleURL is shown in the connection panel.
func New(_ context.Context, t theme.Theme, consoleURL string) Model {
	return Model{
		theme:   t,
		items:   resources(),
		console: consoleURL,
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	switch actionForKeystroke(key.Keystroke()) {
	case keyActionUp:
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case keyActionDown:
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}
		return m, nil
	case keyActionConfirm:
		return m.openResource(m.items[m.cursor])
	case keyActionBack:
		return m, navigation.Navigate(navigation.Welcome)
	}

	text := key.Text
	if text == "" && key.Code > 0 && key.Code < 128 {
		text = string(rune(key.Code))
	}
	for i, item := range m.items {
		if text == item.number || text == item.shortcut {
			m.cursor = i
			return m.openResource(item)
		}
	}
	return m, nil
}

func (m Model) openResource(item resource) (Model, tea.Cmd) {
	if item.soon || item.route == "" {
		return m, nil
	}
	return m, navigation.Navigate(item.route)
}

// SetConsoleURL refreshes the connection panel when context changes.
func (m *Model) SetConsoleURL(url string) { m.console = url }
