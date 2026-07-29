// Package services exposes Console service list/get/mutation use cases to
// presentation layers without importing TUI code.
package services

import (
	"context"
	"strings"

	gqlclient "github.com/pluralsh/console/go/client"

	"github.com/pluralsh/plural-cli/pkg/bridge"
	"github.com/pluralsh/plural-cli/pkg/console"
)

const defaultPageSize int64 = 50

// Cluster is a credential-free Console cluster summary.
type Cluster struct {
	ID     string
	Name   string
	Handle string
}

// Summary is the credential-free list row for a service deployment.
type Summary struct {
	ID        string
	Name      string
	Namespace string
	Status    string
	GitRef    string
	GitFolder string
}

// ServiceError is a redacted Console component/sync error.
type ServiceError struct {
	Source  string
	Message string
}

// Detail is the credential-free detail payload for a service deployment.
type Detail struct {
	Summary
	ClusterID     string
	ClusterName   string
	ClusterHandle string
	RevisionSHA   string
	RevisionRef   string
	Components    int
	Synced        int
	Errors        []ServiceError
}

// Page is one cursor page of service summaries for a single cluster.
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

func (s *Service) ListClusters(ctx context.Context, query string) ([]Cluster, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	client, err := s.client(ctx)
	if err != nil {
		return nil, err
	}
	result, err := client.ListClusters()
	if err != nil {
		return nil, err
	}
	if result == nil || result.Clusters == nil {
		return nil, nil
	}
	clusters := make([]Cluster, 0, len(result.Clusters.Edges))
	for _, edge := range result.Clusters.Edges {
		if edge == nil || edge.Node == nil {
			continue
		}
		cluster := Cluster{ID: edge.Node.ID, Name: edge.Node.Name}
		if edge.Node.Handle != nil {
			cluster.Handle = *edge.Node.Handle
		}
		if !matchesCluster(cluster, query) {
			continue
		}
		clusters = append(clusters, cluster)
	}
	return clusters, nil
}

func (s *Service) List(ctx context.Context, clusterID string, after *string, query string) (Page, error) {
	if err := ctx.Err(); err != nil {
		return Page{}, err
	}
	clusterID = strings.TrimSpace(clusterID)
	if clusterID == "" {
		return Page{}, &bridge.Error{Code: bridge.ErrorInvalid, Err: errMissingCluster}
	}
	client, err := s.client(ctx)
	if err != nil {
		return Page{}, err
	}
	edges, err := client.ListClusterServices(&clusterID, nil)
	if err != nil {
		return Page{}, err
	}
	items := make([]Summary, 0, len(edges))
	for _, edge := range edges {
		if edge == nil || edge.Node == nil {
			continue
		}
		summary := summaryFromBase(edge.Node)
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
	service, err := client.GetClusterService(&id, nil, nil)
	if err != nil {
		return Detail{}, err
	}
	if service == nil {
		return Detail{}, &bridge.Error{Code: bridge.ErrorUnavailable, Err: errMissingService}
	}
	return detailFromExtended(service), nil
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

func summaryFromBase(node *gqlclient.ServiceDeploymentBaseFragment) Summary {
	summary := Summary{
		ID:        node.ID,
		Name:      node.Name,
		Namespace: node.Namespace,
		Status:    string(node.Status),
	}
	if node.Git != nil {
		summary.GitRef = node.Git.Ref
		summary.GitFolder = node.Git.Folder
	}
	return summary
}

func detailFromExtended(service *gqlclient.ServiceDeploymentExtended) Detail {
	detail := Detail{
		Summary: Summary{
			ID:        service.ID,
			Name:      service.Name,
			Namespace: service.Namespace,
			Status:    string(service.Status),
		},
		Components: len(service.Components),
	}
	if service.Git != nil {
		detail.GitRef = service.Git.Ref
		detail.GitFolder = service.Git.Folder
	}
	if service.Cluster != nil {
		detail.ClusterID = service.Cluster.ID
		detail.ClusterName = service.Cluster.Name
		if service.Cluster.Handle != nil {
			detail.ClusterHandle = *service.Cluster.Handle
		}
	}
	if service.Revision != nil {
		if service.Revision.Sha != nil {
			detail.RevisionSHA = *service.Revision.Sha
		}
		if detail.RevisionSHA == "" {
			detail.RevisionSHA = service.Revision.ID
		}
		if service.Revision.Git != nil {
			detail.RevisionRef = service.Revision.Git.Ref
		}
	}
	for _, component := range service.Components {
		if component != nil && component.Synced {
			detail.Synced++
		}
	}
	for _, item := range service.Errors {
		if item == nil {
			continue
		}
		detail.Errors = append(detail.Errors, ServiceError{Source: item.Source, Message: item.Message})
	}
	return detail
}

func matchesQuery(summary Summary, query string) bool {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return true
	}
	haystack := strings.ToLower(strings.Join([]string{summary.Name, summary.Namespace, summary.Status, summary.GitRef, summary.GitFolder}, " "))
	return strings.Contains(haystack, query)
}

func matchesCluster(cluster Cluster, query string) bool {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return true
	}
	haystack := strings.ToLower(strings.Join([]string{cluster.Name, cluster.Handle, cluster.ID}, " "))
	return strings.Contains(haystack, query)
}
