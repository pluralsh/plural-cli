package repositories

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
	title := "Repositories"
	if m.mode == modeDetail && m.detail.URL != "" {
		title = "Repositories · " + shortURL(m.detail.URL, 40)
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
		switch m.detail.Health {
		case "PULLABLE":
			return m.theme.Success.Render("PULLABLE")
		case "FAILED":
			return m.theme.Danger.Render("FAILED")
		default:
			return m.theme.Muted.Render(loCoalesce(m.detail.Health, "UNKNOWN"))
		}
	case modeList:
		if m.filter != "" {
			return m.theme.Muted.Render(fmt.Sprintf("%d matching", len(m.page.Items)))
		}
		return m.theme.Success.Render(fmt.Sprintf("%d repositories", len(m.page.Items)))
	default:
		return m.theme.Muted.Render("repositories")
	}
}

func (m Model) bodyAndHelp(width int) (string, string) {
	if m.mode == modeFilter {
		lines := []string{
			m.theme.Muted.Render("Filter by url, id, health, error, or auth method."),
			"",
			m.filterInput.View(),
		}
		return page.Panel(m.theme, "Filter repositories", lines, width, 6, true), "enter apply · esc cancel"
	}
	if m.needsAuth {
		lines := []string{
			m.theme.Warning.Render("○ Console is not connected"),
			m.theme.Muted.Render("  Connect a Console profile to browse repositories."),
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
		return "Repositories · filter “" + m.filter + "”"
	}
	return "Repositories"
}

func (m Model) listLines(width int) []string {
	if m.loading && len(m.page.Items) == 0 {
		return []string{m.theme.Warning.Render("◌ Loading repositories…")}
	}
	if m.err != nil {
		return []string{
			m.theme.Danger.Render("✗ Unable to load repositories"),
			m.theme.Danger.Render("Error  " + m.err.Error()),
			m.theme.Muted.Render("Press r to retry."),
		}
	}
	if len(m.page.Items) == 0 {
		return []string{
			m.theme.Warning.Render("○ No repositories found"),
			m.theme.Muted.Render("  Adjust the filter or connect another Console."),
		}
	}
	urlWidth := max(24, min(48, width*2/3))
	lines := []string{m.theme.Muted.Render("  " + pad("URL", urlWidth) + " " + pad("HEALTH", 10) + " ERROR")}
	start, end := visibleWindow(m.cursor, len(m.page.Items), 8)
	for i := start; i < end; i++ {
		item := m.page.Items[i]
		cursor := "  "
		if i == m.cursor {
			cursor = "› "
		}
		health := loCoalesce(item.Health, "UNKNOWN")
		errText := loCoalesce(item.Error, "—")
		row := cursor + pad(item.URL, urlWidth) + " " + pad(health, 10) + " " + errText
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
		return []string{m.theme.Warning.Render("◌ Loading repository detail…")}
	}
	if m.err != nil {
		return []string{m.theme.Danger.Render("✗ Unable to load repository"), m.theme.Danger.Render(m.err.Error())}
	}
	lines := []string{
		m.labelValue("URL", m.detail.URL),
		m.labelValue("Health", loCoalesce(m.detail.Health, "UNKNOWN")),
		m.labelValue("Auth", loCoalesce(m.detail.AuthMethod, "—")),
		m.labelValue("Decrypt", fmt.Sprintf("%v", m.detail.Decrypt)),
		m.labelValue("ID", m.detail.ID),
	}
	if m.detail.Error != "" {
		lines = append(lines, m.theme.Danger.Render("Error      "+m.detail.Error))
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

func shortURL(url string, maxLen int) string {
	if lipgloss.Width(url) <= maxLen {
		return url
	}
	return ansi.Truncate(url, maxLen, "…")
}
