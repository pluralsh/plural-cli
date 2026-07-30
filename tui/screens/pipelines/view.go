package pipelines

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
	title := "Pipelines"
	if m.mode == modeDetail && m.detail.Name != "" {
		title = "Pipelines · " + m.detail.Name
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
		return m.theme.Success.Render(fmt.Sprintf("%d stages", m.detail.StageCount))
	case modeList:
		if m.filter != "" {
			return m.theme.Muted.Render(fmt.Sprintf("%d matching", len(m.page.Items)))
		}
		return m.theme.Success.Render(fmt.Sprintf("%d pipelines", len(m.page.Items)))
	default:
		return m.theme.Muted.Render("pipelines")
	}
}

func (m Model) bodyAndHelp(width int) (string, string) {
	if m.mode == modeFilter {
		lines := []string{
			m.theme.Muted.Render("Filter by name, project, or id."),
			"",
			m.filterInput.View(),
		}
		return page.Panel(m.theme, "Filter pipelines", lines, width, 6, true), "enter apply · esc cancel"
	}
	if m.needsAuth {
		lines := []string{
			m.theme.Warning.Render("○ Console is not connected"),
			m.theme.Muted.Render("  Connect a Console profile to browse pipelines."),
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
		return "Pipelines · filter “" + m.filter + "”"
	}
	return "Pipelines"
}

func (m Model) listLines(width int) []string {
	if m.loading && len(m.page.Items) == 0 {
		return []string{m.theme.Warning.Render("◌ Loading pipelines…")}
	}
	if m.err != nil {
		return []string{
			m.theme.Danger.Render("✗ Unable to load pipelines"),
			m.theme.Danger.Render("Error  " + m.err.Error()),
			m.theme.Muted.Render("Press r to retry."),
		}
	}
	if len(m.page.Items) == 0 {
		return []string{
			m.theme.Warning.Render("○ No pipelines found"),
			m.theme.Muted.Render("  Adjust the filter or connect another Console."),
		}
	}
	nameWidth := max(16, min(28, width/3))
	projectWidth := max(10, min(20, width/4))
	lines := []string{m.theme.Muted.Render("  " + pad("NAME", nameWidth) + " " + pad("PROJECT", projectWidth) + " STAGES")}
	start, end := visibleWindow(m.cursor, len(m.page.Items), 8)
	for i := start; i < end; i++ {
		item := m.page.Items[i]
		cursor := "  "
		if i == m.cursor {
			cursor = "› "
		}
		project := loCoalesce(item.Project, "—")
		row := cursor + pad(item.Name, nameWidth) + " " + pad(project, projectWidth) + " " + fmt.Sprintf("%d", item.StageCount)
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
		return []string{m.theme.Warning.Render("◌ Loading pipeline detail…")}
	}
	if m.err != nil {
		return []string{m.theme.Danger.Render("✗ Unable to load pipeline"), m.theme.Danger.Render(m.err.Error())}
	}
	lines := []string{
		m.labelValue("Name", m.detail.Name),
		m.labelValue("Project", loCoalesce(m.detail.Project, "—")),
		m.labelValue("Stages", fmt.Sprintf("%d", m.detail.StageCount)),
		m.labelValue("ID", m.detail.ID),
	}
	if len(m.detail.Stages) > 0 {
		lines = append(lines, "", m.theme.Muted.Render("Stages"))
		for _, stage := range m.detail.Stages {
			services := strings.Join(stage.Services, ", ")
			if services == "" {
				services = "—"
			}
			lines = append(lines, "  "+stage.Name+"  "+m.theme.Muted.Render(services))
		}
	}
	if len(m.detail.Edges) > 0 {
		lines = append(lines, "", m.theme.Muted.Render("Edges"))
		for _, edge := range m.detail.Edges {
			lines = append(lines, "  "+edge.From+" → "+edge.To)
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
