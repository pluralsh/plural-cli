package providers

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
	title := "Providers"
	if m.mode == modeDetail && m.detail.Name != "" {
		title = "Providers · " + m.detail.Name
	}
	body, help := m.bodyAndHelp(contentWidth)
	return page.Render(m.theme, width, height, title, m.headerStatus(), body, help)
}

func (m Model) headerStatus() string {
	if m.loading {
		return m.theme.Warning.Render("◌ loading")
	}
	if m.needsAuth {
		return m.theme.Warning.Render("○ connect Console")
	}
	if m.err != nil {
		return m.theme.Danger.Render("✗ load failed")
	}
	switch m.mode {
	case modeDetail:
		return m.theme.Success.Render(loCoalesce(m.detail.Cloud, "provider"))
	case modeList:
		if m.filter != "" {
			return m.theme.Muted.Render(fmt.Sprintf("%d matching", len(m.page.Items)))
		}
		return m.theme.Success.Render(fmt.Sprintf("%d providers", len(m.page.Items)))
	default:
		return m.theme.Muted.Render("providers")
	}
}

func (m Model) bodyAndHelp(width int) (string, string) {
	if m.mode == modeFilter {
		lines := []string{
			m.theme.Muted.Render("Filter by name, cloud, namespace, or id."),
			"",
			m.filterInput.View(),
		}
		return page.Panel(m.theme, "Filter providers", lines, width, 6, true), "enter apply · esc cancel"
	}
	if m.needsAuth {
		lines := []string{
			m.theme.Warning.Render("○ Console is not connected"),
			m.theme.Muted.Render("  Connect a Console profile to browse providers."),
			"",
			m.theme.Body.Render("Press c to open Access."),
		}
		return page.Panel(m.theme, "Console required", lines, width, 8, true), "c connect · esc back · ctrl+c quit"
	}
	if m.mode == modeDetail {
		help := "r refresh · esc list · ctrl+c quit"
		return page.Panel(m.theme, "Summary", m.detailLines(), width, 16, true), help
	}
	help := "↑/↓ select · enter open · / filter · n/p page · r refresh · esc back"
	if width < 100 {
		help = "↑/↓ · enter · / · n/p page · esc back"
	}
	return page.Panel(m.theme, m.listTitle(), m.listLines(width), width, 14, true), help
}

func (m Model) listTitle() string {
	if m.filter != "" {
		return "Providers · filter “" + m.filter + "”"
	}
	return "Providers"
}

func (m Model) listLines(width int) []string {
	if m.loading && len(m.page.Items) == 0 {
		return []string{m.theme.Warning.Render("◌ Loading providers…")}
	}
	if m.err != nil {
		return []string{
			m.theme.Danger.Render("✗ Unable to load providers"),
			m.theme.Danger.Render("Error  " + m.err.Error()),
			m.theme.Muted.Render("Press r to retry."),
		}
	}
	if len(m.page.Items) == 0 {
		return []string{
			m.theme.Warning.Render("○ No providers found"),
			m.theme.Muted.Render("  Adjust the filter or connect another Console."),
		}
	}
	nameWidth := max(14, min(24, width/4))
	cloudWidth := max(6, min(10, width/8))
	editWidth := 8
	lines := []string{m.theme.Muted.Render("  " + pad("NAME", nameWidth) + " " + pad("CLOUD", cloudWidth) + " " + pad("EDITABLE", editWidth) + " REPO")}
	start, end := visibleWindow(m.cursor, len(m.page.Items), 8)
	for i := start; i < end; i++ {
		item := m.page.Items[i]
		cursor := "  "
		if i == m.cursor {
			cursor = "› "
		}
		row := cursor + pad(item.Name, nameWidth) + " " + pad(item.Cloud, cloudWidth) + " " + pad(loCoalesce(item.Editable, "—"), editWidth) + " " + loCoalesce(item.RepoURL, "—")
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
		return []string{m.theme.Warning.Render("◌ Loading provider detail…")}
	}
	if m.err != nil {
		return []string{m.theme.Danger.Render("✗ Unable to load provider"), m.theme.Danger.Render(m.err.Error())}
	}
	lines := []string{
		m.labelValue("Name", m.detail.Name),
		m.labelValue("Cloud", loCoalesce(m.detail.Cloud, "—")),
		m.labelValue("Namespace", loCoalesce(m.detail.Namespace, "—")),
		m.labelValue("Editable", loCoalesce(m.detail.Editable, "—")),
		m.labelValue("Repo", loCoalesce(m.detail.RepoURL, "—")),
		m.labelValue("Service", loCoalesce(m.detail.Service, "—")),
		m.labelValue("ID", m.detail.ID),
	}
	if m.detail.DeletedAt != "" {
		lines = append(lines, m.labelValue("Deleted", m.detail.DeletedAt))
	}
	if len(m.detail.Credentials) > 0 {
		lines = append(lines, "", m.theme.Muted.Render("Credentials"))
		for _, credential := range m.detail.Credentials {
			lines = append(lines, "  "+credential.Name+"  "+m.theme.Muted.Render(credential.Namespace+" · "+credential.Kind))
		}
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
