// Package services implements Console service browsing and contextual actions.
package services

import (
	"context"
	"fmt"
	"path/filepath"
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
	modeReview
	modeOperating
	modeResult
	modeDeleteConfirm
	modeTarball
	modeCreate
	modeEdit
	modeClone
	modeCloneCluster
	modeWorkbench
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
	keyActionCreate
	keyActionBackground
)

var keyActionKeystrokes = map[keyAction][]string{
	keyActionBack:           {"esc"},
	keyActionMoveUp:         {"up"},
	keyActionMoveDown:       {"down"},
	keyActionConfirm:        {"enter"},
	keyActionRefresh:        {"r"},
	keyActionFilter:         {"/"},
	keyActionNextPage:       {"right", "]"},
	keyActionPrevPage:       {"left", "p", "["},
	keyActionConnectConsole: {"c"},
	keyActionCreate:         {"n"},
	keyActionBackground:     {"b"},
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
type opDoneMsg struct {
	err     error
	detail  servicesbridge.Detail
	path    string
	request uint64
	kind    actionKind
}

// Model owns Services-screen interaction state.
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

	detail       servicesbridge.Detail
	detailID     string
	listCursor   int
	listAfter    *string
	listFilter   string
	listPrev     []string
	actionCursor int

	pending pendingOp
	opLog   []string
	result  string

	formInput   textinput.Model
	formFields  []formField
	formIndex   int
	formValues  map[string]string
	formDryRun  bool
	wbTemplate  bool
	confirmName string

	pickingCloneDest bool
	cloneDest        servicesbridge.Cluster
}

type formField struct {
	label string
	key   string
}

func New(ctx context.Context, loader servicesbridge.Loader, t theme.Theme) Model {
	input := textinput.New()
	input.Prompt = "› "
	input.Placeholder = "filter"
	input.CharLimit = 256
	styles := textinput.DefaultDarkStyles()
	styles.Focused.Text = t.Body
	styles.Focused.Prompt = t.Title
	styles.Focused.Placeholder = t.Muted
	styles.Blurred = styles.Focused
	input.SetStyles(styles)
	form := input
	form.Placeholder = ""
	return Model{
		ctx: ctx, loader: loader, theme: t, loading: loader != nil,
		filterInput: input, formInput: form, mode: modeClusters, wbTemplate: true,
	}
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
	detailID := m.detailID
	return func() tea.Msg {
		switch op.kind {
		case actionKick:
			detail, err := loader.Kick(ctx, op.kickID)
			return opDoneMsg{err: err, detail: detail, request: request, kind: op.kind}
		case actionDelete:
			err := loader.Delete(ctx, op.deleteID)
			return opDoneMsg{err: err, request: request, kind: op.kind}
		case actionTarball:
			path, err := loader.DownloadTarball(ctx, detailID, op.tarball)
			return opDoneMsg{err: err, path: path, request: request, kind: op.kind}
		case actionEdit:
			detail, err := loader.Update(ctx, *op.update)
			return opDoneMsg{err: err, detail: detail, request: request, kind: op.kind}
		case actionClone:
			detail, err := loader.Clone(ctx, *op.clone)
			return opDoneMsg{err: err, detail: detail, request: request, kind: op.kind}
		case actionCreate:
			detail, err := loader.Create(ctx, *op.create)
			return opDoneMsg{err: err, detail: detail, request: request, kind: op.kind}
		default:
			return opDoneMsg{err: fmt.Errorf("unsupported action"), request: request, kind: op.kind}
		}
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
		m.pickingCloneDest = false
		m.cloneDest = servicesbridge.Cluster{}
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
			if m.pickingCloneDest {
				m.mode = modeCloneCluster
			} else {
				m.mode = modeClusters
			}
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
			m.actionCursor = 0
			m.mode = modeDetail
		}
		return m, nil
	case opDoneMsg:
		if msg.request != m.request {
			return m, nil
		}
		m.loading = false
		m.err = msg.err
		m.mode = modeResult
		if msg.err != nil {
			m.result = "failed"
			m.opLog = append(m.opLog, "✗ "+msg.err.Error())
			return m, nil
		}
		m.result = "ok"
		switch msg.kind {
		case actionKick, actionEdit, actionCreate:
			m.detail = msg.detail
			m.detailID = msg.detail.ID
			m.opLog = append(m.opLog, "✓ completed")
		case actionClone:
			m.detail = msg.detail
			m.detailID = msg.detail.ID
			m.opLog = append(m.opLog, "✓ cloned "+msg.detail.Name)
		case actionDelete:
			m.opLog = append(m.opLog, "✓ deleted")
		case actionTarball:
			m.opLog = append(m.opLog, "✓ wrote "+msg.path)
			m.result = msg.path
		default:
			m.opLog = append(m.opLog, "✓ completed")
		}
		return m, nil
	case tea.KeyPressMsg:
		return m.updateKey(msg)
	}
	switch m.mode {
	case modeFilter, modeDeleteConfirm, modeTarball, modeCreate, modeEdit, modeClone, modeWorkbench:
		var cmd tea.Cmd
		m.formInput, cmd = m.formInput.Update(msg)
		if m.mode == modeFilter {
			m.filterInput = m.formInput
		}
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
	case modeReview:
		return m.updateReview(action)
	case modeOperating:
		return m, nil
	case modeResult:
		return m.updateResult(action)
	case modeDeleteConfirm:
		return m.updateDeleteConfirm(action, key)
	case modeTarball:
		return m.updateTarball(action, key)
	case modeCreate, modeEdit, modeClone:
		return m.updateForm(action, key)
	case modeCloneCluster:
		return m.updateCloneCluster(action)
	case modeWorkbench:
		return m.updateWorkbench(action, key, text)
	case modeDetail:
		return m.updateDetail(action, text)
	}

	if action == keyActionBack {
		switch m.mode {
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
	if m.needsAuth && (action == keyActionConnectConsole || text == "c") {
		return m, navigation.Navigate(navigation.Access)
	}
	if m.mode == modeClusters {
		return m.updateClusters(action)
	}
	return m.updateList(action, text)
}

func (m Model) updateFilter(action keyAction, key tea.KeyPressMsg) (Model, tea.Cmd) {
	switch action {
	case keyActionBack:
		m.filterInput.Blur()
		if m.pickingCloneDest {
			m.mode = modeCloneCluster
		} else if m.filteringCluster {
			m.mode = modeClusters
		} else {
			m.mode = modeList
		}
		return m, nil
	case keyActionConfirm:
		value := strings.TrimSpace(m.filterInput.Value())
		m.filterInput.Blur()
		if m.pickingCloneDest || m.filteringCluster {
			m.clusterFilter = value
			m.clusterCursor = 0
			if m.pickingCloneDest {
				m.mode = modeCloneCluster
			} else {
				m.mode = modeClusters
			}
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

func (m Model) updateDetail(action keyAction, text string) (Model, tea.Cmd) {
	if action == keyActionBack {
		m.mode = modeList
		m.err = nil
		m.cursor = m.listCursor
		m.after = m.listAfter
		m.serviceFilter = m.listFilter
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
	switch action {
	case keyActionMoveUp:
		m.actionCursor = clampCursor(m.actionCursor-1, len(actions))
		return m, nil
	case keyActionMoveDown:
		m.actionCursor = clampCursor(m.actionCursor+1, len(actions))
		return m, nil
	case keyActionConfirm:
		return m.openAction(actions[m.actionCursor])
	}
	for i, a := range actions {
		if text == a.shortcut {
			m.actionCursor = i
			return m.openAction(a)
		}
	}
	return m, nil
}

func (m Model) openAction(a detailAction) (Model, tea.Cmd) {
	switch a.kind {
	case actionKick:
		m.pending = m.kickPlan()
		m.mode = modeReview
		return m, nil
	case actionDelete:
		m.mode = modeDeleteConfirm
		m.confirmName = ""
		m.formInput.SetValue("")
		m.formInput.Placeholder = m.detail.Name
		m.formInput.Focus()
		return m, nil
	case actionTarball:
		m.mode = modeTarball
		dir := filepath.Join(".", m.detail.Name+"-tarball")
		m.formInput.SetValue(dir)
		m.formInput.Placeholder = dir
		m.formInput.Focus()
		return m, nil
	case actionEdit:
		return m.beginEditForm(), nil
	case actionClone:
		return m.beginClone()
	case actionWorkbench:
		m.mode = modeWorkbench
		m.wbTemplate = true
		m.formInput.SetValue("")
		m.formInput.Placeholder = "./values.yaml.liquid"
		m.formInput.Focus()
		return m, nil
	}
	return m, nil
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
		if m.pending.kind == actionDelete && m.result == "ok" {
			m.mode = modeList
			return m, m.beginList(m.after)
		}
		if m.pending.create != nil && m.result == "ok" {
			m.mode = modeDetail
			return m, nil
		}
		m.mode = modeDetail
		return m, nil
	case keyActionConfirm:
		if m.result == "ok" {
			if m.pending.kind == actionDelete {
				m.mode = modeList
				return m, m.beginList(m.after)
			}
			m.mode = modeDetail
			if m.detailID != "" {
				return m, m.beginDetail(m.detailID)
			}
			return m, nil
		}
		m.mode = modeReview
		return m, nil
	case keyActionRefresh:
		if m.result != "ok" {
			m.mode = modeReview
			return m, nil
		}
	}
	return m, nil
}

func (m Model) updateDeleteConfirm(action keyAction, key tea.KeyPressMsg) (Model, tea.Cmd) {
	switch action {
	case keyActionBack:
		m.formInput.Blur()
		m.mode = modeDetail
		return m, nil
	case keyActionConfirm:
		if strings.TrimSpace(m.formInput.Value()) != m.detail.Name {
			m.err = fmt.Errorf("name does not match %q", m.detail.Name)
			return m, nil
		}
		m.formInput.Blur()
		m.err = nil
		m.pending = m.deletePlan()
		m.mode = modeReview
		return m, nil
	}
	var cmd tea.Cmd
	m.formInput, cmd = m.formInput.Update(key)
	return m, cmd
}

func (m Model) updateTarball(action keyAction, key tea.KeyPressMsg) (Model, tea.Cmd) {
	switch action {
	case keyActionBack:
		m.formInput.Blur()
		m.mode = modeDetail
		return m, nil
	case keyActionConfirm:
		dir := strings.TrimSpace(m.formInput.Value())
		if dir == "" {
			dir = filepath.Join(".", m.detail.Name+"-tarball")
		}
		m.formInput.Blur()
		m.pending = pendingOp{
			kind:    actionTarball,
			title:   "Download tarball · " + m.detail.Name,
			cli:     "plural cd services tarball " + m.detail.ID + " --dir " + dir,
			tarball: dir,
			lines: []string{
				"Action      Download tarball",
				"Service     " + m.detail.Name,
				"Directory   " + dir,
			},
		}
		m.mode = modeReview
		return m, nil
	}
	var cmd tea.Cmd
	m.formInput, cmd = m.formInput.Update(key)
	return m, cmd
}

func (m Model) beginCreateForm() Model {
	m.mode = modeCreate
	m.formFields = []formField{
		{label: "Name", key: "name"},
		{label: "Namespace", key: "namespace"},
		{label: "Repo ID", key: "repo"},
		{label: "Git ref", key: "ref"},
		{label: "Git folder", key: "folder"},
		{label: "Kustomize", key: "kustomize"},
		{label: "Version", key: "version"},
	}
	m.formIndex = 0
	m.formDryRun = false
	m.formValues = map[string]string{"namespace": "default", "version": "0.0.1", "ref": "main"}
	m.formInput.SetValue("")
	m.formInput.Placeholder = "service name"
	m.formInput.Focus()
	m.err = nil
	return m
}

func (m Model) beginEditForm() Model {
	m.mode = modeEdit
	m.formFields = []formField{
		{label: "Git ref", key: "ref"},
		{label: "Git folder", key: "folder"},
		{label: "Kustomize", key: "kustomize"},
		{label: "Version", key: "version"},
	}
	m.formIndex = 0
	m.formDryRun = false
	m.formValues = map[string]string{
		"ref":     m.detail.GitRef,
		"folder":  m.detail.GitFolder,
		"version": "0.0.1",
	}
	m.formInput.SetValue(m.formValues["ref"])
	m.formInput.Placeholder = "git ref"
	m.formInput.Focus()
	m.err = nil
	return m
}

func (m Model) beginClone() (Model, tea.Cmd) {
	m.pickingCloneDest = true
	m.cloneDest = servicesbridge.Cluster{}
	m.clusterFilter = ""
	m.clusterCursor = 0
	m.err = nil
	m.mode = modeCloneCluster
	m.loading = true
	return m, m.beginClusters()
}

func (m Model) updateCloneCluster(action keyAction) (Model, tea.Cmd) {
	switch action {
	case keyActionBack:
		m.pickingCloneDest = false
		m.clusterFilter = ""
		m.mode = modeDetail
		m.err = nil
		return m, nil
	case keyActionMoveUp:
		m.clusterCursor = clampCursor(m.clusterCursor-1, len(m.clusters))
	case keyActionMoveDown:
		m.clusterCursor = clampCursor(m.clusterCursor+1, len(m.clusters))
	case keyActionConfirm:
		if len(m.clusters) == 0 {
			return m, nil
		}
		m.cloneDest = m.clusters[m.clusterCursor]
		m.pickingCloneDest = false
		m.clusterFilter = ""
		return m.beginCloneForm(), nil
	case keyActionRefresh:
		return m, m.beginClusters()
	case keyActionFilter:
		m.mode = modeFilter
		m.filteringCluster = true
		m.filterInput.Placeholder = "filter destination clusters"
		m.filterInput.SetValue(m.clusterFilter)
		m.filterInput.Focus()
		m.formInput = m.filterInput
	}
	return m, nil
}

func (m Model) beginCloneForm() Model {
	m.mode = modeClone
	m.formFields = []formField{
		{label: "Name", key: "name"},
		{label: "Namespace", key: "namespace"},
	}
	m.formIndex = 0
	m.formValues = map[string]string{
		"name":      m.detail.Name + "-clone",
		"namespace": loCoalesce(m.detail.Namespace, "default"),
	}
	m.formInput.SetValue(m.formValues["name"])
	m.formInput.Placeholder = "cloned service name"
	m.formInput.Focus()
	m.err = nil
	return m
}

func (m Model) updateForm(action keyAction, key tea.KeyPressMsg) (Model, tea.Cmd) {
	switch action {
	case keyActionBack:
		m.formInput.Blur()
		switch m.mode {
		case modeCreate:
			m.mode = modeList
		case modeClone:
			m.pickingCloneDest = true
			m.mode = modeCloneCluster
		default:
			m.mode = modeDetail
		}
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
	if key.Text == "d" && key.Mod == tea.ModCtrl {
		// ignore
	}
	if key.Keystroke() == "ctrl+d" {
		m.formDryRun = !m.formDryRun
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
	m.formInput.Placeholder = field.label
	m.formInput.Focus()
}

func (m Model) submitForm() (Model, tea.Cmd) {
	m.formInput.Blur()
	switch m.mode {
	case modeCreate:
		input := servicesbridge.CreateInput{
			ClusterID: m.cluster.ID,
			Name:      m.formValues["name"],
			Namespace: m.formValues["namespace"],
			RepoID:    m.formValues["repo"],
			GitRef:    m.formValues["ref"],
			GitFolder: m.formValues["folder"],
			Kustomize: m.formValues["kustomize"],
			Version:   m.formValues["version"],
			DryRun:    m.formDryRun,
		}
		m.pending = pendingOp{
			kind:   actionCreate,
			title:  "Create service · " + input.Name,
			cli:    fmt.Sprintf("plural cd services create %s --name %s --repo-id %s --git-ref %s --git-folder %s", clusterLabel(m.cluster), input.Name, input.RepoID, input.GitRef, input.GitFolder),
			create: &input,
			lines: []string{
				"Action     Create service",
				"Cluster    " + clusterLabel(m.cluster),
				"Name       " + input.Name,
				"Namespace  " + input.Namespace,
				"Repo       " + input.RepoID,
				"Git        " + input.GitRef + " / " + input.GitFolder,
				fmt.Sprintf("Dry-run    %v (Console attribute)", input.DryRun),
			},
		}
		m.mode = modeReview
		return m, nil
	case modeEdit:
		dry := m.formDryRun
		input := servicesbridge.UpdateInput{
			ID:        m.detail.ID,
			GitRef:    m.formValues["ref"],
			GitFolder: m.formValues["folder"],
			Kustomize: m.formValues["kustomize"],
			Version:   m.formValues["version"],
			DryRun:    &dry,
		}
		m.pending = pendingOp{
			kind:   actionEdit,
			title:  "Update · " + m.detail.Name,
			cli:    "plural cd services update " + m.detail.ID,
			update: &input,
			lines: []string{
				"Action     Update service",
				"Service    " + m.detail.Name,
				"Git        " + input.GitRef + " / " + input.GitFolder,
				"Version    " + input.Version,
				fmt.Sprintf("Dry-run    %v", dry),
			},
		}
		m.mode = modeReview
		return m, nil
	case modeClone:
		input := servicesbridge.CloneInput{
			SourceID:      m.detail.ID,
			DestClusterID: m.cloneDest.ID,
			Name:          m.formValues["name"],
			Namespace:     m.formValues["namespace"],
		}
		dest := clusterLabel(m.cloneDest)
		m.pending = pendingOp{
			kind:  actionClone,
			title: "Clone · " + m.detail.Name,
			cli:   fmt.Sprintf("plural cd services clone %s %s --name %s --namespace %s", dest, m.detail.ID, input.Name, input.Namespace),
			clone: &input,
			lines: []string{
				"Action     Clone service",
				"Source     " + m.detail.Name + " · " + clusterLabel(m.cluster),
				"Dest       " + dest,
				"Name       " + input.Name,
				"Namespace  " + input.Namespace,
			},
		}
		m.mode = modeReview
		return m, nil
	}
	return m, nil
}

func (m Model) updateWorkbench(action keyAction, key tea.KeyPressMsg, text string) (Model, tea.Cmd) {
	switch action {
	case keyActionBack:
		m.formInput.Blur()
		m.mode = modeDetail
		return m, nil
	case keyActionConfirm:
		m.mode = modeResult
		m.result = "ok"
		m.pending = pendingOp{kind: actionWorkbench, title: "Workbench · " + m.detail.Name}
		file := strings.TrimSpace(m.formInput.Value())
		mode := "template"
		if !m.wbTemplate {
			mode = "lua"
		}
		m.opLog = []string{
			"Dry-run workbench — rendering is CLI-backed for now.",
			"",
			fmt.Sprintf("  plural cd services %s --file %q --service %s/%s", mode, file, clusterLabel(servicesbridge.Cluster{Handle: m.detail.ClusterHandle, Name: m.detail.ClusterName}), m.detail.Name),
		}
		m.formInput.Blur()
		return m, nil
	}
	if text == "tab" || key.Keystroke() == "tab" {
		m.wbTemplate = !m.wbTemplate
		return m, nil
	}
	var cmd tea.Cmd
	m.formInput, cmd = m.formInput.Update(key)
	return m, cmd
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
		m.formInput = m.filterInput
	}
	return m, nil
}

func (m Model) updateList(action keyAction, text string) (Model, tea.Cmd) {
	if action == keyActionCreate || text == "n" {
		if m.cluster.ID == "" {
			return m, nil
		}
		return m.beginCreateForm(), nil
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
		m.formInput = m.filterInput
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
