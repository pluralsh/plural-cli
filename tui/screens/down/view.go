package down

import (
	"fmt"

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
	return page.Render(m.theme, width, height, "Down", m.headerStatus(), body, help)
}

func (m Model) headerStatus() string {
	switch m.mode {
	case modeAffirm:
		return m.theme.Muted.Render("step · destroy Affirm")
	case modeDestroying:
		return m.theme.Muted.Render("destroying…")
	case modeComplete:
		if m.err != nil {
			return m.theme.Danger.Render("failed")
		}
		return m.theme.Success.Render("destroyed")
	default:
		return m.theme.Muted.Render("step 1 · mode")
	}
}

func (m Model) bodyAndHelp(width int) (string, string) {
	switch m.mode {
	case modeAffirm:
		lines := []string{
			m.theme.Muted.Render(AffirmMessage()),
			m.theme.Muted.Render("Same Affirm as plural down (PLURAL_DOWN_AFFIRM_DESTROY)."),
			"",
			"Mode  " + m.modeLabel(),
			m.theme.Muted.Render("      " + m.cli()),
			"",
		}
		lines = append(lines, m.affirmLines(width)...)
		if m.err != nil {
			lines = append(lines, "", m.theme.Danger.Render(m.err.Error()))
		}
		return page.Panel(m.theme, "Destroy", lines, width, 14, true), "↑/↓ · y/n · enter · esc back"
	case modeDestroying:
		lines := []string{
			"Mode  " + m.modeLabel(),
			"",
			m.spin.View() + " " + m.theme.Muted.Render("Destroying management cluster terraform…"),
			m.theme.Muted.Render("Same path as plural down — terraform output may use this terminal."),
		}
		for _, step := range m.steps {
			lines = append(lines, m.theme.Muted.Render("→ "+step))
		}
		return page.Panel(m.theme, "Destroying", lines, width, 12, true), "please wait"
	case modeComplete:
		lines := []string{}
		if m.err != nil {
			lines = append(lines,
				m.theme.Danger.Render("Destroy failed"),
				"",
				m.theme.Danger.Render(m.err.Error()),
			)
		} else {
			lines = append(lines,
				m.theme.Success.Render("✓ Management cluster destroy finished"),
				"",
				"Mode  "+m.modeLabel(),
				m.theme.Muted.Render("      "+m.cli()),
			)
		}
		if len(m.steps) > 0 {
			lines = append(lines, "", m.theme.Muted.Render("Steps"))
			for _, step := range m.steps {
				lines = append(lines, "  "+step)
			}
		}
		lines = append(lines, "",
			m.theme.Muted.Render("Equivalent CLI"),
			"  "+m.cli(),
		)
		return page.Panel(m.theme, "Complete", lines, width, 14, true), "esc welcome · ctrl+c quit"
	default:
		intro := []string{
			m.theme.Muted.Render("Destroys your management cluster and any apps installed on it."),
			m.theme.Muted.Render("Same as plural down — requires workspace.yaml in the current repo."),
			"",
		}
		lines := append(intro, m.cloudLines(width)...)
		if m.err != nil {
			lines = append(lines, "", m.theme.Danger.Render(m.err.Error()))
		}
		help := "↑/↓ · 1–2 / s/c · enter · esc welcome"
		if width < 100 {
			help = "↑/↓ · enter · esc welcome"
		}
		return page.Panel(m.theme, "Destroy mode", lines, width, 12, true), help
	}
}

func (m Model) cloudLines(width int) []string {
	opts := cloudOptions()
	var lines []string
	for i, o := range opts {
		prefix := "  "
		label := fmt.Sprintf("%d  %s   %-14s  %s", i+1, cloudShortcut(o.id), o.title, o.blurb)
		if i == m.cursor {
			prefix = "› "
			label = m.theme.Title.Render(label)
		} else {
			label = m.theme.Body.Render(fmt.Sprintf("%d  %s   ", i+1, cloudShortcut(o.id))) +
				m.theme.Body.Render(fmt.Sprintf("%-14s  ", o.title)) +
				m.theme.Muted.Render(o.blurb)
		}
		line := prefix + label
		lines = append(lines, ansi.Truncate(line, max(1, width-4), "…"))
	}
	return lines
}

func (m Model) affirmLines(width int) []string {
	opts := affirmOptions()
	var lines []string
	for i, o := range opts {
		prefix := "  "
		label := fmt.Sprintf("%s  %s", o.title, o.blurb)
		if i == m.cursor {
			prefix = "› "
			label = m.theme.Title.Render(o.title) + "  " + m.theme.Muted.Render(o.blurb)
		} else {
			label = m.theme.Body.Render(o.title) + "  " + m.theme.Muted.Render(o.blurb)
		}
		lines = append(lines, ansi.Truncate(prefix+label, max(1, width-4), "…"))
	}
	return lines
}

func (m Model) modeLabel() string {
	if m.cloud {
		return "Plural Cloud"
	}
	return "Self-hosted"
}

func (m Model) cli() string {
	if m.cloud {
		return "plural down --cloud"
	}
	return "plural down"
}
