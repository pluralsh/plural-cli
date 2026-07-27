// Package commandbar provides the reusable command input shown at the bottom
// of TUI screens.
package commandbar

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/samber/lo"

	"github.com/pluralsh/plural-cli/tui/theme"
)

const (
	minimumWidth     = 12
	title            = "Command"
	popupTitle       = "Available commands"
	maximumPopupRows = 6
)

type keyAction uint8

const (
	keyActionNone keyAction = iota
	keyActionNextSuggestion
	keyActionPreviousSuggestion
	keyActionSubmit
	keyActionDismiss
)

var keyActionKeystrokes = map[keyAction]string{
	keyActionNextSuggestion:     "down",
	keyActionPreviousSuggestion: "up",
	keyActionSubmit:             "enter",
	keyActionDismiss:            "esc",
}

func actionForKeystroke(keystroke string) keyAction {
	for action, candidate := range keyActionKeystrokes {
		if keystroke == candidate {
			return action
		}
	}

	return keyActionNone
}

// Model owns command entry, completion, selection, and rendering.
type Model struct {
	theme       theme.Theme
	input       textinput.Model
	selected    string
	suggestions []string
	popupOpen   bool
	popupCursor int
}

// SubmittedMsg is emitted when the user submits a command. The shell or
// screen decides what the command means; the input component only owns entry.
type SubmittedMsg struct{ Command string }

// New creates a focused command bar with the provided completion candidates.
func New(t theme.Theme, suggestions []string) Model {
	input := textinput.New()
	input.Prompt = "› "
	input.Placeholder = "Search commands…"
	input.CharLimit = 80
	input.ShowSuggestions = true
	input.SetSuggestions(suggestions)
	input.SetVirtualCursor(true)

	styles := textinput.DefaultDarkStyles()
	styles.Focused.Text = t.Body
	styles.Focused.Prompt = t.Title
	styles.Focused.Placeholder = t.Muted
	styles.Focused.Suggestion = t.Muted
	styles.Blurred = styles.Focused
	styles.Cursor.Color = t.Colors.Primary
	styles.Cursor.Shape = tea.CursorBar
	styles.Cursor.Blink = false
	input.SetStyles(styles)
	input.Focus()

	return Model{theme: t, input: input, suggestions: suggestions}
}

// Update handles completion, selection, clearing, and text entry.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch actionForKeystroke(key.Keystroke()) {
		case keyActionNextSuggestion:
			matches := m.filteredSuggestions()
			if len(matches) == 0 {
				return m, nil
			}
			if m.popupOpen {
				m.popupCursor = (m.popupCursor + 1) % len(matches)
			} else {
				m.popupOpen = true
				m.popupCursor = 0
			}
			return m, nil
		case keyActionPreviousSuggestion:
			matches := m.filteredSuggestions()
			if len(matches) == 0 {
				return m, nil
			}
			if m.popupOpen {
				m.popupCursor = (m.popupCursor - 1 + len(matches)) % len(matches)
			} else {
				m.popupOpen = true
				m.popupCursor = len(matches) - 1
			}
			return m, nil
		case keyActionSubmit:
			if m.popupOpen {
				matches := m.filteredSuggestions()
				if len(matches) > 0 {
					m.popupCursor = min(m.popupCursor, len(matches)-1)
					m.selected = matches[m.popupCursor]
					m.input.SetValue(m.selected)
				}
				m.popupOpen = false
			} else {
				m.selected = lo.CoalesceOrEmpty(strings.TrimSpace(m.input.Value()), m.input.CurrentSuggestion())
			}
			if m.selected == "" {
				return m, nil
			}
			selected := m.selected
			return m, func() tea.Msg { return SubmittedMsg{Command: selected} }
		case keyActionDismiss:
			if m.popupOpen {
				m.popupOpen = false
				return m, nil
			}
			m.input.Reset()
			m.selected = ""
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	if matches := m.filteredSuggestions(); len(matches) == 0 {
		m.popupCursor = 0
	} else {
		m.popupCursor = min(m.popupCursor, len(matches)-1)
	}
	return m, cmd
}

// Selected returns the most recently selected command.
func (m Model) Selected() string { return m.selected }

// Value returns the current command input value.
func (m Model) Value() string { return m.input.Value() }

// CurrentSuggestion returns the active completion candidate.
func (m Model) CurrentSuggestion() string { return m.input.CurrentSuggestion() }

// View renders the framed input and its contextual key help.
func (m Model) View(width int) string {
	width = max(width, minimumWidth)
	input := m.input
	input.SetWidth(max(1, width-7))

	help := "tab complete · ↑/↓ suggestions · enter select · esc clear · ctrl+c quit"
	if m.popupOpen {
		help = "↑/↓ choose · enter open · esc close · type to filter · ctrl+c quit"
	} else if m.selected != "" {
		help = "Opening “" + m.selected + "”…"
	}
	help = m.theme.Muted.Render(ansi.Truncate(help, max(1, width-2), "…"))

	command := renderBox(input.View(), width) + "\n  " + help
	if !m.popupOpen {
		return command
	}
	return m.renderPopup(width) + "\n" + command
}

func (m Model) filteredSuggestions() []string {
	query := strings.ToLower(strings.TrimSpace(m.input.Value()))
	if query == "" {
		return m.suggestions
	}
	result := make([]string, 0, len(m.suggestions))
	for _, suggestion := range m.suggestions {
		if strings.Contains(strings.ToLower(suggestion), query) {
			result = append(result, suggestion)
		}
	}
	return result
}

func (m Model) renderPopup(width int) string {
	matches := m.filteredSuggestions()
	rows := min(maximumPopupRows, len(matches))
	popupWidth := min(width, 38)
	innerWidth := popupWidth - 4
	title := ansi.Truncate(popupTitle, max(1, popupWidth-5), "…")
	rule := strings.Repeat("─", max(0, popupWidth-5-lipgloss.Width(title)))
	lines := []string{"╭─ " + title + " " + rule + "╮"}
	start := 0
	if m.popupCursor >= rows {
		start = m.popupCursor - rows + 1
	}
	for i := 0; i < rows; i++ {
		index := start + i
		line := "  " + matches[index]
		if index == m.popupCursor {
			line = m.theme.Title.Render("› " + matches[index])
		}
		line = ansi.Truncate(line, innerWidth, "…")
		lines = append(lines, "│ "+line+strings.Repeat(" ", max(0, innerWidth-lipgloss.Width(line)))+" │")
	}
	lines = append(lines, "╰"+strings.Repeat("─", popupWidth-2)+"╯")
	return strings.Join(lines, "\n")
}

// renderBox draws the frame directly so text input escape sequences remain on
// one line and its width stays predictable.
func renderBox(line string, width int) string {
	innerWidth := width - 4
	topRule := strings.Repeat("─", max(0, width-5-lipgloss.Width(title)))
	top := "╭─ " + title + " " + topRule + "╮"

	line = ansi.Truncate(line, innerWidth, "…")
	body := "│ " + line + strings.Repeat(" ", max(0, innerWidth-lipgloss.Width(line))) + " │"
	bottom := "╰" + strings.Repeat("─", width-2) + "╯"
	return strings.Join([]string{top, body, bottom}, "\n")
}
