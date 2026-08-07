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
	body, help := m.bodyAndHelp(contentWidth)
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
		if m.consoleTokenMode {
			return m.theme.Muted.Render("step · console token")
		}
		return m.theme.Muted.Render("step · console credentials")
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
	case modeSetupGit:
		return m.theme.Muted.Render("step · git repository")
	case modeSelectSCM:
		return m.theme.Muted.Render("step · scm provider")
	case modeAppDomain:
		return m.theme.Muted.Render("step · app domain")
	case modeAffirmDeploy:
		return m.theme.Muted.Render("step · deploy Affirm")
	case modeSelected:
		if m.cloudInstance.Name != "" {
			return m.theme.Success.Render(m.cloudInstance.Name)
		}
		if m.provider.ID != "" {
			return m.theme.Success.Render(m.provider.ID)
		}
		return m.theme.Success.Render("ready")
	case modeRunning:
		return m.theme.Muted.Render("running Flush + Generate…")
	case modeDone:
		if m.runErr != nil {
			return m.theme.Danger.Render("failed")
		}
		return m.theme.Success.Render("generated")
	case modeCommitMsg:
		return m.theme.Muted.Render("step · commit message")
	case modeDeploying:
		return m.theme.Muted.Render("deploying…")
	case modeComplete:
		if m.deployErr != nil {
			return m.theme.Danger.Render("deploy failed")
		}
		return m.theme.Success.Render("deployed")
	case modeCLITip:
		return m.theme.Muted.Render(m.flow.ID)
	default:
		return m.theme.Muted.Render("step 1 · mode")
	}
}

func (m Model) bodyAndHelp(width int) (string, string) {
	switch m.mode {
	case modeSelected:
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
		if m.scm.ID != "" {
			lines = append(lines, "SCM          "+m.scm.Title+" (scm.Setup)")
		} else if m.inGitRepo {
			lines = append(lines, "Git          already inside a work tree")
		}
		domain := m.appDomain
		if domain == "" {
			domain = "(skipped)"
		}
		if !m.flow.DryRun {
			lines = append(lines, "App domain   "+domain)
		}
		next := "Enter to Flush workspace.yaml + Generate, then Deploy."
		if m.flow.Cloud {
			next = "Enter to Flush + ImportCluster + Generate, then Deploy."
		}
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
	case modeRunning:
		lines := []string{
			m.spinner.View() + " " + m.theme.Muted.Render("Running Flush + Generate…"),
			m.theme.Muted.Render("Next: commit message → Deploy."),
			"",
		}
		if len(m.runSteps) == 0 {
			lines = append(lines, m.theme.Muted.Render("Starting…"))
		} else {
			for _, s := range m.runSteps {
				lines = append(lines, "  · "+s)
			}
		}
		return page.Panel(m.theme, "Generating", lines, width, 12, true), "please wait"
	case modeDone:
		lines := []string{}
		if m.runErr != nil {
			lines = append(lines,
				m.theme.Danger.Render("Generate failed"),
				"",
				m.theme.Danger.Render(m.runErr.Error()),
				"",
				m.theme.Muted.Render("Esc returns to Plan to retry."),
			)
			return page.Panel(m.theme, "Done", lines, width, 14, true), "esc plan · ctrl+c quit"
		}
		lines = append(lines,
			m.theme.Success.Render("✓ Finished generating the repo"),
			"",
		)
		if m.flow.Cloud {
			lines = append(lines, "Console  "+m.cloudInstance.Name)
		}
		if m.provider.Title != "" {
			lines = append(lines, "Provider "+m.provider.Title)
		}
		for _, s := range m.runSteps {
			lines = append(lines, m.theme.Muted.Render("  · "+s))
		}
		if m.flow.DryRun {
			lines = append(lines, "",
				m.theme.Muted.Render("Dry-run: no Deploy will run."),
			)
			return page.Panel(m.theme, "Done", lines, width, 14, true), "esc plan · ctrl+c quit"
		}
		lines = append(lines, "",
			m.theme.Muted.Render("Enter to continue to Deploy (commit message, then terraform)."),
			m.theme.Muted.Render("Equivalent CLI: "+m.cli()),
		)
		return page.Panel(m.theme, "Generated", lines, width, 16, true), "enter deploy · esc plan · ctrl+c quit"
	case modeCommitMsg:
		lines := []string{
			m.theme.Muted.Render("Enter a commit message to push your configuration."),
			m.theme.Muted.Render("Leave empty to skip commit/push (same as plural up CommitMsg)."),
			"",
			"› Message",
			"  " + m.formInput.View(),
		}
		if m.err != nil {
			lines = append(lines, "", m.theme.Danger.Render(m.err.Error()))
		}
		return page.Panel(m.theme, "Git commit", lines, width, 12, true), "enter deploy · esc back"
	case modeDeploying:
		lines := []string{
			m.spinner.View() + " " + m.theme.Muted.Render("Running Deploy (terraform / import / apps)…"),
			"",
		}
		if len(m.runSteps) == 0 {
			lines = append(lines, m.theme.Muted.Render("Starting…"))
		} else {
			for _, s := range m.runSteps {
				lines = append(lines, "  · "+s)
			}
		}
		return page.Panel(m.theme, "Deploying", lines, width, 12, true), "please wait"
	case modeComplete:
		lines := []string{}
		if m.deployErr != nil {
			lines = append(lines,
				m.theme.Danger.Render("Deploy failed"),
				"",
				m.theme.Danger.Render(m.deployErr.Error()),
				"",
				m.theme.Muted.Render("Esc returns to Generated to retry Deploy."),
			)
		} else {
			lines = append(lines,
				m.theme.Success.Render("✓ Finished setting up your management cluster!"),
				"",
			)
			if m.flow.Cloud {
				lines = append(lines, "Console  "+m.cloudInstance.Name)
			}
			if m.provider.Title != "" {
				lines = append(lines, "Provider "+m.provider.Title)
			}
			if m.commitMsg != "" {
				lines = append(lines, "Commit   "+truncate(m.commitMsg, max(20, width-12)))
			} else {
				lines = append(lines, "Commit   (skipped)")
			}
			for _, s := range m.runSteps {
				lines = append(lines, m.theme.Muted.Render("  · "+s))
			}
			if m.provider.ID == "byok" && m.flow.Cloud {
				lines = append(lines, "",
					m.theme.Muted.Render("BYOK cloud: configure IAM for plrl-deploy-operator/stacks"),
					m.theme.Muted.Render("(operator reinstall not run from TUI yet)."),
				)
			} else {
				lines = append(lines, "",
					m.theme.Muted.Render("Use terraform as usual; gitops lives under bootstrap/."),
				)
			}
			lines = append(lines, "",
				m.theme.Muted.Render("Equivalent CLI"),
				"  "+m.cli(),
			)
		}
		return page.Panel(m.theme, "Complete", lines, width, 16, true), "esc back · ctrl+c quit"
	case modeCLITip:
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
	case modeLoadInstances:
		lines := []string{
			"Mode  " + m.flow.Title,
			"",
			m.spinner.View() + " " + m.theme.Muted.Render("Fetching Console instances (GetConsoleInstances)…"),
			m.theme.Muted.Render("Same list plural up --cloud uses in choseCluster."),
		}
		return page.Panel(m.theme, "Plural Cloud", lines, width, 10, true), "esc cancel"
	case modeSelectInstance:
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
	case modeConsoleLogin:
		if m.consoleTokenMode {
			lines := []string{
				"Instance  " + m.cloudInstance.Name,
				m.theme.Muted.Render("          "+truncate(m.cloudInstance.URL, max(20, width-12))),
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
			m.theme.Muted.Render("          "+truncate(m.cloudInstance.URL, max(20, width-12))),
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
	case modeSetupGit:
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
	case modeAffirmDeploy:
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
	case modeIgnoreContinue:
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
			m.theme.Muted.Render("Press enter to continue to the git Affirm (next step)."),
			m.theme.Muted.Render("Then: scm.Setup (if needed) → app domain → deploy Affirm → plan."),
		}
		return page.Panel(m.theme, "Continuing with warning", lines, width, 14, true), "enter continue · esc back"
	case modeRunPreflights:
		lines := []string{
			"Provider  " + m.provider.Title,
			"",
			m.spinner.View() + " " + m.theme.Muted.Render("Running provider.Preflights()…"),
			m.theme.Muted.Render("IAM / permissions checks — skipped on failure if --ignore-preflights."),
		}
		return page.Panel(m.theme, "Preflight checks", lines, width, 10, true), "esc cancel"
	case modeSelectSCM:
		lines := []string{
			m.theme.Muted.Render("Select the SCM provider to use for your repository:"),
			m.theme.Muted.Render("Same first prompt as scm.Setup() in plural up."),
			"",
		}
		lines = append(lines, m.scmLines(width)...)
		help := "↑/↓ · 1–3 / letter · enter · esc git"
		return page.Panel(m.theme, "SCM provider", lines, width, 12, true), help
	case modeAppDomain:
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
	case modeProbing:
		lines := []string{
			"Provider  " + m.provider.Title,
			"",
			m.spinner.View() + " " + m.theme.Muted.Render("Checking credentials and fetching regions…"),
			m.theme.Muted.Render("Same checks plural up runs before the provider survey."),
		}
		return page.Panel(m.theme, "Provider setup", lines, width, 10, true), "esc cancel"
	case modeProviderForm:
		return m.formView(width)
	case modeSelectProvider:
		intro := []string{
			"Mode         " + m.flow.Title,
			"Preflights   " + yesNoLabel(m.ignorePreflights),
			m.theme.Muted.Render("             " + m.cli()),
			"",
			m.theme.Muted.Render("Select the cloud provider (same list as plural up init)."),
			m.theme.Muted.Render("Next: verify credentials · fetch regions/projects."),
			"",
		}
		lines := append(intro, m.providerLines(width)...)
		if m.err != nil {
			lines = append(lines, "", m.theme.Danger.Render(m.err.Error()))
			lines = append(lines, m.theme.Muted.Render("Fix credentials, or choose Ignore to warn and continue without the region survey."))
		}
		help := "↑/↓ select · 1–4 / letter · enter · esc preflights"
		if width < 100 {
			help = "↑/↓ · enter · esc preflights"
		}
		return page.Panel(m.theme, "Cloud provider", lines, width, 16, true), help
	case modeIgnorePreflights:
		intro := []string{
			"Mode  " + m.flow.Title,
			m.theme.Muted.Render("      " + m.flow.CLI(false)),
			"",
			m.theme.Muted.Render("After provider setup, run provider.Preflights() (IAM, permissions, …)?"),
			m.theme.Muted.Render("Ignore = warn and continue — same as plural up --ignore-preflights."),
			m.theme.Muted.Render("Credential login + region survey still run first (CLI GetProvider)."),
			"",
		}
		lines := append(intro, m.ignoreLines(width)...)
		help := "↑/↓ select · 1/r run · 2/i ignore · enter · esc mode"
		if width < 100 {
			help = "↑/↓ · enter · esc mode"
		}
		return page.Panel(m.theme, "Preflight checks", lines, width, 14, true), help
	default:
		intro := []string{
			m.theme.Muted.Render("Sets up your repository and an initial management cluster."),
			m.theme.Muted.Render("Only self-hosted continues with the provider survey for now."),
			"",
		}
		lines := append(intro, m.flowLines(width)...)
		help := "↑/↓ select · 1–4 / letter · enter · esc welcome"
		if width < 100 {
			help = "↑/↓ · enter · esc welcome"
		}
		return page.Panel(m.theme, "Setup mode", lines, width, 14, true), help
	}
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
