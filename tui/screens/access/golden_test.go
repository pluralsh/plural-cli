package access

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"

	accessbridge "github.com/pluralsh/plural-cli/pkg/bridge/access"
	"github.com/pluralsh/plural-cli/tui/theme"
)

func TestAccessGoldens(t *testing.T) {
	personal := accessbridge.Profile{ID: "app-personal", Name: "personal", Email: "alex@acme.io", Endpoint: "app.plural.sh"}
	consulting := accessbridge.Profile{ID: "app-consulting", Name: "consulting", Email: "alex@consulting.dev", Endpoint: "cloud.plural.example"}
	production := accessbridge.ConsoleProfile{ID: "console-production", Name: "production", URL: "https://console.acme.io"}
	staging := accessbridge.ConsoleProfile{ID: "console-staging", Name: "staging", URL: "https://console.staging.acme.io"}
	snapshot := accessbridge.Snapshot{
		State: accessbridge.State{
			Profiles: []accessbridge.Profile{personal, consulting}, ActiveProfileID: personal.ID,
			ConsoleProfiles: []accessbridge.ConsoleProfile{production, staging}, ActiveConsoleID: production.ID,
		},
		Context: accessbridge.AuthContext{Base: &personal, Acting: &accessbridge.Identity{Email: "deploy@acme.io", ServiceAccount: true}, Console: &production},
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
			golden := filepath.Join("testdata", "access-"+strconv.Itoa(width)+".golden")
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("read golden: %v\nactual:\n%s", err, got)
			}
			if got != strings.TrimSuffix(string(want), "\n") {
				t.Fatalf("view changed\nwant:\n%s\n\ngot:\n%s", want, got)
			}
			assertGoldenDimensions(t, got, width, height)
			if strings.Contains(got, "super-secret") {
				t.Fatal("Access golden exposed a credential")
			}
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
