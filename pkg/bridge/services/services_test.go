package services

import (
	"context"
	"testing"

	gqlclient "github.com/pluralsh/console/go/client"

	"github.com/pluralsh/plural-cli/pkg/bridge"
)

type fakeResolver struct {
	url, token string
	err        error
}

func (f fakeResolver) ActiveConsole(context.Context) (string, string, error) {
	return f.url, f.token, f.err
}

type fakeAPI struct {
	clusters *gqlclient.ListClusters
	edges    []*gqlclient.ServiceDeploymentEdgeFragment
	listErr  error
	detail   *gqlclient.ServiceDeploymentExtended
	getErr   error
	clusterID string
}

func (f *fakeAPI) ListClusters() (*gqlclient.ListClusters, error) {
	return f.clusters, f.listErr
}
func (f *fakeAPI) ListClusterServices(clusterId, _ *string) ([]*gqlclient.ServiceDeploymentEdgeFragment, error) {
	if clusterId != nil {
		f.clusterID = *clusterId
	}
	return f.edges, f.listErr
}
func (f *fakeAPI) GetClusterService(*string, *string, *string) (*gqlclient.ServiceDeploymentExtended, error) {
	return f.detail, f.getErr
}

func TestListClustersAndScopedServices(t *testing.T) {
	handle := "prod-eu"
	api := &fakeAPI{
		clusters: &gqlclient.ListClusters{Clusters: &gqlclient.ListClusters_Clusters{Edges: []*gqlclient.ClusterEdgeFragment{
			{Node: &gqlclient.ClusterFragment{ID: "c1", Name: "production", Handle: &handle}},
			{Node: &gqlclient.ClusterFragment{ID: "c2", Name: "staging"}},
		}}},
		edges: []*gqlclient.ServiceDeploymentEdgeFragment{
			{Node: &gqlclient.ServiceDeploymentBaseFragment{ID: "1", Name: "api", Namespace: "default", Status: gqlclient.ServiceDeploymentStatusHealthy}},
			{Node: &gqlclient.ServiceDeploymentBaseFragment{ID: "2", Name: "worker", Namespace: "jobs", Status: gqlclient.ServiceDeploymentStatusFailed}},
		},
	}
	service := &Service{
		resolve:   fakeResolver{url: "https://console.example.com", token: "token"},
		newClient: func(string, string) (API, error) { return api, nil },
	}

	clusters, err := service.ListClusters(t.Context(), "prod")
	if err != nil || len(clusters) != 1 || clusters[0].Handle != "prod-eu" {
		t.Fatalf("ListClusters() = %#v, %v", clusters, err)
	}

	page, err := service.List(t.Context(), "c1", nil, "api")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if api.clusterID != "c1" || len(page.Items) != 1 || page.Items[0].Name != "api" {
		t.Fatalf("scoped page = %#v cluster=%q", page.Items, api.clusterID)
	}
}

func TestListRequiresCluster(t *testing.T) {
	service := &Service{
		resolve:   fakeResolver{url: "https://console.example.com", token: "token"},
		newClient: func(string, string) (API, error) { return &fakeAPI{}, nil },
	}
	_, err := service.List(t.Context(), "", nil, "")
	if !bridge.IsCode(err, bridge.ErrorInvalid) {
		t.Fatalf("List() error = %v", err)
	}
}

func TestGetMapsDetail(t *testing.T) {
	handle := "prod-eu"
	sha := "abc123"
	api := &fakeAPI{detail: &gqlclient.ServiceDeploymentExtended{
		ID: "svc-1", Name: "api", Namespace: "default", Status: gqlclient.ServiceDeploymentStatusFailed,
		Git:     &gqlclient.GitRefFragment{Ref: "main", Folder: "services/api"},
		Cluster: &gqlclient.BaseClusterFragment{Name: "prod", Handle: &handle},
		Revision: &gqlclient.RevisionFragment{ID: "rev-1", Sha: &sha, Git: &gqlclient.RevisionFragment_Git{Ref: "main"}},
		Components: []*gqlclient.ServiceDeploymentExtended_Components{
			{Synced: true},
			{Synced: false},
		},
		Errors: []*gqlclient.ErrorFragment{{Source: "sync", Message: "rollout timed out"}},
	}}
	service := &Service{
		resolve:   fakeResolver{url: "https://console.example.com", token: "token"},
		newClient: func(string, string) (API, error) { return api, nil },
	}
	detail, err := service.Get(t.Context(), "svc-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if detail.ClusterHandle != "prod-eu" || detail.RevisionSHA != "abc123" || detail.Components != 2 || detail.Synced != 1 {
		t.Fatalf("detail = %#v", detail)
	}
}

func TestMissingConsoleReturnsTypedError(t *testing.T) {
	service := NewService(fakeResolver{err: &bridge.Error{Code: bridge.ErrorUnauthenticated, Err: errNoConsole}})
	_, err := service.ListClusters(t.Context(), "")
	if !bridge.IsCode(err, bridge.ErrorUnauthenticated) {
		t.Fatalf("ListClusters() error = %v", err)
	}
}
