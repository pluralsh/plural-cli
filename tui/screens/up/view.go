package up

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"

	upbridge "github.com/pluralsh/plural-cli/pkg/bridge/up"
	"github.com/pluralsh/plural-cli/tui/components/page"
)

func (m Model) View(width, height int) string {
	width, height = page.Size(width, height)
	if width < page.MinimumWidth || height < page.MinimumHeight {
		return page.Unsupported(m.theme, width, height)
	}
	contentWidth := page.ContentWidth(width)
	body, help := m.bodyAndHelp(contentWidth, height)
	return page.Render(m.theme, width, height, "Up", m.headerStatus(), body, help)
}

func (m Model) headerStatus() string {
	switch m.mode {
	case modeIgnorePreflights:
		return m.theme.Muted.Render("step 2 · preflights")
	case modeLoadInstances:
		return m.theme.Muted.Render("loading Console instances…")
	case modeSelectInstance:
		return m.theme.Muted.Render("step · Console instance")
	case modeConsoleLogin:
		return m.headerConsoleLogin()
	case modeSelectProvider:
		return m.theme.Muted.Render("step 3 · provider")
	case modeProbing:
		return m.theme.Muted.Render("checking credentials…")
	case modeProviderForm:
		return m.theme.Muted.Render("step 4 · " + m.provider.ID)
	case modeRunPreflights:
		return m.theme.Muted.Render("running preflights…")
	case modeIgnoreContinue:
		return m.theme.Muted.Render("continuing · ignored failures")
	case modeAlreadyInit:
		return m.theme.Muted.Render("step · already initialized")
	case modeEnsuringInit:
		return m.theme.Muted.Render("checking workspace…")
	case modeBucketPrefix:
		return m.theme.Muted.Render("step · bucket naming")
	case modePluralSubdomain:
		return m.theme.Muted.Render("step · onplural.sh")
	case modeSetupGit:
		return m.theme.Muted.Render("step · git repository")
	case modeSelectSCM:
		return m.theme.Muted.Render("step · scm provider")
	case modeSCMSetup:
		return m.theme.Muted.Render("scm · authenticate / create / clone")
	case modeAppDomain:
		return m.theme.Muted.Render("step · app domain")
	case modeAffirmDeploy:
		return m.theme.Muted.Render("step · deploy Affirm")
	case modeSelected:
		return m.headerSelected()
	case modeRunning:
		return m.theme.Muted.Render("running Flush + Generate…")
	case modeDone:
		return m.headerDone()
	case modeDeploying:
		return m.theme.Muted.Render("deploying…")
	case modeDeployCommit:
		return m.theme.Muted.Render("commit checkpoint")
	case modeComplete:
		return m.headerComplete()
	case modeCLITip:
		return m.theme.Muted.Render(m.flow.ID)
	default:
		return m.theme.Muted.Render("step 1 · mode")
	}
}

func (m Model) headerConsoleLogin() string {
	if m.consoleTokenMode {
		return m.theme.Muted.Render("step · console token")
	}
	return m.theme.Muted.Render("step · console credentials")
}

func (m Model) headerSelected() string {
	if m.cloudInstance.Name != "" {
		return m.theme.Success.Render(m.cloudInstance.Name)
	}
	if m.provider.ID != "" {
		return m.theme.Success.Render(m.provider.ID)
	}
	return m.theme.Success.Render("ready")
}

func (m Model) headerDone() string {
	if m.runErr != nil {
		return m.theme.Danger.Render("failed")
	}
	return m.theme.Success.Render("generated")
}

func (m Model) headerComplete() string {
	if m.deployErr != nil {
		return m.theme.Danger.Render("deploy failed")
	}
	return m.theme.Success.Render("deployed")
}

func (m Model) viewPlan(width int) (string, string) {
	lines := []string{
		m.theme.Success.Render("✓ Continuing plural up init"),
		"",
		"Mode         " + m.flow.Title,
		"Preflights   " + yesNoLabel(m.ignorePreflights),
	}
	if m.cloudInstance.Name != "" {
		lines = append(lines, "Console      "+m.cloudInstance.Name)
		if m.cloudInstance.URL != "" {
			lines = append(lines, "             "+truncate(m.cloudInstance.URL, max(20, width-16)))
		}
	}
	if m.provider.Title != "" {
		lines = append(lines, "Provider     "+m.provider.Title+" ("+m.provider.ID+")")
	}
	if m.credSummary != "" {
		lines = append(lines, "Credentials  "+truncate(m.credSummary, max(20, width-16)))
	}
	if m.probeWarn != "" {
		for _, w := range strings.Split(m.probeWarn, "\n") {
			if strings.TrimSpace(w) != "" {
				lines = append(lines, m.theme.Danger.Render("Warning: "+truncate(w, max(20, width-12))))
			}
		}
	}
	if m.domainNote != "" {
		lines = append(lines, m.theme.Muted.Render("Domain setup ignored: "+truncate(m.domainNote, max(20, width-24))))
	}
	for _, field := range m.formFields {
		label := field.Label + strings.Repeat(" ", max(1, 12-len(field.Label)))
		lines = append(lines, label+" "+formValue(m.formValues, field.Key))
	}
	if !m.flow.Cloud {
		if m.bucketPrefix != "" {
			lines = append(lines, "Bucket       "+m.bucketPrefix)
		}
		if m.pluralDNS != "" {
			lines = append(lines, "Plural DNS   "+m.pluralDNS)
		}
	}
	switch {
	case m.scm.ID != "":
		scmLine := "SCM          " + m.scm.Title
		if m.scmRepo != "" {
			scmLine += " → " + m.scmRepo
		}
		lines = append(lines, scmLine)
	case m.alreadyInit:
		lines = append(lines, "Git          workspace.yaml present · init skipped")
	case m.inGitRepo:
		lines = append(lines, "Git          already inside a work tree")
	}
	domain := m.appDomain
	if domain == "" {
		domain = "(skipped)"
	}
	if !m.flow.DryRun {
		lines = append(lines, "App domain   "+domain)
	}
	next := m.planNextHint()
	lines = append(lines,
		"",
		m.theme.Muted.Render("Equivalent CLI"),
		"  "+m.cli(),
		"",
		m.theme.Muted.Render(next),
	)
	help := "enter run · esc back · ctrl+c quit"
	if m.err != nil {
		lines = append(lines, "", m.theme.Danger.Render(m.err.Error()))
	}
	return page.Panel(m.theme, "Plan", lines, width, 18, true), help
}

func (m Model) planNextHint() string {
	switch {
	case m.flow.DryRun && m.alreadyInit && m.flow.Cloud:
		return "Enter to ImportCluster + Generate only (skip Flush, no Deploy)."
	case m.flow.DryRun && m.alreadyInit:
		return "Enter to Generate only (skip Flush, no Deploy — --dry-run)."
	case m.flow.DryRun && m.flow.Cloud:
		return "Enter to Flush + ImportCluster + Generate only (no Deploy — --cloud --dry-run)."
	case m.flow.DryRun:
		return "Enter to Flush + Generate only (no Deploy — --dry-run)."
	case m.alreadyInit && m.flow.Cloud:
		return "Enter to ImportCluster + Generate (skip Flush), then Deploy."
	case m.alreadyInit:
		return "Enter to Generate (skip Flush — workspace.yaml exists), then Deploy."
	case m.flow.Cloud:
		return "Enter to Flush + ImportCluster + Generate, then Deploy."
	default:
		return "Enter to Flush workspace.yaml + Generate, then Deploy."
	}
}

func (m Model) bodyAndHelp(width, height int) (string, string) {
	switch m.mode {
	case modeSelected:
		return m.viewPlan(width)
	case modeRunning:
		return m.viewGenerating(width, height)
	case modeDone:
		return m.viewGenerateDone(width, height)
	case modeDeploying:
		return m.viewDeploying(width, height)
	case modeDeployCommit:
		return m.viewDeployCommit(width, height)
	case modeComplete:
		return m.viewDeployComplete(width, height)
	case modeCLITip:
		return m.viewCLITip(width)
	case modeLoadInstances:
		return m.viewLoadInstances(width)
	case modeSelectInstance:
		return m.viewSelectInstance(width)
	case modeConsoleLogin:
		return m.viewConsoleLogin(width)
	case modeSetupGit:
		return m.viewSetupGit(width)
	case modeAffirmDeploy:
		return m.viewAffirmDeploy(width)
	case modeIgnoreContinue:
		return m.viewIgnoreContinue(width)
	case modeAlreadyInit:
		return m.viewAlreadyInit(width)
	case modeEnsuringInit:
		return m.viewEnsuringInit(width)
	case modeBucketPrefix:
		return m.viewBucketPrefix(width)
	case modePluralSubdomain:
		return m.viewPluralSubdomain(width)
	case modeRunPreflights:
		return m.viewRunPreflights(width)
	case modeSelectSCM:
		return m.viewSelectSCM(width)
	case modeSCMSetup:
		return m.viewSCMSetup(width)
	case modeAppDomain:
		return m.viewAppDomain(width)
	case modeProbing:
		return m.viewProbing(width)
	case modeProviderForm:
		return m.formView(width)
	case modeSelectProvider:
		return m.viewSelectProvider(width)
	case modeIgnorePreflights:
		return m.viewIgnorePreflights(width)
	default:
		return m.viewSelectFlow(width)
	}
}

func (m Model) viewGenerating(width, height int) (string, string) {
	panelH, logN := logPanelBudget(height, 4)
	lines := []string{
		m.spinner.View() + " " + m.theme.Muted.Render("Running Flush + Generate…"),
		"",
	}
	if m.flow.DryRun {
		lines = append(lines,
			m.theme.Muted.Render("Dry-run: generate only — output streams below (no Deploy)."),
			"",
		)
	} else {
		lines = append(lines,
			m.theme.Muted.Render("Terraform / generation output streams below (TUI stays open)."),
			"",
		)
	}
	lines = append(lines, m.opLogLines(logN, width)...)
	return page.Panel(m.theme, "Generating", lines, width, panelH, true), "↑/↓ · pgup/pgdn scroll · end follow"
}

func (m Model) viewGenerateDone(width, height int) (string, string) {
	panelH, logN := logPanelBudget(height, 6)
	lines := []string{}
	switch {
	case m.runErr != nil:
		lines = append(lines,
			m.theme.Danger.Render("Generate failed — scroll logs below, then esc to Plan"),
			m.theme.Danger.Render(truncate(m.runErr.Error(), max(20, width-4))),
			"",
		)
	case m.flow.DryRun:
		lines = append(lines,
			m.theme.Success.Render("✓ Dry-run finished — no Deploy will run"),
			m.theme.Muted.Render("Scroll logs below · esc returns to Plan"),
			"",
		)
	default:
		lines = append(lines,
			m.theme.Success.Render("✓ Finished generating the repo"),
			m.theme.Muted.Render("Scroll logs below · enter Deploy · esc Plan"),
			"",
		)
	}
	lines = append(lines, m.opLogExportHint(width))
	lines = append(lines, m.opLogScrollHint(width))
	lines = append(lines, m.opLogLines(logN, width)...)
	help := "↑/↓ scroll · e export · esc plan"
	if m.runErr == nil && !m.flow.DryRun {
		help = "↑/↓ scroll · e export · enter deploy · esc plan"
	}
	return page.Panel(m.theme, "Generate complete", lines, width, panelH, true), help
}

func (m Model) viewDeploying(width, height int) (string, string) {
	panelH, logN := logPanelBudget(height, 4)
	lines := make([]string, 0, 4+logN)
	lines = append(lines,
		m.spinner.View()+" "+m.theme.Muted.Render("Running Deploy (terraform / import / apps)…"),
		"",
		m.theme.Muted.Render("Terraform output streams below. Commit is prompted after mgmt apply."),
		"",
	)
	lines = append(lines, m.opLogLines(logN, width)...)
	return page.Panel(m.theme, "Deploying", lines, width, panelH, true), "↑/↓ · pgup/pgdn scroll · end follow"
}

func (m Model) viewDeployCommit(width, height int) (string, string) {
	panelH, logN := logPanelBudget(height, 6)
	lines := make([]string, 0, 6+logN)
	lines = append(lines,
		m.theme.Muted.Render("==> Enter a commit message to push your configuration"),
		m.theme.Muted.Render("Same checkpoint as plural up (after management terraform)."),
		"",
		"› Message",
		"  "+m.formInput.View(),
		"",
	)
	lines = append(lines, m.opLogLines(logN, width)...)
	return page.Panel(m.theme, "Commit", lines, width, panelH, true), "enter continue · esc skip commit"
}

func (m Model) viewDeployComplete(width, height int) (string, string) {
	panelH, logN := logPanelBudget(height, 6)
	lines := []string{}
	if m.deployErr != nil {
		lines = append(lines,
			m.theme.Danger.Render("Deploy failed — scroll logs below, then esc to retry"),
			m.theme.Danger.Render(truncate(m.deployErr.Error(), max(20, width-4))),
			"",
		)
	} else {
		lines = append(lines,
			m.theme.Success.Render("✓ Finished setting up your management cluster!"),
			m.theme.Muted.Render("Scroll logs below · esc back"),
			"",
		)
	}
	lines = append(lines, m.opLogExportHint(width))
	lines = append(lines, m.opLogScrollHint(width))
	lines = append(lines, m.opLogLines(logN, width)...)
	return page.Panel(m.theme, "Deploy complete", lines, width, panelH, true), "↑/↓ scroll · e export · esc back"
}

func (m Model) viewCLITip(width int) (string, string) {
	lines := []string{
		m.theme.Muted.Render(m.flow.Title + " is not fully wired in the TUI yet."),
		m.theme.Muted.Render("Use the CLI for this path, or pick Self-hosted / Plural Cloud."),
		"",
		"Mode         " + m.flow.Title,
		m.theme.Muted.Render("             " + m.flow.Blurb),
		"Preflights   " + yesNoLabel(m.ignorePreflights),
		"",
		m.theme.Muted.Render("Equivalent CLI"),
		"  " + m.cli(),
		"",
		m.theme.Muted.Render("Dry-run wizards land in a later step."),
	}
	return page.Panel(m.theme, "Coming next", lines, width, 14, true), "esc change preflights · ctrl+c quit"
}

func (m Model) viewLoadInstances(width int) (string, string) {
	lines := []string{
		"Mode  " + m.flow.Title,
		"",
		m.spinner.View() + " " + m.theme.Muted.Render("Fetching Console instances (GetConsoleInstances)…"),
		m.theme.Muted.Render("Same list plural up --cloud uses in choseCluster."),
	}
	return page.Panel(m.theme, "Plural Cloud", lines, width, 10, true), "esc cancel"
}

func (m Model) viewSelectInstance(width int) (string, string) {
	lines := []string{
		m.theme.Muted.Render("Select one of the following clusters:"),
		m.theme.Muted.Render("Same survey as plural up --cloud choseCluster."),
		"",
	}
	lines = append(lines, m.instanceLines(width)...)
	if m.err != nil {
		lines = append(lines, "", m.theme.Danger.Render(m.err.Error()))
	}
	help := "↑/↓ · 1–n · enter · esc preflights"
	return page.Panel(m.theme, "Console instance", lines, width, 14, true), help
}

func (m Model) viewConsoleLogin(width int) (string, string) {
	if m.consoleTokenMode {
		lines := []string{
			"Instance  " + m.cloudInstance.Name,
			m.theme.Muted.Render("          " + truncate(m.cloudInstance.URL, max(20, width-12))),
			"",
			m.theme.Muted.Render("Enter your console access token (plural cd login)."),
			"",
			"› Token",
			"  " + m.formInput.View(),
		}
		if m.err != nil {
			lines = append(lines, "", m.theme.Danger.Render(m.err.Error()))
		}
		return page.Panel(m.theme, "Console login", lines, width, 14, true), "enter continue · esc back"
	}
	priorURL, _ := m.readPriorConsole()
	lines := []string{
		"Instance  " + m.cloudInstance.Name,
		m.theme.Muted.Render("          " + truncate(m.cloudInstance.URL, max(20, width-12))),
		"",
		m.theme.Muted.Render(fmt.Sprintf("You've already configured your console at %s,", truncate(priorURL, max(24, width-8)))),
		m.theme.Muted.Render("continue using those credentials?"),
		m.theme.Muted.Render("Same Affirm as HandleCdLogin (PLURAL_CD_USE_EXISTING_CREDENTIALS)."),
		"",
	}
	lines = append(lines, m.consoleCredLines(width)...)
	if m.err != nil {
		lines = append(lines, "", m.theme.Danger.Render(m.err.Error()))
	}
	return page.Panel(m.theme, "Console credentials", lines, width, 14, true), "↑/↓ · y/n · enter · esc back"
}

func (m Model) viewSetupGit(width int) (string, string) {
	lines := []string{}
	if m.probeWarn != "" {
		lines = append(lines,
			m.theme.Muted.Render("Preflight checks failed, but continuing because --ignore-preflights was specified."),
			m.theme.Muted.Render("Please note that you may encounter issues later on during provisioning."),
			m.theme.Danger.Render("Warning: "+truncate(strings.Split(m.probeWarn, "\n")[0], max(20, width-12))),
			"",
		)
	}
	if m.inGitRepo {
		lines = append(lines,
			m.theme.Muted.Render("Already inside a git work tree — plural up skips Affirm / scm.Setup."),
			m.theme.Muted.Render("Continue with the rest of init (domain / workspace)?"),
			"",
		)
	} else {
		lines = append(lines,
			m.theme.Muted.Render(upbridge.SetupGitPrompt),
			m.theme.Muted.Render("Same Affirm as plural up / init (PLURAL_INIT_AFFIRM_SETUP_REPO)."),
			"",
		)
	}
	lines = append(lines, m.setupGitLines(width)...)
	if m.err != nil {
		lines = append(lines, "", m.theme.Danger.Render(m.err.Error()))
	}
	help := "↑/↓ · y/n · enter · esc back"
	return page.Panel(m.theme, "Git repository", lines, width, 16, true), help
}

func (m Model) viewAffirmDeploy(width int) (string, string) {
	lines := []string{
		m.theme.Muted.Render("Are you ready to set up your initial management cluster?"),
		m.theme.Muted.Render("You can check the generated terraform/helm to confirm everything looks good first."),
		m.theme.Muted.Render("Same Affirm as plural up (PLURAL_UP_AFFIRM_DEPLOY)."),
		"",
	}
	lines = append(lines, m.affirmDeployLines(width)...)
	if m.err != nil {
		lines = append(lines, "", m.theme.Danger.Render(m.err.Error()))
	}
	return page.Panel(m.theme, "Deploy", lines, width, 14, true), "↑/↓ · y/n · enter · esc domain"
}

func (m Model) viewIgnoreContinue(width int) (string, string) {
	lines := []string{
		m.theme.Muted.Render("Preflight checks failed, but continuing because --ignore-preflights was specified."),
		m.theme.Muted.Render("Please note that you may encounter issues later on during provisioning."),
		"",
		m.theme.Danger.Render("Warning: " + truncate(m.probeWarn, max(20, width-12))),
		"",
		"Mode         " + m.flow.Title,
		"Provider     " + m.provider.Title,
		"Preflights   " + yesNoLabel(true),
		"",
		m.theme.Muted.Render("Press enter to continue."),
		m.theme.Muted.Render("Next: Configure (self-hosted) → git Affirm → scm.Setup → domain → plan."),
	}
	return page.Panel(m.theme, "Continuing with warning", lines, width, 14, true), "enter continue · esc back"
}

func (m Model) viewAlreadyInit(width int) (string, string) {
	lines := []string{
		m.theme.Success.Render("Found workspace.yaml, skipping init as this repo has already been initialized"),
		"",
		m.theme.Muted.Render("Same path as plural up when workspace.yaml is present."),
		m.theme.Muted.Render("Next: ensure domain / branch → app domain → deploy Affirm → Generate."),
		"",
		"Mode         " + m.flow.Title,
		"Provider     " + m.provider.Title + " (" + m.provider.ID + ")",
	}
	if cluster := formValue(m.formValues, "cluster"); cluster != "" {
		lines = append(lines, "Cluster      "+cluster)
	}
	if m.pluralDNS != "" {
		lines = append(lines, "Plural DNS   "+m.pluralDNS)
	}
	if m.err != nil {
		lines = append(lines, "", m.theme.Danger.Render(m.err.Error()))
	}
	return page.Panel(m.theme, "Already initialized", lines, width, 14, true), "enter continue · esc back"
}

func (m Model) viewEnsuringInit(width int) (string, string) {
	lines := []string{
		m.spinner.View() + " " + m.theme.Muted.Render("Checking domain…"),
		m.theme.Muted.Render("ensureWorkspace — Plural DNS / branch / .gitignore"),
	}
	return page.Panel(m.theme, "Workspace check", lines, width, 10, true), "please wait"
}

func (m Model) viewBucketPrefix(width int) (string, string) {
	lines := []string{
		m.theme.Muted.Render(upbridge.BucketPrefixPrompt),
		m.theme.Muted.Render("Same as plural up Configure (workspace bucket naming)."),
		"",
		"› Prefix",
		"  " + m.formInput.View(),
	}
	if m.err != nil {
		lines = append(lines, "", m.theme.Danger.Render(m.err.Error()))
	}
	return page.Panel(m.theme, "Bucket naming", lines, width, 12, true), "enter · esc back"
}

func (m Model) viewPluralSubdomain(width int) (string, string) {
	lines := []string{
		m.theme.Muted.Render(upbridge.PluralSubdomainPrompt),
		m.theme.Muted.Render("Registers subdomain.onplural.sh (CreateDomain)."),
		"",
		"› Subdomain",
		"  " + m.formInput.View(),
	}
	if m.err != nil {
		lines = append(lines, "", m.theme.Danger.Render(m.err.Error()))
	}
	return page.Panel(m.theme, "Plural DNS", lines, width, 12, true), "enter · esc back"
}

func (m Model) viewRunPreflights(width int) (string, string) {
	lines := []string{
		"Provider  " + m.provider.Title,
		"",
		m.spinner.View() + " " + m.theme.Muted.Render("Running provider.Preflights()…"),
		m.theme.Muted.Render("IAM / permissions checks — skipped on failure if --ignore-preflights."),
	}
	return page.Panel(m.theme, "Preflight checks", lines, width, 10, true), "esc cancel"
}

func (m Model) viewSelectSCM(width int) (string, string) {
	lines := make([]string, 0, 3+len(m.scms))
	lines = append(lines,
		m.theme.Muted.Render("Select the SCM provider to use for your repository:"),
		m.theme.Muted.Render("Same first prompt as scm.Setup() in plural up."),
		"",
	)
	lines = append(lines, m.scmLines(width)...)
	help := "↑/↓ · 1–3 / letter · enter · esc git"
	return page.Panel(m.theme, "SCM provider", lines, width, 12, true), help
}

func (m Model) viewSCMSetup(width int) (string, string) {
	lines := []string{
		"SCM  " + m.scm.Title,
		"",
		m.spinner.View() + " " + m.theme.Muted.Render("Device login · create repo · clone…"),
		m.theme.Muted.Render("Terminal released for GitHub/GitLab/Bitbucket oauth (same as plural up)."),
		m.theme.Muted.Render("Follow the one-time code prompt in the terminal, then return here."),
	}
	if m.err != nil {
		lines = append(lines, "", m.theme.Danger.Render(m.err.Error()))
	}
	return page.Panel(m.theme, "SCM setup", lines, width, 12, true), "wait for browser / device flow"
}

func (m Model) viewAppDomain(width int) (string, string) {
	lines := []string{
		m.theme.Muted.Render("Application domain (askAppDomain)."),
		m.theme.Muted.Render("None / empty skips — same as plural up."),
		"",
	}
	if len(m.domainOpts) == 0 && m.err == nil {
		// still loading select options, or free-text mode after load
		if m.formInput.Focused() {
			lines = append(lines, "› Domain")
			lines = append(lines, "  "+m.formInput.View())
		} else {
			lines = append(lines, m.spinner.View()+" "+m.theme.Muted.Render("Fetching DNS zones…"))
		}
	} else if m.domainIsSelect() {
		lines = append(lines, "› Select hosted / DNS zone")
		start, end := optionWindow(m.optionCursor, len(m.domainOpts), 8)
		for j := start; j < end; j++ {
			mark := "  "
			style := m.theme.Muted
			if j == m.optionCursor {
				mark = "• "
				style = m.theme.Title
			}
			lines = append(lines, "  "+style.Render(mark+truncate(m.domainOpts[j], max(12, width-8))))
		}
	}
	if m.err != nil {
		lines = append(lines, "", m.theme.Danger.Render(m.err.Error()))
	}
	help := "enter confirm · esc back"
	if m.domainIsSelect() {
		help = "↑/↓ · enter · esc back"
	}
	return page.Panel(m.theme, "App domain", lines, width, 16, true), help
}

func (m Model) viewProbing(width int) (string, string) {
	lines := []string{
		"Provider  " + m.provider.Title,
		"",
		m.spinner.View() + " " + m.theme.Muted.Render("Checking credentials and fetching regions…"),
		m.theme.Muted.Render("Same checks plural up runs before the provider survey."),
	}
	return page.Panel(m.theme, "Provider setup", lines, width, 10, true), "esc cancel"
}

func (m Model) viewSelectProvider(width int) (string, string) {
	intro := make([]string, 0, 7+8)
	intro = append(intro,
		"Mode         "+m.flow.Title,
		"Preflights   "+yesNoLabel(m.ignorePreflights),
		m.theme.Muted.Render("             "+m.cli()),
		"",
		m.theme.Muted.Render("Select the cloud provider (same list as plural up init)."),
		m.theme.Muted.Render("Next: verify credentials · fetch regions/projects."),
		"",
	)
	intro = append(intro, m.providerLines(width)...)
	if m.err != nil {
		intro = append(intro, "", m.theme.Danger.Render(m.err.Error()))
		intro = append(intro, m.theme.Muted.Render("Fix credentials, or choose Ignore to warn and continue without the region survey."))
	}
	help := "↑/↓ select · 1–4 / letter · enter · esc preflights"
	if width < 100 {
		help = "↑/↓ · enter · esc preflights"
	}
	return page.Panel(m.theme, "Cloud provider", intro, width, 16, true), help
}

func (m Model) viewIgnorePreflights(width int) (string, string) {
	intro := make([]string, 0, 7+4)
	intro = append(intro,
		"Mode  "+m.flow.Title,
		m.theme.Muted.Render("      "+m.flow.CLI(false)),
		"",
		m.theme.Muted.Render("After provider setup, run provider.Preflights() (IAM, permissions, …)?"),
		m.theme.Muted.Render("Ignore = warn and continue — same as plural up --ignore-preflights."),
		m.theme.Muted.Render("Credential login + region survey still run first (CLI GetProvider)."),
		"",
	)
	intro = append(intro, m.ignoreLines(width)...)
	help := "↑/↓ select · 1/r run · 2/i ignore · enter · esc mode"
	if width < 100 {
		help = "↑/↓ · enter · esc mode"
	}
	return page.Panel(m.theme, "Preflight checks", intro, width, 14, true), help
}

func (m Model) viewSelectFlow(width int) (string, string) {
	intro := make([]string, 0, 3+len(m.flows))
	intro = append(intro,
		m.theme.Muted.Render("Sets up your repository and an initial management cluster."),
		m.theme.Muted.Render("Self-hosted and dry-run run the provider survey; cloud paths pick a Console first."),
		"",
	)
	intro = append(intro, m.flowLines(width)...)
	help := "↑/↓ select · 1–4 / letter · enter · esc welcome"
	if width < 100 {
		help = "↑/↓ · enter · esc welcome"
	}
	return page.Panel(m.theme, "Setup mode", intro, width, 14, true), help
}

func (m Model) formView(width int) (string, string) {
	lines := []string{
		"Provider  " + m.provider.Title,
	}
	if m.credSummary != "" {
		lines = append(lines, m.theme.Muted.Render("          "+truncate(m.credSummary, max(24, width-12))))
	} else {
		lines = append(lines, m.theme.Muted.Render("          Matches plural up / provider init prompts."))
	}
	if m.probeWarn != "" {
		lines = append(lines, m.theme.Danger.Render("⚠ "+truncate(m.probeWarn, max(24, width-4))))
	}
	lines = append(lines, "")

	for i, field := range m.formFields {
		cursor := "  "
		active := i == m.formIndex
		if active {
			cursor = "› "
		}
		if active && m.currentIsSelect() {
			lines = append(lines, cursor+field.Label)
			opts := field.Options
			start, end := optionWindow(m.optionCursor, len(opts), 6)
			for j := start; j < end; j++ {
				mark := "  "
				style := m.theme.Muted
				if j == m.optionCursor {
					mark = "• "
					style = m.theme.Title
				}
				lines = append(lines, "  "+style.Render(mark+truncate(opts[j], max(12, width-8))))
			}
			if start > 0 || end < len(opts) {
				lines = append(lines, m.theme.Muted.Render(fmt.Sprintf("  (%d/%d)", m.optionCursor+1, len(opts))))
			}
			continue
		}
		if active {
			lines = append(lines, cursor+field.Label)
			lines = append(lines, "  "+m.formInput.View())
			continue
		}
		val := formValue(m.formValues, field.Key)
		lines = append(lines, cursor+field.Label+"  "+m.theme.Muted.Render(truncate(val, max(8, width-24))))
	}
	if m.err != nil {
		lines = append(lines, "", m.theme.Danger.Render(m.err.Error()))
	}
	help := "↑/↓ · enter next/done · esc providers"
	if m.currentIsSelect() {
		help = "↑/↓ options · enter select · esc providers"
	}
	return page.Panel(m.theme, "Provider setup", lines, width, 18, true), help
}

func optionWindow(cursor, total, size int) (int, int) {
	if total <= size {
		return 0, total
	}
	start := cursor - size/2
	if start < 0 {
		start = 0
	}
	end := start + size
	if end > total {
		end = total
		start = end - size
	}
	return start, end
}

func (m Model) flowLines(width int) []string {
	lines := make([]string, 0, len(m.flows))
	for i, f := range m.flows {
		cursor := "  "
		if i == m.cursor {
			cursor = "› "
		}
		left := fmt.Sprintf("%d  %s   %-14s  %s", i+1, flowShortcut(f.ID), f.Title, f.Blurb)
		var row string
		if i == m.cursor {
			row = cursor + m.theme.Title.Render(left)
		} else {
			row = cursor + m.theme.Body.Render(left)
		}
		lines = append(lines, ansi.Truncate(row, max(1, width-2), "…"))
	}
	return lines
}

func (m Model) ignoreLines(width int) []string {
	opts := ignorePreflightOptions()
	lines := make([]string, 0, len(opts))
	for i, opt := range opts {
		cursor := "  "
		if i == m.cursor {
			cursor = "› "
		}
		check := "[ ]"
		if i == m.cursor {
			check = "[x]"
		}
		shortcut := "r"
		if opt.value {
			shortcut = "i"
		}
		left := fmt.Sprintf("%s %s   %-10s  %s", check, shortcut, opt.title, opt.blurb)
		var row string
		if i == m.cursor {
			row = cursor + m.theme.Title.Render(left)
		} else {
			row = cursor + m.theme.Body.Render(left)
		}
		lines = append(lines, ansi.Truncate(row, max(1, width-2), "…"))
	}
	return lines
}

func (m Model) setupGitLines(width int) []string {
	opts := m.gitAffirmOptions()
	lines := make([]string, 0, len(opts))
	for i, opt := range opts {
		cursor := "  "
		if i == m.cursor {
			cursor = "› "
		}
		check := "[ ]"
		if i == m.cursor {
			check = "[x]"
		}
		shortcut := "y"
		if !opt.value {
			shortcut = "n"
		}
		left := fmt.Sprintf("%s %s   %-4s  %s", check, shortcut, opt.title, opt.blurb)
		var row string
		if i == m.cursor {
			row = cursor + m.theme.Title.Render(left)
		} else {
			row = cursor + m.theme.Body.Render(left)
		}
		lines = append(lines, ansi.Truncate(row, max(1, width-2), "…"))
	}
	return lines
}

func (m Model) consoleCredLines(width int) []string {
	priorURL, _ := m.readPriorConsole()
	opts := consoleCredOptions(priorURL)
	lines := make([]string, 0, len(opts))
	for i, opt := range opts {
		cursor := "  "
		if i == m.cursor {
			cursor = "› "
		}
		check := "[ ]"
		if i == m.cursor {
			check = "[x]"
		}
		shortcut := "y"
		if !opt.value {
			shortcut = "n"
		}
		left := fmt.Sprintf("%s %s   %-4s  %s", check, shortcut, opt.title, opt.blurb)
		var row string
		if i == m.cursor {
			row = cursor + m.theme.Title.Render(left)
		} else {
			row = cursor + m.theme.Body.Render(left)
		}
		lines = append(lines, ansi.Truncate(row, max(1, width-2), "…"))
	}
	return lines
}

func (m Model) instanceLines(width int) []string {
	lines := make([]string, 0, len(m.instances))
	start, end := optionWindow(m.cursor, len(m.instances), 8)
	for i := start; i < end; i++ {
		inst := m.instances[i]
		cursor := "  "
		if i == m.cursor {
			cursor = "› "
		}
		left := fmt.Sprintf("%d   %-20s  %s", i+1, inst.Name, truncate(inst.URL, max(12, width-28)))
		var row string
		if i == m.cursor {
			row = cursor + m.theme.Title.Render(left)
		} else {
			row = cursor + m.theme.Body.Render(left)
		}
		lines = append(lines, ansi.Truncate(row, max(1, width-2), "…"))
	}
	return lines
}

func (m Model) affirmDeployLines(width int) []string {
	opts := affirmDeployOptions()
	lines := make([]string, 0, len(opts))
	for i, opt := range opts {
		cursor := "  "
		if i == m.cursor {
			cursor = "› "
		}
		check := "[ ]"
		if i == m.cursor {
			check = "[x]"
		}
		shortcut := "y"
		if !opt.value {
			shortcut = "n"
		}
		left := fmt.Sprintf("%s %s   %-4s  %s", check, shortcut, opt.title, opt.blurb)
		var row string
		if i == m.cursor {
			row = cursor + m.theme.Title.Render(left)
		} else {
			row = cursor + m.theme.Body.Render(left)
		}
		lines = append(lines, ansi.Truncate(row, max(1, width-2), "…"))
	}
	return lines
}

func (m Model) scmLines(width int) []string {
	lines := make([]string, 0, len(m.scms))
	for i, s := range m.scms {
		cursor := "  "
		if i == m.cursor {
			cursor = "› "
		}
		left := fmt.Sprintf("%d  %s   %-10s  %s", i+1, scmShortcut(s.ID), s.Title, s.Blurb)
		var row string
		if i == m.cursor {
			row = cursor + m.theme.Title.Render(left)
		} else {
			row = cursor + m.theme.Body.Render(left)
		}
		lines = append(lines, ansi.Truncate(row, max(1, width-2), "…"))
	}
	return lines
}

func (m Model) providerLines(width int) []string {
	lines := make([]string, 0, len(m.providers))
	for i, p := range m.providers {
		cursor := "  "
		if i == m.cursor {
			cursor = "› "
		}
		left := fmt.Sprintf("%d  %s   %-6s  %s", i+1, providerShortcut(p.ID), p.Title, p.Blurb)
		var row string
		if i == m.cursor {
			row = cursor + m.theme.Title.Render(left)
		} else {
			row = cursor + m.theme.Body.Render(left)
		}
		lines = append(lines, ansi.Truncate(row, max(1, width-2), "…"))
	}
	return lines
}

func (m Model) opLogLines(limit, width int) []string {
	if limit <= 0 {
		limit = 12
	}
	if len(m.opLog) == 0 {
		return []string{m.theme.Muted.Render("Waiting for output…")}
	}
	wrapped := wrapOpLog(m.opLog, width)
	start := m.opLogStart(limit, len(wrapped))
	end := start + limit
	if end > len(wrapped) {
		end = len(wrapped)
	}
	out := make([]string, 0, end-start)
	for _, line := range wrapped[start:end] {
		out = append(out, m.theme.Muted.Render(line))
	}
	return out
}

// opLogInnerWidth is the text width inside a page.Panel (borders + padding).
func opLogInnerWidth(panelWidth int) int {
	return max(20, panelWidth-4)
}

func wrapOpLog(lines []string, width int) []string {
	maxW := opLogInnerWidth(width)
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			out = append(out, "")
			continue
		}
		out = append(out, strings.Split(ansi.Wrap(line, maxW, ""), "\n")...)
	}
	return out
}

func opLogContentWidth(termWidth int) int {
	if termWidth <= 0 {
		termWidth = page.DefaultWidth
	}
	return page.ContentWidth(termWidth)
}

func (m Model) opLogStart(limit, total int) int {
	if total == 0 {
		return 0
	}
	maxStart := max(0, total-limit)
	if m.opLogFollow {
		return maxStart
	}
	if m.opLogY < 0 {
		return 0
	}
	if m.opLogY > maxStart {
		return maxStart
	}
	return m.opLogY
}

func (m Model) opLogScrollHint(width int) string {
	if len(m.opLog) == 0 {
		return m.theme.Muted.Render("No log lines captured.")
	}
	limit := 12
	if m.viewH > 0 {
		_, limit = logPanelBudget(m.viewH, 5)
	}
	wrapped := wrapOpLog(m.opLog, width)
	start := m.opLogStart(limit, len(wrapped))
	end := min(len(wrapped), start+limit)
	label := fmt.Sprintf("Logs  %d–%d / %d", start+1, end, len(wrapped))
	if m.opLogFollow {
		label += "  · following"
	}
	return m.theme.Muted.Render(ansi.Truncate(label, max(1, width-4), "…"))
}

func (m Model) opLogExportHint(width int) string {
	if m.logExportErr != nil {
		return m.theme.Danger.Render(ansi.Truncate("Could not save logs: "+m.logExportErr.Error()+" · e retry", max(1, width-4), "…"))
	}
	if m.logExportPath != "" {
		return m.theme.Muted.Render(ansi.Truncate("Saved "+m.logExportPath+" · e to save again", max(1, width-4), "…"))
	}
	return m.theme.Muted.Render("e exports full logs to a file (ctrl+c quits the TUI)")
}

// logPanelBudget sizes the streaming log panel to fill most of the terminal.
// chrome is the number of intro lines above the log (excluding panel borders).
func logPanelBudget(termHeight, chrome int) (panelHeight, logLines int) {
	// header (2) + blank (1) + help (1) + min separation (2)
	panelHeight = termHeight - 6
	if panelHeight < 18 {
		panelHeight = 18
	}
	if chrome < 0 {
		chrome = 0
	}
	logLines = panelHeight - 2 - chrome
	if logLines < 12 {
		logLines = 12
	}
	return panelHeight, logLines
}
