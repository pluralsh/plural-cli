// Package workbenches implements interactive workbench job follow-ups.
package workbenches

import (
	"context"
	"strings"
	"time"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/pluralsh/plural-cli/pkg/bridge"
	workbenchesbridge "github.com/pluralsh/plural-cli/pkg/bridge/workbenches"
	"github.com/pluralsh/plural-cli/tui/navigation"
	"github.com/pluralsh/plural-cli/tui/theme"
)

type mode uint8

const (
	modeList mode = iota
	modeDetail
	modeFilter
	modePrompt
	modeReview
	modeOperating
	modeResult
)

type initMsg struct{}
type listedMsg struct {
	page    workbenchesbridge.Page
	err     error
	request uint64
}
type detailMsg struct {
	detail  workbenchesbridge.Detail
	err     error
	request uint64
}
type followedMsg struct {
	result  workbenchesbridge.PromptResult
	err     error
	request uint64
}

type Model struct {
	ctx         context.Context
	loader      workbenchesbridge.Loader
	theme       theme.Theme
	mode        mode
	loading     bool
	err         error
	needsAuth   bool
	request     uint64
	page        workbenchesbridge.Page
	cursor      int
	filter      string
	filterInput textinput.Model
	prompt      textarea.Model
	detail      workbenchesbridge.Detail
	result      workbenchesbridge.PromptResult
}

func New(ctx context.Context, loader workbenchesbridge.Loader, t theme.Theme) Model {
	filter := textinput.New()
	filter.Prompt = "› "
	filter.Placeholder = "filter workbench jobs"
	styles := textinput.DefaultDarkStyles()
	styles.Focused.Text, styles.Focused.Prompt, styles.Focused.Placeholder = t.Body, t.Title, t.Muted
	styles.Blurred = styles.Focused
	filter.SetStyles(styles)
	prompt := textarea.New()
	prompt.Prompt = "│ "
	prompt.Placeholder = "Follow-up prompt"
	prompt.SetWidth(70)
	prompt.SetHeight(5)
	return Model{ctx: ctx, loader: loader, theme: t, filterInput: filter, prompt: prompt}
}

func (m Model) Init() tea.Cmd { return func() tea.Msg { return initMsg{} } }
func (m *Model) beginList() tea.Cmd {
	m.loading = true
	m.request++
	r, l, c, q := m.request, m.loader, m.ctx, m.filter
	return func() tea.Msg { p, e := l.List(c, nil, q); return listedMsg{p, e, r} }
}
func (m *Model) beginDetail(id string) tea.Cmd {
	m.loading = true
	m.request++
	r, l, c := m.request, m.loader, m.ctx
	return func() tea.Msg { d, e := l.Get(c, id); return detailMsg{d, e, r} }
}
func (m *Model) beginFollowup() tea.Cmd {
	m.mode = modeOperating
	m.loading = true
	m.request++
	r, l, c, id, p := m.request, m.loader, m.ctx, m.detail.ID, strings.TrimSpace(m.prompt.Value())
	return func() tea.Msg { result, e := l.FollowUp(c, id, p, 0); return followedMsg{result, e, r} }
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case initMsg:
		m.mode, m.err, m.needsAuth = modeList, nil, false
		if m.loader == nil {
			return m, nil
		}
		return m, m.beginList()
	case listedMsg:
		if msg.request != m.request {
			return m, nil
		}
		m.loading, m.err = false, msg.err
		m.needsAuth = bridge.IsCode(msg.err, bridge.ErrorUnauthenticated)
		if msg.err == nil {
			m.page = msg.page
			m.cursor = clamp(m.cursor, len(msg.page.Items))
			m.mode = modeList
		}
		return m, nil
	case detailMsg:
		if msg.request != m.request {
			return m, nil
		}
		m.loading, m.err = false, msg.err
		if msg.err == nil {
			m.detail = msg.detail
			m.mode = modeDetail
		}
		return m, nil
	case followedMsg:
		if msg.request != m.request {
			return m, nil
		}
		m.loading, m.err = false, msg.err
		m.result = msg.result
		m.mode = modeResult
		return m, nil
	case tea.KeyPressMsg:
		return m.updateKey(msg)
	}
	if m.mode == modeFilter {
		var cmd tea.Cmd
		m.filterInput, cmd = m.filterInput.Update(msg)
		return m, cmd
	}
	if m.mode == modePrompt {
		var cmd tea.Cmd
		m.prompt, cmd = m.prompt.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) updateKey(key tea.KeyPressMsg) (Model, tea.Cmd) {
	s := key.Keystroke()
	if m.mode == modeFilter {
		switch s {
		case "esc":
			m.mode = modeList
			m.filterInput.Blur()
			return m, nil
		case "enter":
			m.filter = strings.TrimSpace(m.filterInput.Value())
			m.filterInput.Blur()
			return m, m.beginList()
		}
		var cmd tea.Cmd
		m.filterInput, cmd = m.filterInput.Update(key)
		return m, cmd
	}
	if m.mode == modePrompt {
		switch s {
		case "esc":
			m.mode = modeDetail
			m.prompt.Blur()
			return m, nil
		case "ctrl+s":
			if strings.TrimSpace(m.prompt.Value()) != "" {
				m.mode = modeReview
				m.prompt.Blur()
			}
			return m, nil
		}
		var cmd tea.Cmd
		m.prompt, cmd = m.prompt.Update(key)
		return m, cmd
	}
	if m.mode == modeReview {
		if s == "esc" {
			m.mode = modePrompt
			m.prompt.Focus()
			return m, nil
		}
		if s == "enter" {
			return m, m.beginFollowup()
		}
		return m, nil
	}
	if m.mode == modeOperating {
		return m, nil
	}
	if m.mode == modeResult {
		if s == "esc" || s == "enter" {
			m.mode = modeDetail
		}
		return m, nil
	}
	if m.mode == modeDetail {
		switch s {
		case "esc":
			m.mode = modeList
		case "f":
			m.mode = modePrompt
			m.prompt.Reset()
			m.prompt.Focus()
		}
		return m, nil
	}
	if s == "esc" {
		return m, navigation.Navigate(navigation.AI)
	}
	if m.loading {
		return m, nil
	}
	if m.needsAuth && s == "c" {
		return m, navigation.Navigate(navigation.Access)
	}
	switch s {
	case "up", "k":
		m.cursor = clamp(m.cursor-1, len(m.page.Items))
	case "down", "j":
		m.cursor = clamp(m.cursor+1, len(m.page.Items))
	case "enter":
		if len(m.page.Items) > 0 {
			return m, m.beginDetail(m.page.Items[m.cursor].ID)
		}
	case "/":
		m.mode = modeFilter
		m.filterInput.SetValue(m.filter)
		m.filterInput.Focus()
	case "r":
		return m, m.beginList()
	}
	return m, nil
}

func clamp(cursor, count int) int {
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

var _ = time.Second
