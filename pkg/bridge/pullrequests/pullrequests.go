// Package pullrequests exposes read-only Console PR automation list/get
// use cases to presentation layers without importing TUI code.
package pullrequests

import (
	"context"
	"errors"
	"strings"

	gqlclient "github.com/pluralsh/console/go/client"
	"github.com/samber/lo"

	"github.com/pluralsh/plural-cli/pkg/bridge"
	"github.com/pluralsh/plural-cli/pkg/console"
)

const defaultPageSize int64 = 10

var (
	errNoConsole         = errors.New("connect a Console profile before browsing Console resources")
	errMissingID         = errors.New("pr automation id is required")
	errMissingAutomation = errors.New("pr automation was not found")
)

// Summary is a credential-free list row for a PR automation.
type Summary struct {
	ID         string
	Name       string
	Title      string
	Addon      string
	Identifier string
}

// Detail is the credential-free detail payload for a PR automation.
type Detail struct {
	Summary
	Message    string
	InsertedAt string
	UpdatedAt  string
}

// Page is one cursor page of PR automation summaries.
type Page struct {
	Items      []Summary
	EndCursor  string
	HasNext    bool
	TotalShown int
}

// ConsoleResolver supplies the active Console URL and token.
type ConsoleResolver interface {
	ActiveConsole(ctx context.Context) (url, token string, err error)
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
	result, err := client.ListPrAutomations()
	if err != nil {
		return Page{}, err
	}
	if result == nil || result.PrAutomations == nil {
		return Page{}, nil
	}
	items := make([]Summary, 0, len(result.PrAutomations.Edges))
	for _, edge := range result.PrAutomations.Edges {
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
	automation, err := client.GetPrAutomation(id)
	if err != nil {
		return Detail{}, err
	}
	if automation == nil {
		return Detail{}, &bridge.Error{Code: bridge.ErrorUnavailable, Err: errMissingAutomation}
	}
	return detailFromFragment(automation), nil
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

func summaryFromFragment(node *gqlclient.PrAutomationFragment) Summary {
	return Summary{
		ID:         node.ID,
		Name:       node.Name,
		Title:      lo.FromPtr(node.Title),
		Addon:      lo.FromPtr(node.Addon),
		Identifier: lo.FromPtr(node.Identifier),
	}
}

func detailFromFragment(node *gqlclient.PrAutomationFragment) Detail {
	return Detail{
		Summary:    summaryFromFragment(node),
		Message:    lo.FromPtr(node.Message),
		InsertedAt: lo.FromPtr(node.InsertedAt),
		UpdatedAt:  lo.FromPtr(node.UpdatedAt),
	}
}

func matchesQuery(summary Summary, query string) bool {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return true
	}
	haystack := strings.ToLower(strings.Join([]string{
		summary.Name, summary.Title, summary.Addon, summary.Identifier, summary.ID,
	}, " "))
	return strings.Contains(haystack, query)
}
