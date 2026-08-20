package agents

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
	title := "Agent runs"
	if m.mode != modeList && m.detail.ID != "" {
		title += " · " + m.detail.ID
	}
	return page.Render(m.theme, width, height, title, m.status(), body, help)
}

func (m Model) status() string {
	if m.loading {
		return m.theme.Warning.Render("◌ loading")
	}
	if m.needsAuth {
		return m.theme.Warning.Render("○ connect Console")
	}
	if m.err != nil {
		return m.theme.Danger.Render("✗ failed")
	}
	return m.theme.Success.Render(fmt.Sprintf("%d resumable", len(m.page.Items)))
}

func (m Model) bodyAndHelp(width int) (string, string) {
	if m.mode == modeFilter {
		return page.Panel(m.theme, "Filter agent runs", []string{m.input.View()}, width, 5, true), "enter apply · esc cancel"
	}
	if m.needsAuth {
		return page.Panel(m.theme, "Console required", []string{"Connect a Console profile to browse agent runs.", "", "Press c to open Access."}, width, 7, true), "c connect · esc AI hub"
	}
	if m.mode == modeRepoPath {
		return page.Panel(m.theme, "Choose local clone", []string{"Existing clone for " + m.detail.Repository, "", m.input.View()}, width, 7, true), "enter continue · esc cancel"
	}
	if m.mode == modeResult {
		lines := []string{m.theme.Success.Render("✓ Resume complete"), "", m.result}
		if m.err != nil {
			lines = []string{m.theme.Danger.Render("✗ Resume failed"), "", m.err.Error()}
		}
		return page.Panel(m.theme, "Agent resume", lines, width, 9, true), "enter/esc detail"
	}
	if m.mode == modeDetail {
		lines := []string{
			"Repository  " + value(m.detail.Repository),
			"Branch      " + value(m.detail.Branch),
			"Provider    " + value(m.detail.Provider),
			"PR ref      " + value(m.detail.PRRef),
			"Prompt      " + value(m.detail.Prompt),
			"Run ID      " + m.detail.ID,
		}
		if len(m.detail.PullRequests) > 1 {
			lines = append(lines, "", fmt.Sprintf("%d pull request branches available", len(m.detail.PullRequests)))
		}
		return page.Panel(m.theme, "Run detail", lines, width, 12, true), "r resume interactively · esc list"
	}
	if m.loading && len(m.page.Items) == 0 {
		return page.Panel(m.theme, "Resumable runs", []string{"◌ Loading agent runs…"}, width, 14, true), "esc AI hub"
	}
	if m.err != nil {
		return page.Panel(m.theme, "Resumable runs", []string{"Unable to load agent runs", m.err.Error()}, width, 14, true), "r retry · esc AI hub"
	}
	lines := []string{m.theme.Muted.Render("  REPOSITORY            PROVIDER     BRANCH / PR                    PROMPT")}
	for i, item := range m.page.Items {
		cursor := "  "
		if i == m.cursor {
			cursor = "› "
		}
		row := fmt.Sprintf("%s%-20s %-12s %-30s %s", cursor, repoName(item.Repository), value(item.Provider), value(first(item.PRRef, item.Branch)), value(item.Prompt))
		lines = append(lines, ansi.Truncate(row, width-2, "…"))
	}
	if len(m.page.Items) == 0 {
		lines = append(lines, "No resumable agent runs found.")
	}
	return page.Panel(m.theme, "Resumable agent runs", lines, width, 14, true), "↑/↓ select · enter detail · / filter · r refresh · esc AI hub"
}

func value(v string) string {
	if strings.TrimSpace(v) == "" {
		return "—"
	}
	return v
}
func first(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
func repoName(v string) string {
	v = strings.TrimSuffix(strings.TrimSpace(v), ".git")
	if i := strings.LastIndexAny(v, "/:"); i >= 0 {
		return v[i+1:]
	}
	return v
}
