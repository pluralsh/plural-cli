package ai

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"

	aibridge "github.com/pluralsh/plural-cli/pkg/bridge/ai"
	"github.com/pluralsh/plural-cli/tui/components/page"
)

func (m Model) View(width, height int) string {
	width, height = page.Size(width, height)
	if width < page.MinimumWidth || height < page.MinimumHeight {
		return page.Unsupported(m.theme, width, height)
	}
	contentWidth := page.ContentWidth(width)
	body, help := m.bodyAndHelp(contentWidth, height)
	return page.Render(m.theme, width, height, m.title(), m.headerStatus(), body, help)
}

func (m Model) title() string {
	if m.mode == modeChat {
		return "AI · Chat"
	}
	return "AI"
}

func (m Model) headerStatus() string {
	if m.mode != modeChat {
		return m.theme.Success.Render(fmt.Sprintf("%d commands", len(items)))
	}
	if m.thinking {
		return m.theme.Warning.Render("◌ thinking")
	}
	if m.needsAuth {
		return m.theme.Warning.Render("○ connect App")
	}
	if m.err != nil {
		return m.theme.Danger.Render("✗ failed")
	}
	return m.theme.Success.Render(fmt.Sprintf("%d messages", len(m.history)))
}

func (m Model) bodyAndHelp(width, height int) (string, string) {
	if m.mode == modeChat {
		return m.chatBodyAndHelp(width, height)
	}
	lines := make([]string, 0, 2+len(items)+2)
	lines = append(lines, m.theme.Muted.Render("Choose Chat, Agents, or Workbenches."), "")
	for i, item := range items {
		cursor := "  "
		if i == m.cursor {
			cursor = "› "
		}
		lines = append(lines, cursor+fmt.Sprintf("%s  %-12s %s", item.number, item.title, item.blurb))
	}
	lines = append(lines, "", m.theme.Muted.Render("Chat uses Plural App; Agents and Workbenches use Console."))
	return page.Panel(m.theme, "AI workspaces", lines, width, 10, true), "↑/↓ select · enter open · 1-3 shortcut · esc back"
}

func (m Model) chatBodyAndHelp(width, height int) (string, string) {
	if m.needsAuth {
		lines := []string{
			m.theme.Warning.Render("○ Plural App is not connected"),
			m.theme.Muted.Render("  Chat uses your app.plural.sh token, not Console."),
			"",
			m.theme.Body.Render("Press c to open Access."),
		}
		return page.Panel(m.theme, "App login required", lines, width, 8, true), "c connect · esc hub"
	}
	inputH := 5
	convH := max(8, height-inputH-6)
	inner := max(1, convH-2)
	lines := m.transcriptLines(width - 4)
	if m.thinking {
		lines = append(lines, m.theme.Warning.Render("◌ Thinking…"))
	}
	if m.err != nil {
		lines = append(lines, m.theme.Danger.Render(m.err.Error()))
	}
	offset := conversationOffset(len(lines), inner, m.chatFromEnd)
	end := min(len(lines), offset+inner)
	visible := lines[offset:end]
	conversation := page.Panel(m.theme, "Conversation", visible, width, convH, true)
	input := m.input
	input.SetWidth(max(8, width-8))
	message := page.Panel(m.theme, "Message", []string{input.View()}, width, inputH, true)
	help := "enter send · ctrl+n new · pgup/pgdn · esc hub"
	if m.thinking {
		help = "esc cancel · ctrl+c cancel"
	}
	if width < 100 {
		help = "enter send · ctrl+n · pgup/pgdn · esc"
		if m.thinking {
			help = "esc/ctrl+c cancel"
		}
	}
	return conversation + "\n" + message, help
}

func (m Model) transcriptLines(innerWidth int) []string {
	if innerWidth < 1 {
		innerWidth = 1
	}
	lines := make([]string, 0, len(m.history)*4)
	for i, message := range m.history {
		if i > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, m.speaker(message.Role))
		content := strings.TrimSpace(message.Content)
		if content == "" {
			continue
		}
		lines = append(lines, strings.Split(ansi.Wrap(content, innerWidth, ""), "\n")...)
	}
	return lines
}

func (m Model) speaker(role string) string {
	if strings.EqualFold(role, aibridge.RoleUser) {
		return m.theme.Title.Render("You")
	}
	return m.theme.Success.Render("Plural AI")
}

func conversationOffset(total, inner, fromEnd int) int {
	maxOff := max(0, total-inner)
	if fromEnd <= 0 {
		return maxOff
	}
	return max(0, maxOff-fromEnd)
}

func normalizeView(view string) string {
	lines := strings.Split(ansi.Strip(view), "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " ")
	}
	return strings.Join(lines, "\n")
}
