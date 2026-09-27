package notifications

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
	notificationsbridge "github.com/pluralsh/plural-cli/pkg/bridge/notifications"
	"github.com/pluralsh/plural-cli/tui/navigation"
	"github.com/pluralsh/plural-cli/tui/theme"
)

type fakeLoader struct {
	page   notificationsbridge.Page
	detail notificationsbridge.Detail
	err    error
}

func (f *fakeLoader) List(context.Context, *string, string) (notificationsbridge.Page, error) {
	return f.page, f.err
}
func (f *fakeLoader) Get(context.Context, string) (notificationsbridge.Detail, error) {
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

func TestOpenSinkDetailAndBack(t *testing.T) {
	loader := &fakeLoader{
		page: notificationsbridge.Page{Items: []notificationsbridge.Summary{
			{ID: "s1", Name: "ops-slack", Type: "SLACK", URL: "https://hooks.slack.com/x"},
			{ID: "s2", Name: "ops-teams", Type: "TEAMS", URL: "https://teams.example/hook"},
		}},
		detail: notificationsbridge.Detail{
			Summary: notificationsbridge.Summary{ID: "s1", Name: "ops-slack", Type: "SLACK", URL: "https://hooks.slack.com/x"},
			Bindings: []notificationsbridge.Binding{
				{Kind: "user", Name: "ops@acme.io"},
				{Kind: "group", Name: "platform"},
			},
		},
	}
	model := loadList(t, New(t.Context(), loader, theme.New(colorprofile.ASCII)))
	if model.mode != modeList || len(model.page.Items) != 2 {
		t.Fatalf("list state = mode=%d count=%d", model.mode, len(model.page.Items))
	}
	if !strings.Contains(model.View(80, 24), "ops-slack") {
		t.Fatalf("list missing name:\n%s", model.View(80, 24))
	}

	model, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(cmd())
	if model.mode != modeDetail || model.detail.Name != "ops-slack" {
		t.Fatalf("detail = %#v mode=%d", model.detail, model.mode)
	}
	if !strings.Contains(model.View(80, 24), "ops@acme.io") {
		t.Fatalf("detail missing binding:\n%s", model.View(80, 24))
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
		page: notificationsbridge.Page{
			Items:     []notificationsbridge.Summary{{ID: "s1", Name: "a", Type: "SLACK"}},
			EndCursor: "s1",
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
	loader.page = notificationsbridge.Page{Items: []notificationsbridge.Summary{{ID: "s2", Name: "b", Type: "TEAMS"}}}
	model, _ = model.Update(cmd())
	if model.after == nil || *model.after != "s1" || len(model.prevCursors) != 1 {
		t.Fatalf("after page turn after=%v prev=%v", model.after, model.prevCursors)
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

func TestNotificationsGoldens(t *testing.T) {
	list := New(t.Context(), nil, theme.New(colorprofile.ASCII))
	list.loading = false
	list.mode = modeList
	list.page = notificationsbridge.Page{Items: []notificationsbridge.Summary{
		{ID: "s1", Name: "ops-slack", Type: "SLACK", URL: "https://hooks.slack.com/services/T/B/x"},
		{ID: "s2", Name: "ops-teams", Type: "TEAMS", URL: "https://teams.example/webhook"},
		{ID: "s3", Name: "plural-inbox", Type: "PLURAL"},
	}, HasNext: true, EndCursor: "s3"}

	detail := list
	detail.mode = modeDetail
	detail.detail = notificationsbridge.Detail{
		Summary: notificationsbridge.Summary{ID: "s1", Name: "ops-slack", Type: "SLACK", URL: "https://hooks.slack.com/services/T/B/x"},
		Bindings: []notificationsbridge.Binding{
			{Kind: "user", Name: "ops@acme.io"},
			{Kind: "group", Name: "platform"},
		},
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
			golden := filepath.Join("testdata", "notifications-"+tc.name+".golden")
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

func TestWriteNotificationsGoldens(t *testing.T) {
	if os.Getenv("UPDATE_GOLDEN") == "" {
		t.Skip("set UPDATE_GOLDEN=1 to refresh fixtures")
	}
	list := New(t.Context(), nil, theme.New(colorprofile.ASCII))
	list.loading = false
	list.mode = modeList
	list.page = notificationsbridge.Page{Items: []notificationsbridge.Summary{
		{ID: "s1", Name: "ops-slack", Type: "SLACK", URL: "https://hooks.slack.com/services/T/B/x"},
		{ID: "s2", Name: "ops-teams", Type: "TEAMS", URL: "https://teams.example/webhook"},
		{ID: "s3", Name: "plural-inbox", Type: "PLURAL"},
	}, HasNext: true, EndCursor: "s3"}
	detail := list
	detail.mode = modeDetail
	detail.detail = notificationsbridge.Detail{
		Summary: notificationsbridge.Summary{ID: "s1", Name: "ops-slack", Type: "SLACK", URL: "https://hooks.slack.com/services/T/B/x"},
		Bindings: []notificationsbridge.Binding{
			{Kind: "user", Name: "ops@acme.io"},
			{Kind: "group", Name: "platform"},
		},
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
		if err := os.WriteFile(filepath.Join("testdata", "notifications-"+tc.name+".golden"), []byte(got), 0o644); err != nil {
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
