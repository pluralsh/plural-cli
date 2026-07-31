package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"

	welcomebridge "github.com/pluralsh/plural-cli/pkg/bridge/welcome"
	"github.com/pluralsh/plural-cli/tui/navigation"
	"github.com/pluralsh/plural-cli/tui/theme"
)

type welcomeLoaderFunc func(context.Context) (welcomebridge.Snapshot, error)

func (f welcomeLoaderFunc) Load(ctx context.Context) (welcomebridge.Snapshot, error) { return f(ctx) }

func TestModelHandlesResizeAndQuit(t *testing.T) {
	model := New(t.Context(), theme.New(colorprofile.ASCII), Dependencies{})
	updated, cmd := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	if cmd != nil {
		t.Fatal("resize returned a command")
	}
	resized := updated.(Model)
	if resized.width != 100 || resized.height != 30 {
		t.Fatalf("size = %dx%d", resized.width, resized.height)
	}

	_, cmd = resized.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if cmd == nil {
		t.Fatal("quit key did not return a command")
	}
	if msg := cmd(); msg == nil {
		t.Fatal("quit command returned nil")
	}
}

func TestModelRoutesScreensWithoutRebuildingShell(t *testing.T) {
	model := New(t.Context(), theme.New(colorprofile.ASCII), Dependencies{})
	updated, _ := model.Update(navigation.NavigateMsg{Route: navigation.Diagnostics})
	routed := updated.(Model)
	if routed.route != navigation.Diagnostics || !strings.Contains(routed.View().Content, "Diagnostics") {
		t.Fatalf("route/view = %q\n%s", routed.route, routed.View().Content)
	}
	updated, _ = routed.Update(navigation.NavigateMsg{Route: navigation.Deployments})
	routed = updated.(Model)
	if routed.route != navigation.Deployments || !strings.Contains(routed.View().Content, "CD / Deployments") {
		t.Fatalf("deployments route/view = %q\n%s", routed.route, routed.View().Content)
	}
	updated, _ = routed.Update(navigation.NavigateMsg{Route: navigation.Services})
	routed = updated.(Model)
	if routed.route != navigation.Services || !strings.Contains(routed.View().Content, "Services") {
		t.Fatalf("services route/view = %q\n%s", routed.route, routed.View().Content)
	}
	updated, _ = routed.Update(navigation.NavigateMsg{Route: navigation.Clusters})
	routed = updated.(Model)
	if routed.route != navigation.Clusters || !strings.Contains(routed.View().Content, "Clusters") {
		t.Fatalf("clusters route/view = %q\n%s", routed.route, routed.View().Content)
	}
	updated, _ = routed.Update(navigation.NavigateMsg{Route: navigation.Repositories})
	routed = updated.(Model)
	if routed.route != navigation.Repositories || !strings.Contains(routed.View().Content, "Repositories") {
		t.Fatalf("repositories route/view = %q\n%s", routed.route, routed.View().Content)
	}
	updated, _ = routed.Update(navigation.NavigateMsg{Route: navigation.Pipelines})
	routed = updated.(Model)
	if routed.route != navigation.Pipelines || !strings.Contains(routed.View().Content, "Pipelines") {
		t.Fatalf("pipelines route/view = %q\n%s", routed.route, routed.View().Content)
	}
	updated, _ = routed.Update(navigation.NavigateMsg{Route: navigation.AI})
	routed = updated.(Model)
	if routed.route != navigation.AI || !strings.Contains(routed.View().Content, "AI workspaces") {
		t.Fatalf("ai route/view = %q\n%s", routed.route, routed.View().Content)
	}
	updated, _ = routed.Update(navigation.NavigateMsg{Route: navigation.Welcome})
	if got := updated.(Model).route; got != navigation.Welcome {
		t.Fatalf("route = %q", got)
	}
}

func TestModelRoutesAIDedicatedScreen(t *testing.T) {
	model := New(t.Context(), theme.New(colorprofile.ASCII), Dependencies{})
	updated, _ := model.Update(navigation.NavigateMsg{Route: navigation.AI})
	routed := updated.(Model)
	if routed.route != navigation.AI || !strings.Contains(routed.View().Content, "AI workspaces") {
		t.Fatalf("ai route/view = %q\n%s", routed.route, routed.View().Content)
	}
}

func TestRunRejectsMissingTerminal(t *testing.T) {
	if err := Run(t.Context(), nil, nil, Dependencies{}); !errors.Is(err, ErrNoTerminal) {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestModelLoadsWelcomeSnapshot(t *testing.T) {
	loader := welcomeLoaderFunc(func(context.Context) (welcomebridge.Snapshot, error) {
		return welcomebridge.Snapshot{
			Version: "v1.0.0",
			App:     welcomebridge.AppProfile{Configured: true, Email: "dev@example.com"},
		}, nil
	})
	model := New(t.Context(), theme.New(colorprofile.ASCII), Dependencies{Welcome: loader})
	cmd := model.Init()
	if cmd == nil {
		t.Fatal("Init() did not load the welcome snapshot")
	}
	updated := tea.Model(model)
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, batchCmd := range batch {
			updated, _ = updated.Update(batchCmd())
		}
	} else {
		updated, _ = updated.Update(msg)
	}
	loaded := updated.(Model)
	loaded.width = 80
	if got := loaded.View().Content; !strings.Contains(got, "dev@example.com") {
		t.Fatalf("welcome view does not contain loaded identity:\n%s", got)
	}
}
