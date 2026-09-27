package providers

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
	providers *gqlclient.ListProviders
	listErr   error
	detail    *gqlclient.ClusterProviderFragment
	getErr    error
}

func (f *fakeAPI) ListProviders() (*gqlclient.ListProviders, error) { return f.providers, f.listErr }
func (f *fakeAPI) GetProvider(string) (*gqlclient.ClusterProviderFragment, error) {
	return f.detail, f.getErr
}

func TestListAndGet(t *testing.T) {
	api := &fakeAPI{
		providers: &gqlclient.ListProviders{ClusterProviders: &gqlclient.ListProviders_ClusterProviders{
			Edges: []*gqlclient.ListProviders_ClusterProviders_Edges{
				{Node: &gqlclient.ClusterProviderFragment{
					ID: "pr1", Name: "aws-west", Cloud: "aws", Namespace: "infra",
					Editable:   lo.ToPtr(true),
					Repository: &gqlclient.GitRepositoryFragment{URL: "https://github.com/acme/infra"},
				}},
				{Node: &gqlclient.ClusterProviderFragment{ID: "pr2", Name: "gcp-east", Cloud: "gcp"}},
			},
		}},
		detail: &gqlclient.ClusterProviderFragment{
			ID: "pr1", Name: "aws-west", Cloud: "aws", Namespace: "infra",
			Editable:   lo.ToPtr(true),
			Repository: &gqlclient.GitRepositoryFragment{URL: "https://github.com/acme/infra"},
			Service:    &gqlclient.ServiceDeploymentFragment{Name: "provider", Namespace: "infra"},
			Credentials: []*gqlclient.ProviderCredentialFragment{
				{Name: "aws-creds", Namespace: "infra", Kind: "Secret"},
			},
		},
	}
	service := &Service{
		resolve:   fakeResolver{url: "https://console.example.com", token: "token"},
		newClient: func(string, string) (API, error) { return api, nil },
		pageSize:  10,
	}

	page, err := service.List(t.Context(), nil, "aws")
	if err != nil || len(page.Items) != 1 || page.Items[0].Name != "aws-west" || page.Items[0].Cloud != "aws" {
		t.Fatalf("List() = %#v, %v", page, err)
	}
	if page.Items[0].Editable != "true" || page.Items[0].RepoURL != "https://github.com/acme/infra" {
		t.Fatalf("summary = %#v", page.Items[0])
	}

	detail, err := service.Get(t.Context(), "pr1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if detail.Service != "infra/provider" || len(detail.Credentials) != 1 || detail.Credentials[0].Name != "aws-creds" {
		t.Fatalf("detail = %#v", detail)
	}
}

func TestListPages(t *testing.T) {
	edges := make([]*gqlclient.ListProviders_ClusterProviders_Edges, 0, 3)
	for _, id := range []string{"pr1", "pr2", "pr3"} {
		edges = append(edges, &gqlclient.ListProviders_ClusterProviders_Edges{
			Node: &gqlclient.ClusterProviderFragment{ID: id, Name: id, Cloud: "aws"},
		})
	}
	api := &fakeAPI{providers: &gqlclient.ListProviders{ClusterProviders: &gqlclient.ListProviders_ClusterProviders{Edges: edges}}}
	service := &Service{
		resolve:   fakeResolver{url: "https://console.example.com", token: "token"},
		newClient: func(string, string) (API, error) { return api, nil },
		pageSize:  2,
	}
	first, err := service.List(t.Context(), nil, "")
	if err != nil || len(first.Items) != 2 || !first.HasNext || first.EndCursor != "pr2" {
		t.Fatalf("first = %#v, %v", first, err)
	}
	after := first.EndCursor
	second, err := service.List(t.Context(), &after, "")
	if err != nil || len(second.Items) != 1 || second.HasNext || second.Items[0].ID != "pr3" {
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
