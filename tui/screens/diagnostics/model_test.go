package diagnostics

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"

	bridge "github.com/pluralsh/plural-cli/pkg/bridge/welcome"
	"github.com/pluralsh/plural-cli/tui/navigation"
	"github.com/pluralsh/plural-cli/tui/theme"
)

type loaderFunc func(context.Context) (bridge.Snapshot, error)

func (f loaderFunc) Load(ctx context.Context) (bridge.Snapshot, error) { return f(ctx) }

func TestDiagnosticsLoadsContextAndReturnsToWelcome(t *testing.T) {
	model := New(t.Context(), loaderFunc(func(context.Context) (bridge.Snapshot, error) {
		return bridge.Snapshot{App: bridge.AppProfile{Configured: true, Email: "dev@example.com"}, Diagnostics: []string{"workspace: invalid"}}, nil
	}), theme.New(colorprofile.ASCII))
	model, _ = model.Update(model.Init()())
	view := model.View(100, 30)
	if !strings.Contains(view, "dev@example.com") || !strings.Contains(view, "workspace: invalid") {
		t.Fatalf("diagnostics missing context:\n%s", view)
	}
	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil || cmd().(navigation.NavigateMsg).Route != navigation.Welcome {
		t.Fatal("esc did not return to welcome")
	}
}
