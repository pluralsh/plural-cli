package workbenches

import (
	"context"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"

	workbenchesbridge "github.com/pluralsh/plural-cli/pkg/bridge/workbenches"
	"github.com/pluralsh/plural-cli/tui/theme"
)

type fakeLoader struct {
	page   workbenchesbridge.Page
	detail workbenchesbridge.Detail
}

func (f fakeLoader) List(context.Context, *string, string) (workbenchesbridge.Page, error) {
	return f.page, nil
}
func (f fakeLoader) Get(context.Context, string) (workbenchesbridge.Detail, error) {
	return f.detail, nil
}
func (f fakeLoader) FollowUp(context.Context, string, string, time.Duration) (workbenchesbridge.PromptResult, error) {
	return workbenchesbridge.PromptResult{ID: "prompt-1", WorkbenchID: "job-1"}, nil
}

func TestFollowUpFlowIsInteractive(t *testing.T) {
	loader := fakeLoader{page: workbenchesbridge.Page{Items: []workbenchesbridge.Summary{{ID: "job-1", WorkbenchName: "triage", Prompt: "investigate"}}}, detail: workbenchesbridge.Detail{Summary: workbenchesbridge.Summary{ID: "job-1", WorkbenchName: "triage"}}}
	model := New(t.Context(), loader, theme.New(colorprofile.ASCII))
	model, cmd := model.Update(model.Init()())
	model, _ = model.Update(cmd())
	model, cmd = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(cmd())
	model, _ = model.Update(tea.KeyPressMsg{Code: 'f'})
	if model.mode != modePrompt {
		t.Fatalf("expected prompt mode, got %d", model.mode)
	}
	model.prompt.SetValue("verify the fix")
	model, _ = model.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	if model.mode != modeReview {
		t.Fatalf("expected review mode, got %d", model.mode)
	}
	model, cmd = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(cmd())
	if model.mode != modeResult || model.result.ID != "prompt-1" {
		t.Fatalf("unexpected result: %#v", model)
	}
}
