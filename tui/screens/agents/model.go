// Package agents implements the interactive agent-run browser.
package agents

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/pluralsh/plural-cli/pkg/bridge"
	agentsbridge "github.com/pluralsh/plural-cli/pkg/bridge/agents"
	"github.com/pluralsh/plural-cli/tui/navigation"
	"github.com/pluralsh/plural-cli/tui/theme"
)

type mode uint8

const (
	modeList mode = iota
	modeDetail
	modeFilter
	modeRepoPath
	modeResult
)

type initMsg struct{}
type listedMsg struct {
	page    agentsbridge.Page
	err     error
	request uint64
}
type detailMsg struct {
	detail  agentsbridge.Detail
	err     error
	request uint64
}
type resumedMsg struct{ err error }

type Model struct {
	ctx        context.Context
	loader     agentsbridge.Loader
	theme      theme.Theme
	mode       mode
	loading    bool
	err        error
	needsAuth  bool
	request    uint64
	page       agentsbridge.Page
	cursor     int
	filter     string
	input      textinput.Model
	detail     agentsbridge.Detail
	result     string
	execResume bool
	getwd      func() (string, error)
}

func New(ctx context.Context, loader agentsbridge.Loader, t theme.Theme) Model {
	input := textinput.New()
	input.Prompt = "› "
	input.Placeholder = "filter agent runs"
	input.CharLimit = 256
	styles := textinput.DefaultDarkStyles()
	styles.Focused.Text, styles.Focused.Prompt, styles.Focused.Placeholder = t.Body, t.Title, t.Muted
	styles.Blurred = styles.Focused
	input.SetStyles(styles)
	return Model{ctx: ctx, loader: loader, theme: t, input: input, execResume: true, getwd: os.Getwd}
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
			m.page, m.cursor, m.mode = msg.page, clamp(m.cursor, len(msg.page.Items)), modeList
		}
		return m, nil
	case detailMsg:
		if msg.request != m.request {
			return m, nil
		}
		m.loading, m.err = false, msg.err
		m.needsAuth = bridge.IsCode(msg.err, bridge.ErrorUnauthenticated)
		if msg.err == nil {
			m.detail, m.mode = msg.detail, modeDetail
		}
		return m, nil
	case resumedMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.result = "Agent session finished and the TUI resumed."
		}
		m.mode = modeResult
		return m, nil
	case tea.KeyPressMsg:
		return m.updateKey(msg)
	}
	if m.mode == modeFilter || m.mode == modeRepoPath {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
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
			m.input.Blur()
			return m, nil
		case "enter":
			m.filter = strings.TrimSpace(m.input.Value())
			m.input.Blur()
			return m, m.beginList()
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(key)
		return m, cmd
	}
	if m.mode == modeRepoPath {
		switch stroke {
		case "esc":
			m.mode = modeDetail
			m.input.Blur()
			return m, nil
		case "enter":
			path := strings.TrimSpace(m.input.Value())
			if path == "" {
				path = m.workingDirectory()
			}
			if path == "" {
				m.err = errors.New("enter the path to an existing clone of " + m.detail.Repository)
				return m, nil
			}
			m.input.Blur()
			m.loading = true
			m.err = nil
			m.result = "Launching agent resume for " + m.detail.ID + " in " + path
			return m, m.beginResume(path)
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(key)
		return m, cmd
	}
	if m.mode == modeResult {
		if stroke == "esc" || stroke == "enter" {
			m.mode = modeDetail
		}
		return m, nil
	}
	if m.mode == modeDetail {
		switch stroke {
		case "esc":
			m.mode = modeList
		case "r":
			m.mode = modeRepoPath
			m.err = nil
			cwd := m.workingDirectory()
			m.input.CharLimit = 1024
			m.input.SetValue(cwd)
			m.input.Placeholder = cwd
			if m.input.Placeholder == "" {
				m.input.Placeholder = "absolute path to existing clone"
			}
			m.input.Focus()
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
		m.input.Placeholder = "filter agent runs"
		m.input.SetValue(m.filter)
		m.input.Focus()
	case "r":
		return m, m.beginList()
	}
	return m, nil
}

func (m Model) beginResume(path string) tea.Cmd {
	if m.loader == nil {
		return func() tea.Msg { return resumedMsg{err: errors.New("agent services are unavailable")} }
	}
	id, prRef, loader, ctx := m.detail.ID, m.detail.PRRef, m.loader, m.ctx
	run := func() error { return loader.Resume(ctx, id, path, prRef) }
	if !m.execResume {
		return func() tea.Msg { return resumedMsg{err: run()} }
	}
	return tea.Exec(&resumeExecCommand{run: run}, func(err error) tea.Msg {
		return resumedMsg{err: err}
	})
}

// resumeExecCommand runs RestoreAndResume after the TUI releases the terminal
// so the provider agent can use stdin/stdout.
type resumeExecCommand struct {
	run func() error
}

func (c *resumeExecCommand) Run() error {
	if c.run == nil {
		return errors.New("agent resume is not configured")
	}
	return c.run()
}

func (c *resumeExecCommand) SetStdin(io.Reader)  {}
func (c *resumeExecCommand) SetStdout(io.Writer) {}
func (c *resumeExecCommand) SetStderr(io.Writer) {}

func (m Model) workingDirectory() string {
	getwd := m.getwd
	if getwd == nil {
		getwd = os.Getwd
	}
	dir, err := getwd()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(dir)
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
