package deployments

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"

	"github.com/pluralsh/plural-cli/tui/navigation"
	"github.com/pluralsh/plural-cli/tui/theme"
)

func TestDeploymentsGoldens(t *testing.T) {
	model := New(t.Context(), theme.New(colorprofile.ASCII), "https://console.acme.io")
	for _, width := range []int{80, 120} {
		t.Run(strconv.Itoa(width), func(t *testing.T) {
			height := 24
			if width == 120 {
				height = 30
			}
			got := normalizeView(model.View(width, height))
			golden := filepath.Join("testdata", "deployments-"+strconv.Itoa(width)+".golden")
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("read golden: %v\nactual:\n%s", err, got)
			}
			if got != strings.TrimSuffix(string(want), "\n") {
				t.Fatalf("view changed\nwant:\n%s\n\ngot:\n%s", want, got)
			}
		})
	}
}

func TestUpdateGoldens(t *testing.T) {
	if os.Getenv("UPDATE_GOLDEN") == "" {
		t.Skip("set UPDATE_GOLDEN=1 to refresh fixtures")
	}
	model := New(t.Context(), theme.New(colorprofile.ASCII), "https://console.acme.io")
	_ = os.MkdirAll("testdata", 0o755)
	for _, width := range []int{80, 120} {
		height := 24
		if width == 120 {
			height = 30
		}
		got := normalizeView(model.View(width, height)) + "\n"
		if err := os.WriteFile(filepath.Join("testdata", "deployments-"+strconv.Itoa(width)+".golden"), []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestServicesShortcutNavigates(t *testing.T) {
	model := New(t.Context(), theme.New(colorprofile.ASCII), "https://console.acme.io")
	_, cmd := model.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
	if cmd == nil {
		t.Fatal("expected navigation")
	}
	if msg := cmd(); msg != (navigation.NavigateMsg{Route: navigation.Services}) {
		t.Fatalf("msg = %#v", msg)
	}
}

func TestClustersNavigates(t *testing.T) {
	model := New(t.Context(), theme.New(colorprofile.ASCII), "https://console.acme.io")
	_, cmd := model.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
	if cmd == nil {
		t.Fatal("expected navigation")
	}
	if msg := cmd(); msg != (navigation.NavigateMsg{Route: navigation.Clusters}) {
		t.Fatalf("msg = %#v", msg)
	}
}

func TestRepositoriesNavigates(t *testing.T) {
	model := New(t.Context(), theme.New(colorprofile.ASCII), "https://console.acme.io")
	_, cmd := model.Update(tea.KeyPressMsg{Code: 'r', Text: "r"})
	if cmd == nil {
		t.Fatal("expected navigation")
	}
	if msg := cmd(); msg != (navigation.NavigateMsg{Route: navigation.Repositories}) {
		t.Fatalf("msg = %#v", msg)
	}
}

func TestPipelinesNavigates(t *testing.T) {
	model := New(t.Context(), theme.New(colorprofile.ASCII), "https://console.acme.io")
	_, cmd := model.Update(tea.KeyPressMsg{Code: 'p', Text: "p"})
	if cmd == nil {
		t.Fatal("expected navigation")
	}
	if msg := cmd(); msg != (navigation.NavigateMsg{Route: navigation.Pipelines}) {
		t.Fatalf("msg = %#v", msg)
	}
}

func TestSoonResourceDoesNotNavigate(t *testing.T) {
	model := New(t.Context(), theme.New(colorprofile.ASCII), "")
	model.cursor = 4 // notifications [soon]
	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd != nil {
		t.Fatalf("unexpected cmd %#v", cmd())
	}
}

func TestEscReturnsWelcome(t *testing.T) {
	model := New(t.Context(), theme.New(colorprofile.ASCII), "")
	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected welcome navigation")
	}
	if msg := cmd(); msg != (navigation.NavigateMsg{Route: navigation.Welcome}) {
		t.Fatalf("msg = %#v", msg)
	}
}

func normalizeView(view string) string {
	lines := strings.Split(ansi.Strip(view), "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " ")
	}
	return strings.Join(lines, "\n")
}
