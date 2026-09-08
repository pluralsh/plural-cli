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
	clusters  *gqlclient.ListClusters
	edges     []*gqlclient.ServiceDeploymentEdgeFragment
	listErr   error
	detail    *gqlclient.ServiceDeploymentExtended
	getErr    error
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
func (f *fakeAPI) KickClusterService(*string, *string, *string) (*gqlclient.ServiceDeploymentExtended, error) {
	return f.detail, f.getErr
}
func (f *fakeAPI) DeleteClusterService(string) (*gqlclient.DeleteServiceDeployment, error) {
	return &gqlclient.DeleteServiceDeployment{}, f.getErr
}
func (f *fakeAPI) CreateClusterService(*string, *string, gqlclient.ServiceDeploymentAttributes) (*gqlclient.ServiceDeploymentExtended, error) {
	return f.detail, f.getErr
}
func (f *fakeAPI) UpdateClusterService(*string, *string, *string, gqlclient.ServiceUpdateAttributes) (*gqlclient.ServiceDeploymentExtended, error) {
	return f.detail, f.getErr
}
func (f *fakeAPI) CloneService(string, *string, *string, *string, gqlclient.ServiceCloneAttributes) (*gqlclient.ServiceDeploymentFragment, error) {
	if f.detail == nil {
		return nil, f.getErr
	}
	return &gqlclient.ServiceDeploymentFragment{ID: f.detail.ID, Name: f.detail.Name, Namespace: f.detail.Namespace}, f.getErr
}
func (f *fakeAPI) GetDeployToken(*string, *string) (string, error) { return "token", f.getErr }

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

func TestListPages(t *testing.T) {
	edges := make([]*gqlclient.ServiceDeploymentEdgeFragment, 0, 12)
	for i := 0; i < 12; i++ {
		id := string(rune('a' + i))
		edges = append(edges, &gqlclient.ServiceDeploymentEdgeFragment{
			Node: &gqlclient.ServiceDeploymentBaseFragment{
				ID: id, Name: "svc-" + id, Namespace: "default", Status: gqlclient.ServiceDeploymentStatusHealthy,
			},
		})
	}
	api := &fakeAPI{edges: edges}
	service := &Service{
		resolve:   fakeResolver{url: "https://console.example.com", token: "token"},
		newClient: func(string, string) (API, error) { return api, nil },
		pageSize:  10,
	}
	first, err := service.List(t.Context(), "c1", nil, "")
	if err != nil || len(first.Items) != 10 || !first.HasNext || first.EndCursor != "j" {
		t.Fatalf("first = %#v, %v", first, err)
	}
	after := first.EndCursor
	second, err := service.List(t.Context(), "c1", &after, "")
	if err != nil || len(second.Items) != 2 || second.HasNext || second.Items[0].ID != "k" {
		t.Fatalf("second = %#v, %v", second, err)
	}
}

func TestGetMapsDetail(t *testing.T) {
	handle := "prod-eu"
	sha := "abc123"
	tarball := "https://console.example.com/tarball/svc-1"
	deleted := "2026-09-08T10:00:00Z"
	ns := "default"
	version := "apps/v1"
	state := gqlclient.ComponentStateRunning
	health := gqlclient.GitHealthPullable
	auth := gqlclient.AuthMethodSSH
	dryRun := false
	templated := true
	kustomize := "overlays/prod"
	api := &fakeAPI{detail: &gqlclient.ServiceDeploymentExtended{
		ID: "svc-1", Name: "api", Namespace: "default", Version: "0.1.4",
		Status:    gqlclient.ServiceDeploymentStatusFailed,
		Tarball:   &tarball,
		DeletedAt: &deleted,
		DryRun:    &dryRun,
		Templated: &templated,
		Git:       &gqlclient.GitRefFragment{Ref: "main", Folder: "services/api"},
		Kustomize: &gqlclient.KustomizeFragment{Path: kustomize},
		Cluster:   &gqlclient.BaseClusterFragment{Name: "prod", Handle: &handle},
		Revision:  &gqlclient.RevisionFragment{ID: "rev-1", Sha: &sha, Git: &gqlclient.RevisionFragment_Git{Ref: "main"}},
		Repository: &gqlclient.GitRepositoryFragment{
			ID: "repo-1", URL: "https://github.com/acme/fleet.git", AuthMethod: &auth, Health: &health,
		},
		Configuration: []*gqlclient.ServiceDeploymentExtended_Configuration{
			{Name: "cluster", Value: "prod"},
			{Name: "replicas", Value: "3"},
		},
		Components: []*gqlclient.ServiceDeploymentExtended_Components{
			{ID: "cmp-1", Name: "api", Namespace: &ns, Kind: "Deployment", Version: &version, State: &state, Synced: true},
			{ID: "cmp-2", Name: "api", Kind: "Service", Synced: false},
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
	if detail.ClusterHandle != "prod-eu" || detail.RevisionSHA != "abc123" || detail.RevisionID != "rev-1" {
		t.Fatalf("revision/cluster = %#v", detail)
	}
	if detail.Version != "0.1.4" || detail.Tarball == "" || !detail.Templated || detail.DryRun || detail.DeletedAt == "" {
		t.Fatalf("identity = %#v", detail)
	}
	if detail.KustomizePath != "overlays/prod" || detail.Repository == nil || detail.Repository.URL == "" {
		t.Fatalf("git = %#v", detail)
	}
	if len(detail.Configuration) != 2 || detail.Configuration[0].Name != "cluster" {
		t.Fatalf("configuration = %#v", detail.Configuration)
	}
	if len(detail.Components) != 2 || detail.Synced != 1 || detail.Components[0].Kind != "Deployment" {
		t.Fatalf("components = %#v", detail.Components)
	}
	if len(detail.Errors) != 1 || detail.Errors[0].Source != "sync" {
		t.Fatalf("errors = %#v", detail.Errors)
	}
}

func TestMissingConsoleReturnsTypedError(t *testing.T) {
	service := NewService(fakeResolver{err: &bridge.Error{Code: bridge.ErrorUnauthenticated, Err: errNoConsole}})
	_, err := service.ListClusters(t.Context(), "")
	if !bridge.IsCode(err, bridge.ErrorUnauthenticated) {
		t.Fatalf("ListClusters() error = %v", err)
	}
}
