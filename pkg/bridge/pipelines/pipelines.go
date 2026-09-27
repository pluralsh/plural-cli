// Package pipelines exposes read-only Console pipeline list/get use cases to
// presentation layers without importing TUI code.
package pipelines

import (
	"context"
	"errors"
	"strings"

	gqlclient "github.com/pluralsh/console/go/client"

	"github.com/pluralsh/plural-cli/pkg/bridge"
	"github.com/pluralsh/plural-cli/pkg/console"
)

const defaultPageSize int64 = 10

var (
	errNoConsole       = errors.New("connect a Console profile before browsing Console resources")
	errMissingID       = errors.New("pipeline id is required")
	errMissingPipeline = errors.New("pipeline was not found")
)

// Summary is a credential-free list row for a Console pipeline.
type Summary struct {
	ID         string
	Name       string
	Project    string
	StageCount int
}

// Stage is a credential-free pipeline stage summary.
type Stage struct {
	Name     string
	Services []string
}

// Edge is a credential-free stage promotion edge.
type Edge struct {
	From string
	To   string
}

// Detail is the credential-free detail payload for a Console pipeline.
type Detail struct {
	Summary
	Stages []Stage
	Edges  []Edge
}

// Page is one cursor page of pipeline summaries.
type Page struct {
	Items      []Summary
	EndCursor  string
	HasNext    bool
	TotalShown int
}

// Loader is the narrow contract consumed by the Pipelines screen.
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
	ListPipelines() (*gqlclient.GetPipelines, error)
	GetPipeline(id string) (*gqlclient.PipelineFragment, error)
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
	result, err := client.ListPipelines()
	if err != nil {
		return Page{}, err
	}
	if result == nil || result.Pipelines == nil {
		return Page{}, nil
	}
	items := make([]Summary, 0, len(result.Pipelines.Edges))
	for _, edge := range result.Pipelines.Edges {
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
	pipeline, err := client.GetPipeline(id)
	if err != nil {
		return Detail{}, err
	}
	if pipeline == nil {
		return Detail{}, &bridge.Error{Code: bridge.ErrorUnavailable, Err: errMissingPipeline}
	}
	return detailFromFragment(pipeline), nil
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

func summaryFromFragment(node *gqlclient.PipelineFragment) Summary {
	summary := Summary{ID: node.ID, Name: node.Name, StageCount: len(node.Stages)}
	if node.Project != nil {
		summary.Project = node.Project.Name
	}
	return summary
}

func detailFromFragment(node *gqlclient.PipelineFragment) Detail {
	detail := Detail{Summary: summaryFromFragment(node)}
	for _, stage := range node.Stages {
		if stage == nil {
			continue
		}
		item := Stage{Name: stage.Name}
		for _, svc := range stage.Services {
			if svc == nil || svc.Service == nil {
				continue
			}
			name := svc.Service.Name
			if svc.Service.Namespace != "" {
				name = svc.Service.Namespace + "/" + name
			}
			item.Services = append(item.Services, name)
		}
		detail.Stages = append(detail.Stages, item)
	}
	for _, edge := range node.Edges {
		if edge == nil {
			continue
		}
		detail.Edges = append(detail.Edges, Edge{From: edge.From.Name, To: edge.To.Name})
	}
	return detail
}

func matchesQuery(summary Summary, query string) bool {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return true
	}
	haystack := strings.ToLower(strings.Join([]string{summary.Name, summary.Project, summary.ID}, " "))
	return strings.Contains(haystack, query)
}
