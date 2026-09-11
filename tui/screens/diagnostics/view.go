package diagnostics

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/pluralsh/plural-cli/tui/components/page"
)

func (m Model) View(width, height int) string {
	width, height = page.Size(width, height)
	if width < page.MinimumWidth || height < page.MinimumHeight {
		return page.Unsupported(m.theme, width, height)
	}
	contentWidth := page.ContentWidth(width)
	status := m.theme.Success.Render("✓ local context ready")
	if m.loading {
		status = m.theme.Warning.Render("◌ loading")
	}
	if m.err != nil {
		status = m.theme.Danger.Render("✗ load failed")
	}

	contextLines, checkLines := m.viewLines()
	body := page.Panel(m.theme, "Local context", contextLines, contentWidth, 8, false) + "\n\n" +
		page.Panel(m.theme, "Checks", checkLines, contentWidth, 5, len(m.snapshot.Diagnostics) > 0 || m.err != nil)
	return page.Render(m.theme, width, height, "Diagnostics", status, body, "r refresh · esc back · ctrl+c quit")
}

func (m Model) viewLines() ([]string, []string) {
	if m.loading {
		return []string{m.theme.Warning.Render("◌ Loading credential-free local context…")}, []string{m.theme.Muted.Render("Checks begin after local context loads.")}
	}
	if m.err != nil {
		return []string{m.theme.Danger.Render("✗ Unable to read local context")}, []string{m.theme.Danger.Render("Error  " + m.err.Error()), m.theme.Muted.Render("Press r to retry.")}
	}
	contextLines := []string{
		m.contextLine("Plural App", m.snapshot.App.Configured, m.snapshot.App.Email),
		m.contextLine("Console", m.snapshot.Console.Configured, m.snapshot.Console.URL),
		m.contextLine("Workspace", m.snapshot.Workspace.Configured, m.snapshot.Workspace.Path),
		m.contextLine("Kubernetes", m.snapshot.KubeContext != "", m.snapshot.KubeContext),
	}
	checks := []string{m.theme.Success.Render("✓ No local diagnostics reported")}
	if len(m.snapshot.Diagnostics) > 0 {
		checks = make([]string, 0, len(m.snapshot.Diagnostics))
		for _, diagnostic := range m.snapshot.Diagnostics {
			checks = append(checks, m.theme.Warning.Render("! WARN")+"  "+diagnostic)
		}
	}
	return contextLines, checks
}

func (m Model) contextLine(label string, configured bool, detail string) string {
	status := m.theme.Warning.Render("○ NOT CONFIGURED")
	if configured {
		status = m.theme.Success.Render("✓ OK")
	}
	if detail == "" {
		detail = "—"
	}
	label += strings.Repeat(" ", max(1, 12-len(label)))
	status += strings.Repeat(" ", max(1, 16-lipgloss.Width(status)))
	return label + " " + status + " " + detail
}
