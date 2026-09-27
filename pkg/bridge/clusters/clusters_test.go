package clusters

import (
	"context"
	"testing"

	gqlclient "github.com/pluralsh/console/go/client"
	"github.com/samber/lo"

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
	listErr  error
	detail   *gqlclient.ClusterFragment
	getErr   error
}

func (f *fakeAPI) ListClusters() (*gqlclient.ListClusters, error) { return f.clusters, f.listErr }
func (f *fakeAPI) GetCluster(*string, *string) (*gqlclient.ClusterFragment, error) {
	return f.detail, f.getErr
}

func TestListAndGet(t *testing.T) {
	handle := "prod-eu"
	version := "1.30.2"
	distro := gqlclient.ClusterDistroEks
	pinged := "2026-07-29T10:00:00Z"
	api := &fakeAPI{
		clusters: &gqlclient.ListClusters{Clusters: &gqlclient.ListClusters_Clusters{Edges: []*gqlclient.ClusterEdgeFragment{
			{Node: &gqlclient.ClusterFragment{
				ID: "c1", Name: "production", Handle: &handle,
				CurrentVersion: &version, Distro: &distro,
			}},
			{Node: &gqlclient.ClusterFragment{ID: "c2", Name: "staging"}},
		}}},
		detail: &gqlclient.ClusterFragment{
			ID: "c1", Name: "production", Handle: &handle,
			CurrentVersion: &version, Distro: &distro,
			Self: lo.ToPtr(true), PingedAt: &pinged, Protect: lo.ToPtr(false),
			Project:   &gqlclient.TinyProjectFragment{Name: "acme"},
			Tags:      []*gqlclient.ClusterTags{{Name: "env", Value: "prod"}},
			NodePools: []*gqlclient.NodePoolFragment{{}, {}},
		},
	}
	service := &Service{
		resolve:   fakeResolver{url: "https://console.example.com", token: "token"},
		newClient: func(string, string) (API, error) { return api, nil },
		pageSize:  50,
	}

	page, err := service.List(t.Context(), nil, "prod")
	if err != nil || len(page.Items) != 1 || page.Items[0].Handle != "prod-eu" || page.Items[0].Version != "1.30.2" {
		t.Fatalf("List() = %#v, %v", page, err)
	}

	detail, err := service.Get(t.Context(), "c1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !detail.Self || detail.Project != "acme" || detail.NodePools != 2 || len(detail.Tags) != 1 {
		t.Fatalf("detail = %#v", detail)
	}
}

func TestListPages(t *testing.T) {
	api := &fakeAPI{
		clusters: &gqlclient.ListClusters{Clusters: &gqlclient.ListClusters_Clusters{Edges: []*gqlclient.ClusterEdgeFragment{
			{Node: &gqlclient.ClusterFragment{ID: "c1", Name: "a"}},
			{Node: &gqlclient.ClusterFragment{ID: "c2", Name: "b"}},
			{Node: &gqlclient.ClusterFragment{ID: "c3", Name: "c"}},
		}}},
	}
	service := &Service{
		resolve:   fakeResolver{url: "https://console.example.com", token: "token"},
		newClient: func(string, string) (API, error) { return api, nil },
		pageSize:  2,
	}
	first, err := service.List(t.Context(), nil, "")
	if err != nil || len(first.Items) != 2 || !first.HasNext || first.EndCursor != "c2" {
		t.Fatalf("first = %#v, %v", first, err)
	}
	after := first.EndCursor
	second, err := service.List(t.Context(), &after, "")
	if err != nil || len(second.Items) != 1 || second.HasNext || second.Items[0].ID != "c3" {
		t.Fatalf("second = %#v, %v", second, err)
	}
}

func TestGetRequiresID(t *testing.T) {
	service := &Service{
		resolve:   fakeResolver{url: "https://console.example.com", token: "token"},
		newClient: func(string, string) (API, error) { return &fakeAPI{}, nil },
	}
	_, err := service.Get(t.Context(), "")
	if !bridge.IsCode(err, bridge.ErrorInvalid) {
		t.Fatalf("Get() error = %v", err)
	}
}
