package down

import (
	"context"
	"os"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"

	upbridge "github.com/pluralsh/plural-cli/pkg/bridge/up"
	"github.com/pluralsh/plural-cli/tui/navigation"
	"github.com/pluralsh/plural-cli/tui/theme"
)

type stubRunner struct {
	calls []upbridge.DestroyInput
	err   error
}

func (s *stubRunner) Run(context.Context, upbridge.RunInput, upbridge.ProgressFunc) (upbridge.RunResult, error) {
	return upbridge.RunResult{}, nil
}

func (s *stubRunner) Deploy(context.Context, upbridge.DeployInput, upbridge.ProgressFunc) error {
	return nil
}

func (s *stubRunner) Destroy(_ context.Context, in upbridge.DestroyInput, progress upbridge.ProgressFunc) error {
	s.calls = append(s.calls, in)
	if progress != nil {
		progress("Building destroy context…")
		progress("Destroying management cluster terraform…")
	}
	return s.err
}

func testModel(t *testing.T) (Model, *stubRunner) {
	t.Helper()
	runner := &stubRunner{}
	model := New(t.Context(), theme.New(colorprofile.ASCII))
	model.runner = runner
	model.exportDir = t.TempDir()
	return model, runner
}

func TestDownSelectsSelfHostedThenAffirmsAndDestroys(t *testing.T) {
	model, runner := testModel(t)
	if model.ModeName() != "select-cloud" {
		t.Fatalf("mode = %s", model.ModeName())
	}
	view := model.View(80, 24)
	if !strings.Contains(view, "Destroy mode") || !strings.Contains(view, "Self-hosted") {
		t.Fatalf("select view:\n%s", view)
	}

	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if model.ModeName() != "affirm" {
		t.Fatalf("after enter mode = %s", model.ModeName())
	}
	if model.Cloud() {
		t.Fatal("expected self-hosted")
	}
	if !strings.Contains(model.View(80, 24), "Are you ready to destroy") {
		t.Fatalf("affirm view:\n%s", model.View(80, 24))
	}

	model, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if model.ModeName() != "destroying" {
		t.Fatalf("after affirm mode = %s", model.ModeName())
	}
	if cmd == nil {
		t.Fatal("expected destroy cmd")
	}
	// drain batch: spinner tick + destroyCmd
	msg := model.destroyCmd(upbridge.DestroyInput{Cloud: false})()
	model, _ = model.Update(msg)
	if model.ModeName() != "complete" {
		t.Fatalf("after destroy mode = %s err=%v", model.ModeName(), model.err)
	}
	if model.err != nil {
		t.Fatal(model.err)
	}
	if len(runner.calls) != 1 || runner.calls[0].Cloud {
		t.Fatalf("calls = %#v", runner.calls)
	}
	if !strings.Contains(model.View(80, 24), "destroy finished") {
		t.Fatalf("complete view:\n%s", model.View(80, 24))
	}
}

func TestDownCloudShortcut(t *testing.T) {
	model, runner := testModel(t)
	model, _ = model.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
	if model.ModeName() != "affirm" || !model.Cloud() {
		t.Fatalf("mode=%s cloud=%v", model.ModeName(), model.Cloud())
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})
	msg := model.destroyCmd(upbridge.DestroyInput{Cloud: true})()
	model, _ = model.Update(msg)
	if len(runner.calls) != 1 || !runner.calls[0].Cloud {
		t.Fatalf("calls = %#v", runner.calls)
	}
}

func TestDownAffirmNoCancels(t *testing.T) {
	model, _ := testModel(t)
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
	if model.ModeName() != "select-cloud" {
		t.Fatalf("mode = %s", model.ModeName())
	}
	if model.err == nil || !strings.Contains(model.err.Error(), "cancelled") {
		t.Fatalf("err = %v", model.err)
	}
}

func TestDownEscReturnsWelcome(t *testing.T) {
	model, _ := testModel(t)
	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected navigate")
	}
	if got := cmd().(navigation.NavigateMsg).Route; got != navigation.Welcome {
		t.Fatalf("route = %q", got)
	}
}

func TestDownDestroyFailure(t *testing.T) {
	model, runner := testModel(t)
	runner.err = context.Canceled
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})
	msg := model.destroyCmd(upbridge.DestroyInput{})()
	model, _ = model.Update(msg)
	if model.ModeName() != "complete" || model.err == nil {
		t.Fatalf("mode=%s err=%v", model.ModeName(), model.err)
	}
	if !strings.Contains(model.View(80, 24), "Destroy failed") {
		t.Fatalf("view:\n%s", model.View(80, 24))
	}
	if model.logExportPath == "" {
		t.Fatal("expected auto-export on destroy error")
	}
	if !strings.Contains(model.View(80, 24), "Saved ") {
		t.Fatalf("view should show export path:\n%s", model.View(80, 24))
	}
}

func TestExportLogsOnComplete(t *testing.T) {
	model, _ := testModel(t)
	model.mode = modeComplete
	model.opLog = []string{"destroy-log-line"}
	model, _ = model.Update(tea.KeyPressMsg{Code: 'e', Text: "e"})
	if model.logExportPath == "" || model.logExportErr != nil {
		t.Fatalf("export path=%q err=%v", model.logExportPath, model.logExportErr)
	}
	data, err := os.ReadFile(model.logExportPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "destroy-log-line") {
		t.Fatalf("file:\n%s", data)
	}
}

func TestShouldNotExecStubRunner(t *testing.T) {
	if shouldExecLiveRunner(&stubRunner{}) {
		t.Fatal("stub should stay in-process")
	}
	if !shouldExecLiveRunner(upbridge.LiveRunner{}) {
		t.Fatal("live should stream")
	}
}

func TestOpLogLinesWithLongBuffer(t *testing.T) {
	model, _ := testModel(t)
	model.mode = modeDestroying
	for i := 0; i < 100; i++ {
		model.opLog = append(model.opLog, "line")
	}
	// Must not panic (regression: make cap used limit-start which went negative).
	_ = model.View(80, 24)
	lines := model.opLogLines(12, 80)
	if len(lines) != 12 {
		t.Fatalf("len=%d want 12", len(lines))
	}
}

func TestOpLogLinesUsePanelWidth(t *testing.T) {
	model, _ := testModel(t)
	model.mode = modeComplete
	long := strings.Repeat("abcdefghij", 20) // 200 chars
	model.opLog = []string{long}
	view := ansi.Strip(model.View(160, 24))
	if !strings.Contains(view, strings.Repeat("abcdefghij", 12)) {
		t.Fatalf("expected wrapped log to keep the start of the line, got:\n%s", view)
	}
	if !strings.Contains(view, long[len(long)-40:]) {
		t.Fatalf("expected long log line to wrap instead of truncate, got:\n%s", view)
	}
	if strings.Contains(view, long) {
		t.Fatal("expected wrap (newline) so the 200-char line is not a single row")
	}
}

func TestOpLogScrollOnComplete(t *testing.T) {
	model, _ := testModel(t)
	model.mode = modeComplete
	model.viewH = 24
	model.opLogFollow = true
	for i := 0; i < 40; i++ {
		model.opLog = append(model.opLog, "line")
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyHome})
	if model.opLogFollow || model.opLogY != 0 {
		t.Fatalf("home: follow=%v y=%d", model.opLogFollow, model.opLogY)
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnd})
	if !model.opLogFollow {
		t.Fatal("end should follow")
	}
	view := model.View(80, 24)
	if !strings.Contains(view, "destroy finished") || !strings.Contains(view, "Scroll logs") {
		t.Fatalf("complete should keep logs:\n%s", view)
	}
}

func TestCompleteWaitsForUser(t *testing.T) {
	model, _ := testModel(t)
	model.mode = modeComplete
	model.opLog = []string{"destroy-log-line"}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if model.ModeName() != "complete" {
		t.Fatalf("scroll must stay on complete, mode=%s", model.ModeName())
	}
	if !strings.Contains(model.View(80, 24), "destroy-log-line") {
		t.Fatalf("logs should remain:\n%s", model.View(80, 24))
	}
	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter should navigate welcome")
	}
	msg := cmd()
	nav, ok := msg.(navigation.NavigateMsg)
	if !ok || nav.Route != navigation.Welcome {
		t.Fatalf("nav = %#v", msg)
	}
}

func TestResetClearsCompletedRun(t *testing.T) {
	model, runner := testModel(t)
	model.mode = modeComplete
	model.err = context.Canceled
	model.opLog = []string{"old-destroy-log"}
	model.cloud = true
	model = model.Reset()
	if model.ModeName() != "select-cloud" {
		t.Fatalf("mode=%s", model.ModeName())
	}
	if model.err != nil || model.Cloud() || len(model.opLog) != 0 {
		t.Fatalf("stale state: err=%v cloud=%v logs=%d", model.err, model.Cloud(), len(model.opLog))
	}
	if model.runner != runner {
		t.Fatal("reset should keep the runner")
	}
	if strings.Contains(model.View(80, 24), "old-destroy-log") || strings.Contains(model.View(80, 24), "Destroy complete") {
		t.Fatalf("expected a fresh destroy picker:\n%s", model.View(80, 24))
	}
}
