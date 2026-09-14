package pullrequests

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
	pullrequestsbridge "github.com/pluralsh/plural-cli/pkg/bridge/pullrequests"
	"github.com/pluralsh/plural-cli/tui/navigation"
	"github.com/pluralsh/plural-cli/tui/theme"
)

type fakeLoader struct {
	page      pullrequestsbridge.Page
	detail    pullrequestsbridge.Detail
	created   pullrequestsbridge.CreatedPR
	err       error
	createErr error
}

func (f *fakeLoader) List(context.Context, *string, string) (pullrequestsbridge.Page, error) {
	return f.page, f.err
}
func (f *fakeLoader) Get(context.Context, string) (pullrequestsbridge.Detail, error) {
	return f.detail, f.err
}
func (f *fakeLoader) CreatePR(context.Context, pullrequestsbridge.CreatePRInput) (pullrequestsbridge.CreatedPR, error) {
	return f.created, f.createErr
}
func (f *fakeLoader) TriggerPR(context.Context, pullrequestsbridge.TriggerPRInput) (pullrequestsbridge.CreatedPR, error) {
	return f.created, f.createErr
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

func loadDetail(t *testing.T, loader *fakeLoader) Model {
	t.Helper()
	model := loadList(t, New(t.Context(), loader, theme.New(colorprofile.ASCII)))
	model, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(cmd())
	if model.mode != modeDetail {
		t.Fatalf("mode = %d", model.mode)
	}
	return model
}

func TestOpenAutomationDetailAndBack(t *testing.T) {
	loader := &fakeLoader{
		page: pullrequestsbridge.Page{Items: []pullrequestsbridge.Summary{
			{ID: "pra1", Name: "cluster-create", Title: "Create cluster", Addon: "cluster"},
			{ID: "pra2", Name: "service-bump", Title: "Bump service"},
		}},
		detail: pullrequestsbridge.Detail{
			Summary: pullrequestsbridge.Summary{ID: "pra1", Name: "cluster-create", Title: "Create cluster", Addon: "cluster", Identifier: "ops/cluster-create"},
			Message: "Opens a PR to provision a new cluster",
		},
	}
	model := loadDetail(t, loader)
	if !strings.Contains(model.View(80, 24), "cluster-create") {
		t.Fatalf("detail missing name:\n%s", model.View(80, 24))
	}
	if !strings.Contains(model.View(80, 24), "Create") || !strings.Contains(model.View(80, 24), "Trigger") {
		t.Fatalf("detail missing actions:\n%s", model.View(80, 24))
	}

	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if model.mode != modeList {
		t.Fatalf("mode after detail esc = %d", model.mode)
	}
	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if cmd == nil || cmd() != (navigation.NavigateMsg{Route: navigation.Deployments}) {
		t.Fatalf("expected deployments navigation")
	}
}

func TestCreatePRFlow(t *testing.T) {
	loader := &fakeLoader{
		page: pullrequestsbridge.Page{Items: []pullrequestsbridge.Summary{
			{ID: "pra1", Name: "cluster-create", Title: "Create cluster", Addon: "cluster"},
		}},
		detail: pullrequestsbridge.Detail{
			Summary: pullrequestsbridge.Summary{ID: "pra1", Name: "cluster-create", Title: "Create cluster"},
		},
		created: pullrequestsbridge.CreatedPR{ID: "pr1", URL: "https://github.com/acme/fleet/pull/1", Title: "Create cluster", Status: "OPEN"},
	}
	model := loadDetail(t, loader)
	model, _ = model.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
	if model.mode != modeCreateForm {
		t.Fatalf("mode = %d", model.mode)
	}
	model.formInput.SetValue("feat/cluster")
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter}) // next field
	model.formInput.SetValue(`{"cluster":"demo"}`)
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter}) // review
	if model.mode != modeReview || model.pending.create == nil {
		t.Fatalf("review = mode=%d pending=%#v", model.mode, model.pending)
	}
	model, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(cmd())
	if model.mode != modeResult || model.result != "ok" || !strings.Contains(strings.Join(model.opLog, "\n"), "pr1") {
		t.Fatalf("result = mode=%d result=%s log=%v", model.mode, model.result, model.opLog)
	}
}

func TestTriggerPRFlow(t *testing.T) {
	loader := &fakeLoader{
		page: pullrequestsbridge.Page{Items: []pullrequestsbridge.Summary{
			{ID: "pra1", Name: "cluster-create"},
		}},
		detail:  pullrequestsbridge.Detail{Summary: pullrequestsbridge.Summary{ID: "pra1", Name: "cluster-create"}},
		created: pullrequestsbridge.CreatedPR{ID: "pr2", URL: "https://github.com/acme/fleet/pull/2"},
	}
	model := loadDetail(t, loader)
	model, _ = model.Update(tea.KeyPressMsg{Code: 't', Text: "t"})
	if model.mode != modeTriggerForm {
		t.Fatalf("mode = %d", model.mode)
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter}) // skip branch
	model.formInput.SetValue("cluster=demo")
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if model.mode != modeReview || model.pending.trigger == nil || model.pending.trigger.Configuration["cluster"] != "demo" {
		t.Fatalf("review = %#v", model.pending)
	}
	model, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(cmd())
	if model.mode != modeResult || model.result != "ok" {
		t.Fatalf("result = mode=%d result=%s", model.mode, model.result)
	}
}

func TestCLITipTemplate(t *testing.T) {
	loader := &fakeLoader{
		page:   pullrequestsbridge.Page{Items: []pullrequestsbridge.Summary{{ID: "pra1", Name: "cluster-create"}}},
		detail: pullrequestsbridge.Detail{Summary: pullrequestsbridge.Summary{ID: "pra1", Name: "cluster-create"}},
	}
	model := loadDetail(t, loader)
	model, _ = model.Update(tea.KeyPressMsg{Code: 'm', Text: "m"})
	if model.mode != modeCLITip {
		t.Fatalf("mode = %d", model.mode)
	}
	model.formInput.SetValue("./pra.yaml")
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if model.mode != modeResult || !strings.Contains(model.pending.cli, "plural pr template") {
		t.Fatalf("cli tip = %#v log=%v", model.pending, model.opLog)
	}
}

func TestNextPrevPage(t *testing.T) {
	loader := &fakeLoader{
		page: pullrequestsbridge.Page{
			Items:     []pullrequestsbridge.Summary{{ID: "pra1", Name: "a", Addon: "cluster"}},
			EndCursor: "pra1",
			HasNext:   true,
		},
	}
	model := loadList(t, New(t.Context(), loader, theme.New(colorprofile.ASCII)))
	model, cmd := model.Update(tea.KeyPressMsg{Code: 'n'})
	if cmd == nil {
		t.Fatal("expected next-page list command")
	}
	loader.page = pullrequestsbridge.Page{Items: []pullrequestsbridge.Summary{{ID: "pra2", Name: "b"}}}
	model, _ = model.Update(cmd())
	if model.after == nil || *model.after != "pra1" {
		t.Fatalf("after=%v", model.after)
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

func TestPullRequestsGoldens(t *testing.T) {
	list := New(t.Context(), nil, theme.New(colorprofile.ASCII))
	list.loading = false
	list.mode = modeList
	list.page = pullrequestsbridge.Page{Items: []pullrequestsbridge.Summary{
		{ID: "pra1", Name: "cluster-create", Title: "Create cluster", Addon: "cluster", Identifier: "ops/cluster-create"},
		{ID: "pra2", Name: "service-bump", Title: "Bump service chart", Addon: "service"},
		{ID: "pra3", Name: "stack-plan", Title: "Stack plan PR"},
	}, HasNext: true, EndCursor: "pra3"}

	detail := list
	detail.mode = modeDetail
	detail.detail = pullrequestsbridge.Detail{
		Summary: pullrequestsbridge.Summary{ID: "pra1", Name: "cluster-create", Title: "Create cluster", Addon: "cluster", Identifier: "ops/cluster-create"},
		Message: "Opens a PR to provision a new cluster",
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
			golden := filepath.Join("testdata", "pullrequests-"+tc.name+".golden")
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

func TestWritePullRequestsGoldens(t *testing.T) {
	if os.Getenv("UPDATE_GOLDEN") == "" {
		t.Skip("set UPDATE_GOLDEN=1 to refresh fixtures")
	}
	list := New(t.Context(), nil, theme.New(colorprofile.ASCII))
	list.loading = false
	list.mode = modeList
	list.page = pullrequestsbridge.Page{Items: []pullrequestsbridge.Summary{
		{ID: "pra1", Name: "cluster-create", Title: "Create cluster", Addon: "cluster", Identifier: "ops/cluster-create"},
		{ID: "pra2", Name: "service-bump", Title: "Bump service chart", Addon: "service"},
		{ID: "pra3", Name: "stack-plan", Title: "Stack plan PR"},
	}, HasNext: true, EndCursor: "pra3"}
	detail := list
	detail.mode = modeDetail
	detail.detail = pullrequestsbridge.Detail{
		Summary: pullrequestsbridge.Summary{ID: "pra1", Name: "cluster-create", Title: "Create cluster", Addon: "cluster", Identifier: "ops/cluster-create"},
		Message: "Opens a PR to provision a new cluster",
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
		if err := os.WriteFile(filepath.Join("testdata", "pullrequests-"+tc.name+".golden"), []byte(got), 0o644); err != nil {
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
