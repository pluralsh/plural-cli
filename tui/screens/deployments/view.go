package deployments

import (
	"fmt"

	"github.com/charmbracelet/x/ansi"
	"github.com/samber/lo"

	"github.com/pluralsh/plural-cli/tui/components/page"
)

func (m Model) View(width, height int) string {
	width, height = page.Size(width, height)
	if width < page.MinimumWidth || height < page.MinimumHeight {
		return page.Unsupported(m.theme, width, height)
	}
	contentWidth := page.ContentWidth(width)
	status := m.theme.Muted.Render("no console")
	if m.console != "" {
		status = m.theme.Success.Render(ansi.Truncate(m.console, 40, "…"))
	}

	body := page.Panel(m.theme, "Resources", m.resourceLines(contentWidth-4), contentWidth, 8, true) + "\n\n" +
		page.Panel(m.theme, "Connection", m.connectionLines(), contentWidth, 4, false)

	help := "1–6 / letter · enter open · esc welcome · ctrl+c quit"
	return page.Render(m.theme, width, height, "CD / Deployments", status, body, help)
}

func (m Model) resourceLines(innerWidth int) []string {
	lines := make([]string, 0, len(m.items))
	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = "› "
		}
		soon := ""
		if item.soon {
			soon = "  " + m.theme.Muted.Render("[soon]")
		}
		left := fmt.Sprintf("%s  %s   %-14s  %s", item.number, item.shortcut, item.title, item.blurb)
		var row string
		switch {
		case i == m.cursor && !item.soon:
			row = cursor + m.theme.Title.Render(left) + soon
		case item.soon:
			row = cursor + m.theme.Muted.Render(left) + soon
		default:
			row = cursor + m.theme.Body.Render(left)
		}
		lines = append(lines, ansi.Truncate(row, max(1, innerWidth), "…"))
	}
	return lines
}

func (m Model) connectionLines() []string {
	url := lo.CoalesceOrEmpty(m.console, "not connected")
	display := ansi.Truncate(url, 56, "…")
	if m.console == "" {
		display = m.theme.Warning.Render(display)
	}
	return []string{
		"Console   " + display,
		m.theme.Muted.Render("Tip       plural cd … remains the automation API"),
	}
}
