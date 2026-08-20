package commandbar

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"

	"github.com/pluralsh/plural-cli/tui/theme"
)

func TestCompletesAndSubmitsSelection(t *testing.T) {
	model := New(theme.New(colorprofile.ASCII), []string{"access", "diagnostics"})
	model, _ = model.Update(tea.KeyPressMsg{Code: 'd', Text: "d"})
	if got := model.CurrentSuggestion(); got != "diagnostics" {
		t.Fatalf("suggestion = %q, want diagnostics", got)
	}

	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if got := model.Value(); got != "diagnostics" {
		t.Fatalf("completed value = %q, want diagnostics", got)
	}

	var cmd tea.Cmd
	model, cmd = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("selecting a command did not emit submission")
	}
	if got := cmd().(SubmittedMsg).Command; got != "diagnostics" {
		t.Fatalf("submitted command = %q, want diagnostics", got)
	}
	if got := model.Selected(); got != "diagnostics" {
		t.Fatalf("selected command = %q, want diagnostics", got)
	}
}

func TestViewIsStandaloneAndWidthBounded(t *testing.T) {
	model := New(theme.New(colorprofile.ASCII), []string{"access"})
	view := model.View(76)
	lines := strings.Split(ansi.Strip(view), "\n")
	if len(lines) != 4 {
		t.Fatalf("view height = %d, want 4", len(lines))
	}
	if !strings.Contains(lines[0], "Command") || !strings.Contains(lines[3], "ctrl+c quit") {
		t.Fatalf("command bar is incomplete:\n%s", view)
	}
	for _, line := range lines {
		if got := lipgloss.Width(line); got > 76 {
			t.Fatalf("line width %d exceeds 76: %q", got, line)
		}
	}
}

func TestArrowKeysOpenAndSelectCommandPopup(t *testing.T) {
	model := New(theme.New(colorprofile.ASCII), []string{"access", "diagnostics", "profiles"})
	model, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if cmd != nil || !model.popupOpen || model.popupCursor != 0 {
		t.Fatalf("first down did not open popup: %#v", model)
	}
	view := ansi.Strip(model.View(76))
	if !strings.Contains(view, "Available commands") || !strings.Contains(view, "› access") || !strings.Contains(view, "  diagnostics") {
		t.Fatalf("command popup is incomplete:\n%s", view)
	}

	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	model, cmd = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("popup selection did not submit")
	}
	if got := cmd().(SubmittedMsg).Command; got != "diagnostics" {
		t.Fatalf("submitted command = %q, want diagnostics", got)
	}
	if model.popupOpen {
		t.Fatal("popup remained open after submit")
	}
}

func TestCommandPopupFiltersAndEscClosesBeforeClearing(t *testing.T) {
	model := New(theme.New(colorprofile.ASCII), []string{"access", "diagnostics", "profiles"})
	model, _ = model.Update(tea.KeyPressMsg{Code: 'p', Text: "p"})
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	view := ansi.Strip(model.View(76))
	if !strings.Contains(view, "› profiles") || strings.Contains(view, "diagnostics") {
		t.Fatalf("filtered popup is incorrect:\n%s", view)
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if model.popupOpen || model.Value() != "p" {
		t.Fatalf("first esc should only close popup: open=%v value=%q", model.popupOpen, model.Value())
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if model.Value() != "" {
		t.Fatalf("second esc did not clear input: %q", model.Value())
	}
}
