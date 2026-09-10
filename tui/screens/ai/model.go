package ai

import (
	"context"
	"errors"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/pluralsh/plural-cli/pkg/bridge"
	aibridge "github.com/pluralsh/plural-cli/pkg/bridge/ai"
	"github.com/pluralsh/plural-cli/tui/navigation"
	"github.com/pluralsh/plural-cli/tui/theme"
)

type mode uint8

const (
	modeHub mode = iota
	modeChat
)

type keyAction uint8

const (
	keyActionNone keyAction = iota
	keyActionBack
	keyActionMoveUp
	keyActionMoveDown
	keyActionConfirm
	keyActionNewChat
	keyActionPgUp
	keyActionPgDown
	keyActionCancel
	keyActionConnect
)

var keyActionKeystrokes = map[keyAction][]string{
	keyActionBack:     {"esc"},
	keyActionMoveUp:   {"up", "k"},
	keyActionMoveDown: {"down", "j"},
	keyActionConfirm:  {"enter"},
	keyActionNewChat:  {"ctrl+n"},
	keyActionPgUp:     {"pgup"},
	keyActionPgDown:   {"pgdown"},
	keyActionCancel:   {"ctrl+c"},
	keyActionConnect:  {"c"},
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
	{number: "1", shortcut: "c", title: "Chat", blurb: "Plural App assistant", command: "plural ai", usage: "Chat with Plural App about setup, open source, or Kubernetes."},
	{number: "2", shortcut: "a", title: "Agents", blurb: "list and resume runs", command: "plural agents resume [run-id]", usage: "Resume or inspect an agent run from the console-backed TUI flow."},
	{number: "3", shortcut: "w", title: "Workbenches", blurb: "PR follow-up prompts", command: "plural workbenches pr-followup --prompt ...", usage: "Send a follow-up prompt to a workbench-backed pull request."},
}

var errNoApp = errors.New("connect a Plural App profile before chatting")

type replyMsg struct {
	message aibridge.Message
	err     error
	request uint64
}

type initMsg struct{}

// Model owns the AI hub and the in-place Chat conversation.
type Model struct {
	ctx         context.Context
	chat        aibridge.Client
	theme       theme.Theme
	mode        mode
	cursor      int
	history     []aibridge.Message
	input       textinput.Model
	thinking    bool
	needsAuth   bool
	err         error
	request     uint64
	cancel      context.CancelFunc
	chatFromEnd int
}

func New(ctx context.Context, chat aibridge.Client, t theme.Theme) Model {
	input := textinput.New()
	input.Prompt = "› "
	input.Placeholder = "Ask about Plural, open source, or Kubernetes"
	input.CharLimit = 4000
	styles := textinput.DefaultDarkStyles()
	styles.Focused.Text, styles.Focused.Prompt, styles.Focused.Placeholder = t.Body, t.Title, t.Muted
	styles.Blurred = styles.Focused
	input.SetStyles(styles)
	return Model{ctx: ctx, chat: chat, theme: t, input: input}
}

func (m Model) Init() tea.Cmd { return func() tea.Msg { return initMsg{} } }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case initMsg:
		m.needsAuth = false
		m.err = nil
		if m.mode == modeChat {
			m.input.Focus()
		}
		return m, nil
	case replyMsg:
		return m.applyReply(msg)
	case tea.KeyPressMsg:
		if m.mode == modeChat {
			return m.updateChat(msg)
		}
		return m.updateHub(msg)
	}
	if m.mode == modeChat && !m.thinking && !m.needsAuth {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) updateHub(key tea.KeyPressMsg) (Model, tea.Cmd) {
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

func (m Model) updateChat(key tea.KeyPressMsg) (Model, tea.Cmd) {
	action := actionForKeystroke(key.Keystroke())
	if m.thinking {
		if action == keyActionBack || action == keyActionCancel {
			return m.stopThinking(), nil
		}
		return m, nil
	}
	if m.needsAuth {
		if action == keyActionConnect {
			return m, navigation.Navigate(navigation.Access)
		}
		if action == keyActionBack {
			return m.closeChat(), nil
		}
		return m, nil
	}
	switch action {
	case keyActionBack:
		return m.closeChat(), nil
	case keyActionConfirm:
		return m.send()
	case keyActionNewChat:
		m.resetConversation()
		return m, nil
	case keyActionPgUp:
		m.chatFromEnd += 8
		return m, nil
	case keyActionPgDown:
		m.chatFromEnd = max(0, m.chatFromEnd-8)
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(key)
	return m, cmd
}

func (m Model) open() (Model, tea.Cmd) {
	switch m.cursor {
	case 0:
		return m.openChat()
	case 1:
		return m, navigation.Navigate(navigation.Agents)
	default:
		return m, navigation.Navigate(navigation.Workbenches)
	}
}

func (m Model) openChat() (Model, tea.Cmd) {
	m.mode = modeChat
	m.needsAuth = false
	m.err = nil
	if len(m.history) == 0 {
		m.resetConversation()
	}
	m.input.Focus()
	return m, nil
}

func (m Model) closeChat() Model {
	m.mode = modeHub
	m.input.Blur()
	m.thinking = false
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	return m
}

func (m *Model) resetConversation() {
	m.history = []aibridge.Message{{Role: aibridge.RoleSystem, Content: aibridge.Intro}}
	m.chatFromEnd = 0
	m.err = nil
	m.needsAuth = false
	m.input.SetValue("")
}

func (m Model) send() (Model, tea.Cmd) {
	prompt := strings.TrimSpace(m.input.Value())
	if prompt == "" {
		return m, nil
	}
	m.input.SetValue("")
	m.history = append(m.history, aibridge.Message{Role: aibridge.RoleUser, Content: prompt})
	m.err = nil
	m.chatFromEnd = 0
	return m, m.beginChat()
}

func (m *Model) beginChat() tea.Cmd {
	m.thinking = true
	m.request++
	request, history := m.request, append([]aibridge.Message(nil), m.history...)
	chatCtx, cancel := context.WithCancel(m.ctx)
	m.cancel = cancel
	client := m.chat
	return func() tea.Msg {
		defer cancel()
		if client == nil {
			return replyMsg{
				err:     &bridge.Error{Code: bridge.ErrorUnauthenticated, Err: errNoApp},
				request: request,
			}
		}
		message, err := client.Chat(chatCtx, history)
		return replyMsg{message: message, err: err, request: request}
	}
}

func (m Model) applyReply(msg replyMsg) (Model, tea.Cmd) {
	if msg.request != m.request {
		return m, nil
	}
	m.thinking = false
	m.cancel = nil
	if msg.err != nil {
		m.err = msg.err
		m.needsAuth = bridge.IsCode(msg.err, bridge.ErrorUnauthenticated)
		if m.needsAuth {
			m.input.Blur()
		}
		return m, nil
	}
	m.history = append(m.history, msg.message)
	m.chatFromEnd = 0
	return m, nil
}

func (m Model) stopThinking() Model {
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	m.request++
	m.thinking = false
	return m
}

// HasCancellableOperation lets the shell route Ctrl+C here while a reply is in flight.
func (m Model) HasCancellableOperation() bool { return m.thinking && m.cancel != nil }

func (m Model) Snapshot() item { return items[m.cursor] }
