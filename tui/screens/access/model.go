// Package access implements the Phase 1 identity and connection screen. It
// depends only on access.Manager and can be developed independently of
// the root shell.
package access

import (
	"context"
	"errors"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	accessbridge "github.com/pluralsh/plural-cli/pkg/bridge/access"
	"github.com/pluralsh/plural-cli/tui/navigation"
	"github.com/pluralsh/plural-cli/tui/theme"
)

type loadedMsg struct {
	snapshot accessbridge.Snapshot
	err      error
}
type changedMsg struct{ err error }
type accountsMsg struct {
	accounts []accessbridge.ServiceAccount
	err      error
}
type authorizedMsg struct {
	authorization accessbridge.DeviceAuthorization
	requestID     uint64
	err           error
}
type loggedInMsg struct {
	profile   accessbridge.Profile
	requestID uint64
	err       error
}

type mode uint8

const (
	modeProfiles mode = iota
	modeAccounts
	modeConsoleForm
	modeDeviceLogin
)

type keyAction uint8

const (
	keyActionNone keyAction = iota
	keyActionBack
	keyActionCancelOperation
	keyActionNextPanel
	keyActionPreviousPanel
	keyActionMoveUp
	keyActionMoveDown
	keyActionConfirm
	keyActionRefresh
	keyActionDeviceLogin
	keyActionAddConsole
	keyActionImpersonate
	keyActionStopImpersonating
)

var keyActionKeystrokes = map[keyAction][]string{
	keyActionBack:              {"esc"},
	keyActionCancelOperation:   {"ctrl+c"},
	keyActionNextPanel:         {"tab"},
	keyActionPreviousPanel:     {"shift+tab"},
	keyActionMoveUp:            {"up", "k"},
	keyActionMoveDown:          {"down", "j"},
	keyActionConfirm:           {"enter"},
	keyActionRefresh:           {"r"},
	keyActionDeviceLogin:       {"n"},
	keyActionAddConsole:        {"c"},
	keyActionImpersonate:       {"i"},
	keyActionStopImpersonating: {"x"},
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

// Model owns only Access-screen interaction state.
type Model struct {
	ctx           context.Context
	manager       accessbridge.Manager
	theme         theme.Theme
	loading       bool
	snapshot      accessbridge.Snapshot
	err           error
	panel         int
	appCursor     int
	consoleCursor int
	accountCursor int
	mode          mode
	authorization accessbridge.DeviceAuthorization
	operationCtx  context.Context
	cancel        context.CancelFunc
	form          []textinput.Model
	formIndex     int
	loginRequest  uint64
}

func New(ctx context.Context, manager accessbridge.Manager, t theme.Theme) Model {
	return Model{ctx: ctx, manager: manager, theme: t, loading: manager != nil, form: newConsoleForm(t)}
}

func newConsoleForm(t theme.Theme) []textinput.Model {
	values := make([]textinput.Model, 3)
	for i, placeholder := range []string{"Profile name", "https://console.example.com", "Console token"} {
		values[i] = textinput.New()
		values[i].Prompt = "› "
		values[i].Placeholder = placeholder
		values[i].CharLimit = 256
		styles := textinput.DefaultDarkStyles()
		styles.Focused.Text = t.Body
		styles.Focused.Prompt = t.Title
		styles.Focused.Placeholder = t.Muted
		styles.Blurred = styles.Focused
		values[i].SetStyles(styles)
	}
	values[2].EchoMode = textinput.EchoPassword
	return values
}

func (m Model) Init() tea.Cmd {
	if m.manager == nil {
		return nil
	}
	return m.load
}
func (m Model) load() tea.Msg {
	snapshot, err := m.manager.Load(m.ctx)
	return loadedMsg{snapshot, err}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case loadedMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.snapshot = msg.snapshot
		}
		return m, nil
	case changedMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.mode = modeProfiles
			return m, m.load
		}
		return m, nil
	case accountsMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.snapshot.ServiceAccounts = msg.accounts
			m.mode = modeAccounts
			m.accountCursor = 0
		}
		return m, nil
	case authorizedMsg:
		if msg.requestID != m.loginRequest {
			return m, nil
		}
		m.loading = false
		m.err = msg.err
		if msg.err != nil {
			m.cancel = nil
			m.operationCtx = nil
			m.mode = modeProfiles
			return m, nil
		}
		m.authorization = msg.authorization
		m.mode = modeDeviceLogin
		m.loading = true
		loginCtx := m.operationCtx
		if loginCtx == nil {
			var cancel context.CancelFunc
			loginCtx, cancel = context.WithCancel(m.ctx)
			m.operationCtx, m.cancel = loginCtx, cancel
		}
		return m, func() tea.Msg {
			profile, err := m.manager.CompleteDeviceLogin(loginCtx, "default", msg.authorization, "app.plural.sh")
			return loggedInMsg{profile: profile, requestID: msg.requestID, err: err}
		}
	case loggedInMsg:
		if msg.requestID != m.loginRequest {
			return m, nil
		}
		m.loading = false
		m.cancel = nil
		m.operationCtx = nil
		m.err = msg.err
		m.mode = modeProfiles
		if msg.err == nil {
			return m, m.load
		}
		return m, nil
	case tea.KeyPressMsg:
		return m.updateKey(msg)
	}
	if m.mode == modeConsoleForm {
		return m.updateForm(msg)
	}
	return m, nil
}

func (m Model) updateKey(key tea.KeyPressMsg) (Model, tea.Cmd) {
	action := actionForKeystroke(key.Keystroke())
	if action == keyActionBack || (action == keyActionCancelOperation && m.cancel != nil) {
		if m.cancel != nil {
			m.cancel()
			m.loginRequest++
			m.cancel = nil
			m.operationCtx = nil
			m.loading = false
			m.mode = modeProfiles
			return m, nil
		}
		if m.mode != modeProfiles {
			m.mode = modeProfiles
			m.err = nil
			return m, nil
		}
		return m, navigation.Navigate(navigation.Welcome)
	}
	if m.loading {
		return m, nil
	}
	if m.manager == nil {
		m.err = errors.New("access services are unavailable")
		return m, nil
	}
	if m.mode == modeConsoleForm {
		return m.updateForm(key)
	}
	if m.mode == modeAccounts {
		return m.updateAccounts(key)
	}
	switch action {
	case keyActionNextPanel:
		m.panel = (m.panel + 1) % 2
	case keyActionPreviousPanel:
		m.panel = (m.panel + 1) % 2
	case keyActionMoveUp:
		m = m.move(-1)
	case keyActionMoveDown:
		m = m.move(1)
	case keyActionConfirm:
		return m.activate()
	case keyActionRefresh:
		m.loading = true
		return m, m.load
	case keyActionDeviceLogin:
		m.loading = true
		m.loginRequest++
		requestID := m.loginRequest
		loginCtx, cancel := context.WithCancel(m.ctx)
		m.operationCtx, m.cancel = loginCtx, cancel
		return m, func() tea.Msg {
			authorization, err := m.manager.BeginDeviceLogin(loginCtx, "app.plural.sh")
			return authorizedMsg{authorization: authorization, requestID: requestID, err: err}
		}
	case keyActionAddConsole:
		m.mode = modeConsoleForm
		m.formIndex = 0
		for i := range m.form {
			m.form[i].Reset()
			m.form[i].Blur()
		}
		m.form[0].Focus()
	case keyActionImpersonate:
		m.loading = true
		return m, func() tea.Msg {
			accounts, err := m.manager.SearchServiceAccounts(m.ctx, "")
			return accountsMsg{accounts, err}
		}
	case keyActionStopImpersonating:
		m.manager.StopImpersonating()
		m.loading = true
		return m, m.load
	}
	return m, nil
}

// HasCancellableOperation lets the shell route Ctrl+C to this screen before
// applying its global quit binding.
func (m Model) HasCancellableOperation() bool { return m.cancel != nil }

func (m Model) updateAccounts(key tea.KeyPressMsg) (Model, tea.Cmd) {
	count := len(m.snapshot.ServiceAccounts)
	switch actionForKeystroke(key.Keystroke()) {
	case keyActionMoveUp:
		m.accountCursor = clampCursor(m.accountCursor-1, count)
	case keyActionMoveDown:
		m.accountCursor = clampCursor(m.accountCursor+1, count)
	case keyActionConfirm:
		if count > 0 {
			email := m.snapshot.ServiceAccounts[m.accountCursor].Email
			m.loading = true
			return m, func() tea.Msg { return changedMsg{err: m.manager.Impersonate(m.ctx, email)} }
		}
	}
	return m, nil
}

func (m Model) updateForm(msg tea.Msg) (Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok && actionForKeystroke(key.Keystroke()) == keyActionConfirm {
		if m.formIndex < len(m.form)-1 {
			m.form[m.formIndex].Blur()
			m.formIndex++
			m.form[m.formIndex].Focus()
			return m, nil
		}
		name, url, token := m.form[0].Value(), m.form[1].Value(), m.form[2].Value()
		m.loading = true
		m.mode = modeProfiles
		return m, func() tea.Msg {
			_, err := m.manager.AddConsoleProfile(m.ctx, name, url, token)
			return changedMsg{err: err}
		}
	}
	var cmd tea.Cmd
	m.form[m.formIndex], cmd = m.form[m.formIndex].Update(msg)
	return m, cmd
}

func (m Model) move(delta int) Model {
	if m.panel == 0 {
		m.appCursor = clampCursor(m.appCursor+delta, len(m.snapshot.State.Profiles))
	} else {
		m.consoleCursor = clampCursor(m.consoleCursor+delta, len(m.snapshot.State.ConsoleProfiles))
	}
	return m
}
func clampCursor(cursor, count int) int {
	if count == 0 {
		return 0
	}
	if cursor < 0 {
		return count - 1
	}
	if cursor >= count {
		return 0
	}
	return cursor
}
func (m Model) activate() (Model, tea.Cmd) {
	if m.panel == 0 && len(m.snapshot.State.Profiles) > 0 {
		id := m.snapshot.State.Profiles[m.appCursor].ID
		m.loading = true
		return m, func() tea.Msg { return changedMsg{err: m.manager.ActivateProfile(m.ctx, id)} }
	}
	if m.panel == 1 && len(m.snapshot.State.ConsoleProfiles) > 0 {
		id := m.snapshot.State.ConsoleProfiles[m.consoleCursor].ID
		m.loading = true
		return m, func() tea.Msg { return changedMsg{err: m.manager.ActivateConsole(m.ctx, id)} }
	}
	return m, nil
}
