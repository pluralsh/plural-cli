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
	}

	items, err := service.List(t.Context(), "prod")
	if err != nil || len(items) != 1 || items[0].Handle != "prod-eu" || items[0].Version != "1.30.2" {
		t.Fatalf("List() = %#v, %v", items, err)
	}

	detail, err := service.Get(t.Context(), "c1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !detail.Self || detail.Project != "acme" || detail.NodePools != 2 || len(detail.Tags) != 1 {
		t.Fatalf("detail = %#v", detail)
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
