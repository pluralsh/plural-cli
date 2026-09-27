package ai

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"

	"github.com/pluralsh/plural-cli/pkg/bridge"
	aibridge "github.com/pluralsh/plural-cli/pkg/bridge/ai"
	"github.com/pluralsh/plural-cli/tui/navigation"
	"github.com/pluralsh/plural-cli/tui/theme"
)

type fakeChat struct {
	history []aibridge.Message
	reply   aibridge.Message
	err     error
}

func (f *fakeChat) Chat(_ context.Context, history []aibridge.Message) (aibridge.Message, error) {
	f.history = append([]aibridge.Message(nil), history...)
	return f.reply, f.err
}

func TestAIHubRoutesToInteractiveScreens(t *testing.T) {
	model := New(t.Context(), nil, theme.New(colorprofile.ASCII))
	got := normalizeView(model.View(80, 24))
	if !strings.Contains(got, "AI workspaces") || !strings.Contains(got, "Chat") || !strings.Contains(got, "Agents") {
		t.Fatalf("hub view missing entries:\n%s", got)
	}
	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd != nil {
		t.Fatal("chat selection should stay on the AI screen")
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	_, cmd = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil || cmd() != (navigation.NavigateMsg{Route: navigation.Agents}) {
		t.Fatal("agents selection did not navigate to the interactive screen")
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	_, cmd = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil || cmd() != (navigation.NavigateMsg{Route: navigation.Workbenches}) {
		t.Fatal("workbenches selection did not navigate to the interactive screen")
	}
}

func TestAIHubChatShortcutOpensConversation(t *testing.T) {
	model := New(t.Context(), nil, theme.New(colorprofile.ASCII))
	model, cmd := model.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
	if cmd != nil {
		t.Fatal("opening chat emitted navigation")
	}
	if model.mode != modeChat {
		t.Fatalf("mode = %d, want chat", model.mode)
	}
	got := normalizeView(model.View(80, 24))
	if !strings.Contains(got, "Conversation") || !strings.Contains(got, "What can we do to help you with Plural") {
		t.Fatalf("chat view missing intro:\n%s", got)
	}
}

func TestAIChatSendsAndAppendsReply(t *testing.T) {
	client := &fakeChat{reply: aibridge.Message{Role: aibridge.RoleAssistant, Content: "run plural up"}}
	model := New(t.Context(), client, theme.New(colorprofile.ASCII))
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model.input.SetValue("how do I bootstrap?")
	model, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("send did not start chat")
	}
	if !model.thinking || !model.HasCancellableOperation() {
		t.Fatal("expected in-flight chat")
	}
	model, _ = model.Update(cmd())
	if model.thinking {
		t.Fatal("reply left thinking state")
	}
	if len(model.history) != 3 || model.history[2].Content != "run plural up" {
		t.Fatalf("history = %#v", model.history)
	}
	if len(client.history) != 2 || client.history[1].Content != "how do I bootstrap?" {
		t.Fatalf("client history = %#v", client.history)
	}
	got := normalizeView(model.View(80, 24))
	if !strings.Contains(got, "run plural up") || !strings.Contains(got, "You") {
		t.Fatalf("transcript missing reply:\n%s", got)
	}
}

func TestAIChatUnauthenticatedOpensAccess(t *testing.T) {
	client := &fakeChat{err: &bridge.Error{Code: bridge.ErrorUnauthenticated, Err: errors.New("connect")}}
	model := New(t.Context(), client, theme.New(colorprofile.ASCII))
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model.input.SetValue("hello")
	model, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(cmd())
	if !model.needsAuth {
		t.Fatal("expected app login required")
	}
	got := normalizeView(model.View(80, 24))
	if !strings.Contains(got, "App login required") {
		t.Fatalf("missing auth panel:\n%s", got)
	}
	_, nav := model.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
	if nav == nil || nav() != (navigation.NavigateMsg{Route: navigation.Access}) {
		t.Fatal("c did not open Access")
	}
	model, _ = model.Update(model.Init()())
	if model.needsAuth || model.mode != modeChat {
		t.Fatal("returning to AI should retry chat after App login")
	}
}

func TestAIChatEscReturnsToHub(t *testing.T) {
	model := New(t.Context(), nil, theme.New(colorprofile.ASCII))
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if cmd != nil {
		t.Fatal("esc from chat should stay on the AI screen")
	}
	if model.mode != modeHub {
		t.Fatalf("mode = %d, want hub", model.mode)
	}
}

func TestAIChatIgnoresStaleReplyAfterCancel(t *testing.T) {
	model := New(t.Context(), &fakeChat{reply: aibridge.Message{Role: aibridge.RoleAssistant, Content: "late"}}, theme.New(colorprofile.ASCII))
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model.input.SetValue("hello")
	model, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if model.thinking {
		t.Fatal("cancel left thinking state")
	}
	model, _ = model.Update(cmd())
	if len(model.history) != 2 {
		t.Fatalf("stale reply appended: %#v", model.history)
	}
}
