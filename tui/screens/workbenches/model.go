// Package workbenches implements interactive workbench browsing and job follow-up.
package workbenches

import (
	"context"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/pluralsh/plural-cli/pkg/bridge"
	workbenchesbridge "github.com/pluralsh/plural-cli/pkg/bridge/workbenches"
	"github.com/pluralsh/plural-cli/tui/components/page"
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

// Model is the Workbenches screen: browse jobs and queue a follow-up on one.
type Model struct {
	ctx          context.Context
	loader       workbenchesbridge.Loader
	theme        theme.Theme
	mode         mode
	loading      bool
	err          error
	needsAuth    bool
	request      uint64
	page         workbenchesbridge.Page
	cursor       int
	filter       string
	filterInput  textinput.Model
	prompt       textinput.Model
	detail       workbenchesbridge.Detail
	result       workbenchesbridge.PromptResult
	returnTo     mode
	detailOffset int
	viewW        int
	viewH        int
}

// New constructs the workbenches screen.
func New(ctx context.Context, loader workbenchesbridge.Loader, t theme.Theme) Model {
	filter := textinput.New()
	filter.Prompt = "› "
	filter.Placeholder = "filter workbench jobs"
	styles := textinput.DefaultDarkStyles()
	styles.Focused.Text, styles.Focused.Prompt, styles.Focused.Placeholder = t.Body, t.Title, t.Muted
	styles.Blurred = styles.Focused
	filter.SetStyles(styles)
	prompt := textinput.New()
	prompt.Prompt = "› "
	prompt.Placeholder = "follow-up prompt"
	prompt.CharLimit = 4000
	prompt.SetStyles(styles)
	return Model{ctx: ctx, loader: loader, theme: t, filterInput: filter, prompt: prompt}
}

func (m Model) Init() tea.Cmd { return func() tea.Msg { return initMsg{} } }

func (m *Model) beginList() tea.Cmd {
	m.loading = true
	m.request++
	request, loader, ctx, query := m.request, m.loader, m.ctx, m.filter
	return func() tea.Msg {
		page, err := loader.List(ctx, nil, query)
		return listedMsg{page: page, err: err, request: request}
	}
}

func (m *Model) beginDetail(id string) tea.Cmd {
	m.loading = true
	m.request++
	request, loader, ctx := m.request, m.loader, m.ctx
	return func() tea.Msg {
		detail, err := loader.Get(ctx, id)
		return detailMsg{detail: detail, err: err, request: request}
	}
}

func (m *Model) beginFollowup() tea.Cmd {
	m.mode = modeOperating
	m.loading = true
	m.err = nil
	m.request++
	request, loader, ctx := m.request, m.loader, m.ctx
	jobID := m.detail.ID
	prompt := strings.TrimSpace(m.prompt.Value())
	return func() tea.Msg {
		result, err := loader.FollowUp(ctx, jobID, prompt, 0)
		return followedMsg{result: result, err: err, request: request}
	}
}

func (m Model) startFollowup() Model {
	if m.mode == modeList {
		if len(m.page.Items) == 0 {
			return m
		}
		m.detail = workbenchesbridge.Detail{Summary: m.page.Items[m.cursor]}
	}
	if m.detail.ID == "" {
		return m
	}
	m.returnTo = m.mode
	m.err = nil
	m.mode = modePrompt
	m.prompt.SetValue("")
	m.prompt.Focus()
	return m
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
			m.detailOffset = 0
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
	case tea.WindowSizeMsg:
		m.viewW, m.viewH = msg.Width, msg.Height
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
	stroke := key.Keystroke()
	if m.mode == modeFilter {
		switch stroke {
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
		switch stroke {
		case "esc":
			m.mode = m.backMode()
			m.prompt.Blur()
			return m, nil
		case "enter":
			if strings.TrimSpace(m.prompt.Value()) == "" {
				return m, nil
			}
			m.mode = modeReview
			m.prompt.Blur()
			return m, nil
		}
		var cmd tea.Cmd
		m.prompt, cmd = m.prompt.Update(key)
		return m, cmd
	}
	if m.mode == modeReview {
		if stroke == "esc" {
			m.mode = modePrompt
			m.prompt.Focus()
			return m, nil
		}
		if stroke == "enter" {
			return m, m.beginFollowup()
		}
		return m, nil
	}
	if m.mode == modeOperating {
		return m, nil
	}
	if m.mode == modeResult {
		if stroke == "esc" || stroke == "enter" {
			m.mode = m.backMode()
		}
		return m, nil
	}
	if m.mode == modeDetail {
		switch stroke {
		case "esc":
			m.mode = modeList
			m.detailOffset = 0
		case "f":
			return m.startFollowup(), nil
		case "up", "k":
			m.scrollDetail(-1)
		case "down", "j":
			m.scrollDetail(1)
		case "pgup":
			m.scrollDetail(-m.detailVisible())
		case "pgdown":
			m.scrollDetail(m.detailVisible())
		case "home":
			m.detailOffset = 0
		case "end":
			m.scrollDetail(len(m.detailLines(m.contentWidth())))
		}
		return m, nil
	}
	if stroke == "esc" {
		return m, navigation.Navigate(navigation.AI)
	}
	if m.loading {
		return m, nil
	}
	if m.needsAuth && stroke == "c" {
		return m, navigation.Navigate(navigation.Access)
	}
	switch stroke {
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
	case "f":
		return m.startFollowup(), nil
	}
	return m, nil
}

func (m Model) backMode() mode {
	if m.returnTo == modeDetail {
		return modeDetail
	}
	return modeList
}

func (m *Model) scrollDetail(delta int) {
	inner := m.detailVisible()
	maxOff := max(0, len(m.detailLines(m.contentWidth()))-inner)
	m.detailOffset = min(max(0, m.detailOffset+delta), maxOff)
}

func (m Model) detailVisible() int {
	return max(1, detailPanelHeight(m.viewH)-2)
}

func (m Model) contentWidth() int {
	width := m.viewW
	if width <= 0 {
		width = page.DefaultWidth
	}
	return page.ContentWidth(width)
}

func detailPanelHeight(height int) int {
	if height <= 0 {
		height = page.DefaultHeight
	}
	return max(12, height-10)
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
