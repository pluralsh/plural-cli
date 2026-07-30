// Package notifications exposes read-only Console notification sink list/get
// use cases to presentation layers without importing TUI code.
package notifications

import (
	"context"
	"errors"
	"strings"

	gqlclient "github.com/pluralsh/console/go/client"

	"github.com/pluralsh/plural-cli/pkg/bridge"
	"github.com/pluralsh/plural-cli/pkg/console"
)

const (
	defaultPageSize int64 = 10
	fetchPageSize   int64 = 100
)

var (
	errNoConsole   = errors.New("connect a Console profile before browsing Console resources")
	errMissingID   = errors.New("notification sink id is required")
	errMissingSink = errors.New("notification sink was not found")
)

// Summary is a credential-free list row for a notification sink.
type Summary struct {
	ID   string
	Name string
	Type string
	URL  string
}

// Binding is a credential-free notification binding (user or group).
type Binding struct {
	Kind string
	Name string
}

// Detail is the credential-free detail payload for a notification sink.
type Detail struct {
	Summary
	Bindings []Binding
}

// Page is one cursor page of sink summaries.
type Page struct {
	Items      []Summary
	EndCursor  string
	HasNext    bool
	TotalShown int
}

// Loader is the narrow contract consumed by the Notifications screen.
type Loader interface {
	List(ctx context.Context, after *string, query string) (Page, error)
	Get(ctx context.Context, id string) (Detail, error)
}

// ConsoleResolver supplies the active Console URL and token.
type ConsoleResolver interface {
	ActiveConsole(ctx context.Context) (url, token string, err error)
}

// API is the Console surface required by this package.
type API interface {
	ListNotificationSinks(after *string, first *int64) (*gqlclient.ListNotificationSinks_NotificationSinks, error)
	GetNotificationSink(id string) (*gqlclient.NotificationSinkFragment, error)
}

// ClientFactory builds a Console API for an authenticated endpoint.
type ClientFactory func(token, url string) (API, error)

// Service implements Loader against Console GraphQL.
type Service struct {
	resolve   ConsoleResolver
	newClient ClientFactory
	pageSize  int64
}

// NewService wires production Console credentials and client construction.
func NewService(resolve ConsoleResolver) *Service {
	return &Service{
		resolve: resolve,
		newClient: func(token, url string) (API, error) {
			return console.NewConsoleClient(token, url)
		},
		pageSize: defaultPageSize,
	}
}

func (s *Service) client(ctx context.Context) (API, error) {
	if s.resolve == nil {
		return nil, &bridge.Error{Code: bridge.ErrorUnauthenticated, Err: errNoConsole}
	}
	url, token, err := s.resolve.ActiveConsole(ctx)
	if err != nil {
		return nil, err
	}
	factory := s.newClient
	if factory == nil {
		factory = func(token, url string) (API, error) {
			return console.NewConsoleClient(token, url)
		}
	}
	return factory(token, url)
}

func (s *Service) List(ctx context.Context, after *string, query string) (Page, error) {
	if err := ctx.Err(); err != nil {
		return Page{}, err
	}
	client, err := s.client(ctx)
	if err != nil {
		return Page{}, err
	}
	first := fetchPageSize
	result, err := client.ListNotificationSinks(nil, &first)
	if err != nil {
		return Page{}, err
	}
	if result == nil {
		return Page{}, nil
	}
	items := make([]Summary, 0, len(result.Edges))
	for _, edge := range result.Edges {
		if edge == nil || edge.Node == nil {
			continue
		}
		summary := summaryFromFragment(edge.Node)
		if !matchesQuery(summary, query) {
			continue
		}
		items = append(items, summary)
	}
	return pageItems(items, after, s.pageSize), nil
}

func (s *Service) Get(ctx context.Context, id string) (Detail, error) {
	if err := ctx.Err(); err != nil {
		return Detail{}, err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return Detail{}, &bridge.Error{Code: bridge.ErrorInvalid, Err: errMissingID}
	}
	client, err := s.client(ctx)
	if err != nil {
		return Detail{}, err
	}
	sink, err := client.GetNotificationSink(id)
	if err != nil {
		return Detail{}, err
	}
	if sink == nil {
		return Detail{}, &bridge.Error{Code: bridge.ErrorUnavailable, Err: errMissingSink}
	}
	return detailFromFragment(sink), nil
}

func pageItems(items []Summary, after *string, pageSize int64) Page {
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	start := 0
	if after != nil && *after != "" {
		for i, item := range items {
			if item.ID == *after {
				start = i + 1
				break
			}
		}
	}
	if start > len(items) {
		start = len(items)
	}
	end := start + int(pageSize)
	if end > len(items) {
		end = len(items)
	}
	page := Page{Items: items[start:end], TotalShown: end - start, HasNext: end < len(items)}
	if len(page.Items) > 0 {
		page.EndCursor = page.Items[len(page.Items)-1].ID
	}
	return page
}

func summaryFromFragment(node *gqlclient.NotificationSinkFragment) Summary {
	return Summary{
		ID:   node.ID,
		Name: node.Name,
		Type: string(node.Type),
		URL:  sinkURL(node),
	}
}

func detailFromFragment(node *gqlclient.NotificationSinkFragment) Detail {
	detail := Detail{Summary: summaryFromFragment(node)}
	for _, binding := range node.NotificationBindings {
		if binding == nil {
			continue
		}
		switch {
		case binding.User != nil:
			name := binding.User.Email
			if name == "" {
				name = binding.User.Name
			}
			detail.Bindings = append(detail.Bindings, Binding{Kind: "user", Name: name})
		case binding.Group != nil:
			detail.Bindings = append(detail.Bindings, Binding{Kind: "group", Name: binding.Group.Name})
		}
	}
	return detail
}

func sinkURL(node *gqlclient.NotificationSinkFragment) string {
	if node.Configuration.Slack != nil {
		return node.Configuration.Slack.URL
	}
	if node.Configuration.Teams != nil {
		return node.Configuration.Teams.URL
	}
	return ""
}

func matchesQuery(summary Summary, query string) bool {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return true
	}
	haystack := strings.ToLower(strings.Join([]string{summary.Name, summary.Type, summary.URL, summary.ID}, " "))
	return strings.Contains(haystack, query)
}
