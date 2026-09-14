// Package ai exposes Plural App chat to presentation layers.
package ai

import (
	"context"
	"errors"
	"strings"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/bridge"
	"github.com/pluralsh/plural-cli/pkg/config"
)

// Intro is the system prompt used by `plural ai` and the TUI chat screen.
const Intro = "What can we do to help you with Plural, using open source, or kubernetes?"

const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

var errNoApp = errors.New("connect a Plural App profile before chatting")

// Message is one turn in an App-API chat history.
type Message struct {
	Name    string
	Content string
	Role    string
}

// Client is the TUI-facing chat surface.
type Client interface {
	Chat(context.Context, []Message) (Message, error)
}

// API is the App GraphQL chat method used by Service.
type API interface {
	Chat(history []*api.ChatMessage) (*api.ChatMessage, error)
}

// ClientFactory constructs an App chat client that honors caller cancellation.
type ClientFactory func(ctx context.Context) (API, error)

// Service implements Client against the Plural App API.
type Service struct {
	newClient ClientFactory
}

// NewService chats with the active Plural App token from ~/.plural/config.yml.
func NewService() *Service {
	return &Service{newClient: defaultClient}
}

func defaultClient(ctx context.Context) (API, error) {
	if !config.Exists() || strings.TrimSpace(config.Read().Token) == "" {
		return nil, &bridge.Error{Code: bridge.ErrorUnauthenticated, Err: errNoApp}
	}
	conf := config.Read()
	return api.FromConfigWithContext(ctx, &conf), nil
}

func (s *Service) client(ctx context.Context) (API, error) {
	if s == nil || s.newClient == nil {
		return nil, &bridge.Error{Code: bridge.ErrorUnauthenticated, Err: errNoApp}
	}
	return s.newClient(ctx)
}

// Chat sends history to Plural App and returns the assistant reply.
func (s *Service) Chat(ctx context.Context, history []Message) (Message, error) {
	if err := ctx.Err(); err != nil {
		return Message{}, err
	}
	client, err := s.client(ctx)
	if err != nil {
		return Message{}, err
	}
	hist := make([]*api.ChatMessage, len(history))
	for i, message := range history {
		hist[i] = &api.ChatMessage{Name: message.Name, Content: message.Content, Role: message.Role}
	}
	reply, err := client.Chat(hist)
	if err != nil {
		return Message{}, err
	}
	if reply == nil {
		return Message{}, &bridge.Error{Code: bridge.ErrorUnavailable, Err: errors.New("empty chat reply")}
	}
	return Message{Name: reply.Name, Content: reply.Content, Role: reply.Role}, nil
}
