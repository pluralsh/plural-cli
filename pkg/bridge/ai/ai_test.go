package ai

import (
	"context"
	"errors"
	"testing"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/bridge"
)

type fakeAPI struct {
	history []*api.ChatMessage
	reply   *api.ChatMessage
	err     error
}

func (f *fakeAPI) Chat(history []*api.ChatMessage) (*api.ChatMessage, error) {
	f.history = history
	return f.reply, f.err
}

func TestChatMapsHistoryAndReply(t *testing.T) {
	fake := &fakeAPI{reply: &api.ChatMessage{Role: RoleAssistant, Name: "plural", Content: "try plural up"}}
	service := &Service{newClient: func(context.Context) (API, error) { return fake, nil }}
	reply, err := service.Chat(t.Context(), []Message{
		{Role: RoleSystem, Content: Intro},
		{Role: RoleUser, Content: "how do I bootstrap?"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if reply.Content != "try plural up" || reply.Role != RoleAssistant {
		t.Fatalf("reply = %#v", reply)
	}
	if len(fake.history) != 2 || fake.history[1].Content != "how do I bootstrap?" {
		t.Fatalf("history = %#v", fake.history)
	}
}

func TestChatRequiresAppToken(t *testing.T) {
	service := &Service{}
	_, err := service.Chat(t.Context(), []Message{{Role: RoleUser, Content: "hi"}})
	if !bridge.IsCode(err, bridge.ErrorUnauthenticated) {
		t.Fatalf("err = %v", err)
	}
}

func TestChatHonorsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	service := &Service{newClient: func(context.Context) (API, error) {
		t.Fatal("factory should not run after cancel")
		return nil, nil
	}}
	if _, err := service.Chat(ctx, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v", err)
	}
}

func TestChatRejectsEmptyReply(t *testing.T) {
	service := &Service{newClient: func(context.Context) (API, error) {
		return &fakeAPI{}, nil
	}}
	_, err := service.Chat(t.Context(), nil)
	if !bridge.IsCode(err, bridge.ErrorUnavailable) {
		t.Fatalf("err = %v", err)
	}
}
