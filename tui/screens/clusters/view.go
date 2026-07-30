package clusters

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
	title := "Clusters"
	if m.mode == modeDetail && m.detail.Name != "" {
		title = "Clusters · " + m.detail.Name
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
		if m.detail.DeletedAt != "" {
			return m.theme.Danger.Render("terminating")
		}
		if m.detail.Self {
			return m.theme.Success.Render("self")
		}
		return m.theme.Success.Render(loCoalesce(m.detail.Distro, "ready"))
	case modeList:
		if m.filter != "" {
			return m.theme.Muted.Render(fmt.Sprintf("%d matching", len(m.items)))
		}
		return m.theme.Success.Render(fmt.Sprintf("%d clusters", len(m.items)))
	default:
		return m.theme.Muted.Render("clusters")
	}
}

func (m Model) bodyAndHelp(width int) (string, string) {
	if m.mode == modeFilter {
		lines := []string{
			m.theme.Muted.Render("Filter by handle, name, id, version, or distro."),
			"",
			m.filterInput.View(),
		}
		return page.Panel(m.theme, "Filter clusters", lines, width, 6, true), "enter apply · esc cancel"
	}
	if m.needsAuth {
		lines := []string{
			m.theme.Warning.Render("○ Console is not connected"),
			m.theme.Muted.Render("  Connect a Console profile to browse clusters."),
			"",
			m.theme.Body.Render("Press c to open Access."),
		}
		return page.Panel(m.theme, "Console required", lines, width, 8, true), "c connect · esc back · ctrl+c quit"
	}
	if m.mode == modeDetail {
		help := "r refresh · esc list · ctrl+c quit"
		return page.Panel(m.theme, "Summary", m.detailLines(), width, 16, true), help
	}
	help := "↑/↓ select · enter open · / filter · r refresh · esc back"
	if width < 100 {
		help = "↑/↓ · enter · / filter · esc back"
	}
	return page.Panel(m.theme, m.listTitle(), m.listLines(width), width, 14, true), help
}

func (m Model) listTitle() string {
	if m.filter != "" {
		return "Clusters · filter “" + m.filter + "”"
	}
	return "Clusters"
}

func (m Model) listLines(width int) []string {
	if m.loading && len(m.items) == 0 {
		return []string{m.theme.Warning.Render("◌ Loading clusters…")}
	}
	if m.err != nil {
		return []string{m.theme.Danger.Render("✗ Unable to load clusters"), m.theme.Danger.Render("Error  " + m.err.Error()), m.theme.Muted.Render("Press r to retry.")}
	}
	if len(m.items) == 0 {
		return []string{m.theme.Warning.Render("○ No clusters found"), m.theme.Muted.Render("  Adjust the filter or connect another Console.")}
	}
	handleWidth := max(12, min(20, width/4))
	nameWidth := max(12, min(24, width/3))
	lines := []string{m.theme.Muted.Render("  " + pad("HANDLE", handleWidth) + " " + pad("NAME", nameWidth) + " " + pad("VERSION", 10) + " DISTRO")}
	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = "› "
		}
		handle := item.Handle
		if handle == "" {
			handle = "—"
		} else {
			handle = "@" + handle
		}
		version := loCoalesce(item.Version, "—")
		distro := loCoalesce(item.Distro, "—")
		row := cursor + pad(handle, handleWidth) + " " + pad(item.Name, nameWidth) + " " + pad(version, 10) + " " + distro
		lines = append(lines, ansi.Truncate(row, width-2, "…"))
	}
	return lines
}

func (m Model) detailLines() []string {
	if m.loading {
		return []string{m.theme.Warning.Render("◌ Loading cluster detail…")}
	}
	if m.err != nil {
		return []string{m.theme.Danger.Render("✗ Unable to load cluster"), m.theme.Danger.Render(m.err.Error())}
	}
	handle := m.detail.Handle
	if handle != "" {
		handle = "@" + handle
	} else {
		handle = "—"
	}
	lines := []string{
		m.labelValue("Handle", handle),
		m.labelValue("Name", m.detail.Name),
		m.labelValue("Version", loCoalesce(m.detail.Version, "—")),
		m.labelValue("Distro", loCoalesce(m.detail.Distro, "—")),
		m.labelValue("Project", loCoalesce(m.detail.Project, "—")),
		m.labelValue("Provider", loCoalesce(m.detail.Provider, "—")),
		m.labelValue("Pinged", loCoalesce(m.detail.PingedAt, "—")),
		m.labelValue("Self", fmt.Sprintf("%v", m.detail.Self)),
		m.labelValue("Protect", fmt.Sprintf("%v", m.detail.Protect)),
		m.labelValue("Node pools", fmt.Sprintf("%d", m.detail.NodePools)),
		m.labelValue("ID", m.detail.ID),
	}
	if m.detail.DeletedAt != "" {
		lines = append(lines, m.theme.Danger.Render("Deleted   "+m.detail.DeletedAt))
	}
	if len(m.detail.Tags) > 0 {
		lines = append(lines, "", m.theme.Muted.Render("Tags"))
		for _, tag := range m.detail.Tags {
			lines = append(lines, "  "+tag.Name+"="+tag.Value)
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
