package ai

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"

	"github.com/pluralsh/plural-cli/tui/navigation"
	"github.com/pluralsh/plural-cli/tui/theme"
)

func TestAIHubRoutesToInteractiveScreens(t *testing.T) {
	model := New(theme.New(colorprofile.ASCII))
	if got := normalizeView(model.View(80, 24)); !strings.Contains(got, "AI workspaces") || !strings.Contains(got, "Agents") {
		t.Fatalf("hub view missing entries:\n%s", got)
	}
	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil || cmd() != (navigation.NavigateMsg{Route: navigation.Agents}) {
		t.Fatal("agents selection did not navigate to the interactive screen")
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	_, cmd = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil || cmd() != (navigation.NavigateMsg{Route: navigation.Workbenches}) {
		t.Fatal("workbenches selection did not navigate to the interactive screen")
	}
}
