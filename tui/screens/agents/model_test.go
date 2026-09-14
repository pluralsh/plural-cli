package agents

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"

	agentsbridge "github.com/pluralsh/plural-cli/pkg/bridge/agents"
	"github.com/pluralsh/plural-cli/tui/theme"
)

type fakeLoader struct {
	page       agentsbridge.Page
	detail     agentsbridge.Detail
	resumeID   string
	resumePath string
	resumePR   string
	resumeErr  error
}

func (f *fakeLoader) List(context.Context, *string, string) (agentsbridge.Page, error) {
	return f.page, nil
}
func (f *fakeLoader) Get(context.Context, string) (agentsbridge.Detail, error) { return f.detail, nil }
func (f *fakeLoader) Resume(_ context.Context, id, path, prRef string) error {
	f.resumeID, f.resumePath, f.resumePR = id, path, prRef
	return f.resumeErr
}

func TestSelectRunOpensInteractiveDetail(t *testing.T) {
	loader := &fakeLoader{page: agentsbridge.Page{Items: []agentsbridge.Summary{{ID: "run-1", Repository: "acme/repo", Provider: "codex"}}}, detail: agentsbridge.Detail{Summary: agentsbridge.Summary{ID: "run-1", Repository: "acme/repo", Provider: "codex"}}}
	model := New(t.Context(), loader, theme.New(colorprofile.ASCII))
	model, cmd := model.Update(model.Init()())
	model, _ = model.Update(cmd())
	model, cmd = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(cmd())
	if model.mode != modeDetail || model.detail.ID != "run-1" {
		t.Fatalf("unexpected detail state: %#v", model)
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: 'r'})
	if model.mode != modeRepoPath {
		t.Fatalf("expected repo path step, got %d", model.mode)
	}
}

func TestRepoPathSuggestsWorkingDirectory(t *testing.T) {
	loader := &fakeLoader{
		page:   agentsbridge.Page{Items: []agentsbridge.Summary{{ID: "run-1", Repository: "acme/repo"}}},
		detail: agentsbridge.Detail{Summary: agentsbridge.Summary{ID: "run-1", Repository: "git@github.com:acme/tf-test.git"}},
	}
	model := New(t.Context(), loader, theme.New(colorprofile.ASCII))
	model.execResume = false
	model.getwd = func() (string, error) { return "/home/lukasz/plural/tf-test", nil }
	model, cmd := model.Update(model.Init()())
	model, _ = model.Update(cmd())
	model, cmd = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(cmd())
	model, _ = model.Update(tea.KeyPressMsg{Code: 'r'})
	if model.input.Value() != "/home/lukasz/plural/tf-test" {
		t.Fatalf("suggested path = %q", model.input.Value())
	}
	got := model.View(80, 24)
	if !strings.Contains(got, "Current dir") || !strings.Contains(got, "/home/lukasz/plural/tf-test") {
		t.Fatalf("path view missing cwd suggestion:\n%s", got)
	}
	model, cmd = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter did not resume from suggested cwd")
	}
	_, _ = model.Update(cmd())
	if loader.resumePath != "/home/lukasz/plural/tf-test" {
		t.Fatalf("resume path = %q", loader.resumePath)
	}
}

func TestResumeCallsBridgeAndShowsError(t *testing.T) {
	loader := &fakeLoader{
		page:      agentsbridge.Page{Items: []agentsbridge.Summary{{ID: "run-1", Repository: "acme/repo", Provider: "codex", PRRef: "feat/fix"}}},
		detail:    agentsbridge.Detail{Summary: agentsbridge.Summary{ID: "run-1", Repository: "acme/repo", Provider: "codex", PRRef: "feat/fix"}},
		resumeErr: errors.New("not a git checkout for git@github.com:acme/repo.git"),
	}
	model := New(t.Context(), loader, theme.New(colorprofile.ASCII))
	model.execResume = false
	model, cmd := model.Update(model.Init()())
	model, _ = model.Update(cmd())
	model, cmd = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(cmd())
	model, _ = model.Update(tea.KeyPressMsg{Code: 'r'})
	model.input.SetValue("/work/repo")
	model, cmd = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("resume did not start")
	}
	model, _ = model.Update(cmd())
	if loader.resumeID != "run-1" || loader.resumePath != "/work/repo" || loader.resumePR != "feat/fix" {
		t.Fatalf("resume args = %s %s %s", loader.resumeID, loader.resumePath, loader.resumePR)
	}
	if model.mode != modeResult || model.err == nil {
		t.Fatalf("expected failed result, got mode=%d err=%v", model.mode, model.err)
	}
	got := model.View(80, 24)
	if !strings.Contains(got, "Resume failed") || !strings.Contains(got, "not a git checkout") {
		t.Fatalf("result view missing resume error:\n%s", got)
	}
}
