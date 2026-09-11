package workbenches

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"

	workbenchesbridge "github.com/pluralsh/plural-cli/pkg/bridge/workbenches"
	"github.com/pluralsh/plural-cli/tui/theme"
)

type fakeLoader struct {
	page      workbenchesbridge.Page
	detail    workbenchesbridge.Detail
	gotJob    string
	gotPrompt string
	result    workbenchesbridge.PromptResult
}

func (f *fakeLoader) List(context.Context, *string, string) (workbenchesbridge.Page, error) {
	return f.page, nil
}
func (f *fakeLoader) Get(context.Context, string) (workbenchesbridge.Detail, error) {
	return f.detail, nil
}
func (f *fakeLoader) FollowUp(_ context.Context, jobID, prompt string, _ time.Duration) (workbenchesbridge.PromptResult, error) {
	f.gotJob, f.gotPrompt = jobID, prompt
	return f.result, nil
}

func TestJobDetailScrollsPrompt(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 40; i++ {
		fmt.Fprintf(&b, "LINE-%02d unique-content\n", i)
	}
	loader := &fakeLoader{
		page: workbenchesbridge.Page{Items: []workbenchesbridge.Summary{{ID: "job-1", WorkbenchName: "triage"}}},
		detail: workbenchesbridge.Detail{Summary: workbenchesbridge.Summary{
			ID:     "job-1",
			Prompt: b.String(),
		}},
	}
	model := listed(t, loader)
	model, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(cmd())
	top := normalizeView(model.View(80, 24))
	if !strings.Contains(top, "LINE-00 unique-content") {
		t.Fatalf("expected start of prompt:\n%s", top)
	}
	if strings.Contains(top, "LINE-39 unique-content") {
		t.Fatalf("entire prompt fit without scrolling:\n%s", top)
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyPgDown})
	mid := normalizeView(model.View(80, 24))
	if strings.Contains(mid, "LINE-00 unique-content") {
		t.Fatalf("pgdown did not move the prompt:\n%s", mid)
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnd})
	bottom := normalizeView(model.View(80, 24))
	if !strings.Contains(bottom, "LINE-39 unique-content") {
		t.Fatalf("end did not reveal the end of the prompt:\n%s", bottom)
	}
}

func TestJobDetailWrapsMultilinePrompt(t *testing.T) {
	prompt := "You've been assigned the following issue from github, please investigate everything necessary to complete the task. The issue will be described below:\n\n## Comment on something"
	loader := &fakeLoader{
		page: workbenchesbridge.Page{Items: []workbenchesbridge.Summary{{
			ID:            "21968068-f35f-41a3-9a27-3a1d5ef56df5",
			WorkbenchName: "DevOps Automation",
			Status:        "successful",
			Prompt:        prompt,
		}}},
		detail: workbenchesbridge.Detail{Summary: workbenchesbridge.Summary{
			ID:            "21968068-f35f-41a3-9a27-3a1d5ef56df5",
			WorkbenchName: "DevOps Automation",
			Status:        "successful",
			Prompt:        prompt,
			InsertedAt:    "2026-09-08T12:52:44.988123Z",
		}},
	}
	model := listed(t, loader)
	got := normalizeView(model.View(80, 24))
	if strings.Contains(got, "\n## Comment") {
		t.Fatalf("list row should collapse prompt newlines:\n%s", got)
	}
	model, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(cmd())
	got = normalizeView(model.View(80, 24))
	jobAt := strings.Index(got, "21968068-f35f-41a3-9a27-3a1d5ef56df5")
	promptAt := strings.Index(got, "You've been assigned")
	headingAt := strings.Index(got, "## Comment on something")
	if jobAt < 0 || promptAt < 0 || headingAt < 0 {
		t.Fatalf("detail missing fields:\n%s", got)
	}
	if !(jobAt < promptAt && promptAt < headingAt) {
		t.Fatalf("expected Job ID, then wrapped prompt:\n%s", got)
	}
	if !strings.Contains(got, "2026-09-08 12:52 UTC") {
		t.Fatalf("timestamp not formatted:\n%s", got)
	}
}

func TestListStillBrowsesJobs(t *testing.T) {
	model := listed(t, &fakeLoader{
		page: workbenchesbridge.Page{Items: []workbenchesbridge.Summary{{ID: "job-1", WorkbenchName: "triage", Prompt: "investigate"}}},
	})
	got := normalizeView(model.View(80, 24))
	if !strings.Contains(got, "Recent workbench jobs") || !strings.Contains(got, "triage") {
		t.Fatalf("list missing jobs:\n%s", got)
	}
	model, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter did not open detail")
	}
}

func TestFollowUpPromptLooksLikeOtherInputs(t *testing.T) {
	model := listed(t, &fakeLoader{
		page: workbenchesbridge.Page{Items: []workbenchesbridge.Summary{{ID: "job-1"}}},
	})
	model, _ = model.Update(tea.KeyPressMsg{Code: 'f', Text: "f"})
	model.prompt.SetValue("hello")
	got := normalizeView(model.View(80, 24))
	if strings.Contains(got, " 1 hello") || strings.Contains(got, "ctrl+s") {
		t.Fatalf("prompt still looks like a textarea:\n%s", got)
	}
	if !strings.Contains(got, "› hello") {
		t.Fatalf("expected single-line prompt input:\n%s", got)
	}
	if !strings.Contains(got, "enter review") {
		t.Fatalf("expected enter to review:\n%s", got)
	}
}

func TestFollowUpQueuesSelectedJob(t *testing.T) {
	loader := &fakeLoader{
		page: workbenchesbridge.Page{Items: []workbenchesbridge.Summary{{ID: "job-1", WorkbenchName: "triage"}}},
		result: workbenchesbridge.PromptResult{
			ID:          "prompt-1",
			WorkbenchID: "job-1",
			DequeueAt:   "2026-09-10T10:00:00Z",
		},
	}
	model := listed(t, loader)
	model, _ = model.Update(tea.KeyPressMsg{Code: 'f', Text: "f"})
	if model.mode != modePrompt || model.detail.ID != "job-1" {
		t.Fatalf("expected prompt for selected job, mode=%d id=%q", model.mode, model.detail.ID)
	}
	model.prompt.SetValue("verify the fix")
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if model.mode != modeReview {
		t.Fatalf("expected review mode, got %d", model.mode)
	}
	got := normalizeView(model.View(80, 24))
	if !strings.Contains(got, "verify the fix") || !strings.Contains(got, "job-1") {
		t.Fatalf("review missing job fields:\n%s", got)
	}
	model, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(cmd())
	if model.mode != modeResult || model.result.ID != "prompt-1" {
		t.Fatalf("unexpected result: %#v", model.result)
	}
	if loader.gotJob != "job-1" || loader.gotPrompt != "verify the fix" {
		t.Fatalf("follow-up job=%q prompt=%q", loader.gotJob, loader.gotPrompt)
	}
}

func listed(t *testing.T, loader workbenchesbridge.Loader) Model {
	t.Helper()
	model := New(t.Context(), loader, theme.New(colorprofile.ASCII))
	model, cmd := model.Update(model.Init()())
	if cmd == nil {
		t.Fatal("init did not list jobs")
	}
	model, _ = model.Update(cmd())
	return model
}

func normalizeView(view string) string {
	lines := strings.Split(ansi.Strip(view), "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " ")
	}
	return strings.Join(lines, "\n")
}
