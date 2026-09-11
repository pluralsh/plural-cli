package page

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"

	"github.com/pluralsh/plural-cli/tui/theme"
)

func TestRenderAnchorsSharedChromeAndHelp(t *testing.T) {
	theme := theme.New(colorprofile.ASCII)
	body := Panel(theme, "Content", []string{"one", "two"}, ContentWidth(80), 6, true)
	view := ansi.Strip(Render(theme, 80, 24, "Screen", "✓ ready", body, "esc back"))
	lines := strings.Split(view, "\n")
	if len(lines) != 24 {
		t.Fatalf("height = %d, want 24", len(lines))
	}
	if !strings.HasPrefix(lines[0], "  Plural  Screen") || !strings.Contains(lines[len(lines)-1], "esc back") {
		t.Fatalf("shared chrome is incomplete:\n%s", view)
	}
	for _, line := range lines {
		if got := lipgloss.Width(line); got > 80 {
			t.Fatalf("line width %d exceeds 80: %q", got, line)
		}
	}
}

func TestUnsupportedUsesWelcomeMinimum(t *testing.T) {
	view := ansi.Strip(Render(theme.New(colorprofile.ASCII), 79, 23, "Screen", "", "", ""))
	if !strings.Contains(view, "79×23") || !strings.Contains(view, "80×24") {
		t.Fatalf("unsupported dimensions missing:\n%s", view)
	}
}
