package workbenches

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
	body, help := m.bodyAndHelp(page.ContentWidth(width))
	title := "Workbenches"
	if m.detail.ID != "" && m.mode != modeList {
		title += " · " + m.detail.WorkbenchName
	}
	return page.Render(m.theme, width, height, title, m.status(), body, help)
}
func (m Model) status() string {
	if m.loading {
		return m.theme.Warning.Render("◌ working")
	}
	if m.needsAuth {
		return m.theme.Warning.Render("○ connect Console")
	}
	if m.err != nil {
		return m.theme.Danger.Render("✗ failed")
	}
	return m.theme.Success.Render(fmt.Sprintf("%d jobs", len(m.page.Items)))
}
func (m Model) bodyAndHelp(width int) (string, string) {
	if m.mode == modeFilter {
		return page.Panel(m.theme, "Filter workbench jobs", []string{m.filterInput.View()}, width, 5, true), "enter apply · esc cancel"
	}
	if m.needsAuth {
		return page.Panel(m.theme, "Console required", []string{"Connect a Console profile to browse workbench jobs.", "", "Press c to open Access."}, width, 7, true), "c connect · esc AI hub"
	}
	if m.mode == modePrompt {
		return page.Panel(m.theme, "Follow-up prompt", []string{"Job        " + m.detail.ID, "Workbench  " + m.detail.WorkbenchName, "", m.prompt.View()}, width, 11, true), "ctrl+s review · esc cancel"
	}
	if m.mode == modeReview {
		return page.Panel(m.theme, "Review follow-up", []string{"Workbench  " + m.detail.WorkbenchName, "Job        " + m.detail.ID, "", "Prompt", m.prompt.Value(), "", m.theme.Muted.Render("This queues the prompt for the selected running or settled job.")}, width, 12, true), "enter queue · esc edit"
	}
	if m.mode == modeOperating {
		return page.Panel(m.theme, "Queue follow-up", []string{m.theme.Warning.Render("◌ Queueing prompt…")}, width, 7, true), "ctrl+c quit"
	}
	if m.mode == modeResult {
		lines := []string{m.theme.Success.Render("✓ Prompt queued"), "", "Prompt ID   " + m.result.ID, "Job ID      " + m.result.WorkbenchID, "Dequeues     " + m.result.DequeueAt}
		if m.err != nil {
			lines = []string{m.theme.Danger.Render("✗ Follow-up failed"), "", m.err.Error()}
		}
		return page.Panel(m.theme, "Result", lines, width, 10, true), "enter/esc detail"
	}
	if m.mode == modeDetail {
		return page.Panel(m.theme, "Job detail", []string{"Workbench  " + value(m.detail.WorkbenchName), "Status     " + value(m.detail.Status), "Prompt     " + value(m.detail.Prompt), "Job ID     " + m.detail.ID, "Started    " + value(m.detail.InsertedAt)}, width, 11, true), "f follow up · esc list"
	}
	if m.loading && len(m.page.Items) == 0 {
		return page.Panel(m.theme, "Recent workbench jobs", []string{"◌ Loading workbench jobs…"}, width, 14, true), "esc AI hub"
	}
	if m.err != nil {
		return page.Panel(m.theme, "Recent workbench jobs", []string{"Unable to load workbench jobs", m.err.Error()}, width, 14, true), "r retry · esc AI hub"
	}
	lines := []string{m.theme.Muted.Render("  WORKBENCH           STATUS       PROMPT")}
	for i, item := range m.page.Items {
		cursor := "  "
		if i == m.cursor {
			cursor = "› "
		}
		row := fmt.Sprintf("%s%-19s %-12s %s", cursor, value(item.WorkbenchName), value(item.Status), value(item.Prompt))
		lines = append(lines, ansi.Truncate(row, width-2, "…"))
	}
	if len(m.page.Items) == 0 {
		lines = append(lines, "No workbench jobs found.")
	}
	return page.Panel(m.theme, "Recent workbench jobs", lines, width, 14, true), "↑/↓ select · enter detail · / filter · r refresh · esc AI hub"
}
func value(v string) string {
	if strings.TrimSpace(v) == "" {
		return "—"
	}
	return v
}
