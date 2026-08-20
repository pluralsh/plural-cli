package agents

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"

	agentsbridge "github.com/pluralsh/plural-cli/pkg/bridge/agents"
	"github.com/pluralsh/plural-cli/tui/theme"
)

type fakeLoader struct {
	page   agentsbridge.Page
	detail agentsbridge.Detail
}

func (f fakeLoader) List(context.Context, *string, string) (agentsbridge.Page, error) {
	return f.page, nil
}
func (f fakeLoader) Get(context.Context, string) (agentsbridge.Detail, error) { return f.detail, nil }

func TestSelectRunOpensInteractiveDetail(t *testing.T) {
	loader := fakeLoader{page: agentsbridge.Page{Items: []agentsbridge.Summary{{ID: "run-1", Repository: "acme/repo", Provider: "codex"}}}, detail: agentsbridge.Detail{Summary: agentsbridge.Summary{ID: "run-1", Repository: "acme/repo", Provider: "codex"}}}
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
