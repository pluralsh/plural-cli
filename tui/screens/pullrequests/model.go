// Package pullrequests implements the Console PR automations browser and actions.
package pullrequests

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/pluralsh/plural-cli/pkg/bridge"
	pullrequestsbridge "github.com/pluralsh/plural-cli/pkg/bridge/pullrequests"
	"github.com/pluralsh/plural-cli/tui/navigation"
	"github.com/pluralsh/plural-cli/tui/theme"
)

type mode uint8

const (
	modeList mode = iota
	modeDetail
	modeFilter
	modeCreateForm
	modeTriggerForm
	modeCLITip
	modeReview
	modeOperating
	modeResult
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
	keyActionConnectConsole
	keyActionNextPage
	keyActionPrevPage
)

var keyActionKeystrokes = map[keyAction][]string{
	keyActionBack:           {"esc"},
	keyActionMoveUp:         {"up", "k"},
	keyActionMoveDown:       {"down", "j"},
	keyActionConfirm:        {"enter"},
	keyActionRefresh:        {"r"},
	keyActionFilter:         {"/"},
	keyActionConnectConsole: {"c"},
	keyActionNextPage:       {"n", "right", "]"},
	keyActionPrevPage:       {"p", "left", "["},
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

type formField struct {
	label string
	key   string
}

type initMsg struct{}
type listedMsg struct {
	page    pullrequestsbridge.Page
	err     error
	request uint64
}
type detailMsg struct {
	detail  pullrequestsbridge.Detail
	err     error
	request uint64
}
type opDoneMsg struct {
	err     error
	created pullrequestsbridge.CreatedPR
	request uint64
	kind    actionKind
}

// Model owns Pull-requests-screen interaction state.
type Model struct {
	ctx       context.Context
	loader    pullrequestsbridge.Loader
	theme     theme.Theme
	mode      mode
	loading   bool
	err       error
	needsAuth bool
	request   uint64

	page        pullrequestsbridge.Page
	cursor      int
	filter      string
	filterInput textinput.Model
	after       *string
	prevCursors []string

	detail       pullrequestsbridge.Detail
	detailID     string
	listCursor   int
	listAfter    *string
	listFilter   string
	listPrev     []string
	actionCursor int

	formInput  textinput.Model
	formFields []formField
	formIndex  int
	formValues map[string]string

	pending pendingOp
	opLog   []string
	result  string
	cliKind actionKind
}

func New(ctx context.Context, loader pullrequestsbridge.Loader, t theme.Theme) Model {
	input := textinput.New()
	input.Prompt = "› "
	input.Placeholder = "filter PR automations"
	input.CharLimit = 256
	styles := textinput.DefaultDarkStyles()
	styles.Focused.Text = t.Body
	styles.Focused.Prompt = t.Title
	styles.Focused.Placeholder = t.Muted
	styles.Blurred = styles.Focused
	input.SetStyles(styles)
	form := input
	form.Placeholder = ""
	return Model{ctx: ctx, loader: loader, theme: t, loading: loader != nil, filterInput: input, formInput: form, mode: modeList}
}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg { return initMsg{} }
}

func (m *Model) beginList(after *string) tea.Cmd {
	m.loading = true
	m.request++
	request := m.request
	query := m.filter
	loader := m.loader
	ctx := m.ctx
	return func() tea.Msg {
		page, err := loader.List(ctx, after, query)
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

func (m *Model) beginPending() tea.Cmd {
	m.mode = modeOperating
	m.loading = true
	m.err = nil
	m.opLog = []string{"starting…"}
	m.request++
	request := m.request
	loader := m.loader
	ctx := m.ctx
	op := m.pending
	return func() tea.Msg {
		switch op.kind {
		case actionCreate:
			created, err := loader.CreatePR(ctx, *op.create)
			return opDoneMsg{err: err, created: created, request: request, kind: op.kind}
		case actionTrigger:
			created, err := loader.TriggerPR(ctx, *op.trigger)
			return opDoneMsg{err: err, created: created, request: request, kind: op.kind}
		default:
			return opDoneMsg{err: fmt.Errorf("unsupported action"), request: request, kind: op.kind}
		}
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case initMsg:
		m.mode = modeList
		m.page = pullrequestsbridge.Page{}
		m.cursor = 0
		m.after = nil
		m.prevCursors = nil
		m.err = nil
		m.needsAuth = false
		if m.loader == nil {
			m.loading = false
			return m, nil
		}
		return m, m.beginList(nil)
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
			m.actionCursor = 0
		}
		return m, nil
	case opDoneMsg:
		if msg.request != m.request {
			return m, nil
		}
		m.loading = false
		m.mode = modeResult
		if msg.err != nil {
			m.result = "failed"
			m.err = msg.err
			m.opLog = []string{msg.err.Error()}
			return m, nil
		}
		m.result = "ok"
		m.err = nil
		m.opLog = []string{
			"PR ID    " + msg.created.ID,
			"URL      " + loCoalesce(msg.created.URL, "—"),
			"Title    " + loCoalesce(msg.created.Title, "—"),
			"Status   " + loCoalesce(msg.created.Status, "—"),
			"Ref      " + loCoalesce(msg.created.Ref, "—"),
		}
		return m, nil
	case tea.KeyPressMsg:
		return m.updateKey(msg)
	}
	switch m.mode {
	case modeFilter:
		var cmd tea.Cmd
		m.filterInput, cmd = m.filterInput.Update(msg)
		return m, cmd
	case modeCreateForm, modeTriggerForm, modeCLITip:
		var cmd tea.Cmd
		m.formInput, cmd = m.formInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) updateKey(key tea.KeyPressMsg) (Model, tea.Cmd) {
	action := actionForKeystroke(key.Keystroke())
	text := key.Text
	if text == "" && key.Code > 0 && key.Code < 128 {
		text = string(rune(key.Code))
	}

	switch m.mode {
	case modeFilter:
		return m.updateFilter(action, key)
	case modeCreateForm, modeTriggerForm:
		return m.updateForm(action, key)
	case modeCLITip:
		return m.updateCLITip(action, key)
	case modeReview:
		return m.updateReview(action)
	case modeResult:
		return m.updateResult(action)
	case modeOperating:
		return m, nil
	case modeDetail:
		return m.updateDetail(action, text)
	default:
		return m.updateList(action)
	}
}

func (m Model) updateFilter(action keyAction, key tea.KeyPressMsg) (Model, tea.Cmd) {
	switch action {
	case keyActionBack:
		m.mode = modeList
		m.filterInput.Blur()
		return m, nil
	case keyActionConfirm:
		m.filter = strings.TrimSpace(m.filterInput.Value())
		m.filterInput.Blur()
		m.mode = modeList
		m.cursor = 0
		m.after = nil
		m.prevCursors = nil
		return m, m.beginList(nil)
	}
	var cmd tea.Cmd
	m.filterInput, cmd = m.filterInput.Update(key)
	return m, cmd
}

func (m Model) updateList(action keyAction) (Model, tea.Cmd) {
	if action == keyActionBack {
		return m, navigation.Navigate(navigation.Deployments)
	}
	if m.loading {
		return m, nil
	}
	if m.needsAuth && action == keyActionConnectConsole {
		return m, navigation.Navigate(navigation.Access)
	}
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
		m.listFilter = m.filter
		m.listPrev = append([]string(nil), m.prevCursors...)
		m.detailID = m.page.Items[m.cursor].ID
		return m, m.beginDetail(m.detailID)
	case keyActionRefresh:
		return m, m.beginList(m.after)
	case keyActionFilter:
		m.mode = modeFilter
		m.filterInput.SetValue(m.filter)
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

func (m Model) updateDetail(action keyAction, text string) (Model, tea.Cmd) {
	if action == keyActionBack {
		m.mode = modeList
		m.err = nil
		m.cursor = m.listCursor
		m.after = m.listAfter
		m.filter = m.listFilter
		m.prevCursors = append([]string(nil), m.listPrev...)
		return m, nil
	}
	if m.loading {
		return m, nil
	}
	if action == keyActionRefresh && m.detailID != "" {
		return m, m.beginDetail(m.detailID)
	}
	actions := detailActions()
	for i, a := range actions {
		if text == a.shortcut {
			m.actionCursor = i
			return m.openAction(a)
		}
	}
	switch action {
	case keyActionMoveUp:
		m.actionCursor = clampCursor(m.actionCursor-1, len(actions))
	case keyActionMoveDown:
		m.actionCursor = clampCursor(m.actionCursor+1, len(actions))
	case keyActionConfirm:
		return m.openAction(actions[m.actionCursor])
	}
	return m, nil
}

func (m Model) openAction(a detailAction) (Model, tea.Cmd) {
	switch a.kind {
	case actionCreate:
		return m.beginCreateForm(), nil
	case actionTrigger:
		return m.beginTriggerForm(), nil
	case actionTemplate, actionTest, actionContracts:
		m.cliKind = a.kind
		m.mode = modeCLITip
		m.formInput.SetValue("./automation.yaml")
		m.formInput.Placeholder = "path to file"
		m.formInput.Focus()
		m.err = nil
		return m, nil
	}
	return m, nil
}

func (m Model) beginCreateForm() Model {
	m.mode = modeCreateForm
	m.formFields = []formField{
		{label: "Branch", key: "branch"},
		{label: "Context JSON", key: "context"},
	}
	m.formIndex = 0
	m.formValues = map[string]string{}
	m.formInput.SetValue("")
	m.formInput.Placeholder = "optional branch"
	m.formInput.Focus()
	m.err = nil
	return m
}

func (m Model) beginTriggerForm() Model {
	m.mode = modeTriggerForm
	m.formFields = []formField{
		{label: "Branch", key: "branch"},
		{label: "Configuration", key: "configuration"},
	}
	m.formIndex = 0
	m.formValues = map[string]string{}
	m.formInput.SetValue("")
	m.formInput.Placeholder = "optional branch"
	m.formInput.Focus()
	m.err = nil
	return m
}

func (m Model) updateForm(action keyAction, key tea.KeyPressMsg) (Model, tea.Cmd) {
	switch action {
	case keyActionBack:
		m.formInput.Blur()
		m.mode = modeDetail
		m.err = nil
		return m, nil
	case keyActionConfirm:
		m.saveFormField()
		if m.formIndex < len(m.formFields)-1 {
			m.formIndex++
			m.loadFormField()
			return m, nil
		}
		return m.submitForm()
	case keyActionMoveDown:
		m.saveFormField()
		if m.formIndex < len(m.formFields)-1 {
			m.formIndex++
			m.loadFormField()
		}
		return m, nil
	case keyActionMoveUp:
		m.saveFormField()
		if m.formIndex > 0 {
			m.formIndex--
			m.loadFormField()
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.formInput, cmd = m.formInput.Update(key)
	return m, cmd
}

func (m *Model) saveFormField() {
	if m.formValues == nil {
		m.formValues = map[string]string{}
	}
	if m.formIndex >= 0 && m.formIndex < len(m.formFields) {
		m.formValues[m.formFields[m.formIndex].key] = strings.TrimSpace(m.formInput.Value())
	}
}

func (m *Model) loadFormField() {
	if m.formIndex < 0 || m.formIndex >= len(m.formFields) {
		return
	}
	field := m.formFields[m.formIndex]
	m.formInput.SetValue(m.formValues[field.key])
	switch field.key {
	case "context":
		m.formInput.Placeholder = `optional JSON, e.g. {"cluster":"demo"}`
	case "configuration":
		m.formInput.Placeholder = "key=value pairs, e.g. cluster=demo region=us-east-1"
	default:
		m.formInput.Placeholder = field.label
	}
	m.formInput.Focus()
}

func (m Model) submitForm() (Model, tea.Cmd) {
	m.formInput.Blur()
	switch m.mode {
	case modeCreateForm:
		input := pullrequestsbridge.CreatePRInput{
			AutomationID: m.detail.ID,
			Branch:       m.formValues["branch"],
			Context:      m.formValues["context"],
		}
		m.pending = m.createPlan(input)
		m.mode = modeReview
		m.err = nil
		return m, nil
	case modeTriggerForm:
		cfg, err := parseConfiguration(m.formValues["configuration"])
		if err != nil {
			m.err = err
			m.formIndex = 1
			m.loadFormField()
			return m, nil
		}
		input := pullrequestsbridge.TriggerPRInput{
			AutomationID:  m.detail.ID,
			Name:          m.detail.Name,
			Branch:        m.formValues["branch"],
			Configuration: cfg,
		}
		m.pending = m.triggerPlan(input)
		m.mode = modeReview
		m.err = nil
		return m, nil
	}
	return m, nil
}

func (m Model) updateCLITip(action keyAction, key tea.KeyPressMsg) (Model, tea.Cmd) {
	switch action {
	case keyActionBack:
		m.formInput.Blur()
		m.mode = modeDetail
		return m, nil
	case keyActionConfirm:
		m.formInput.Blur()
		m.pending = cliTipPlan(m.cliKind, m.detail.Name, m.formInput.Value())
		m.mode = modeResult
		m.result = "ok"
		m.opLog = append([]string{}, m.pending.lines...)
		m.opLog = append(m.opLog, "", "Equivalent CLI", "  "+m.pending.cli)
		return m, nil
	}
	var cmd tea.Cmd
	m.formInput, cmd = m.formInput.Update(key)
	return m, cmd
}

func (m Model) updateReview(action keyAction) (Model, tea.Cmd) {
	switch action {
	case keyActionBack:
		m.mode = modeDetail
		return m, nil
	case keyActionConfirm:
		return m, m.beginPending()
	}
	return m, nil
}

func (m Model) updateResult(action keyAction) (Model, tea.Cmd) {
	switch action {
	case keyActionBack:
		m.mode = modeDetail
		return m, nil
	case keyActionConfirm:
		if m.result == "failed" {
			m.mode = modeReview
			return m, nil
		}
		m.mode = modeDetail
		return m, nil
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
