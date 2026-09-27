package diagnostics

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"

	bridge "github.com/pluralsh/plural-cli/pkg/bridge/welcome"
	"github.com/pluralsh/plural-cli/tui/theme"
)

func TestDiagnosticsGoldens(t *testing.T) {
	snapshot := bridge.Snapshot{
		Version:     "v0.13.0",
		App:         bridge.AppProfile{Configured: true, Name: "personal", Email: "alex@acme.io", Endpoint: "https://app.plural.sh"},
		Console:     bridge.ConsoleConnection{Configured: true, URL: "https://console.acme.io"},
		Workspace:   bridge.Workspace{Configured: true, Path: "/work/path/to/a/very/long/workspace", Name: "plrl-dev-aws", Provider: "aws", Region: "eu-west-1"},
		KubeContext: "plural-platform-prod",
		Diagnostics: []string{"workspace owner does not match the active identity"},
	}
	for _, width := range []int{80, 120} {
		t.Run(strconv.Itoa(width), func(t *testing.T) {
			model := New(t.Context(), nil, theme.New(colorprofile.ASCII))
			model.snapshot = snapshot
			height := 24
			if width == 120 {
				height = 30
			}
			got := normalizeGoldenView(model.View(width, height))
			golden := filepath.Join("testdata", "diagnostics-"+strconv.Itoa(width)+".golden")
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("read golden: %v\nactual:\n%s", err, got)
			}
			if got != strings.TrimSuffix(string(want), "\n") {
				t.Fatalf("view changed\nwant:\n%s\n\ngot:\n%s", want, got)
			}
			assertGoldenDimensions(t, got, width, height)
		})
	}
}

func normalizeGoldenView(view string) string {
	lines := strings.Split(ansi.Strip(view), "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " ")
	}
	return strings.Join(lines, "\n")
}
func assertGoldenDimensions(t *testing.T, view string, width, height int) {
	t.Helper()
	lines := strings.Split(view, "\n")
	if len(lines) != height {
		t.Fatalf("view height = %d, want %d", len(lines), height)
	}
	for _, line := range lines {
		if got := lipgloss.Width(line); got > width {
			t.Fatalf("line width %d exceeds %d: %q", got, width, line)
		}
	}
}
