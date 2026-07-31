// Package stacks exposes read-only Console infrastructure stack list/get
// use cases to presentation layers without importing TUI code.
package stacks

import (
	"context"
	"errors"
	"strconv"
	"strings"

	gqlclient "github.com/pluralsh/console/go/client"
	"github.com/samber/lo"

	"github.com/pluralsh/plural-cli/pkg/bridge"
	"github.com/pluralsh/plural-cli/pkg/console"
)

const defaultPageSize int64 = 10

var (
	errNoConsole    = errors.New("connect a Console profile before browsing Console resources")
	errMissingID    = errors.New("stack id is required")
	errMissingStack = errors.New("infrastructure stack was not found")
)

// Summary is a credential-free list row for an infrastructure stack.
type Summary struct {
	ID       string
	Name     string
	Type     string
	Project  string
	Cluster  string
	Approval string
	RepoURL  string
}

// Detail is the credential-free detail payload for an infrastructure stack.
// Environment and output values are omitted; only names are exposed.
type Detail struct {
	Summary
	Workdir       string
	ManageState   string
	GitRef        string
	GitFolder     string
	ConfigVersion string
	DeletedAt     string
	EnvNames      []string
	OutputNames   []string
}

// Page is one cursor page of stack summaries.
type Page struct {
	Items      []Summary
	EndCursor  string
	HasNext    bool
	TotalShown int
}

// Loader is the narrow contract consumed by the Stacks screen.
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
	ListStacks() (*gqlclient.ListInfrastructureStacks, error)
	GetStack(id string) (*gqlclient.InfrastructureStackFragment, error)
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
	result, err := client.ListStacks()
	if err != nil {
		return Page{}, err
	}
	if result == nil || result.InfrastructureStacks == nil {
		return Page{}, nil
	}
	items := make([]Summary, 0, len(result.InfrastructureStacks.Edges))
	for _, edge := range result.InfrastructureStacks.Edges {
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
	stack, err := client.GetStack(id)
	if err != nil {
		return Detail{}, err
	}
	if stack == nil {
		return Detail{}, &bridge.Error{Code: bridge.ErrorUnavailable, Err: errMissingStack}
	}
	return detailFromFragment(stack), nil
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

func summaryFromFragment(node *gqlclient.InfrastructureStackFragment) Summary {
	summary := Summary{
		ID:   lo.FromPtr(node.ID),
		Name: node.Name,
		Type: string(node.Type),
	}
	if node.Approval != nil {
		summary.Approval = strconv.FormatBool(*node.Approval)
	}
	if node.Project != nil {
		summary.Project = node.Project.Name
	}
	if node.Cluster != nil {
		summary.Cluster = node.Cluster.Name
	}
	if node.Repository != nil {
		summary.RepoURL = node.Repository.URL
	}
	return summary
}

func detailFromFragment(node *gqlclient.InfrastructureStackFragment) Detail {
	detail := Detail{
		Summary:   summaryFromFragment(node),
		GitRef:    node.Git.Ref,
		GitFolder: node.Git.Folder,
	}
	if node.Workdir != nil {
		detail.Workdir = *node.Workdir
	}
	if node.ManageState != nil {
		detail.ManageState = strconv.FormatBool(*node.ManageState)
	}
	if node.Configuration.Version != nil {
		detail.ConfigVersion = *node.Configuration.Version
	}
	if node.DeletedAt != nil {
		detail.DeletedAt = *node.DeletedAt
	}
	for _, env := range node.Environment {
		if env == nil || env.Name == "" {
			continue
		}
		detail.EnvNames = append(detail.EnvNames, env.Name)
	}
	for _, out := range node.Output {
		if out == nil || out.Name == "" {
			continue
		}
		name := out.Name
		if out.Secret != nil && *out.Secret {
			name += " (secret)"
		}
		detail.OutputNames = append(detail.OutputNames, name)
	}
	return detail
}

func matchesQuery(summary Summary, query string) bool {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return true
	}
	haystack := strings.ToLower(strings.Join([]string{
		summary.Name, summary.Type, summary.Project, summary.Cluster, summary.RepoURL, summary.Approval, summary.ID,
	}, " "))
	return strings.Contains(haystack, query)
}
