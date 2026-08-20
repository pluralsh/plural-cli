// Package up implements the plural-up setup wizard.
package up

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	upbridge "github.com/pluralsh/plural-cli/pkg/bridge/up"
	"github.com/pluralsh/plural-cli/pkg/console"
	"github.com/pluralsh/plural-cli/pkg/provider"
	pluralspinner "github.com/pluralsh/plural-cli/tui/components/spinner"
	"github.com/pluralsh/plural-cli/tui/navigation"
	"github.com/pluralsh/plural-cli/tui/theme"
)

type mode uint8

const (
	modeSelectFlow mode = iota
	modeIgnorePreflights
	modeLoadInstances  // GetConsoleInstances
	modeSelectInstance // choseCluster survey
	modeConsoleLogin   // HandleCdLogin Affirm or token
	modeSelectProvider
	modeProbing
	modeProviderForm
	modeRunPreflights  // provider.Preflights() after survey
	modeIgnoreContinue // failed check + --ignore-preflights → Enter to continue
	modeSetupGit       // CLI Affirm: setup git repo here? (Y/n)
	modeSelectSCM      // scm.Setup: github / gitlab / bitbucket
	modeAppDomain      // askAppDomain parity
	modeAffirmDeploy   // common.AffirmUp before deploy
	modeSelected       // Plan summary
	modeRunning        // Flush + Generate
	modeDone           // generate finished (or failed)
	modeCommitMsg      // optional commit before Deploy
	modeDeploying      // up.Context.Deploy
	modeComplete       // deploy finished
	modeCLITip
)

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

type yesNoOption struct {
	value bool
	title string
	blurb string
}

func ignorePreflightOptions() []yesNoOption {
	return []yesNoOption{
		{value: false, title: "Run checks", blurb: "stop if provider.Preflights() fail (default)"},
		{value: true, title: "Ignore", blurb: "warn and continue (--ignore-preflights)"},
	}
}

func setupGitOptions() []yesNoOption {
	return []yesNoOption{
		{value: true, title: "Yes", blurb: "create a git repo here (default)"},
		{value: false, title: "No", blurb: "cancel — clone a repo first"},
	}
}

func setupGitContinueOptions() []yesNoOption {
	return []yesNoOption{
		{value: true, title: "Yes", blurb: "continue init (domain → deploy Affirm)"},
		{value: false, title: "No", blurb: "cancel"},
	}
}

func affirmDeployOptions() []yesNoOption {
	return []yesNoOption{
		{value: true, title: "Yes", blurb: "ready to set up the management cluster (default)"},
		{value: false, title: "No", blurb: "cancel deploy — review generated terraform/helm first"},
	}
}

func consoleCredOptions(priorURL string) []yesNoOption {
	return []yesNoOption{
		{value: true, title: "Yes", blurb: "keep credentials for " + truncate(priorURL, 40)},
		{value: false, title: "No", blurb: "enter a new console access token"},
	}
}

type probeMsg struct {
	result upbridge.ProbeResult
	err    error
}

type optionsMsg struct {
	fieldKey string
	options  []string
	err      error
}

type domainMsg struct {
	options []string
	err     error
	text    bool // free-text domain entry (GCP / default)
	skip    bool // CLI ignored domain setup
}

type preflightMsg struct {
	err error
}

type instancesMsg struct {
	items []upbridge.ConsoleInstance
	err   error
}

type runDoneMsg struct {
	err             error
	steps           []string
	importClusterID string
}

type deployDoneMsg struct {
	err   error
	steps []string
}

// Model owns Up-wizard interaction state.
type Model struct {
	theme  theme.Theme
	prober upbridge.Prober
	ctx    context.Context

	mode      mode
	flows     []upbridge.Flow
	providers []upbridge.Provider
	cursor    int
	flow      upbridge.Flow
	provider  upbridge.Provider

	ignorePreflights bool
	ignoreAsked      bool

	credSummary string
	probeWarn   string // shown when continuing after ignored preflight failure
	inGitRepo   bool
	scm         upbridge.SCMProvider
	scms        []upbridge.SCMProvider
	appDomain   string
	domainOpts  []string
	domainNote  string // zone fetch ignored (CLI "ignoring domain setup...")
	spinner     spinner.Model

	instances        []upbridge.ConsoleInstance
	cloudInstance    upbridge.ConsoleInstance
	consoleTokenMode bool // true = token textinput; false = use-existing Affirm
	instanceLister   upbridge.InstanceLister
	priorConsole     func() (url, token string)
	saveConsole      func(url, token string) error
	runner          upbridge.Runner
	runSteps        []string
	runErr          error
	importClusterID string
	commitMsg       string
	deployErr       error

	formInput     textinput.Model
	formFields    []upbridge.FormField
	formIndex     int
	formValues    map[string]string
	optionCursor  int
	freeTextKeys  map[string]bool // Azure "Create new…" → free text
	err           error
	gitChecker    func() bool
	domainLoader  func() domainMsg // tests stub zone listing
	gitAffirmOpen bool             // Esc from domain returns here when Affirm was shown
}

// New creates the Up wizard starting at setup-flow selection.
func New(ctx context.Context, t theme.Theme) Model {
	return NewWithProber(ctx, t, upbridge.DefaultProber())
}

// NewWithProber is New with an injectable credential/region prober (tests).
func NewWithProber(ctx context.Context, t theme.Theme, prober upbridge.Prober) Model {
	input := textinput.New()
	input.Prompt = "› "
	input.CharLimit = 256
	styles := textinput.DefaultDarkStyles()
	styles.Focused.Text = t.Body
	styles.Focused.Prompt = t.Title
	styles.Focused.Placeholder = t.Muted
	styles.Blurred = styles.Focused
	input.SetStyles(styles)
	if prober == nil {
		prober = upbridge.DefaultProber()
	}
	return Model{
		theme:          t,
		prober:         prober,
		ctx:            ctx,
		mode:           modeSelectFlow,
		flows:          upbridge.Flows(),
		providers:      upbridge.CloudProviders(),
		scms:           upbridge.SCMProviders(),
		formInput:      input,
		spinner:        pluralspinner.New(t),
		gitChecker:     upbridge.InGitRepo,
		instanceLister: upbridge.DefaultInstanceLister(),
		runner:         upbridge.DefaultRunner(),
		priorConsole: func() (string, string) {
			c := upbridge.ReadPriorConsole()
			return c.Url, c.Token
		},
		saveConsole: upbridge.SaveConsoleConfig,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case probeMsg:
		return m.applyProbe(msg)
	case preflightMsg:
		return m.applyPreflight(msg)
	case optionsMsg:
		return m.applyOptions(msg)
	case domainMsg:
		return m.applyDomain(msg)
	case instancesMsg:
		return m.applyInstances(msg)
	case runDoneMsg:
		return m.applyRunDone(msg)
	case deployDoneMsg:
		return m.applyDeployDone(msg)
	case spinner.TickMsg:
		if m.mode != modeProbing && m.mode != modeAppDomain && m.mode != modeRunPreflights && m.mode != modeLoadInstances && m.mode != modeRunning && m.mode != modeDeploying {
			return m, nil
		}
		if m.mode == modeAppDomain && (len(m.domainOpts) > 0 || m.formInput.Focused()) {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		switch m.mode {
		case modeProviderForm:
			if !m.currentIsSelect() {
				var cmd tea.Cmd
				m.formInput, cmd = m.formInput.Update(msg)
				return m, cmd
			}
		case modeAppDomain:
			if !m.domainIsSelect() {
				var cmd tea.Cmd
				m.formInput, cmd = m.formInput.Update(msg)
				return m, cmd
			}
		case modeConsoleLogin:
			if m.consoleTokenMode {
				var cmd tea.Cmd
				m.formInput, cmd = m.formInput.Update(msg)
				return m, cmd
			}
		case modeCommitMsg:
			var cmd tea.Cmd
			m.formInput, cmd = m.formInput.Update(msg)
			return m, cmd
		}
		return m, nil
	}
	action := actionForKeystroke(key.Keystroke())

	switch m.mode {
	case modeSelected:
		return m.updateSelected(action)
	case modeRunning, modeDeploying:
		return m, nil // ignore keys while running
	case modeDone:
		return m.updateDone(action)
	case modeCommitMsg:
		return m.updateCommitMsg(action, key)
	case modeComplete:
		return m.updateComplete(action)
	case modeIgnoreContinue:
		return m.updateIgnoreContinue(action)
	case modeAffirmDeploy:
		return m.updateAffirmDeploy(action, key)
	case modeAppDomain:
		return m.updateAppDomain(action, key)
	case modeSelectSCM:
		return m.updateSelectSCM(action, key)
	case modeSetupGit:
		return m.updateSetupGit(action, key)
	case modeCLITip:
		return m.updateCLITip(action)
	case modeProviderForm:
		return m.updateProviderForm(action, key)
	case modeRunPreflights, modeProbing, modeLoadInstances:
		if action == keyActionBack {
			if m.mode == modeLoadInstances {
				return m.updateLoadInstances(action)
			}
			return m.updateProbing(action)
		}
		return m, nil
	case modeSelectInstance:
		return m.updateSelectInstance(action, key)
	case modeConsoleLogin:
		return m.updateConsoleLogin(action, key)
	case modeSelectProvider:
		return m.updateSelectProvider(action, key)
	case modeIgnorePreflights:
		return m.updateIgnorePreflights(action, key)
	default:
		return m.updateSelectFlow(action, key)
	}
}

func (m Model) updateSelectFlow(action keyAction, key tea.KeyPressMsg) (Model, tea.Cmd) {
	switch action {
	case keyActionBack:
		return m, navigation.Navigate(navigation.Welcome)
	case keyActionUp:
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case keyActionDown:
		if m.cursor < len(m.flows)-1 {
			m.cursor++
		}
		return m, nil
	case keyActionConfirm:
		return m.selectFlow(m.flows[m.cursor])
	}

	text := keyText(key)
	for i, f := range m.flows {
		if text == flowShortcut(f.ID) || text == string(rune('1'+i)) {
			m.cursor = i
			return m.selectFlow(f)
		}
	}
	return m, nil
}

func (m Model) updateIgnorePreflights(action keyAction, key tea.KeyPressMsg) (Model, tea.Cmd) {
	opts := ignorePreflightOptions()
	switch action {
	case keyActionBack:
		m.mode = modeSelectFlow
		m.flow = upbridge.Flow{}
		m.ignorePreflights = false
		m.ignoreAsked = false
		m.cursor = 0
		return m, nil
	case keyActionUp:
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case keyActionDown:
		if m.cursor < len(opts)-1 {
			m.cursor++
		}
		return m, nil
	case keyActionConfirm:
		return m.chooseIgnorePreflights(opts[m.cursor].value)
	}

	text := keyText(key)
	switch text {
	case "1", "r":
		m.cursor = 0
		return m.chooseIgnorePreflights(false)
	case "2", "i":
		m.cursor = 1
		return m.chooseIgnorePreflights(true)
	}
	return m, nil
}

func (m Model) updateSelectProvider(action keyAction, key tea.KeyPressMsg) (Model, tea.Cmd) {
	switch action {
	case keyActionBack:
		m.err = nil
		if m.flow.ID == "cloud" {
			m.resetConsoleInput()
			if len(m.instances) > 1 {
				m.mode = modeSelectInstance
				priorURL, _ := m.readPriorConsole()
				m.cursor = upbridge.DefaultInstanceIndex(m.instances, priorURL)
				return m, nil
			}
			m.mode = modeIgnorePreflights
			m.cursor = 0
			if m.ignorePreflights {
				m.cursor = 1
			}
			return m, nil
		}
		m.mode = modeIgnorePreflights
		m.cursor = 0
		if m.ignorePreflights {
			m.cursor = 1
		}
		return m, nil
	case keyActionUp:
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case keyActionDown:
		if m.cursor < len(m.providers)-1 {
			m.cursor++
		}
		return m, nil
	case keyActionConfirm:
		return m.beginProviderForm(m.providers[m.cursor])
	}

	text := keyText(key)
	for i, p := range m.providers {
		if text == providerShortcut(p.ID) || text == string(rune('1'+i)) {
			m.cursor = i
			return m.beginProviderForm(p)
		}
	}
	return m, nil
}

func (m Model) updateProbing(action keyAction) (Model, tea.Cmd) {
	if action == keyActionBack {
		m.mode = modeSelectProvider
		m.provider = upbridge.Provider{}
		m.err = nil
		m.cursor = 0
		return m, nil
	}
	return m, nil
}

func (m Model) updateIgnoreContinue(action keyAction) (Model, tea.Cmd) {
	switch action {
	case keyActionBack:
		if len(m.formFields) > 0 {
			m.mode = modeProviderForm
			m.applyLoadFormField()
			return m, nil
		}
		m.mode = modeSelectProvider
		m.cursor = 0
		return m, nil
	case keyActionConfirm:
		return m.continueAfterIgnoredFailure()
	}
	return m, nil
}

// continueAfterIgnoredFailure always opens the git Affirm after Enter so the
// warning gate never feels like a dead end. When already in a work tree, Yes
// skips scm.Setup (CLI Affirm is skipped entirely in that case — we still ask
// once here so the TUI has a clear next step).
func (m Model) continueAfterIgnoredFailure() (Model, tea.Cmd) {
	check := m.gitChecker
	if check == nil {
		check = upbridge.InGitRepo
	}
	m.inGitRepo = check()
	m.err = nil
	m.mode = modeSetupGit
	m.gitAffirmOpen = true
	m.cursor = 0 // Yes
	return m, nil
}

func (m Model) updateProviderForm(action keyAction, key tea.KeyPressMsg) (Model, tea.Cmd) {
	switch action {
	case keyActionBack:
		m.formInput.Blur()
		m.mode = modeSelectProvider
		m.provider = upbridge.Provider{}
		m.formFields = nil
		m.formValues = nil
		m.formIndex = 0
		m.credSummary = ""
		m.probeWarn = ""
		m.freeTextKeys = nil
		m.err = nil
		m.cursor = 0
		return m, nil
	case keyActionConfirm:
		if m.currentIsSelect() {
			return m.confirmSelectOption()
		}
		m.saveFormField()
		return m.advanceForm()
	case keyActionDown:
		if m.currentIsSelect() {
			opts := m.currentOptions()
			if m.optionCursor < len(opts)-1 {
				m.optionCursor++
			}
			return m, nil
		}
		m.saveFormField()
		if m.formIndex < len(m.formFields)-1 {
			m.formIndex++
			m.applyLoadFormField()
		}
		return m, nil
	case keyActionUp:
		if m.currentIsSelect() {
			if m.optionCursor > 0 {
				m.optionCursor--
			}
			return m, nil
		}
		m.saveFormField()
		if m.formIndex > 0 {
			m.formIndex--
			m.applyLoadFormField()
		}
		return m, nil
	}
	if !m.currentIsSelect() {
		var cmd tea.Cmd
		m.formInput, cmd = m.formInput.Update(key)
		return m, cmd
	}
	return m, nil
}

func (m Model) updateSelected(action keyAction) (Model, tea.Cmd) {
	switch action {
	case keyActionConfirm:
		return m.beginRun()
	case keyActionBack:
		if m.flow.Cloud {
			// Cloud skips Affirm — Esc returns to domain (or provider when dry-run).
			if !m.flow.DryRun {
				m.mode = modeAppDomain
				if m.domainIsSelect() {
					m.formInput.Blur()
				} else {
					m.formInput.SetValue(m.appDomain)
					m.formInput.Focus()
				}
				m.err = nil
				return m, nil
			}
		}
		if !m.flow.DryRun {
			m.mode = modeAffirmDeploy
			m.cursor = 0
			m.err = nil
			return m, nil
		}
		if len(m.formFields) > 0 {
			m.mode = modeProviderForm
			m.applyLoadFormField()
			return m, nil
		}
		if m.scm.ID != "" {
			m.mode = modeSelectSCM
			m.cursor = 0
			return m, nil
		}
		m.mode = modeSelectProvider
		m.cursor = 0
		return m, nil
	}
	return m, nil
}

func (m Model) updateDone(action keyAction) (Model, tea.Cmd) {
	switch action {
	case keyActionBack:
		m.mode = modeSelected
		m.runErr = nil
		m.runSteps = nil
		return m, nil
	case keyActionConfirm:
		if m.runErr != nil || m.flow.DryRun {
			return m, nil
		}
		return m.beginCommitMsg()
	}
	return m, nil
}

func (m Model) updateComplete(action keyAction) (Model, tea.Cmd) {
	if action == keyActionBack {
		m.mode = modeDone
		m.deployErr = nil
		return m, nil
	}
	return m, nil
}

func (m Model) beginCommitMsg() (Model, tea.Cmd) {
	m.mode = modeCommitMsg
	m.err = nil
	m.formInput.EchoMode = textinput.EchoNormal
	m.formInput.SetValue(m.commitMsg)
	m.formInput.Placeholder = "empty to skip git commit/push"
	m.formInput.Focus()
	return m, nil
}

func (m Model) updateCommitMsg(action keyAction, key tea.KeyPressMsg) (Model, tea.Cmd) {
	switch action {
	case keyActionBack:
		m.formInput.Blur()
		m.mode = modeDone
		return m, nil
	case keyActionConfirm:
		m.commitMsg = strings.TrimSpace(m.formInput.Value())
		m.formInput.Blur()
		return m.beginDeploy()
	}
	var cmd tea.Cmd
	m.formInput, cmd = m.formInput.Update(key)
	return m, cmd
}

func (m Model) beginDeploy() (Model, tea.Cmd) {
	m.mode = modeDeploying
	m.deployErr = nil
	m.runSteps = nil
	m.err = nil
	return m, tea.Batch(m.spinner.Tick, m.deployCmd())
}

func (m Model) deployCmd() tea.Cmd {
	runner := m.runner
	if runner == nil {
		runner = upbridge.DefaultRunner()
	}
	in := upbridge.DeployInput{
		Cloud:            m.flow.Cloud,
		CloudCluster:     m.cloudInstance.Name,
		ImportClusterID:  m.importClusterID,
		IgnorePreflights: m.ignorePreflights || m.flow.DryRun,
		CommitMsg:        m.commitMsg,
	}
	ctx := m.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return func() tea.Msg {
		var steps []string
		err := runner.Deploy(ctx, in, func(step string) {
			steps = append(steps, step)
		})
		return deployDoneMsg{err: err, steps: steps}
	}
}

func (m Model) applyDeployDone(msg deployDoneMsg) (Model, tea.Cmd) {
	if m.mode != modeDeploying {
		return m, nil
	}
	if len(msg.steps) > 0 {
		m.runSteps = msg.steps
	}
	m.deployErr = msg.err
	m.mode = modeComplete
	if msg.err != nil {
		m.err = msg.err
	} else {
		m.err = nil
	}
	return m, nil
}

func (m Model) beginRun() (Model, tea.Cmd) {
	if len(m.formValues) == 0 {
		m.err = fmt.Errorf("provider survey values are required to write workspace.yaml (complete credentials/region first)")
		return m, nil
	}
	m.err = nil
	m.runErr = nil
	m.deployErr = nil
	m.runSteps = nil
	m.importClusterID = ""
	m.mode = modeRunning
	return m, tea.Batch(m.spinner.Tick, m.runCmd())
}

func (m Model) runCmd() tea.Cmd {
	runner := m.runner
	if runner == nil {
		runner = upbridge.DefaultRunner()
	}
	in := upbridge.RunInput{
		Flush: upbridge.FlushInput{
			ProviderID: m.provider.ID,
			Values:     copyStringMap(m.formValues),
			AppDomain:  m.appDomain,
			Cloud:      m.flow.Cloud,
		},
		Generate: upbridge.GenerateInput{
			Cloud:            m.flow.Cloud,
			CloudCluster:     m.cloudInstance.Name,
			IgnorePreflights: m.ignorePreflights || m.flow.DryRun,
		},
	}
	ctx := m.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return func() tea.Msg {
		var steps []string
		res, err := runner.Run(ctx, in, func(step string) {
			steps = append(steps, step)
		})
		return runDoneMsg{err: err, steps: steps, importClusterID: res.ImportClusterID}
	}
}

func (m Model) applyRunDone(msg runDoneMsg) (Model, tea.Cmd) {
	if m.mode != modeRunning {
		return m, nil
	}
	if len(msg.steps) > 0 {
		m.runSteps = msg.steps
	}
	m.importClusterID = msg.importClusterID
	m.runErr = msg.err
	m.mode = modeDone
	if msg.err != nil {
		m.err = msg.err
	} else {
		m.err = nil
	}
	return m, nil
}

func (m Model) updateSetupGit(action keyAction, key tea.KeyPressMsg) (Model, tea.Cmd) {
	opts := m.gitAffirmOptions()
	switch action {
	case keyActionBack:
		m.err = nil
		if m.probeWarn != "" {
			m.mode = modeIgnoreContinue
			return m, nil
		}
		if len(m.formFields) > 0 {
			m.mode = modeProviderForm
			m.applyLoadFormField()
			return m, nil
		}
		m.mode = modeSelectProvider
		m.cursor = 0
		return m, nil
	case keyActionUp:
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case keyActionDown:
		if m.cursor < len(opts)-1 {
			m.cursor++
		}
		return m, nil
	case keyActionConfirm:
		return m.chooseSetupGit(opts[m.cursor].value)
	}
	text := keyText(key)
	switch text {
	case "1", "y", "Y":
		m.cursor = 0
		return m.chooseSetupGit(true)
	case "2", "n", "N":
		m.cursor = 1
		return m.chooseSetupGit(false)
	}
	return m, nil
}

func (m Model) gitAffirmOptions() []yesNoOption {
	if m.inGitRepo {
		return setupGitContinueOptions()
	}
	return setupGitOptions()
}

func (m Model) updateSelectSCM(action keyAction, key tea.KeyPressMsg) (Model, tea.Cmd) {
	switch action {
	case keyActionBack:
		m.mode = modeSetupGit
		m.cursor = 0
		m.scm = upbridge.SCMProvider{}
		return m, nil
	case keyActionUp:
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case keyActionDown:
		if m.cursor < len(m.scms)-1 {
			m.cursor++
		}
		return m, nil
	case keyActionConfirm:
		return m.chooseSCM(m.scms[m.cursor])
	}
	text := keyText(key)
	for i, s := range m.scms {
		if text == string(rune('1'+i)) || text == scmShortcut(s.ID) {
			m.cursor = i
			return m.chooseSCM(s)
		}
	}
	return m, nil
}

func (m Model) updateAppDomain(action keyAction, key tea.KeyPressMsg) (Model, tea.Cmd) {
	switch action {
	case keyActionBack:
		if m.scm.ID != "" {
			m.mode = modeSelectSCM
			m.cursor = 0
			for i, s := range m.scms {
				if s.ID == m.scm.ID {
					m.cursor = i
					break
				}
			}
			return m, nil
		}
		// Affirm was shown this run — Esc returns there (ignore-continue or !inGit).
		if m.gitAffirmOpen {
			m.mode = modeSetupGit
			m.cursor = 0
			return m, nil
		}
		if len(m.formFields) > 0 {
			m.mode = modeProviderForm
			m.applyLoadFormField()
			return m, nil
		}
		m.mode = modeSelectProvider
		return m, nil
	case keyActionConfirm:
		return m.confirmAppDomain()
	case keyActionDown:
		if m.domainIsSelect() && m.optionCursor < len(m.domainOpts)-1 {
			m.optionCursor++
		}
		return m, nil
	case keyActionUp:
		if m.domainIsSelect() && m.optionCursor > 0 {
			m.optionCursor--
		}
		return m, nil
	}
	if !m.domainIsSelect() {
		var cmd tea.Cmd
		m.formInput, cmd = m.formInput.Update(key)
		return m, cmd
	}
	return m, nil
}

func (m Model) updateAffirmDeploy(action keyAction, key tea.KeyPressMsg) (Model, tea.Cmd) {
	opts := affirmDeployOptions()
	switch action {
	case keyActionBack:
		m.err = nil
		m.mode = modeAppDomain
		if m.domainIsSelect() {
			m.formInput.Blur()
		} else {
			m.formInput.SetValue(m.appDomain)
			m.formInput.Focus()
		}
		return m, nil
	case keyActionUp:
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case keyActionDown:
		if m.cursor < len(opts)-1 {
			m.cursor++
		}
		return m, nil
	case keyActionConfirm:
		return m.chooseAffirmDeploy(opts[m.cursor].value)
	}
	text := keyText(key)
	switch text {
	case "1", "y", "Y":
		m.cursor = 0
		return m.chooseAffirmDeploy(true)
	case "2", "n", "N":
		m.cursor = 1
		return m.chooseAffirmDeploy(false)
	}
	return m, nil
}

func (m Model) updateCLITip(action keyAction) (Model, tea.Cmd) {
	if action == keyActionBack {
		m.mode = modeIgnorePreflights
		m.cursor = 0
		if m.ignorePreflights {
			m.cursor = 1
		}
		return m, nil
	}
	return m, nil
}

func (m Model) selectFlow(f upbridge.Flow) (Model, tea.Cmd) {
	m.flow = f
	m.err = nil
	m.ignorePreflights = false
	m.ignoreAsked = false
	m.mode = modeIgnorePreflights
	m.cursor = 0
	return m, nil
}

func (m Model) chooseIgnorePreflights(ignore bool) (Model, tea.Cmd) {
	m.ignorePreflights = ignore
	m.ignoreAsked = true
	m.err = nil
	if m.flow.ID == "cloud" {
		return m.beginLoadInstances()
	}
	if m.flow.NeedsProvider() {
		m.mode = modeSelectProvider
		m.cursor = 0
		return m, nil
	}
	m.mode = modeCLITip
	return m, nil
}

func (m Model) readPriorConsole() (url, token string) {
	if m.priorConsole != nil {
		return m.priorConsole()
	}
	c := upbridge.ReadPriorConsole()
	return c.Url, c.Token
}

func (m Model) resetConsoleInput() {
	m.formInput.EchoMode = textinput.EchoNormal
	m.formInput.SetValue("")
	m.formInput.Placeholder = ""
	m.formInput.Blur()
	m.consoleTokenMode = false
}

func (m Model) beginLoadInstances() (Model, tea.Cmd) {
	m.mode = modeLoadInstances
	m.instances = nil
	m.cloudInstance = upbridge.ConsoleInstance{}
	m.err = nil
	m.resetConsoleInput()
	return m, tea.Batch(m.spinner.Tick, m.listInstancesCmd())
}

func (m Model) listInstancesCmd() tea.Cmd {
	lister := m.instanceLister
	if lister == nil {
		lister = upbridge.DefaultInstanceLister()
	}
	ctx := m.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return func() tea.Msg {
		items, err := lister.List(ctx)
		return instancesMsg{items: items, err: err}
	}
}

func (m Model) applyInstances(msg instancesMsg) (Model, tea.Cmd) {
	if m.mode != modeLoadInstances {
		return m, nil
	}
	if msg.err != nil {
		m.err = msg.err
		m.mode = modeIgnorePreflights
		m.cursor = 0
		if m.ignorePreflights {
			m.cursor = 1
		}
		return m, nil
	}
	if len(msg.items) == 0 {
		m.err = fmt.Errorf("no cloud instances are available for this account")
		m.mode = modeIgnorePreflights
		m.cursor = 0
		if m.ignorePreflights {
			m.cursor = 1
		}
		return m, nil
	}
	m.instances = msg.items
	m.err = nil
	if len(msg.items) == 1 {
		return m.chooseInstance(msg.items[0])
	}
	priorURL, _ := m.readPriorConsole()
	m.mode = modeSelectInstance
	m.cursor = upbridge.DefaultInstanceIndex(msg.items, priorURL)
	return m, nil
}

func (m Model) updateLoadInstances(action keyAction) (Model, tea.Cmd) {
	if action == keyActionBack {
		m.mode = modeIgnorePreflights
		m.cursor = 0
		if m.ignorePreflights {
			m.cursor = 1
		}
		m.err = nil
		return m, nil
	}
	return m, nil
}

func (m Model) updateSelectInstance(action keyAction, key tea.KeyPressMsg) (Model, tea.Cmd) {
	switch action {
	case keyActionBack:
		m.mode = modeIgnorePreflights
		m.cursor = 0
		if m.ignorePreflights {
			m.cursor = 1
		}
		m.err = nil
		m.cloudInstance = upbridge.ConsoleInstance{}
		return m, nil
	case keyActionUp:
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case keyActionDown:
		if m.cursor < len(m.instances)-1 {
			m.cursor++
		}
		return m, nil
	case keyActionConfirm:
		if m.cursor >= 0 && m.cursor < len(m.instances) {
			return m.chooseInstance(m.instances[m.cursor])
		}
		return m, nil
	}
	text := keyText(key)
	for i := range m.instances {
		if text == string(rune('1'+i)) {
			m.cursor = i
			return m.chooseInstance(m.instances[i])
		}
	}
	return m, nil
}

func (m Model) chooseInstance(inst upbridge.ConsoleInstance) (Model, tea.Cmd) {
	m.cloudInstance = inst
	m.err = nil
	return m.beginConsoleLogin()
}

func (m Model) beginConsoleLogin() (Model, tea.Cmd) {
	m.mode = modeConsoleLogin
	m.err = nil
	priorURL, _ := m.readPriorConsole()
	if upbridge.PriorConsoleMatches(priorURL, m.cloudInstance.URL) {
		m.resetConsoleInput()
		m.consoleTokenMode = false
		m.cursor = 0 // Yes keep existing
		return m, nil
	}
	return m.beginConsoleToken()
}

func (m Model) beginConsoleToken() (Model, tea.Cmd) {
	m.consoleTokenMode = true
	m.mode = modeConsoleLogin
	m.err = nil
	m.formInput.EchoMode = textinput.EchoPassword
	m.formInput.EchoCharacter = '*'
	m.formInput.SetValue("")
	m.formInput.Placeholder = "console access token"
	m.formInput.Focus()
	return m, nil
}

func (m Model) updateConsoleLogin(action keyAction, key tea.KeyPressMsg) (Model, tea.Cmd) {
	if m.consoleTokenMode {
		return m.updateConsoleToken(action, key)
	}
	priorURL, _ := m.readPriorConsole()
	opts := consoleCredOptions(priorURL)
	switch action {
	case keyActionBack:
		m.resetConsoleInput()
		if len(m.instances) > 1 {
			m.mode = modeSelectInstance
			m.cursor = upbridge.DefaultInstanceIndex(m.instances, priorURL)
			return m, nil
		}
		m.mode = modeIgnorePreflights
		m.cursor = 0
		if m.ignorePreflights {
			m.cursor = 1
		}
		return m, nil
	case keyActionUp:
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case keyActionDown:
		if m.cursor < len(opts)-1 {
			m.cursor++
		}
		return m, nil
	case keyActionConfirm:
		return m.chooseConsoleCreds(opts[m.cursor].value)
	}
	text := keyText(key)
	switch text {
	case "1", "y", "Y":
		m.cursor = 0
		return m.chooseConsoleCreds(true)
	case "2", "n", "N":
		m.cursor = 1
		return m.chooseConsoleCreds(false)
	}
	return m, nil
}

func (m Model) chooseConsoleCreds(keep bool) (Model, tea.Cmd) {
	if keep {
		return m.finishConsoleLogin("")
	}
	return m.beginConsoleToken()
}

func (m Model) updateConsoleToken(action keyAction, key tea.KeyPressMsg) (Model, tea.Cmd) {
	switch action {
	case keyActionBack:
		priorURL, _ := m.readPriorConsole()
		if upbridge.PriorConsoleMatches(priorURL, m.cloudInstance.URL) {
			m.resetConsoleInput()
			m.consoleTokenMode = false
			m.cursor = 0
			m.err = nil
			return m, nil
		}
		m.resetConsoleInput()
		if len(m.instances) > 1 {
			m.mode = modeSelectInstance
			m.cursor = upbridge.DefaultInstanceIndex(m.instances, priorURL)
			return m, nil
		}
		m.mode = modeIgnorePreflights
		m.cursor = 0
		if m.ignorePreflights {
			m.cursor = 1
		}
		return m, nil
	case keyActionConfirm:
		token := strings.TrimSpace(m.formInput.Value())
		if token == "" {
			m.err = fmt.Errorf("console access token is required")
			return m, nil
		}
		return m.finishConsoleLogin(token)
	}
	var cmd tea.Cmd
	m.formInput, cmd = m.formInput.Update(key)
	return m, cmd
}

func (m Model) finishConsoleLogin(newToken string) (Model, tea.Cmd) {
	priorURL, priorToken := m.readPriorConsole()
	url := m.cloudInstance.URL
	token := priorToken
	confURL := priorURL

	if newToken != "" {
		token = newToken
		confURL = url
		save := m.saveConsole
		if save == nil {
			save = upbridge.SaveConsoleConfig
		}
		if err := save(url, token); err != nil {
			m.err = err
			return m, nil
		}
	} else if !upbridge.PriorConsoleMatches(priorURL, url) {
		m.err = fmt.Errorf("console credentials do not match the selected instance")
		return m, nil
	}
	if confURL == "" {
		confURL = url
	}

	if err := upbridge.ValidateConsoleConfig(m.instances, console.Config{Url: confURL, Token: token}); err != nil {
		m.err = err
		return m, nil
	}

	m.resetConsoleInput()
	m.err = nil
	m.mode = modeSelectProvider
	m.cursor = 0
	return m, nil
}

func (m Model) beginProviderForm(p upbridge.Provider) (Model, tea.Cmd) {
	m.provider = p
	m.mode = modeProbing
	m.formFields = nil
	m.formValues = nil
	m.formIndex = 0
	m.credSummary = ""
	m.probeWarn = ""
	m.freeTextKeys = nil
	m.err = nil
	return m, tea.Batch(m.spinner.Tick, m.probeCmd())
}

func (m Model) probeCmd() tea.Cmd {
	prober := m.prober
	providerID := m.provider.ID
	ctx := m.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return func() tea.Msg {
		res, err := prober.Probe(ctx, providerID)
		return probeMsg{result: res, err: err}
	}
}

func (m Model) applyProbe(msg probeMsg) (Model, tea.Cmd) {
	if m.mode != modeProbing {
		return m, nil
	}
	if msg.err != nil {
		// Credential/GetProvider failure. With --ignore-preflights the CLI warns and
		// continues to the git Affirm — it does not open a region survey.
		m.probeWarn = msg.err.Error()
		if m.ignorePreflights {
			m.err = nil
			m.formFields = nil
			m.formValues = nil
			m.mode = modeIgnoreContinue
			return m, nil
		}
		m.err = msg.err
		m.mode = modeSelectProvider
		m.cursor = 0
		for i, p := range m.providers {
			if p.ID == m.provider.ID {
				m.cursor = i
				break
			}
		}
		return m, nil
	}
	m.probeWarn = ""
	m.credSummary = msg.result.Summary
	return m.openForm(msg.result.Fields), nil
}

func (m Model) submitProviderForm() (Model, tea.Cmd) {
	m.formInput.Blur()
	if err := upbridge.ValidateProviderForm(m.provider.ID, m.formValues); err != nil {
		m.err = err
		for i, field := range m.formFields {
			if field.Key == "cluster" || (field.Required && strings.TrimSpace(m.formValues[field.Key]) == "") {
				m.formIndex = i
				m.applyLoadFormField()
				break
			}
		}
		return m, nil
	}
	m.err = nil
	m.mode = modeRunPreflights
	return m, tea.Batch(m.spinner.Tick, m.preflightCmd())
}

func (m Model) preflightCmd() tea.Cmd {
	prober := m.prober
	providerID := m.provider.ID
	values := copyStringMap(m.formValues)
	ctx := m.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return func() tea.Msg {
		return preflightMsg{err: prober.Preflights(ctx, providerID, values)}
	}
}

func (m Model) applyPreflight(msg preflightMsg) (Model, tea.Cmd) {
	if m.mode != modeRunPreflights {
		return m, nil
	}
	if msg.err != nil {
		m.probeWarn = msg.err.Error()
		if m.ignorePreflights {
			// CLI: print warning and continue to git Affirm / Flush.
			m.err = nil
			m.mode = modeIgnoreContinue
			return m, nil
		}
		m.err = fmt.Errorf("preflight checks failed: %w (rerun with --ignore-preflights to skip)", msg.err)
		m.mode = modeProviderForm
		m.applyLoadFormField()
		return m, nil
	}
	m.probeWarn = ""
	return m.beginGitStep()
}

func (m Model) beginGitStep() (Model, tea.Cmd) {
	check := m.gitChecker
	if check == nil {
		check = upbridge.InGitRepo
	}
	m.inGitRepo = check()
	m.err = nil
	if m.inGitRepo {
		m.gitAffirmOpen = false
		return m.afterGitReady()
	}
	// Default Yes — same as survey.Confirm Default: true
	m.mode = modeSetupGit
	m.gitAffirmOpen = true
	m.cursor = 0
	return m, nil
}

func (m Model) chooseSetupGit(yes bool) (Model, tea.Cmd) {
	if !yes {
		if m.inGitRepo {
			m.err = fmt.Errorf("cancelled continuing plural up init")
			return m, nil
		}
		m.err = fmt.Errorf("you're not in a git repository, either clone one directly or let us set it up for you")
		return m, nil
	}
	m.err = nil
	if m.inGitRepo {
		// CLI skips Affirm + scm.Setup when already in a work tree.
		return m.afterGitReady()
	}
	// CLI runs scm.Setup() — first prompt is SCM provider select.
	m.mode = modeSelectSCM
	m.cursor = 0
	m.scm = upbridge.SCMProvider{}
	return m, nil
}

func (m Model) chooseSCM(s upbridge.SCMProvider) (Model, tea.Cmd) {
	m.scm = s
	m.err = nil
	return m.afterGitReady()
}

func (m Model) afterGitReady() (Model, tea.Cmd) {
	// CLI handleUp: after init → askAppDomain (unless dry-run) → Affirm deploy.
	if m.flow.DryRun {
		m.mode = modeSelected
		return m, nil
	}
	return m.beginAppDomain()
}

func (m Model) beginAppDomain() (Model, tea.Cmd) {
	m.mode = modeAppDomain
	m.appDomain = ""
	m.domainOpts = nil
	m.optionCursor = 0
	m.err = nil
	m.formInput.SetValue("")
	m.formInput.Placeholder = "leave empty to skip"
	m.formInput.Focus()
	return m, tea.Batch(m.spinner.Tick, m.loadDomainCmd())
}

func (m Model) loadDomainCmd() tea.Cmd {
	if m.domainLoader != nil {
		loader := m.domainLoader
		return func() tea.Msg { return loader() }
	}
	providerID := m.provider.ID
	region := strings.TrimSpace(m.formValues["region"])
	if region == "" {
		region = strings.TrimSpace(m.formValues["location"])
	}
	resourceGroup := strings.TrimSpace(m.formValues["resourceGroup"])
	ctx := m.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return func() tea.Msg {
		switch providerID {
		case "aws":
			if region == "" {
				return domainMsg{text: true}
			}
			zones, err := provider.AWSHostedZones(ctx, region)
			if err != nil {
				// CLI: print error, "ignoring domain setup...", continue
				return domainMsg{err: err, skip: true}
			}
			return domainMsg{options: append([]string{upbridge.DomainNoneOption}, zones...)}
		case "azure":
			if resourceGroup == "" {
				return domainMsg{text: true}
			}
			zones, err := provider.AzureDNSZones(ctx, resourceGroup)
			if err != nil {
				return domainMsg{err: err, skip: true}
			}
			if len(zones) == 0 {
				return domainMsg{skip: true}
			}
			return domainMsg{options: append([]string{upbridge.DomainNoneOption}, zones...)}
		default:
			return domainMsg{text: true}
		}
	}
}

func (m Model) applyDomain(msg domainMsg) (Model, tea.Cmd) {
	if m.mode != modeAppDomain {
		return m, nil
	}
	if msg.skip {
		if msg.err != nil {
			m.domainNote = msg.err.Error()
		}
		m.appDomain = ""
		return m.beginAffirmDeploy()
	}
	if msg.text || len(msg.options) == 0 {
		m.domainOpts = nil
		m.formInput.Focus()
		return m, nil
	}
	m.domainOpts = msg.options
	m.optionCursor = 0
	m.formInput.Blur()
	return m, nil
}

func (m Model) domainIsSelect() bool {
	return len(m.domainOpts) > 0
}

func (m Model) confirmAppDomain() (Model, tea.Cmd) {
	if m.domainIsSelect() {
		chosen := m.domainOpts[m.optionCursor]
		if chosen == upbridge.DomainNoneOption {
			m.appDomain = ""
		} else {
			m.appDomain = chosen
		}
	} else {
		m.appDomain = strings.TrimSpace(m.formInput.Value())
	}
	m.formInput.Blur()
	m.err = nil
	return m.beginAffirmDeploy()
}

func (m Model) beginAffirmDeploy() (Model, tea.Cmd) {
	// CLI skips AffirmUp when --cloud.
	if m.flow.Cloud {
		m.mode = modeSelected
		m.err = nil
		return m, nil
	}
	m.mode = modeAffirmDeploy
	m.cursor = 0 // Yes
	m.err = nil
	return m, nil
}

func (m Model) chooseAffirmDeploy(yes bool) (Model, tea.Cmd) {
	if !yes {
		m.err = fmt.Errorf("cancelled deploy")
		return m, nil
	}
	m.err = nil
	m.mode = modeSelected
	return m, nil
}

func scmShortcut(id string) string {
	switch id {
	case "github":
		return "g"
	case "gitlab":
		return "l"
	case "bitbucket":
		return "b"
	default:
		return ""
	}
}

func (m Model) openForm(fields []upbridge.FormField) Model {
	m.mode = modeProviderForm
	m.formFields = fields
	m.formIndex = 0
	m.formValues = map[string]string{}
	m.freeTextKeys = map[string]bool{}
	for _, field := range fields {
		if field.Default != "" {
			m.formValues[field.Key] = field.Default
		}
	}
	m.err = nil
	m.syncOptionCursor()
	m.applyLoadFormField()
	return m
}

func (m Model) applyOptions(msg optionsMsg) (Model, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		return m, nil
	}
	for i := range m.formFields {
		if m.formFields[i].Key == msg.fieldKey {
			m.formFields[i].Options = msg.options
			if cur := m.formValues[msg.fieldKey]; cur != "" {
				found := false
				for _, opt := range msg.options {
					if opt == cur {
						found = true
						break
					}
				}
				if !found && len(msg.options) > 0 {
					m.formValues[msg.fieldKey] = ""
				}
			}
			if m.formValues[msg.fieldKey] == "" && m.formFields[i].Default != "" {
				for _, opt := range msg.options {
					if opt == m.formFields[i].Default {
						m.formValues[msg.fieldKey] = opt
						break
					}
				}
			}
			break
		}
	}
	if m.currentFieldKey() == msg.fieldKey {
		m.syncOptionCursor()
	}
	m.err = nil
	return m, nil
}

func (m Model) confirmSelectOption() (Model, tea.Cmd) {
	opts := m.currentOptions()
	if len(opts) == 0 {
		return m, nil
	}
	if m.optionCursor < 0 || m.optionCursor >= len(opts) {
		m.optionCursor = 0
	}
	chosen := opts[m.optionCursor]
	key := m.currentFieldKey()
	if chosen == provider.CreateNewOption {
		m.freeTextKeys[key] = true
		m.formValues[key] = ""
		m.applyLoadFormField()
		return m, nil
	}
	m.formValues[key] = chosen
	return m.advanceForm()
}

func (m Model) advanceForm() (Model, tea.Cmd) {
	var refresh tea.Cmd
	if m.provider.ID == "gcp" && m.currentFieldKey() == "project" {
		refresh = m.refreshFieldOptions("region")
	}
	if m.formIndex < len(m.formFields)-1 {
		m.formIndex++
		m.err = nil
		m.applyLoadFormField()
		return m, refresh
	}
	return m.submitProviderForm()
}

func (m Model) refreshFieldOptions(fieldKey string) tea.Cmd {
	prober := m.prober
	providerID := m.provider.ID
	values := copyStringMap(m.formValues)
	ctx := m.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return func() tea.Msg {
		opts, err := prober.FieldOptions(ctx, providerID, fieldKey, values)
		return optionsMsg{fieldKey: fieldKey, options: opts, err: err}
	}
}

func (m *Model) saveFormField() {
	if m.formValues == nil {
		m.formValues = map[string]string{}
	}
	if m.formIndex >= 0 && m.formIndex < len(m.formFields) && !m.currentIsSelect() {
		m.formValues[m.formFields[m.formIndex].Key] = strings.TrimSpace(m.formInput.Value())
	}
}

func (m *Model) applyLoadFormField() {
	if m.formIndex < 0 || m.formIndex >= len(m.formFields) {
		return
	}
	m.syncOptionCursor()
	if m.currentIsSelect() {
		m.formInput.Blur()
		return
	}
	field := m.formFields[m.formIndex]
	m.formInput.SetValue(m.formValues[field.Key])
	m.formInput.Placeholder = field.Placeholder
	if m.formInput.Placeholder == "" {
		m.formInput.Placeholder = field.Label
	}
	m.formInput.Focus()
}

func (m *Model) syncOptionCursor() {
	opts := m.currentOptions()
	cur := m.formValues[m.currentFieldKey()]
	m.optionCursor = 0
	for i, opt := range opts {
		if opt == cur {
			m.optionCursor = i
			return
		}
	}
	field := m.currentField()
	if field.Default != "" {
		for i, opt := range opts {
			if opt == field.Default {
				m.optionCursor = i
				return
			}
		}
	}
}

func (m Model) currentField() upbridge.FormField {
	if m.formIndex < 0 || m.formIndex >= len(m.formFields) {
		return upbridge.FormField{}
	}
	return m.formFields[m.formIndex]
}

func (m Model) currentFieldKey() string {
	return m.currentField().Key
}

func (m Model) currentOptions() []string {
	return m.currentField().Options
}

func (m Model) currentIsSelect() bool {
	key := m.currentFieldKey()
	if m.freeTextKeys[key] {
		return false
	}
	return len(m.currentOptions()) > 0
}

func (m Model) cli() string {
	return m.flow.CLI(m.ignorePreflights)
}

func keyText(key tea.KeyPressMsg) string {
	text := key.Text
	if text == "" && key.Code > 0 && key.Code < 128 {
		text = string(rune(key.Code))
	}
	return text
}

func flowShortcut(id string) string {
	switch id {
	case "self-hosted":
		return "s"
	case "cloud":
		return "c"
	case "dry-run":
		return "d"
	case "cloud-dry-run":
		return "x"
	default:
		return ""
	}
}

func providerShortcut(id string) string {
	switch id {
	case "aws":
		return "a"
	case "azure":
		return "z"
	case "gcp":
		return "g"
	case "byok":
		return "b"
	default:
		return ""
	}
}

func formValue(values map[string]string, key string) string {
	if v := strings.TrimSpace(values[key]); v != "" {
		return v
	}
	return "—"
}

func truncate(v string, n int) string {
	if len(v) <= n {
		return v
	}
	return v[:n-1] + "…"
}

func yesNoLabel(v bool) string {
	if v {
		return "ignored (--ignore-preflights)"
	}
	return "run (default)"
}

func copyStringMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// SelectedFlow returns the chosen flow id, or empty if none yet.
func (m Model) SelectedFlow() string { return m.flow.ID }

// SelectedProvider returns the chosen provider id, or empty if none yet.
func (m Model) SelectedProvider() string { return m.provider.ID }

// Cloud reports whether the chosen flow uses --cloud.
func (m Model) Cloud() bool { return m.flow.Cloud }

// DryRun reports whether the chosen flow uses --dry-run.
func (m Model) DryRun() bool { return m.flow.DryRun }

// IgnorePreflights reports whether --ignore-preflights was chosen.
func (m Model) IgnorePreflights() bool { return m.ignorePreflights }
