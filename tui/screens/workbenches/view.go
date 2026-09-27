package workbenches

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"

	"github.com/pluralsh/plural-cli/tui/components/page"
)

func (m Model) View(width, height int) string {
	width, height = page.Size(width, height)
	if width < page.MinimumWidth || height < page.MinimumHeight {
		return page.Unsupported(m.theme, width, height)
	}
	body, help := m.bodyAndHelp(page.ContentWidth(width), height)
	title := "Workbenches"
	if m.detail.ID != "" && (m.mode == modeDetail || (m.mode == modeResult && m.returnTo == modeDetail)) {
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

func (m Model) bodyAndHelp(width, height int) (string, string) {
	tall := detailPanelHeight(height)
	if m.mode == modeFilter {
		return page.Panel(m.theme, "Filter workbench jobs", []string{m.filterInput.View()}, width, 5, true), "enter apply · esc cancel"
	}
	if m.needsAuth {
		return page.Panel(m.theme, "Console required", []string{"Connect a Console profile to browse workbench jobs.", "", "Press c to open Access."}, width, 7, true), "c connect · esc AI hub"
	}
	if m.mode == modePrompt {
		m.prompt.SetWidth(max(8, width-8))
		return page.Panel(m.theme, "Follow-up prompt", []string{
			"Job        " + value(m.detail.ID),
			"Workbench  " + value(m.detail.WorkbenchName),
			"",
			m.prompt.View(),
		}, width, 9, true), "enter review · esc cancel"
	}
	if m.mode == modeReview {
		inner := max(1, width-4)
		lines := []string{
			"Job        " + value(m.detail.ID),
			"Workbench  " + value(m.detail.WorkbenchName),
			"",
			m.theme.Muted.Render("Prompt"),
		}
		lines = append(lines, wrapLines(m.prompt.Value(), inner)...)
		lines = append(lines, "", m.theme.Muted.Render("Queues a follow-up on the selected workbench job."))
		return page.Panel(m.theme, "Review follow-up", lines, width, 12, true), "enter queue · esc edit"
	}
	if m.mode == modeOperating {
		return page.Panel(m.theme, "Queue follow-up", []string{m.theme.Warning.Render("◌ Queueing prompt…")}, width, 7, true), "ctrl+c quit"
	}
	if m.mode == modeResult {
		if m.err != nil {
			inner := max(1, width-4)
			lines := []string{m.theme.Danger.Render("✗ Follow-up failed"), ""}
			lines = append(lines, wrapLines(m.err.Error(), inner)...)
			return page.Panel(m.theme, "Result", lines, width, 12, true), "enter/esc back"
		}
		return page.Panel(m.theme, "Result", []string{
			m.theme.Success.Render("✓ Prompt queued"),
			"",
			"Prompt ID   " + value(m.result.ID),
			"Job ID      " + value(m.result.WorkbenchID),
			"Dequeues    " + value(m.result.DequeueAt),
		}, width, 10, true), "enter/esc back"
	}
	if m.mode == modeDetail {
		lines := windowed(m.detailLines(width), m.detailOffset, tall)
		return page.Panel(m.theme, "Job detail", lines, width, tall, true), "↑/↓ scroll · pgup/pgdn · f follow up · esc list"
	}
	if m.loading && len(m.page.Items) == 0 {
		return page.Panel(m.theme, "Recent workbench jobs", []string{"◌ Loading workbench jobs…"}, width, 14, true), "esc AI hub"
	}
	if m.err != nil {
		inner := max(1, width-4)
		lines := []string{"Unable to load workbench jobs"}
		lines = append(lines, wrapLines(m.err.Error(), inner)...)
		return page.Panel(m.theme, "Recent workbench jobs", lines, width, 14, true), "r retry · esc AI hub"
	}
	lines := []string{m.theme.Muted.Render("  WORKBENCH           STATUS       PROMPT")}
	for i, item := range m.page.Items {
		cursor := "  "
		if i == m.cursor {
			cursor = "› "
		}
		row := fmt.Sprintf("%s%-19s %-12s %s", cursor, value(item.WorkbenchName), value(item.Status), summary(item.Prompt))
		lines = append(lines, ansi.Truncate(row, width-2, "…"))
	}
	if len(m.page.Items) == 0 {
		lines = append(lines, "No workbench jobs found.")
	}
	return page.Panel(m.theme, "Recent workbench jobs", lines, width, 14, true), "↑/↓ select · enter detail · f follow up · / filter · r refresh · esc AI hub"
}

func (m Model) detailLines(width int) []string {
	inner := max(1, width-4)
	lines := []string{
		"Workbench  " + value(m.detail.WorkbenchName),
		"Status     " + value(m.detail.Status),
		"Job ID     " + value(m.detail.ID),
		"Started    " + formatTime(m.detail.InsertedAt),
		"",
		m.theme.Muted.Render("Prompt"),
	}
	return append(lines, wrapLines(m.detail.Prompt, inner)...)
}

func windowed(lines []string, offset, panelH int) []string {
	inner := max(1, panelH-2)
	maxOff := max(0, len(lines)-inner)
	if offset > maxOff {
		offset = maxOff
	}
	if offset < 0 {
		offset = 0
	}
	if offset == 0 {
		return lines
	}
	return lines[offset:]
}

func wrapLines(text string, width int) []string {
	if width < 1 {
		width = 1
	}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.TrimSpace(text)
	if text == "" {
		return []string{"—"}
	}
	out := make([]string, 0, strings.Count(text, "\n")+1)
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimRight(line, " ")
		if line == "" {
			out = append(out, "")
			continue
		}
		out = append(out, strings.Split(ansi.Wrap(line, width, ""), "\n")...)
	}
	return out
}

func formatTime(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "—"
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		parsed, err := time.Parse(layout, v)
		if err == nil {
			return parsed.UTC().Format("2006-01-02 15:04 UTC")
		}
	}
	return v
}

func summary(v string) string {
	v = strings.Join(strings.Fields(v), " ")
	if v == "" {
		return "—"
	}
	return ansi.Truncate(v, 60, "…")
}

func value(v string) string {
	if strings.TrimSpace(v) == "" {
		return "—"
	}
	return strings.Join(strings.Fields(strings.ReplaceAll(v, "\n", " ")), " ")
}
