package services

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
	status := m.headerStatus()
	body, help := m.bodyAndHelp(contentWidth)
	return page.Render(m.theme, width, height, "Services", status, body, help)
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
		return m.statusBadge(m.detail.Status)
	case modeList:
		if m.serviceFilter != "" {
			return m.theme.Muted.Render(fmt.Sprintf("%d matching · %s", len(m.page.Items), clusterLabel(m.cluster)))
		}
		return m.theme.Success.Render(fmt.Sprintf("%d services · %s", len(m.page.Items), clusterLabel(m.cluster)))
	case modeClusters:
		if m.clusterFilter != "" {
			return m.theme.Muted.Render(fmt.Sprintf("%d matching clusters", len(m.clusters)))
		}
		return m.theme.Success.Render(fmt.Sprintf("%d clusters", len(m.clusters)))
	default:
		return m.theme.Muted.Render("select cluster")
	}
}

func (m Model) bodyAndHelp(width int) (string, string) {
	if m.mode == modeFilter {
		title := "Filter services"
		hint := "Filter by name, namespace, status, or git path."
		if m.filteringCluster {
			title = "Filter clusters"
			hint = "Filter by handle, name, or id."
		}
		lines := []string{m.theme.Muted.Render(hint), "", m.filterInput.View()}
		return page.Panel(m.theme, title, lines, width, 6, true), "enter apply · esc cancel"
	}
	if m.needsAuth {
		lines := []string{
			m.theme.Warning.Render("○ Console is not connected"),
			m.theme.Muted.Render("  Connect a Console profile to browse services."),
			"",
			m.theme.Body.Render("Press c to open Access."),
		}
		return page.Panel(m.theme, "Console required", lines, width, 8, true), "c connect · esc back · ctrl+c quit"
	}
	if m.mode == modeDetail {
		return page.Panel(m.theme, m.detail.Name, m.detailLines(), width, 14, true), "r refresh · esc back · ctrl+c quit"
	}
	if m.mode == modeClusters {
		help := "↑/↓ select · enter open cluster · / filter · r refresh · esc back"
		if width < 100 {
			help = "↑/↓ · enter · / filter · esc back"
		}
		return page.Panel(m.theme, m.clusterTitle(), m.clusterLines(width), width, 14, true), help
	}
	help := "↑/↓ select · enter open · / filter · n/p page · r refresh · esc clusters"
	if width < 100 {
		help = "↑/↓ · enter · / filter · n/p page · esc clusters"
	}
	return page.Panel(m.theme, m.listTitle(), m.listLines(width), width, 14, true), help
}

func (m Model) clusterTitle() string {
	if m.clusterFilter != "" {
		return "Clusters · filter “" + m.clusterFilter + "”"
	}
	return "Choose a cluster"
}

func (m Model) listTitle() string {
	title := "Services · " + clusterLabel(m.cluster)
	if m.serviceFilter != "" {
		title += " · filter “" + m.serviceFilter + "”"
	}
	return title
}

func (m Model) clusterLines(width int) []string {
	if m.loading && len(m.clusters) == 0 {
		return []string{m.theme.Warning.Render("◌ Loading clusters…")}
	}
	if m.err != nil {
		return []string{m.theme.Danger.Render("✗ Unable to load clusters"), m.theme.Danger.Render("Error  " + m.err.Error()), m.theme.Muted.Render("Press r to retry.")}
	}
	if len(m.clusters) == 0 {
		return []string{m.theme.Warning.Render("○ No clusters found"), m.theme.Muted.Render("  Adjust the filter or connect another Console.")}
	}
	handleWidth := max(12, min(24, width/3))
	lines := []string{m.theme.Muted.Render("  " + pad("HANDLE", handleWidth) + " " + pad("NAME", 24) + " ID")}
	for i, cluster := range m.clusters {
		cursor := "  "
		if i == m.clusterCursor {
			cursor = "› "
		}
		handle := cluster.Handle
		if handle == "" {
			handle = "—"
		} else {
			handle = "@" + handle
		}
		row := cursor + pad(handle, handleWidth) + " " + pad(cluster.Name, 24) + " " + cluster.ID
		lines = append(lines, ansi.Truncate(row, width-2, "…"))
	}
	return lines
}

func (m Model) listLines(width int) []string {
	if m.loading && len(m.page.Items) == 0 {
		return []string{m.theme.Warning.Render("◌ Loading services for " + clusterLabel(m.cluster) + "…")}
	}
	if m.err != nil {
		return []string{m.theme.Danger.Render("✗ Unable to load services"), m.theme.Danger.Render("Error  " + m.err.Error()), m.theme.Muted.Render("Press r to retry.")}
	}
	if len(m.page.Items) == 0 {
		return []string{m.theme.Warning.Render("○ No services found"), m.theme.Muted.Render("  Adjust the filter or choose another cluster.")}
	}
	nameWidth := max(12, min(28, width/3))
	lines := []string{m.theme.Muted.Render("  " + pad("NAME", nameWidth) + " " + pad("NAMESPACE", 14) + " " + pad("STATUS", 10) + " GIT")}
	for i, item := range m.page.Items {
		cursor := "  "
		if i == m.cursor {
			cursor = "› "
		}
		git := strings.TrimSpace(item.GitRef + " " + item.GitFolder)
		if git == "" {
			git = "—"
		}
		row := cursor + pad(item.Name, nameWidth) + " " + pad(item.Namespace, 14) + " " + pad(item.Status, 10) + " " + git
		lines = append(lines, ansi.Truncate(row, width-2, "…"))
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

func (m Model) detailLines() []string {
	if m.loading {
		return []string{m.theme.Warning.Render("◌ Loading service detail…")}
	}
	if m.err != nil {
		return []string{m.theme.Danger.Render("✗ Unable to load service"), m.theme.Danger.Render(m.err.Error())}
	}
	cluster := m.detail.ClusterName
	if m.detail.ClusterHandle != "" {
		cluster = m.detail.ClusterHandle
		if m.detail.ClusterName != "" && m.detail.ClusterName != m.detail.ClusterHandle {
			cluster = m.detail.ClusterHandle + " · " + m.detail.ClusterName
		}
	}
	if cluster == "" {
		cluster = clusterLabel(m.cluster)
	}
	revision := m.detail.RevisionSHA
	if m.detail.RevisionRef != "" {
		revision = m.detail.RevisionRef
		if m.detail.RevisionSHA != "" {
			revision += " · " + shortSHA(m.detail.RevisionSHA)
		}
	}
	if revision == "" {
		revision = "—"
	}
	git := strings.TrimSpace(m.detail.GitRef + " / " + m.detail.GitFolder)
	if git == " / " || git == "" {
		git = "—"
	}
	lines := []string{
		m.labelValue("Status", m.statusBadge(m.detail.Status)),
		m.labelValue("Namespace", m.detail.Namespace),
		m.labelValue("Cluster", cluster),
		m.labelValue("Revision", revision),
		m.labelValue("Git", git),
		m.labelValue("Components", fmt.Sprintf("%d / %d synced", m.detail.Synced, m.detail.Components)),
		"",
	}
	if len(m.detail.Errors) == 0 {
		lines = append(lines, m.theme.Success.Render("✓ No service errors"))
		return lines
	}
	lines = append(lines, m.theme.Danger.Render("Errors"))
	for _, item := range m.detail.Errors {
		source := item.Source
		if source == "" {
			source = "sync"
		}
		lines = append(lines, m.theme.Danger.Render("✗ "+source)+"  "+item.Message)
	}
	return lines
}

func (m Model) labelValue(label, value string) string {
	label = label + strings.Repeat(" ", max(1, 12-len(label)))
	return label + " " + value
}

func (m Model) statusBadge(status string) string {
	switch strings.ToUpper(status) {
	case "HEALTHY", "SYNCED":
		return m.theme.Success.Render(status)
	case "FAILED":
		return m.theme.Danger.Render(status)
	case "PAUSED", "STALE":
		return m.theme.Warning.Render(status)
	default:
		if status == "" {
			return m.theme.Muted.Render("UNKNOWN")
		}
		return m.theme.Muted.Render(status)
	}
}

func pad(value string, width int) string {
	value = ansi.Truncate(value, width, "…")
	if lipgloss.Width(value) >= width {
		return value
	}
	return value + strings.Repeat(" ", width-lipgloss.Width(value))
}

func shortSHA(value string) string {
	if len(value) > 8 {
		return value[:8]
	}
	return value
}
