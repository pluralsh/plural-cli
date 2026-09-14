package repositories

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"

	"github.com/pluralsh/plural-cli/pkg/bridge"
	repositoriesbridge "github.com/pluralsh/plural-cli/pkg/bridge/repositories"
	"github.com/pluralsh/plural-cli/tui/navigation"
	"github.com/pluralsh/plural-cli/tui/theme"
)

type fakeLoader struct {
	page   repositoriesbridge.Page
	detail repositoriesbridge.Detail
	err    error
}

func (f *fakeLoader) List(context.Context, *string, string) (repositoriesbridge.Page, error) {
	return f.page, f.err
}
func (f *fakeLoader) Get(context.Context, string) (repositoriesbridge.Detail, error) {
	return f.detail, f.err
}

func loadList(t *testing.T, model Model) Model {
	t.Helper()
	cmd := model.Init()
	model, cmd = model.Update(cmd())
	if cmd == nil {
		t.Fatal("expected list command")
	}
	model, _ = model.Update(cmd())
	return model
}

func TestOpenRepositoryDetailAndBack(t *testing.T) {
	loader := &fakeLoader{
		page: repositoriesbridge.Page{Items: []repositoriesbridge.Summary{
			{ID: "r1", URL: "git@github.com:acme/infra.git", Health: "PULLABLE", AuthMethod: "SSH"},
			{ID: "r2", URL: "https://github.com/acme/apps.git", Health: "FAILED", Error: "auth failed"},
		}},
		detail: repositoriesbridge.Detail{
			Summary: repositoriesbridge.Summary{
				ID: "r1", URL: "git@github.com:acme/infra.git", Health: "PULLABLE", AuthMethod: "SSH",
			},
			Decrypt: true,
		},
	}
	model := loadList(t, New(t.Context(), loader, theme.New(colorprofile.ASCII)))
	if model.mode != modeList || len(model.page.Items) != 2 {
		t.Fatalf("list state = mode=%d count=%d", model.mode, len(model.page.Items))
	}
	if !strings.Contains(model.View(80, 24), "git@github.com:acme/infra.git") {
		t.Fatalf("list missing url:\n%s", model.View(80, 24))
	}

	model, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(cmd())
	if model.mode != modeDetail || model.detail.URL != "git@github.com:acme/infra.git" {
		t.Fatalf("detail = %#v mode=%d", model.detail, model.mode)
	}
	if !strings.Contains(model.View(80, 24), "SSH") {
		t.Fatalf("detail view missing auth:\n%s", model.View(80, 24))
	}

	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if model.mode != modeList {
		t.Fatalf("mode after detail esc = %d", model.mode)
	}
	_, cmd = model.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if cmd == nil || cmd() != (navigation.NavigateMsg{Route: navigation.Deployments}) {
		t.Fatalf("expected deployments navigation")
	}
}

func TestNextPrevPage(t *testing.T) {
	loader := &fakeLoader{
		page: repositoriesbridge.Page{
			Items:     []repositoriesbridge.Summary{{ID: "r1", URL: "git@github.com:acme/a.git", Health: "PULLABLE"}},
			EndCursor: "r1",
			HasNext:   true,
		},
	}
	model := loadList(t, New(t.Context(), loader, theme.New(colorprofile.ASCII)))
	if !strings.Contains(model.View(80, 24), "n next") {
		t.Fatalf("missing next pager:\n%s", model.View(80, 24))
	}
	model, cmd := model.Update(tea.KeyPressMsg{Code: 'n'})
	if cmd == nil {
		t.Fatal("expected next-page list command")
	}
	loader.page = repositoriesbridge.Page{
		Items: []repositoriesbridge.Summary{{ID: "r2", URL: "git@github.com:acme/b.git", Health: "PULLABLE"}},
	}
	model, _ = model.Update(cmd())
	if model.after == nil || *model.after != "r1" || len(model.prevCursors) != 1 {
		t.Fatalf("after page turn after=%v prev=%v", model.after, model.prevCursors)
	}
	if !strings.Contains(model.View(80, 24), "p prev") {
		t.Fatalf("missing prev pager:\n%s", model.View(80, 24))
	}
	model, cmd = model.Update(tea.KeyPressMsg{Code: 'p'})
	if cmd == nil {
		t.Fatal("expected prev-page list command")
	}
	model, _ = model.Update(cmd())
	if model.after != nil || len(model.prevCursors) != 0 {
		t.Fatalf("after prev after=%v prev=%v", model.after, model.prevCursors)
	}
}

func TestNoConsoleNavigatesToAccess(t *testing.T) {
	loader := &fakeLoader{err: &bridge.Error{Code: bridge.ErrorUnauthenticated, Err: errors.New("connect")}}
	model := loadList(t, New(t.Context(), loader, theme.New(colorprofile.ASCII)))
	if !model.needsAuth {
		t.Fatal("expected needsAuth")
	}
	_, cmd := model.Update(tea.KeyPressMsg{Code: 'c'})
	if cmd == nil || cmd() != (navigation.NavigateMsg{Route: navigation.Access}) {
		t.Fatalf("expected access navigation")
	}
}

func TestRepositoriesGoldens(t *testing.T) {
	list := New(t.Context(), nil, theme.New(colorprofile.ASCII))
	list.loading = false
	list.mode = modeList
	list.page = repositoriesbridge.Page{Items: []repositoriesbridge.Summary{
		{ID: "r1", URL: "git@github.com:acme/infra.git", Health: "PULLABLE", AuthMethod: "SSH"},
		{ID: "r2", URL: "https://github.com/acme/apps.git", Health: "FAILED", Error: "auth failed"},
		{ID: "r3", URL: "git@gitlab.com:acme/charts.git", Health: "PULLABLE", AuthMethod: "SSH"},
	}, HasNext: true, EndCursor: "r3"}

	detail := list
	detail.mode = modeDetail
	detail.detail = repositoriesbridge.Detail{
		Summary: repositoriesbridge.Summary{
			ID: "r1", URL: "git@github.com:acme/infra.git", Health: "PULLABLE", AuthMethod: "SSH",
		},
		Decrypt: true,
	}

	for _, tc := range []struct {
		name   string
		model  Model
		width  int
		height int
	}{
		{"list-80", list, 80, 24},
		{"list-120", list, 120, 30},
		{"detail-80", detail, 80, 24},
		{"detail-120", detail, 120, 30},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeView(tc.model.View(tc.width, tc.height))
			golden := filepath.Join("testdata", "repositories-"+tc.name+".golden")
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("read golden: %v\nactual:\n%s", err, got)
			}
			if got != strings.TrimSuffix(string(want), "\n") {
				t.Fatalf("view changed\nwant:\n%s\n\ngot:\n%s", want, got)
			}
			lines := strings.Split(got, "\n")
			if len(lines) != tc.height {
				t.Fatalf("height = %d, want %d", len(lines), tc.height)
			}
			for _, line := range lines {
				if w := lipgloss.Width(line); w > tc.width {
					t.Fatalf("line width %d > %d: %q", w, tc.width, line)
				}
			}
		})
	}
}

func TestWriteRepositoriesGoldens(t *testing.T) {
	if os.Getenv("UPDATE_GOLDEN") == "" {
		t.Skip("set UPDATE_GOLDEN=1 to refresh fixtures")
	}
	list := New(t.Context(), nil, theme.New(colorprofile.ASCII))
	list.loading = false
	list.mode = modeList
	list.page = repositoriesbridge.Page{Items: []repositoriesbridge.Summary{
		{ID: "r1", URL: "git@github.com:acme/infra.git", Health: "PULLABLE", AuthMethod: "SSH"},
		{ID: "r2", URL: "https://github.com/acme/apps.git", Health: "FAILED", Error: "auth failed"},
		{ID: "r3", URL: "git@gitlab.com:acme/charts.git", Health: "PULLABLE", AuthMethod: "SSH"},
	}, HasNext: true, EndCursor: "r3"}
	detail := list
	detail.mode = modeDetail
	detail.detail = repositoriesbridge.Detail{
		Summary: repositoriesbridge.Summary{
			ID: "r1", URL: "git@github.com:acme/infra.git", Health: "PULLABLE", AuthMethod: "SSH",
		},
		Decrypt: true,
	}
	_ = os.MkdirAll("testdata", 0o755)
	for _, tc := range []struct {
		name   string
		model  Model
		width  int
		height int
	}{
		{"list-80", list, 80, 24},
		{"list-120", list, 120, 30},
		{"detail-80", detail, 80, 24},
		{"detail-120", detail, 120, 30},
	} {
		got := normalizeView(tc.model.View(tc.width, tc.height)) + "\n"
		if err := os.WriteFile(filepath.Join("testdata", "repositories-"+tc.name+".golden"), []byte(got), 0o644); err != nil {
			t.Fatal(err)
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
