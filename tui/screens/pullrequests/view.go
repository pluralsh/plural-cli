package pullrequests

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/pluralsh/plural-cli/tui/components/page"
)

func (m Model) View(width, height int) string {
	width, height = page.Size(width, height)
	if width < page.MinimumWidth || height < page.MinimumHeight {
		return page.Unsupported(m.theme, width, height)
	}
	contentWidth := page.ContentWidth(width)
	title := "Pull requests"
	if m.mode == modeDetail && m.detail.Name != "" {
		title = "Pull requests · " + m.detail.Name
	}
	if m.mode == modeReview || m.mode == modeOperating || m.mode == modeResult {
		title = m.pending.title
		if title == "" {
			title = "Pull requests"
		}
	}
	body, help := m.bodyAndHelp(contentWidth)
	return page.Render(m.theme, width, height, title, m.headerStatus(), body, help)
}

func (m Model) headerStatus() string {
	if m.loading || m.mode == modeOperating {
		return m.theme.Warning.Render("◌ loading")
	}
	if m.needsAuth {
		return m.theme.Warning.Render("○ connect Console")
	}
	if m.err != nil && m.mode != modeCreateForm && m.mode != modeTriggerForm {
		return m.theme.Danger.Render("✗ load failed")
	}
	switch m.mode {
	case modeDetail:
		return m.theme.Success.Render(loCoalesce(m.detail.Addon, "automation"))
	case modeList:
		if m.filter != "" {
			return m.theme.Muted.Render(fmt.Sprintf("%d matching", len(m.page.Items)))
		}
		return m.theme.Success.Render(fmt.Sprintf("%d automations", len(m.page.Items)))
	case modeReview:
		return m.theme.Warning.Render("review")
	case modeResult:
		if m.result == "failed" {
			return m.theme.Danger.Render("failed")
		}
		return m.theme.Success.Render("done")
	default:
		return m.theme.Muted.Render("pull requests")
	}
}

func (m Model) bodyAndHelp(width int) (string, string) {
	if m.mode == modeFilter {
		lines := []string{
			m.theme.Muted.Render("Filter by name, title, addon, identifier, or id."),
			"",
			m.filterInput.View(),
		}
		return page.Panel(m.theme, "Filter PR automations", lines, width, 6, true), "enter apply · esc cancel"
	}
	if m.needsAuth {
		lines := []string{
			m.theme.Warning.Render("○ Console is not connected"),
			m.theme.Muted.Render("  Connect a Console profile to browse PR automations."),
			"",
			m.theme.Body.Render("Press c to open Access."),
		}
		return page.Panel(m.theme, "Console required", lines, width, 8, true), "c connect · esc back · ctrl+c quit"
	}
	switch m.mode {
	case modeReview:
		lines := append([]string{}, m.pending.lines...)
		lines = append(lines, "", m.theme.Muted.Render("Equivalent CLI"), "  "+m.pending.cli)
		return page.Panel(m.theme, "Plan (immutable)", lines, width, 12, true), "enter confirm · esc back"
	case modeOperating:
		lines := []string{m.theme.Warning.Render("● Running…"), ""}
		lines = append(lines, m.opLog...)
		return page.Panel(m.theme, "Operation", lines, width, 10, true), "ctrl+c quit"
	case modeResult:
		head := m.theme.Success.Render("✓ Success")
		help := "enter detail · esc detail"
		if m.result == "failed" {
			head = m.theme.Danger.Render("✗ Failed")
			help = "enter retry review · esc detail"
		} else if m.pending.kind == actionTemplate || m.pending.kind == actionTest || m.pending.kind == actionContracts {
			head = m.theme.Muted.Render("CLI equivalent")
			help = "esc detail"
		}
		lines := []string{head, ""}
		lines = append(lines, m.opLog...)
		return page.Panel(m.theme, "Result", lines, width, 12, true), help
	case modeCreateForm, modeTriggerForm:
		return m.formView(width)
	case modeCLITip:
		kind := "Template"
		switch m.cliKind {
		case actionTest:
			kind = "Test"
		case actionContracts:
			kind = "Contracts"
		}
		lines := []string{
			m.theme.Muted.Render(kind + " runs locally via the Plural CLI."),
			"",
			"File       " + m.formInput.View(),
			"",
			m.theme.Muted.Render("Enter shows the CLI equivalent."),
		}
		return page.Panel(m.theme, kind+" · CLI", lines, width, 8, true), "enter · esc detail"
	case modeDetail:
		summary := page.Panel(m.theme, "Summary", m.detailLines(), width, 9, false)
		actions := page.Panel(m.theme, "Actions", m.actionLines(width), width, 8, true)
		help := "↑/↓ actions · enter · c t m e o · r refresh · esc list"
		if width < 100 {
			help = "↑/↓ · enter · letters · r · esc"
		}
		return summary + "\n\n" + actions, help
	default:
		help := "↑/↓ select · enter open · / filter · n/p page · r refresh · esc back"
		if width < 100 {
			help = "↑/↓ · enter · / · n/p page · esc back"
		}
		return page.Panel(m.theme, m.listTitle(), m.listLines(width), width, 14, true), help
	}
}

func (m Model) formView(width int) (string, string) {
	title := "Create pull request"
	if m.mode == modeTriggerForm {
		title = "Trigger PR automation"
	}
	lines := []string{
		"Automation  " + m.detail.Name,
		"",
	}
	for i, field := range m.formFields {
		cursor := "  "
		if i == m.formIndex {
			cursor = "› "
			lines = append(lines, cursor+field.label)
			lines = append(lines, "  "+m.formInput.View())
			continue
		}
		val := loCoalesce(m.formValues[field.key], "—")
		lines = append(lines, cursor+field.label+"  "+m.theme.Muted.Render(truncate(val, max(8, width-20))))
	}
	if m.err != nil {
		lines = append(lines, "", m.theme.Danger.Render(m.err.Error()))
	}
	return page.Panel(m.theme, title, lines, width, 12, true), "↑/↓ fields · enter next/review · esc cancel"
}

func (m Model) listTitle() string {
	if m.filter != "" {
		return "PR automations · filter “" + m.filter + "”"
	}
	return "PR automations"
}

func (m Model) listLines(width int) []string {
	if m.loading && len(m.page.Items) == 0 {
		return []string{m.theme.Warning.Render("◌ Loading PR automations…")}
	}
	if m.err != nil {
		return []string{
			m.theme.Danger.Render("✗ Unable to load PR automations"),
			m.theme.Danger.Render("Error  " + m.err.Error()),
			m.theme.Muted.Render("Press r to retry."),
		}
	}
	if len(m.page.Items) == 0 {
		return []string{
			m.theme.Warning.Render("○ No PR automations found"),
			m.theme.Muted.Render("  Adjust the filter or connect another Console."),
		}
	}
	nameWidth := max(12, min(22, width/3))
	addonWidth := max(6, min(12, width/6))
	lines := []string{m.theme.Muted.Render("  " + pad("NAME", nameWidth) + " " + pad("ADDON", addonWidth) + " TITLE")}
	start, end := visibleWindow(m.cursor, len(m.page.Items), 8)
	for i := start; i < end; i++ {
		item := m.page.Items[i]
		cursor := "  "
		if i == m.cursor {
			cursor = "› "
		}
		row := cursor + pad(item.Name, nameWidth) + " " + pad(loCoalesce(item.Addon, "—"), addonWidth) + " " + loCoalesce(item.Title, "—")
		lines = append(lines, ansi.Truncate(row, width-2, "…"))
	}
	if start > 0 || end < len(m.page.Items) {
		lines = append(lines, m.theme.Muted.Render(fmt.Sprintf("  … %d–%d of %d", start+1, end, len(m.page.Items))))
	}
	if m.page.HasNext || len(m.prevCursors) > 0 {
		pager := "page"
		if len(m.prevCursors) > 0 {
			pager += " · p prev"
		}
		if m.page.HasNext {
			pager += " · n next"
		}
		lines = append(lines, "", m.theme.Muted.Render(pager))
	}
	return lines
}

func (m Model) actionLines(width int) []string {
	actions := detailActions()
	lines := make([]string, 0, len(actions))
	for i, a := range actions {
		cursor := "  "
		if i == m.actionCursor {
			cursor = "› "
		}
		suffix := ""
		if a.cliOnly {
			suffix = "  " + m.theme.Muted.Render("CLI")
		}
		row := cursor + a.shortcut + "  " + pad(a.title, 12) + " " + a.blurb + suffix
		lines = append(lines, ansi.Truncate(row, max(1, width-2), "…"))
	}
	return lines
}

func visibleWindow(cursor, count, size int) (start, end int) {
	if count <= 0 {
		return 0, 0
	}
	if size <= 0 {
		size = count
	}
	if count <= size {
		return 0, count
	}
	start = cursor - size/2
	if start < 0 {
		start = 0
	}
	end = start + size
	if end > count {
		end = count
		start = end - size
	}
	return start, end
}

func (m Model) detailLines() []string {
	if m.loading {
		return []string{m.theme.Warning.Render("◌ Loading PR automation detail…")}
	}
	if m.err != nil {
		return []string{m.theme.Danger.Render("✗ Unable to load PR automation"), m.theme.Danger.Render(m.err.Error())}
	}
	lines := []string{
		m.labelValue("Name", m.detail.Name),
		m.labelValue("Title", loCoalesce(m.detail.Title, "—")),
		m.labelValue("Addon", loCoalesce(m.detail.Addon, "—")),
		m.labelValue("Identifier", loCoalesce(m.detail.Identifier, "—")),
		m.labelValue("ID", m.detail.ID),
	}
	if m.detail.Message != "" {
		lines = append(lines, m.labelValue("Message", m.detail.Message))
	}
	return lines
}

func (m Model) labelValue(label, value string) string {
	label = label + strings.Repeat(" ", max(1, 12-len(label)))
	return label + " " + value
}

func pad(value string, width int) string {
	value = ansi.Truncate(value, width, "…")
	if lipgloss.Width(value) >= width {
		return value
	}
	return value + strings.Repeat(" ", width-lipgloss.Width(value))
}

func loCoalesce(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
