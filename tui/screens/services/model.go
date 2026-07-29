// Package services implements the Phase 2 read-only Console services browser.
package services

import (
	"context"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/pluralsh/plural-cli/pkg/bridge"
	servicesbridge "github.com/pluralsh/plural-cli/pkg/bridge/services"
	"github.com/pluralsh/plural-cli/tui/navigation"
	"github.com/pluralsh/plural-cli/tui/theme"
)

type mode uint8

const (
	modeClusters mode = iota
	modeList
	modeDetail
	modeFilter
)

type keyAction uint8

const (
	keyActionNone keyAction = iota
	keyActionBack
	keyActionMoveUp
	keyActionMoveDown
	keyActionConfirm
	keyActionRefresh
	keyActionFilter
	keyActionNextPage
	keyActionPrevPage
	keyActionConnectConsole
)

var keyActionKeystrokes = map[keyAction][]string{
	keyActionBack:           {"esc"},
	keyActionMoveUp:         {"up", "k"},
	keyActionMoveDown:       {"down", "j"},
	keyActionConfirm:        {"enter"},
	keyActionRefresh:        {"r"},
	keyActionFilter:         {"/"},
	keyActionNextPage:       {"n", "right"},
	keyActionPrevPage:       {"p", "left"},
	keyActionConnectConsole: {"c"},
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

type initMsg struct{}
type clustersMsg struct {
	clusters []servicesbridge.Cluster
	err      error
	request  uint64
}
type listedMsg struct {
	page    servicesbridge.Page
	err     error
	request uint64
}
type detailMsg struct {
	detail  servicesbridge.Detail
	err     error
	request uint64
}

// Model owns only Services-screen interaction state.
type Model struct {
	ctx       context.Context
	loader    servicesbridge.Loader
	theme     theme.Theme
	mode      mode
	loading   bool
	err       error
	needsAuth bool

	clusters         []servicesbridge.Cluster
	clusterCursor    int
	cluster          servicesbridge.Cluster
	clusterFilter    string
	serviceFilter    string
	filterInput      textinput.Model
	filteringCluster bool

	page        servicesbridge.Page
	cursor      int
	after       *string
	prevCursors []string
	request     uint64

	detail     servicesbridge.Detail
	detailID   string
	listCursor int
	listAfter  *string
	listFilter string
	listPrev   []string
}

func New(ctx context.Context, loader servicesbridge.Loader, t theme.Theme) Model {
	input := textinput.New()
	input.Prompt = "› "
	input.Placeholder = "filter"
	input.CharLimit = 128
	styles := textinput.DefaultDarkStyles()
	styles.Focused.Text = t.Body
	styles.Focused.Prompt = t.Title
	styles.Focused.Placeholder = t.Muted
	styles.Blurred = styles.Focused
	input.SetStyles(styles)
	return Model{ctx: ctx, loader: loader, theme: t, loading: loader != nil, filterInput: input, mode: modeClusters}
}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg { return initMsg{} }
}

func (m *Model) beginClusters() tea.Cmd {
	m.loading = true
	m.request++
	request := m.request
	query := m.clusterFilter
	loader := m.loader
	ctx := m.ctx
	return func() tea.Msg {
		clusters, err := loader.ListClusters(ctx, query)
		return clustersMsg{clusters: clusters, err: err, request: request}
	}
}

func (m *Model) beginList(after *string) tea.Cmd {
	m.loading = true
	m.request++
	request := m.request
	query := m.serviceFilter
	clusterID := m.cluster.ID
	loader := m.loader
	ctx := m.ctx
	return func() tea.Msg {
		page, err := loader.List(ctx, clusterID, after, query)
		return listedMsg{page: page, err: err, request: request}
	}
}

func (m *Model) beginDetail(id string) tea.Cmd {
	m.loading = true
	m.request++
	request := m.request
	loader := m.loader
	ctx := m.ctx
	return func() tea.Msg {
		detail, err := loader.Get(ctx, id)
		return detailMsg{detail: detail, err: err, request: request}
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case initMsg:
		m.mode = modeClusters
		m.cluster = servicesbridge.Cluster{}
		m.clusters = nil
		m.clusterCursor = 0
		m.page = servicesbridge.Page{}
		m.after = nil
		m.prevCursors = nil
		m.cursor = 0
		m.err = nil
		m.needsAuth = false
		if m.loader == nil {
			m.loading = false
			return m, nil
		}
		return m, m.beginClusters()
	case clustersMsg:
		if msg.request != m.request {
			return m, nil
		}
		m.loading = false
		m.err = msg.err
		m.needsAuth = bridge.IsCode(msg.err, bridge.ErrorUnauthenticated)
		if msg.err == nil {
			m.clusters = msg.clusters
			m.clusterCursor = clampCursor(m.clusterCursor, len(m.clusters))
			m.mode = modeClusters
		}
		return m, nil
	case listedMsg:
		if msg.request != m.request {
			return m, nil
		}
		m.loading = false
		m.err = msg.err
		m.needsAuth = bridge.IsCode(msg.err, bridge.ErrorUnauthenticated)
		if msg.err == nil {
			m.page = msg.page
			m.cursor = clampCursor(m.cursor, len(m.page.Items))
			m.mode = modeList
		}
		return m, nil
	case detailMsg:
		if msg.request != m.request {
			return m, nil
		}
		m.loading = false
		m.err = msg.err
		m.needsAuth = bridge.IsCode(msg.err, bridge.ErrorUnauthenticated)
		if msg.err == nil {
			m.detail = msg.detail
			m.mode = modeDetail
		}
		return m, nil
	case tea.KeyPressMsg:
		return m.updateKey(msg)
	}
	if m.mode == modeFilter {
		var cmd tea.Cmd
		m.filterInput, cmd = m.filterInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) updateKey(key tea.KeyPressMsg) (Model, tea.Cmd) {
	action := actionForKeystroke(key.Keystroke())
	if m.mode == modeFilter {
		switch action {
		case keyActionBack:
			m.mode = modeList
			if m.filteringCluster {
				m.mode = modeClusters
			}
			m.filterInput.Blur()
			return m, nil
		case keyActionConfirm:
			value := strings.TrimSpace(m.filterInput.Value())
			m.filterInput.Blur()
			if m.filteringCluster {
				m.clusterFilter = value
				m.mode = modeClusters
				m.clusterCursor = 0
				return m, m.beginClusters()
			}
			m.serviceFilter = value
			m.mode = modeList
			m.after = nil
			m.prevCursors = nil
			m.cursor = 0
			return m, m.beginList(nil)
		}
		var cmd tea.Cmd
		m.filterInput, cmd = m.filterInput.Update(key)
		return m, cmd
	}
	if action == keyActionBack {
		switch m.mode {
		case modeDetail:
			m.mode = modeList
			m.err = nil
			m.cursor = m.listCursor
			m.after = m.listAfter
			m.serviceFilter = m.listFilter
			m.prevCursors = append([]string(nil), m.listPrev...)
			return m, nil
		case modeList:
			m.mode = modeClusters
			m.page = servicesbridge.Page{}
			m.cluster = servicesbridge.Cluster{}
			m.err = nil
			m.after = nil
			m.prevCursors = nil
			return m, nil
		default:
			return m, navigation.Navigate(navigation.Deployments)
		}
	}
	if m.loading {
		return m, nil
	}
	if m.needsAuth && action == keyActionConnectConsole {
		return m, navigation.Navigate(navigation.Access)
	}
	if m.mode == modeDetail {
		if action == keyActionRefresh && m.detailID != "" {
			return m, m.beginDetail(m.detailID)
		}
		return m, nil
	}
	if m.mode == modeClusters {
		return m.updateClusters(action)
	}
	return m.updateList(action)
}

func (m Model) updateClusters(action keyAction) (Model, tea.Cmd) {
	switch action {
	case keyActionMoveUp:
		m.clusterCursor = clampCursor(m.clusterCursor-1, len(m.clusters))
	case keyActionMoveDown:
		m.clusterCursor = clampCursor(m.clusterCursor+1, len(m.clusters))
	case keyActionConfirm:
		if len(m.clusters) == 0 {
			return m, nil
		}
		m.cluster = m.clusters[m.clusterCursor]
		m.serviceFilter = ""
		m.after = nil
		m.prevCursors = nil
		m.cursor = 0
		return m, m.beginList(nil)
	case keyActionRefresh:
		return m, m.beginClusters()
	case keyActionFilter:
		m.mode = modeFilter
		m.filteringCluster = true
		m.filterInput.Placeholder = "filter clusters"
		m.filterInput.SetValue(m.clusterFilter)
		m.filterInput.Focus()
	}
	return m, nil
}

func (m Model) updateList(action keyAction) (Model, tea.Cmd) {
	switch action {
	case keyActionMoveUp:
		m.cursor = clampCursor(m.cursor-1, len(m.page.Items))
	case keyActionMoveDown:
		m.cursor = clampCursor(m.cursor+1, len(m.page.Items))
	case keyActionConfirm:
		if len(m.page.Items) == 0 {
			return m, nil
		}
		m.listCursor = m.cursor
		m.listAfter = m.after
		m.listFilter = m.serviceFilter
		m.listPrev = append([]string(nil), m.prevCursors...)
		m.detailID = m.page.Items[m.cursor].ID
		return m, m.beginDetail(m.detailID)
	case keyActionRefresh:
		return m, m.beginList(m.after)
	case keyActionFilter:
		m.mode = modeFilter
		m.filteringCluster = false
		m.filterInput.Placeholder = "filter services"
		m.filterInput.SetValue(m.serviceFilter)
		m.filterInput.Focus()
	case keyActionNextPage:
		if !m.page.HasNext || m.page.EndCursor == "" {
			return m, nil
		}
		if m.after != nil {
			m.prevCursors = append(m.prevCursors, *m.after)
		} else {
			m.prevCursors = append(m.prevCursors, "")
		}
		cursor := m.page.EndCursor
		m.after = &cursor
		m.cursor = 0
		return m, m.beginList(m.after)
	case keyActionPrevPage:
		if len(m.prevCursors) == 0 {
			return m, nil
		}
		previous := m.prevCursors[len(m.prevCursors)-1]
		m.prevCursors = m.prevCursors[:len(m.prevCursors)-1]
		if previous == "" {
			m.after = nil
		} else {
			m.after = &previous
		}
		m.cursor = 0
		return m, m.beginList(m.after)
	}
	return m, nil
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

func clusterLabel(cluster servicesbridge.Cluster) string {
	if cluster.Handle != "" {
		return "@" + cluster.Handle
	}
	if cluster.Name != "" {
		return cluster.Name
	}
	return cluster.ID
}
