// Package page provides shared routed-screen chrome: the same two-cell gutter,
// semantic rule, framed surfaces, responsive minimum, and bottom-anchored key
// help used throughout the TUI.
package page

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/pluralsh/plural-cli/tui/theme"
)

const (
	DefaultWidth  = 100
	DefaultHeight = 30
	MinimumWidth  = 80
	MinimumHeight = 24
	SideMargin    = 2
)

// Size applies deterministic defaults for model tests and initial renders.
func Size(width, height int) (int, int) {
	if width <= 0 {
		width = DefaultWidth
	}
	if height <= 0 {
		height = DefaultHeight
	}
	return width, height
}

// ContentWidth returns the width available inside the shared side gutters.
func ContentWidth(width int) int { return max(1, width-2*SideMargin) }

// Render composes routed-screen content and anchors help to the final row.
func Render(t theme.Theme, width, height int, title, status, body, help string) string {
	width, height = Size(width, height)
	if width < MinimumWidth || height < MinimumHeight {
		return Unsupported(t, width, height)
	}
	contentWidth := ContentWidth(width)
	header := renderHeader(t, contentWidth, title, status)
	occupied := lipgloss.Height(header) + 1 + lipgloss.Height(body)
	separation := max(2, height-occupied)
	help = t.Muted.Render(ansi.Truncate(help, contentWidth, "…"))
	content := header + "\n\n" + body + strings.Repeat("\n", separation) + help
	return indent(content, SideMargin)
}

// Panel renders a fixed-height semantic surface. Content is truncated rather
// than allowed to push key help off-screen.
func Panel(t theme.Theme, title string, lines []string, width, height int, focused bool) string {
	width = max(8, width)
	height = max(3, height)
	innerWidth := width - 4
	border := lipgloss.NewStyle().Foreground(t.Colors.Border)
	if focused {
		border = lipgloss.NewStyle().Foreground(t.Colors.Primary)
	}
	styledTitle := t.Body.Render(title)
	if focused {
		styledTitle = t.Title.Render("› " + title)
	}
	ruleWidth := max(1, width-lipgloss.Width(styledTitle)-5)
	result := []string{border.Render("╭─ ") + styledTitle + border.Render(" "+strings.Repeat("─", ruleWidth)+"╮")}
	visible := height - 2
	for i := 0; i < visible; i++ {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		if i == visible-1 && len(lines) > visible {
			line = t.Muted.Render("…")
		}
		line = ansi.Truncate(line, innerWidth, "…")
		result = append(result, border.Render("│")+" "+line+strings.Repeat(" ", max(0, innerWidth-lipgloss.Width(line)))+" "+border.Render("│"))
	}
	result = append(result, border.Render("╰"+strings.Repeat("─", width-2)+"╯"))
	return strings.Join(result, "\n")
}

func Unsupported(t theme.Theme, width, height int) string {
	message := "Unsupported terminal size: " + dimensions(width, height) + " · minimum " + dimensions(MinimumWidth, MinimumHeight)
	message = t.Body.Render(ansi.Truncate(message, max(1, width), "…"))
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, message)
}

func renderHeader(t theme.Theme, width int, title, status string) string {
	left := t.Title.Render("Plural") + "  " + t.Body.Render(title)
	status = ansi.Truncate(status, max(0, width-lipgloss.Width(left)-2), "…")
	gap := strings.Repeat(" ", max(1, width-lipgloss.Width(left)-lipgloss.Width(status)))
	line := ansi.Truncate(left+gap+status, width, "…")
	return line + "\n" + lipgloss.NewStyle().Foreground(t.Colors.Primary).Render(strings.Repeat("─", width))
}

func indent(content string, width int) string {
	padding := strings.Repeat(" ", width)
	return padding + strings.ReplaceAll(content, "\n", "\n"+padding)
}
func dimensions(width, height int) string { return fmt.Sprintf("%d×%d", width, height) }
