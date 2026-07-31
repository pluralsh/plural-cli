package ai

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
	contentWidth := page.ContentWidth(width)
	body, help := m.bodyAndHelp(contentWidth)
	return page.Render(m.theme, width, height, "AI", m.headerStatus(), body, help)
}

func (m Model) headerStatus() string {
	return m.theme.Success.Render(fmt.Sprintf("%d commands", len(items)))
}

func (m Model) bodyAndHelp(width int) (string, string) {
	lines := []string{m.theme.Muted.Render("Choose an AI workspace to open its interactive screen."), ""}
	for i, item := range items {
		cursor := "  "
		if i == m.cursor {
			cursor = "› "
		}
		lines = append(lines, cursor+fmt.Sprintf("%s  %-12s %s", item.number, item.title, item.blurb))
	}
	lines = append(lines, "", m.theme.Muted.Render("Agents browses resumable runs; Workbenches queues follow-up prompts."))
	return page.Panel(m.theme, "AI workspaces", lines, width, 8, true), "↑/↓ select · enter open · 1-2 shortcut · esc back"
}

func normalizeView(view string) string {
	lines := strings.Split(ansi.Strip(view), "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " ")
	}
	return strings.Join(lines, "\n")
}
