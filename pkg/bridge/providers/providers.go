// Package providers exposes read-only Console cluster provider list/get
// use cases to presentation layers without importing TUI code.
package providers

import (
	"context"
	"errors"
	"strconv"
	"strings"

	gqlclient "github.com/pluralsh/console/go/client"

	"github.com/pluralsh/plural-cli/pkg/bridge"
	"github.com/pluralsh/plural-cli/pkg/console"
)

const defaultPageSize int64 = 10

var (
	errNoConsole       = errors.New("connect a Console profile before browsing Console resources")
	errMissingID       = errors.New("provider id is required")
	errMissingProvider = errors.New("cluster provider was not found")
)

// Summary is a credential-free list row for a cluster provider.
type Summary struct {
	ID        string
	Name      string
	Cloud     string
	Namespace string
	Editable  string
	RepoURL   string
}

// Credential is a credential-free provider credential summary.
type Credential struct {
	Name      string
	Namespace string
	Kind      string
}

// Detail is the credential-free detail payload for a cluster provider.
type Detail struct {
	Summary
	Service     string
	DeletedAt   string
	Credentials []Credential
}

// Page is one cursor page of provider summaries.
type Page struct {
	Items      []Summary
	EndCursor  string
	HasNext    bool
	TotalShown int
}

// Loader is the narrow contract consumed by the Providers screen.
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
	ListProviders() (*gqlclient.ListProviders, error)
	GetProvider(id string) (*gqlclient.ClusterProviderFragment, error)
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
	result, err := client.ListProviders()
	if err != nil {
		return Page{}, err
	}
	if result == nil || result.ClusterProviders == nil {
		return Page{}, nil
	}
	items := make([]Summary, 0, len(result.ClusterProviders.Edges))
	for _, edge := range result.ClusterProviders.Edges {
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
	provider, err := client.GetProvider(id)
	if err != nil {
		return Detail{}, err
	}
	if provider == nil {
		return Detail{}, &bridge.Error{Code: bridge.ErrorUnavailable, Err: errMissingProvider}
	}
	return detailFromFragment(provider), nil
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

func summaryFromFragment(node *gqlclient.ClusterProviderFragment) Summary {
	summary := Summary{
		ID:        node.ID,
		Name:      node.Name,
		Cloud:     node.Cloud,
		Namespace: node.Namespace,
	}
	if node.Editable != nil {
		summary.Editable = strconv.FormatBool(*node.Editable)
	}
	if node.Repository != nil {
		summary.RepoURL = node.Repository.URL
	}
	return summary
}

func detailFromFragment(node *gqlclient.ClusterProviderFragment) Detail {
	detail := Detail{Summary: summaryFromFragment(node)}
	if node.DeletedAt != nil {
		detail.DeletedAt = *node.DeletedAt
	}
	if node.Service != nil {
		name := node.Service.Name
		if node.Service.Namespace != "" {
			name = node.Service.Namespace + "/" + name
		}
		detail.Service = name
	}
	for _, credential := range node.Credentials {
		if credential == nil {
			continue
		}
		detail.Credentials = append(detail.Credentials, Credential{
			Name:      credential.Name,
			Namespace: credential.Namespace,
			Kind:      credential.Kind,
		})
	}
	return detail
}

func matchesQuery(summary Summary, query string) bool {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return true
	}
	haystack := strings.ToLower(strings.Join([]string{
		summary.Name, summary.Cloud, summary.Namespace, summary.RepoURL, summary.Editable, summary.ID,
	}, " "))
	return strings.Contains(haystack, query)
}
