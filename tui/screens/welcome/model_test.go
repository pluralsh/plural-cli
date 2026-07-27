package welcome

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	bridge "github.com/pluralsh/plural-cli/pkg/bridge/welcome"
	"github.com/pluralsh/plural-cli/tui/assets"
	"github.com/pluralsh/plural-cli/tui/navigation"
	"github.com/pluralsh/plural-cli/tui/theme"
)

func TestReadOnlyWelcomeGoldens(t *testing.T) {
	snapshot := bridge.Snapshot{
		Version: "v0.13.0",
		App: bridge.AppProfile{
			Configured: true, Name: "personal", Email: "alex@acme.io",
			Endpoint: "https://app.plural.sh", SavedProfiles: 2,
		},
		Console: bridge.ConsoleConnection{Configured: true, URL: "https://console.acme.io"},
		Workspace: bridge.Workspace{
			Configured: true, Path: "/work/path/to/a/very/long/workspace", Name: "plrl-dev-aws",
			Project: "acme", Provider: "aws", Region: "eu-west-1", Owner: "sebastian@plural.sh",
		},
		KubeContext: "plural-platform-prod",
	}

	for _, width := range []int{80, 120} {
		t.Run(strconv.Itoa(width), func(t *testing.T) {
			model := New(t.Context(), nil, theme.New(colorprofile.ASCII))
			model, _ = model.Update(loadedMsg{snapshot: snapshot})
			height := 24
			if width == 120 {
				height = 30
			}
			got := normalizeView(model.View(width, height))
			golden := filepath.Join("testdata", "welcome-"+strconv.Itoa(width)+".golden")
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("read golden: %v\nactual:\n%s", err, got)
			}
			if got != strings.TrimSuffix(string(want), "\n") {
				t.Fatalf("view changed\nwant:\n%s\n\ngot:\n%s", want, got)
			}
			if strings.Contains(got, "secret-token") {
				t.Fatal("welcome view exposed a credential")
			}
		})
	}
}

func TestWelcomeCommandPopupGolden(t *testing.T) {
	model := New(t.Context(), nil, theme.New(colorprofile.ASCII))
	model, _ = model.Update(loadedMsg{snapshot: bridge.Snapshot{
		Version:   "v0.13.0",
		App:       bridge.AppProfile{Configured: true, Name: "personal", Email: "alex@acme.io"},
		Console:   bridge.ConsoleConnection{Configured: true, URL: "https://console.acme.io"},
		Workspace: bridge.Workspace{Configured: true, Path: "/work/plural", Name: "plrl-dev-aws", Provider: "aws", Region: "eu-west-1", Owner: "alex@acme.io"},
	}})
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	got := normalizeView(model.View(80, 24))
	want, err := os.ReadFile(filepath.Join("testdata", "welcome-popup-80.golden"))
	if err != nil {
		t.Fatalf("read golden: %v\nactual:\n%s", err, got)
	}
	if got != strings.TrimSuffix(string(want), "\n") {
		t.Fatalf("view changed\nwant:\n%s\n\ngot:\n%s", want, got)
	}
	if lines := strings.Split(got, "\n"); len(lines) != 24 {
		t.Fatalf("popup view height = %d, want 24", len(lines))
	}
}

func TestConsoleURLStaysOnOneHighlightedHyperlink(t *testing.T) {
	consoleURL := "https://console.production.example.com/deployments/overview"
	model := New(t.Context(), nil, theme.New(colorprofile.TrueColor))
	model, _ = model.Update(loadedMsg{snapshot: bridge.Snapshot{
		App:     bridge.AppProfile{Configured: true, Email: "alex@example.com", Endpoint: "https://app.plural.sh"},
		Console: bridge.ConsoleConnection{Configured: true, URL: consoleURL},
	}})
	view := model.View(100, 30)
	if !strings.Contains(view, "\x1b]8;;"+consoleURL) {
		t.Fatalf("console URL is not an OSC-8 hyperlink:\n%q", view)
	}
	if !strings.Contains(ansi.Strip(view), consoleURL) {
		t.Fatalf("console URL was wrapped or truncated:\n%s", ansi.Strip(view))
	}
}

func TestWorkspacePathUsesEllipsisWhenRightPaneIsNarrow(t *testing.T) {
	workspacePath := "/work/path/to/a/very/long/workspace/directory"
	model := New(t.Context(), nil, theme.New(colorprofile.ASCII))
	model, _ = model.Update(loadedMsg{snapshot: bridge.Snapshot{
		Workspace: bridge.Workspace{
			Configured: true,
			Name:       "plrl-dev-aws",
			Path:       workspacePath,
		},
	}})

	narrow := ansi.Strip(model.View(80, 24))
	if !strings.Contains(narrow, "/work/path/...") {
		t.Fatalf("narrow workspace path has no ellipsis:\n%s", narrow)
	}
	if strings.Contains(narrow, workspacePath) {
		t.Fatalf("narrow workspace path was not truncated:\n%s", narrow)
	}

	wide := ansi.Strip(model.View(160, 30))
	if !strings.Contains(wide, workspacePath) {
		t.Fatalf("wide workspace path did not use available space:\n%s", wide)
	}
}

func TestConnectionGroupsStackWhenURLsNeedTheWidth(t *testing.T) {
	model := New(t.Context(), nil, theme.New(colorprofile.ASCII))
	model, _ = model.Update(loadedMsg{snapshot: bridge.Snapshot{
		App: bridge.AppProfile{Configured: true, Endpoint: "https://app.plural.sh"},
		Console: bridge.ConsoleConnection{
			Configured: true,
			URL:        "https://console.production.example.com/a/long/context/path",
		},
	}})
	view := model.View(80, 30)
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, "Plural App account") && strings.Contains(line, "Console connection") {
			t.Fatalf("connection groups did not stack:\n%s", view)
		}
	}
}

func TestWelcomeLogoUsesStaticEmbeddedAsset(t *testing.T) {
	model := New(t.Context(), nil, theme.New(colorprofile.TrueColor))
	got := ansi.Strip(model.logo())
	want := strings.TrimSpace(assets.Logo)
	if got != want {
		t.Fatalf("welcome logo differs from embedded asset\nwant:\n%s\n\ngot:\n%s", want, got)
	}
	if lipgloss.Width(got) == 0 || lipgloss.Height(got) == 0 {
		t.Fatal("welcome logo is empty")
	}
}

func TestHeroBorderUsesPrimaryColor(t *testing.T) {
	theme := theme.New(colorprofile.TrueColor)
	model := New(t.Context(), nil, theme)
	border := lipgloss.NewStyle().Foreground(theme.Colors.Primary)

	wide := model.renderHero(80, "dev")
	if !strings.HasPrefix(wide, border.Render("╭─ ")) {
		t.Fatalf("wide hero top border does not use the primary color: %q", wide)
	}
	bottom := strings.Split(wide, "\n")[8]
	if !strings.HasPrefix(bottom, border.Render("╰─ ")) || !strings.Contains(bottom, theme.Muted.Render("dev")) {
		t.Fatalf("wide hero bottom border does not use the primary color: %q", wide)
	}

}

func TestStatusAdaptsToTerminalColorCapability(t *testing.T) {
	plain := New(t.Context(), nil, theme.New(colorprofile.ASCII))
	if got := plain.status(true); got != "✓" {
		t.Fatalf("plain success status = %q, want tick", got)
	}
	if got := plain.status(false); got != "✗" {
		t.Fatalf("plain failure status = %q, want cross", got)
	}

	color := New(t.Context(), nil, theme.New(colorprofile.TrueColor))
	if got := ansi.Strip(color.status(true)); got != "●" {
		t.Fatalf("color success status = %q, want dot", got)
	}
	if got := ansi.Strip(color.status(false)); got != "●" {
		t.Fatalf("color failure status = %q, want dot", got)
	}
}

func TestWelcomeForwardsInputToCommandBar(t *testing.T) {
	model := New(t.Context(), nil, theme.New(colorprofile.ASCII))
	model, _ = model.Update(tea.KeyPressMsg{Code: 'd', Text: "d"})
	if got := model.command.CurrentSuggestion(); got != "diagnostics" {
		t.Fatalf("suggestion = %q, want diagnostics", got)
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if got := model.command.Value(); got != "diagnostics" {
		t.Fatalf("completed input = %q, want diagnostics", got)
	}
	model, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("selecting a command did not emit submission")
	}
	model, routeCmd := model.Update(cmd())
	if routeCmd == nil {
		t.Fatal("submitted command did not route")
	}
	if got := routeCmd().(navigation.NavigateMsg).Route; got != navigation.Diagnostics {
		t.Fatalf("route = %q, want diagnostics", got)
	}
	if got := model.command.Selected(); got != "diagnostics" {
		t.Fatalf("selected command = %q, want diagnostics", got)
	}
}

func TestCommandInputIsAnchoredAtBottom(t *testing.T) {
	model := New(t.Context(), nil, theme.New(colorprofile.ASCII))
	lines := strings.Split(normalizeView(model.View(80, 24)), "\n")
	if len(lines) != 24 {
		t.Fatalf("view height = %d, want 24", len(lines))
	}
	if !strings.Contains(lines[len(lines)-4], "Command") {
		t.Fatalf("command bar is not anchored above help:\n%s", strings.Join(lines, "\n"))
	}
	if !strings.Contains(lines[len(lines)-1], "ctrl+c quit") {
		t.Fatalf("keymap is not at the bottom: %q", lines[len(lines)-1])
	}
}

func TestWideLayoutStretchesRightPaneAndStaysLeftAligned(t *testing.T) {
	model := New(t.Context(), nil, theme.New(colorprofile.ASCII))
	model, _ = model.Update(loadedMsg{snapshot: bridge.Snapshot{
		App:       bridge.AppProfile{Configured: true, Name: "personal", Email: "alex@example.com"},
		Console:   bridge.ConsoleConnection{Configured: true, URL: "https://console.example.com"},
		Workspace: bridge.Workspace{Configured: true, Name: "platform"},
	}})

	dividerColumn := -1
	for _, width := range []int{80, 160} {
		lines := strings.Split(normalizeView(model.View(width, 30)), "\n")
		top := []rune(lines[0])
		if len(top) != width-2 {
			t.Fatalf("hero line width at %d columns = %d, want %d", width, len(top), width-2)
		}
		if len(top) < 3 || top[0] != ' ' || top[1] != ' ' || top[2] != '╭' {
			t.Fatalf("hero is not anchored at the two-cell left gutter: %q", lines[0])
		}

		body := []rune(lines[1])
		column := -1
		seenOuterBorder := false
		for i, r := range body {
			if r != '│' {
				continue
			}
			if seenOuterBorder {
				column = i
				break
			}
			seenOuterBorder = true
		}
		if dividerColumn == -1 {
			dividerColumn = column
		} else if column != dividerColumn {
			t.Fatalf("logo rail divider moved from column %d to %d", dividerColumn, column)
		}
	}
}

func TestWelcomeRejectsUnsupportedTerminalSize(t *testing.T) {
	model := New(t.Context(), nil, theme.New(colorprofile.ASCII))

	for _, size := range []struct {
		width  int
		height int
	}{
		{width: 79, height: 24},
		{width: 80, height: 23},
		{width: 54, height: 12},
		{width: 80, height: 2},
	} {
		view := ansi.Strip(model.View(size.width, size.height))
		if !strings.Contains(view, model.dimensions(size.width, size.height)) {
			t.Fatalf("unsupported view does not show detected size %dx%d:\n%s", size.width, size.height, view)
		}
		if !strings.Contains(view, model.dimensions(minimumWidth, minimumHeight)) {
			t.Fatalf("unsupported view does not show minimum size:\n%s", view)
		}
		if strings.Contains(view, "App not connected") {
			t.Fatalf("unsupported terminal rendered the welcome hero:\n%s", view)
		}
		lines := strings.Split(view, "\n")
		if len(lines) != size.height {
			t.Fatalf("unsupported view height = %d, want %d", len(lines), size.height)
		}
		for _, line := range lines {
			if got := lipgloss.Width(line); got > size.width {
				t.Fatalf("unsupported view line width %d exceeds %d: %q", got, size.width, line)
			}
		}
	}
}

func TestWelcomeNeverExceedsSupportedTerminalWidth(t *testing.T) {
	model := New(t.Context(), nil, theme.New(colorprofile.ASCII))
	model, _ = model.Update(loadedMsg{snapshot: bridge.Snapshot{
		App: bridge.AppProfile{Configured: true, Email: "alex@example.com", Endpoint: "https://app.plural.sh"},
		Console: bridge.ConsoleConnection{
			Configured: true,
			URL:        "https://console.production.example.com/a/long/context/path",
		},
	}})
	for _, width := range []int{80, 100, 120, 160} {
		for _, line := range strings.Split(model.View(width, 30), "\n") {
			if got := lipgloss.Width(line); got > width {
				t.Fatalf("line width %d exceeds terminal width %d: %q", got, width, ansi.Strip(line))
			}
		}
	}
}

func normalizeView(view string) string {
	lines := strings.Split(ansi.Strip(view), "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " ")
	}
	return strings.Join(lines, "\n")
}
