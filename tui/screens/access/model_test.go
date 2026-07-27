package access

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"

	accessbridge "github.com/pluralsh/plural-cli/pkg/bridge/access"
	"github.com/pluralsh/plural-cli/tui/navigation"
	"github.com/pluralsh/plural-cli/tui/theme"
)

type fakeManager struct {
	snapshot                                     accessbridge.Snapshot
	activatedApp, activatedConsole, impersonated string
	stopped                                      bool
}

func (f *fakeManager) Load(context.Context) (accessbridge.Snapshot, error) { return f.snapshot, nil }
func (f *fakeManager) BeginDeviceLogin(context.Context, string) (accessbridge.DeviceAuthorization, error) {
	return accessbridge.DeviceAuthorization{LoginURL: "https://login.example.com", DeviceToken: "device"}, nil
}
func (f *fakeManager) CompleteDeviceLogin(context.Context, string, accessbridge.DeviceAuthorization, string) (accessbridge.Profile, error) {
	return accessbridge.Profile{ID: "new"}, nil
}
func (f *fakeManager) AddConsoleProfile(context.Context, string, string, string) (accessbridge.ConsoleProfile, error) {
	return accessbridge.ConsoleProfile{}, nil
}
func (f *fakeManager) ActivateProfile(_ context.Context, id string) error {
	f.activatedApp = id
	return nil
}
func (f *fakeManager) ActivateConsole(_ context.Context, id string) error {
	f.activatedConsole = id
	return nil
}
func (f *fakeManager) SearchServiceAccounts(context.Context, string) ([]accessbridge.ServiceAccount, error) {
	return []accessbridge.ServiceAccount{{ID: "sa", Email: "deploy@example.com"}}, nil
}
func (f *fakeManager) Impersonate(_ context.Context, email string) error {
	f.impersonated = email
	return nil
}
func (f *fakeManager) StopImpersonating() { f.stopped = true }

func loadedModel(t *testing.T, manager *fakeManager) Model {
	model := New(t.Context(), manager, theme.New(colorprofile.ASCII))
	cmd := model.Init()
	if cmd == nil {
		t.Fatal("Init() returned nil")
	}
	model, _ = model.Update(cmd())
	return model
}

func TestFirstRunCanSkipConsoleAndReturnToWelcome(t *testing.T) {
	model := loadedModel(t, &fakeManager{})
	view := model.View(100, 30)
	if !strings.Contains(view, "Skipped for now") || !strings.Contains(view, "Press n to sign in") {
		t.Fatalf("first-run guidance missing:\n%s", view)
	}
	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil || cmd().(navigation.NavigateMsg).Route != navigation.Welcome {
		t.Fatal("esc did not return to welcome")
	}
}

func TestPanelsActivateIndependently(t *testing.T) {
	manager := &fakeManager{snapshot: accessbridge.Snapshot{State: accessbridge.State{
		Profiles: []accessbridge.Profile{{ID: "app-a"}, {ID: "app-b"}}, ActiveProfileID: "app-a",
		ConsoleProfiles: []accessbridge.ConsoleProfile{{ID: "console-a"}, {ID: "console-b"}}, ActiveConsoleID: "console-a",
	}}}
	model := loadedModel(t, manager)
	if model.loading || model.mode != modeProfiles || model.panel != 0 || len(model.snapshot.State.Profiles) != 2 {
		t.Fatalf("loaded model = loading:%v mode:%v panel:%d profiles:%d", model.loading, model.mode, model.panel, len(model.snapshot.State.Profiles))
	}
	down := tea.KeyPressMsg{Code: 'j', Text: "j"}
	model, _ = model.Update(down)
	if model.appCursor != 1 {
		t.Fatalf("App cursor = %d after string=%q keystroke=%q (error %v)", model.appCursor, down.String(), down.Keystroke(), model.err)
	}
	model, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	cmd()
	if manager.activatedApp != "app-b" {
		t.Fatalf("activated App = %q", manager.activatedApp)
	}
	model.loading = false
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	model, _ = model.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	_, cmd = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	cmd()
	if manager.activatedConsole != "console-b" {
		t.Fatalf("activated Console = %q", manager.activatedConsole)
	}
}

func TestServiceAccountPickerCreatesSessionAction(t *testing.T) {
	manager := &fakeManager{snapshot: accessbridge.Snapshot{State: accessbridge.State{Profiles: []accessbridge.Profile{{ID: "app"}}, ActiveProfileID: "app"}}}
	model := loadedModel(t, manager)
	model, cmd := model.Update(tea.KeyPressMsg{Code: 'i', Text: "i"})
	model, _ = model.Update(cmd())
	if model.mode != modeAccounts {
		t.Fatal("service-account picker did not open")
	}
	_, cmd = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	cmd()
	if manager.impersonated != "deploy@example.com" {
		t.Fatalf("impersonated = %q", manager.impersonated)
	}
}

func TestConsoleTokenIsMasked(t *testing.T) {
	model := New(t.Context(), &fakeManager{}, theme.New(colorprofile.ASCII))
	model.mode = modeConsoleForm
	model.form[2].SetValue("super-secret")
	model.formIndex = 2
	model.form[2].Focus()
	view := model.View(100, 30)
	if strings.Contains(view, "super-secret") {
		t.Fatalf("Console token leaked in view:\n%s", view)
	}
}

func TestDeviceLoginCanBeCancelledBeforeGlobalQuit(t *testing.T) {
	model := loadedModel(t, &fakeManager{})
	model, _ = model.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
	if !model.HasCancellableOperation() {
		t.Fatal("device login is not cancellable")
	}
	model, cmd := model.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if cmd != nil || model.HasCancellableOperation() || model.mode != modeProfiles {
		t.Fatalf("cancel left operation active: %#v", model)
	}
}
