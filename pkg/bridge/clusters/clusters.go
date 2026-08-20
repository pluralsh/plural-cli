// Package clusters exposes read-only Console cluster list/get use cases to
// presentation layers without importing TUI code.
package clusters

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
	errNoConsole      = errors.New("connect a Console profile before browsing Console resources")
	errMissingID      = errors.New("cluster id is required")
	errMissingCluster = errors.New("cluster was not found")
)

// Summary is a credential-free list row for a Console cluster.
type Summary struct {
	ID      string
	Name    string
	Handle  string
	Version string
	Distro  string
}

// Tag is a credential-free cluster tag.
type Tag struct {
	Name  string
	Value string
}

// Detail is the credential-free detail payload for a Console cluster.
type Detail struct {
	Summary
	Self      bool
	PingedAt  string
	Protect   bool
	DeletedAt string
	Project   string
	Provider  string
	Tags      []Tag
	NodePools int
}

// Page is one cursor page of cluster summaries.
type Page struct {
	Items      []Summary
	EndCursor  string
	HasNext    bool
	TotalShown int
}

// Loader is the narrow contract consumed by the Clusters screen.
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
	ListClusters() (*gqlclient.ListClusters, error)
	GetCluster(clusterId, clusterName *string) (*gqlclient.ClusterFragment, error)
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
	result, err := client.ListClusters()
	if err != nil {
		return Page{}, err
	}
	if result == nil || result.Clusters == nil {
		return Page{}, nil
	}
	items := make([]Summary, 0, len(result.Clusters.Edges))
	for _, edge := range result.Clusters.Edges {
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
	cluster, err := client.GetCluster(&id, nil)
	if err != nil {
		return Detail{}, err
	}
	if cluster == nil {
		return Detail{}, &bridge.Error{Code: bridge.ErrorUnavailable, Err: errMissingCluster}
	}
	return detailFromFragment(cluster), nil
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

func summaryFromFragment(node *gqlclient.ClusterFragment) Summary {
	summary := Summary{ID: node.ID, Name: node.Name}
	if node.Handle != nil {
		summary.Handle = *node.Handle
	}
	if node.CurrentVersion != nil {
		summary.Version = *node.CurrentVersion
	}
	if node.Distro != nil {
		summary.Distro = string(*node.Distro)
	}
	return summary
}

func detailFromFragment(cluster *gqlclient.ClusterFragment) Detail {
	detail := Detail{Summary: summaryFromFragment(cluster)}
	if cluster.Self != nil {
		detail.Self = *cluster.Self
	}
	if cluster.PingedAt != nil {
		detail.PingedAt = *cluster.PingedAt
	}
	if cluster.Protect != nil {
		detail.Protect = *cluster.Protect
	}
	if cluster.DeletedAt != nil {
		detail.DeletedAt = *cluster.DeletedAt
	}
	if cluster.Project != nil {
		detail.Project = cluster.Project.Name
	}
	if cluster.Provider != nil {
		detail.Provider = cluster.Provider.Name
		if cluster.Provider.Cloud != "" {
			detail.Provider = strings.TrimSpace(detail.Provider + " · " + cluster.Provider.Cloud)
		}
	}
	for _, tag := range cluster.Tags {
		if tag == nil {
			continue
		}
		detail.Tags = append(detail.Tags, Tag{Name: tag.Name, Value: tag.Value})
	}
	detail.NodePools = len(cluster.NodePools)
	return detail
}

func matchesQuery(summary Summary, query string) bool {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return true
	}
	haystack := strings.ToLower(strings.Join([]string{summary.Name, summary.Handle, summary.ID, summary.Version, summary.Distro}, " "))
	return strings.Contains(haystack, query)
}
