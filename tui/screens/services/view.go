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
	title := "Services"
	switch m.mode {
	case modeReview, modeOperating, modeResult:
		title = m.pending.title
		if title == "" {
			title = "Services"
		}
	case modeDeleteConfirm:
		title = "Delete · " + m.detail.Name
	case modeTarball:
		title = "Download tarball · " + m.detail.Name
	case modeCreate:
		title = "Create service · " + clusterLabel(m.cluster)
	case modeEdit:
		title = "Edit · " + m.detail.Name
	case modeClone:
		title = "Clone · " + m.detail.Name
	case modeWorkbench:
		title = "Workbench · " + m.detail.Name
	case modeDetail:
		title = "Services · " + m.detail.Name
	}
	return page.Render(m.theme, width, height, title, status, body, help)
}

func (m Model) headerStatus() string {
	if m.loading || m.mode == modeOperating {
		return m.theme.Warning.Render("◌ working")
	}
	if m.needsAuth {
		return m.theme.Warning.Render("○ connect Console")
	}
	if m.err != nil && m.mode != modeResult {
		return m.theme.Danger.Render("✗ attention")
	}
	switch m.mode {
	case modeReview:
		if m.pending.danger {
			return m.theme.Danger.Render("destructive")
		}
		return m.theme.Warning.Render("review")
	case modeResult:
		if m.result == "ok" || (m.result != "failed" && m.result != "") {
			if m.pending.kind == actionTarball && m.result != "ok" {
				return m.theme.Success.Render("saved")
			}
			if m.result == "failed" {
				return m.theme.Danger.Render("failed")
			}
			return m.theme.Success.Render("done")
		}
		return m.theme.Danger.Render("failed")
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
	case modeWorkbench:
		return m.theme.Muted.Render("dry-run only")
	default:
		return m.theme.Muted.Render("services")
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
		help := "enter detail · esc back"
		if m.result == "failed" {
			head = m.theme.Danger.Render("✗ Failed")
			help = "enter retry review · esc detail"
		} else if m.pending.kind == actionDelete {
			help = "enter list · esc list"
		} else if m.pending.kind == actionTarball {
			head = m.theme.Success.Render("✓ Wrote " + m.result)
		} else if m.pending.kind == actionWorkbench {
			head = m.theme.Muted.Render("CLI equivalent")
			help = "esc detail"
		}
		lines := []string{head, ""}
		lines = append(lines, m.opLog...)
		return page.Panel(m.theme, "Result", lines, width, 12, true), help
	case modeDeleteConfirm:
		lines := []string{
			m.theme.Danger.Render("This permanently deletes the Console service record."),
			m.theme.Muted.Render("Cluster workloads are not automatically uninstalled."),
			"",
			"Service    " + m.detail.Name + " · " + clusterLabel(m.cluster),
			"Type the service name to confirm:",
			"",
			m.formInput.View(),
		}
		if m.err != nil {
			lines = append(lines, "", m.theme.Danger.Render(m.err.Error()))
		}
		return page.Panel(m.theme, "Confirm deletion", lines, width, 11, true), "enter continue · esc cancel"
	case modeTarball:
		lines := []string{
			"Directory  " + m.formInput.View(),
			"",
			m.theme.Muted.Render("Fetches deploy token + tarball and unpacks into that directory."),
		}
		return page.Panel(m.theme, "Destination", lines, width, 7, true), "enter review · esc cancel"
	case modeCreate, modeEdit, modeClone:
		return m.formView(width)
	case modeWorkbench:
		mode := "› Template (.liquid / .tpl)     Lua engine"
		if !m.wbTemplate {
			mode = "  Template (.liquid / .tpl)   › Lua engine"
		}
		lines := []string{
			mode,
			"",
			"File       " + m.formInput.View(),
			"Context    service " + m.detail.Name + " " + clusterLabel(m.cluster),
			"",
			m.theme.Muted.Render("Enter shows the CLI equivalent. Full in-TUI render lands next."),
		}
		return page.Panel(m.theme, "Workbench", lines, width, 10, true), "tab mode · enter · esc detail"
	case modeDetail:
		summary := page.Panel(m.theme, "Summary", m.detailLines(), width, 9, false)
		actions := page.Panel(m.theme, "Actions", m.actionLines(width), width, 8, true)
		help := "↑/↓ actions · enter · k e c t m d · r refresh · esc list"
		if width < 100 {
			help = "↑/↓ · enter · letters · r · esc"
		}
		return summary + "\n\n" + actions, help
	case modeClusters:
		help := "↑/↓ select · enter open cluster · / filter · r refresh · esc back"
		if width < 100 {
			help = "↑/↓ · enter · / filter · esc back"
		}
		return page.Panel(m.theme, m.clusterTitle(), m.clusterLines(width), width, 14, true), help
	default:
		help := "↑/↓ · enter detail · n create · / filter · ]/[ page · r refresh · esc"
		if width < 100 {
			help = "↑/↓ · enter · n create · / · esc"
		}
		return page.Panel(m.theme, m.listTitle(), m.listLines(width), width, 14, true), help
	}
}

func (m Model) formView(width int) (string, string) {
	lines := make([]string, 0, len(m.formFields)+3)
	for i, field := range m.formFields {
		value := m.formValues[field.key]
		cursor := "  "
		if i == m.formIndex {
			cursor = "› "
			value = m.formInput.View()
		}
		lines = append(lines, cursor+pad(field.label, 12)+" "+value)
	}
	lines = append(lines, "", fmt.Sprintf("Dry-run attribute  %v  (ctrl+d toggle)", m.formDryRun))
	step := fmt.Sprintf("field %d/%d", m.formIndex+1, len(m.formFields))
	help := "↑/↓ fields · enter next/review · esc cancel · " + step
	return page.Panel(m.theme, "Form", lines, width, 12, true), help
}

func (m Model) actionLines(width int) []string {
	lines := make([]string, 0, len(detailActions()))
	for i, a := range detailActions() {
		cursor := "  "
		if i == m.actionCursor {
			cursor = "› "
		}
		label := fmt.Sprintf("%s  %-10s  %s", a.shortcut, a.title, a.blurb)
		if a.danger {
			label += "  [destructive]"
			if i == m.actionCursor {
				lines = append(lines, cursor+m.theme.Danger.Render(label))
			} else {
				lines = append(lines, cursor+m.theme.Muted.Render(label))
			}
			continue
		}
		if i == m.actionCursor {
			lines = append(lines, cursor+m.theme.Title.Render(label))
		} else {
			lines = append(lines, cursor+m.theme.Body.Render(label))
		}
		_ = width
	}
	return lines
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
		return []string{m.theme.Warning.Render("○ No services found"), m.theme.Muted.Render("  Press n to create, or adjust the filter.")}
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
			pager += " · [ prev"
		}
		if m.page.HasNext {
			pager += " · ] next"
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
		cluster = "@" + m.detail.ClusterHandle
		if m.detail.ClusterName != "" && m.detail.ClusterName != m.detail.ClusterHandle {
			cluster += " · " + m.detail.ClusterName
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
	}
	if len(m.detail.Errors) > 0 {
		lines = append(lines, m.theme.Danger.Render(fmt.Sprintf("%d errors", len(m.detail.Errors))))
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
