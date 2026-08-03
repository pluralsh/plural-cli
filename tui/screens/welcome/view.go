package welcome

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/samber/lo"

	"github.com/pluralsh/plural-cli/tui/assets"
)

const (
	defaultWidth       = 100
	defaultHeight      = 24
	minimumWidth       = 80
	minimumHeight      = 24
	sideMargin         = 2
	minimumVerticalGap = 1
	defaultVerticalGap = 2
	heroDetailRows     = 7
	logoRailWidth      = 20
	logoIndent         = 4
)

func (m Model) View(width, height int) string {
	width, height = m.viewportSize(width, height)
	if width < minimumWidth || height < minimumHeight {
		return m.renderUnsupportedTerminal(width, height)
	}

	contentWidth := width - 2*sideMargin
	version := lo.CoalesceOrEmpty(m.snapshot.Version, "dev")
	header := m.renderHero(contentWidth, version)
	groups := m.renderGroups(contentWidth)
	gap := m.verticalGap(height, header, groups)

	return m.indent(header+strings.Repeat("\n", gap)+groups, sideMargin)
}

func (m Model) renderGroups(width int) string {
	border := m.primaryBorder()
	title := "Choose an area"
	if m.helpOpen {
		title = "Help"
	}
	topRule := max(1, width-5-lipgloss.Width(title))
	top := border.Render("╭─ " + title + " " + strings.Repeat("─", topRule) + "╮")
	bottom := border.Render("╰" + strings.Repeat("─", width-2) + "╯")

	innerWidth := width - 4
	var body []string
	if m.helpOpen {
		body = []string{
			m.theme.Body.Render("1–6 / letter opens an area"),
			m.theme.Muted.Render("↑/↓ move · enter confirm · esc close help"),
			m.theme.Muted.Render("ctrl+c quit"),
			"",
			m.theme.Muted.Render("More docs will land in a later phase."),
		}
	} else {
		for i, g := range m.groups {
			prefix := "  "
			label := fmt.Sprintf("%s  %s   %-18s  %s", g.number, g.shortcut, g.title, g.blurb)
			if i == m.cursor {
				prefix = "› "
				label = m.theme.Title.Render(label)
			} else {
				label = m.theme.Body.Render(fmt.Sprintf("%s  %s   ", g.number, g.shortcut)) +
					m.theme.Body.Render(fmt.Sprintf("%-18s  ", g.title)) +
					m.theme.Muted.Render(g.blurb)
			}
			line := prefix + label
			line = ansi.Truncate(line, innerWidth, "…")
			body = append(body, line)
		}
		body = append(body, "", m.theme.Muted.Render("Setup · Agents · Develop — later"))
	}

	rows := make([]string, 0, len(body)+2)
	rows = append(rows, top)
	for _, line := range body {
		padded := line + strings.Repeat(" ", max(0, innerWidth-lipgloss.Width(line)))
		rows = append(rows, border.Render("│")+" "+padded+" "+border.Render("│"))
	}
	rows = append(rows, bottom)

	help := m.theme.Muted.Render(ansi.Truncate("1–6 open · letter shortcut · ↑/↓ · enter · ctrl+c quit", max(1, width-2), "…"))
	if m.helpOpen {
		help = m.theme.Muted.Render(ansi.Truncate("any key closes help · ctrl+c quit", max(1, width-2), "…"))
	}
	return strings.Join(rows, "\n") + "\n  " + help
}

func (m Model) viewportSize(width, height int) (int, int) {
	if width <= 0 {
		width = defaultWidth
	}

	if height <= 0 {
		height = defaultHeight
	}

	return width, height
}

func (m Model) verticalGap(height int, blocks ...string) int {
	if height <= 0 {
		return defaultVerticalGap
	}

	occupied := 0
	for _, block := range blocks {
		occupied += lipgloss.Height(block)
	}

	// Joining two blocks with newlines consumes one fewer row than summing
	// their individual heights.
	return max(minimumVerticalGap, height-occupied+len(blocks)-1)
}

func (m Model) indent(content string, width int) string {
	padding := strings.Repeat(" ", width)
	return padding + strings.ReplaceAll(content, "\n", "\n"+padding)
}

func (m Model) renderUnsupportedTerminal(width, height int) string {
	detected := "Unsupported terminal size: " + m.dimensions(width, height)
	required := "Minimum supported size: " + m.dimensions(minimumWidth, minimumHeight)

	if height < 4 {
		message := "Unsupported " + m.dimensions(width, height) + " · minimum " + m.dimensions(minimumWidth, minimumHeight)
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, ansi.Truncate(message, width, "…"))
	}

	content := strings.Join([]string{
		m.theme.Title.Render(ansi.Truncate("Plural", width, "…")),
		"",
		m.theme.Body.Render(ansi.Truncate(detected, width, "…")),
		m.theme.Muted.Render(ansi.Truncate(required, width, "…")),
	}, "\n")
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

func (m Model) dimensions(width, height int) string { return fmt.Sprintf("%d×%d", width, height) }

// renderHero draws the full welcome treatment. The logo rail remains fixed
// while the connection details consume all additional width.
func (m Model) renderHero(width int, version string) string {
	border := m.primaryBorder()
	top := m.renderHeroTop(width, border)
	rightWidth := width - 2 - logoRailWidth - 1
	rightLines := m.heroDetails(rightWidth - 2)
	leftLines := m.logoRail(len(rightLines))

	rows := make([]string, 0, len(rightLines)+2)
	rows = append(rows, top)
	for i, rightLine := range rightLines {
		left := m.fit(leftLines[i], logoRailWidth)
		right := m.fit(rightLine, rightWidth-2)
		rows = append(rows, border.Render("│")+left+border.Render("│")+" "+right+" "+border.Render("│"))
	}

	rows = append(rows, m.renderVersionFooter(width, version, border, m.theme.Muted))
	return strings.Join(rows, "\n")
}

func (m Model) renderHeroTop(width int, border lipgloss.Style) string {
	const trailingRuleWidth = 1

	title := m.theme.Title.Render("Plural")
	titleRail := "─ " + title + " "
	identityWidth := min(44, max(16, width-lipgloss.Width(titleRail)-trailingRuleWidth-7))
	identity := m.heroIdentity(identityWidth)
	identityRail := " " + identity + " " + strings.Repeat("─", trailingRuleWidth)
	middleRuleWidth := max(1, width-2-lipgloss.Width(titleRail)-lipgloss.Width(identityRail))

	return border.Render("╭─ ") + title + border.Render(" "+strings.Repeat("─", middleRuleWidth)) +
		" " + identity + " " + border.Render(strings.Repeat("─", trailingRuleWidth)+"╮")
}

func (m Model) renderVersionFooter(width int, version string, border, versionStyle lipgloss.Style) string {
	ruleWidth := max(1, width-lipgloss.Width(version)-5)
	return border.Render("╰─ ") + versionStyle.Render(version) +
		border.Render(" "+strings.Repeat("─", ruleWidth)+"╯")
}

func (m Model) logoRail(rowCount int) []string {
	rows := make([]string, rowCount)
	for i, line := range strings.Split(strings.TrimSpace(assets.Logo), "\n") {
		row := i + 1
		if row >= rowCount {
			break
		}
		rows[row] = strings.Repeat(" ", logoIndent) + m.theme.Logo.Render(line)
	}
	return rows
}

func (m Model) heroIdentity(maxWidth int) string {
	if !m.snapshot.App.Configured {
		return m.status(false) + " App not connected"
	}
	profile := lo.CoalesceOrEmpty(m.snapshot.App.Name, "default")
	email := lo.CoalesceOrEmpty(m.snapshot.App.Email, m.snapshot.App.Name, "saved account")
	display := ansi.Truncate(profile+" · "+email, max(1, maxWidth-2), "…")
	return m.status(true) + " " + display
}

func (m Model) heroConsole() string {
	if !m.snapshot.Console.Configured {
		return m.status(false) + " Console not connected"
	}
	return m.status(true) + m.theme.Title.Render(" Console")
}

func (m Model) heroDetails(maxWidth int) []string {
	if m.loading {
		return m.padLines([]string{
			m.theme.Body.Render("Local context"),
			m.spinner.View() + " " + m.theme.Muted.Render("Loading…"),
		}, heroDetailRows)
	}
	if m.err != nil {
		return m.padLines([]string{
			m.theme.Body.Render("Local context"),
			m.theme.Danger.Render(m.err.Error()),
		}, heroDetailRows)
	}

	lines := []string{
		m.heroConsole(),
		m.heroConsoleURL(maxWidth),
		m.theme.Title.Render(strings.Repeat("─", max(1, maxWidth))),
	}
	lines = append(lines, m.workspaceDetails(maxWidth)...)
	return m.padLines(lines, heroDetailRows)
}

func (m Model) workspaceDetails(maxWidth int) []string {
	workspace := m.snapshot.Workspace
	if !workspace.Configured {
		return []string{m.status(false) + " No workspace detected"}
	}

	name := lo.CoalesceOrEmpty(workspace.Name, m.filepathBase(workspace.Path))
	prefix := m.status(true) + m.theme.Title.Render(" Workspace ") + "· " + name + " · "
	pathWidth := max(1, maxWidth-lipgloss.Width(prefix))
	if maxWidth < 60 {
		pathWidth = min(pathWidth, 14)
	}
	provider := lo.CoalesceOrEmpty(strings.Join(lo.Compact([]string{workspace.Provider, workspace.Region}), " · "), "—")

	return []string{
		prefix + m.theme.Muted.Render(ansi.Truncate(workspace.Path, pathWidth, "...")),
		"Provider  " + provider,
		"Owner     " + lo.CoalesceOrEmpty(workspace.Owner, "—"),
		"",
	}
}

func (m Model) heroConsoleURL(maxWidth int) string {
	if !m.snapshot.Console.Configured {
		return ""
	}
	prefix := "URL       "
	if lipgloss.Width(prefix)+lipgloss.Width(m.snapshot.Console.URL) > maxWidth {
		prefix = "URL "
	}
	return prefix + m.url(m.snapshot.Console.URL, max(1, maxWidth-lipgloss.Width(prefix)))
}

func (m Model) status(ok bool) string {
	if m.theme.Color {
		if ok {
			return m.theme.Success.Render("●")
		}
		return m.theme.Danger.Render("●")
	}
	if ok {
		return "✓"
	}
	return "✗"
}

func (m Model) url(target string, maxWidth int) string {
	if target == "" {
		return m.theme.Muted.Render("unknown")
	}
	display := ansi.Truncate(target, max(1, maxWidth), "…")
	style := m.theme.Link.Inline(true)
	if m.theme.Hyperlinks {
		style = style.Hyperlink(target)
	}
	return style.Render(display)
}

func (m Model) primaryBorder() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(m.theme.Colors.Primary)
}

// logo is deliberately static. Animation is reserved for the compact spinner
// used while work is in progress, never for the welcome-screen brand mark.
func (m Model) logo() string {
	return m.theme.Logo.Render(strings.TrimSpace(assets.Logo))
}

func (m Model) fit(value string, width int) string {
	value = ansi.Truncate(value, max(1, width), "…")
	return value + strings.Repeat(" ", max(0, width-lipgloss.Width(value)))
}

func (m Model) padLines(lines []string, count int) []string {
	if len(lines) >= count {
		return lines
	}
	return append(lines, make([]string, count-len(lines))...)
}

func (m Model) filepathBase(path string) string {
	path = strings.TrimRight(path, "/\\")
	if i := strings.LastIndexAny(path, "/\\"); i >= 0 {
		return path[i+1:]
	}
	return path
}
