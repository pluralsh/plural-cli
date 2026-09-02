package down

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"

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
}

func TestShouldNotExecStubRunner(t *testing.T) {
	if shouldExecLiveRunner(&stubRunner{}) {
		t.Fatal("stub should stay in-process")
	}
	if !shouldExecLiveRunner(upbridge.LiveRunner{}) {
		t.Fatal("live should tea.Exec")
	}
}
