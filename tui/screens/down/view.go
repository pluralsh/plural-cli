package down

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/pluralsh/plural-cli/tui/components/page"
)

func (m Model) View(width, height int) string {
	width, height = page.Size(width, height)
	if width < page.MinimumWidth || height < page.MinimumHeight {
		return page.Unsupported(m.theme, width, height)
	}
	contentWidth := page.ContentWidth(width)
	body, help := m.bodyAndHelp(contentWidth, height)
	return page.Render(m.theme, width, height, "Down", m.headerStatus(), body, help)
}

func (m Model) headerStatus() string {
	switch m.mode {
	case modeAffirm:
		return m.theme.Muted.Render("step · destroy Affirm")
	case modeDestroying:
		return m.theme.Muted.Render("destroying…")
	case modeComplete:
		if m.err != nil {
			return m.theme.Danger.Render("failed")
		}
		return m.theme.Success.Render("destroyed")
	default:
		return m.theme.Muted.Render("step 1 · mode")
	}
}

func (m Model) bodyAndHelp(width, height int) (string, string) {
	switch m.mode {
	case modeAffirm:
		lines := []string{
			m.theme.Muted.Render(AffirmMessage()),
			m.theme.Muted.Render("Same Affirm as plural down (PLURAL_DOWN_AFFIRM_DESTROY)."),
			"",
			"Mode  " + m.modeLabel(),
			m.theme.Muted.Render("      " + m.cli()),
			"",
		}
		lines = append(lines, m.affirmLines(width)...)
		if m.err != nil {
			lines = append(lines, "", m.theme.Danger.Render(m.err.Error()))
		}
		return page.Panel(m.theme, "Destroy", lines, width, 14, true), "↑/↓ · y/n · enter · esc back"
	case modeDestroying:
		panelH, logN := logPanelBudget(height, 5)
		lines := make([]string, 0, 5+logN)
		lines = append(lines,
			"Mode  "+m.modeLabel(),
			"",
			m.spin.View()+" "+m.theme.Muted.Render("Destroying management cluster terraform…"),
			m.theme.Muted.Render("Terraform output streams below (TUI stays open)."),
			"",
		)
		lines = append(lines, m.opLogLines(logN, width)...)
		return page.Panel(m.theme, "Destroying", lines, width, panelH, true), "↑/↓ · pgup/pgdn scroll · end follow"
	case modeComplete:
		panelH, logN := logPanelBudget(height, 6)
		lines := []string{}
		if m.err != nil {
			lines = append(lines,
				m.theme.Danger.Render("Destroy failed — scroll logs below, then enter/esc welcome"),
				m.theme.Danger.Render(m.err.Error()),
				"",
			)
		} else {
			lines = append(lines,
				m.theme.Success.Render("✓ Management cluster destroy finished"),
				m.theme.Muted.Render("Scroll logs below · enter/esc welcome"),
				"",
			)
		}
		lines = append(lines, m.opLogExportHint(width))
		lines = append(lines, m.opLogScrollHint(width))
		lines = append(lines, m.opLogLines(logN, width)...)
		return page.Panel(m.theme, "Destroy complete", lines, width, panelH, true), "↑/↓ scroll · e export logs · enter/esc welcome"
	default:
		intro := make([]string, 0, 4+len(cloudOptions()))
		intro = append(intro,
			m.theme.Muted.Render("Destroys your management cluster and any apps installed on it."),
			m.theme.Muted.Render("Same as plural down — requires workspace.yaml in the current repo."),
			"",
		)
		intro = append(intro, m.cloudLines(width)...)
		if m.err != nil {
			intro = append(intro, "", m.theme.Danger.Render(m.err.Error()))
		}
		help := "↑/↓ · 1–2 / s/c · enter · esc welcome"
		if width < 100 {
			help = "↑/↓ · enter · esc welcome"
		}
		return page.Panel(m.theme, "Destroy mode", intro, width, 12, true), help
	}
}

func (m Model) cloudLines(width int) []string {
	opts := cloudOptions()
	lines := make([]string, 0, len(opts))
	for i, o := range opts {
		prefix := "  "
		label := fmt.Sprintf("%d  %s   %-14s  %s", i+1, cloudShortcut(o.id), o.title, o.blurb)
		if i == m.cursor {
			prefix = "› "
			label = m.theme.Title.Render(label)
		} else {
			label = m.theme.Body.Render(fmt.Sprintf("%d  %s   ", i+1, cloudShortcut(o.id))) +
				m.theme.Body.Render(fmt.Sprintf("%-14s  ", o.title)) +
				m.theme.Muted.Render(o.blurb)
		}
		line := prefix + label
		lines = append(lines, ansi.Truncate(line, max(1, width-4), "…"))
	}
	return lines
}

func (m Model) affirmLines(width int) []string {
	opts := affirmOptions()
	lines := make([]string, 0, len(opts))
	for i, o := range opts {
		prefix := "  "
		label := m.theme.Body.Render(o.title) + "  " + m.theme.Muted.Render(o.blurb)
		if i == m.cursor {
			prefix = "› "
			label = m.theme.Title.Render(o.title) + "  " + m.theme.Muted.Render(o.blurb)
		}
		lines = append(lines, ansi.Truncate(prefix+label, max(1, width-4), "…"))
	}
	return lines
}

func (m Model) modeLabel() string {
	if m.cloud {
		return "Plural Cloud"
	}
	return "Self-hosted"
}

func (m Model) cli() string {
	if m.cloud {
		return "plural down --cloud"
	}
	return "plural down"
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
	end := start + limit
	if end > len(wrapped) {
		end = len(wrapped)
	}
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
func logPanelBudget(termHeight, chrome int) (panelHeight, logLines int) {
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
