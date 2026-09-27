package repositories

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
	repos   *gqlclient.ListGitRepositories
	listErr error
	detail  *gqlclient.GetGitRepository
	getErr  error
}

func (f *fakeAPI) ListRepositories() (*gqlclient.ListGitRepositories, error) {
	return f.repos, f.listErr
}
func (f *fakeAPI) GetRepository(string) (*gqlclient.GetGitRepository, error) {
	return f.detail, f.getErr
}

func TestListAndGet(t *testing.T) {
	health := gqlclient.GitHealthPullable
	auth := gqlclient.AuthMethodSSH
	errMsg := "auth failed"
	failed := gqlclient.GitHealthFailed
	api := &fakeAPI{
		repos: &gqlclient.ListGitRepositories{GitRepositories: &gqlclient.ListGitRepositories_GitRepositories{
			Edges: []*gqlclient.GitRepositoryEdgeFragment{
				{Node: &gqlclient.GitRepositoryFragment{
					ID: "r1", URL: "git@github.com:acme/infra.git", Health: &health, AuthMethod: &auth,
				}},
				{Node: &gqlclient.GitRepositoryFragment{
					ID: "r2", URL: "https://github.com/acme/apps.git", Health: &failed, Error: &errMsg,
				}},
			},
		}},
		detail: &gqlclient.GetGitRepository{GitRepository: &gqlclient.GitRepositoryFragment{
			ID: "r1", URL: "git@github.com:acme/infra.git", Health: &health, AuthMethod: &auth,
			Decrypt: lo.ToPtr(true),
		}},
	}
	service := &Service{
		resolve:   fakeResolver{url: "https://console.example.com", token: "token"},
		newClient: func(string, string) (API, error) { return api, nil },
		pageSize:  50,
	}

	page, err := service.List(t.Context(), nil, "infra")
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != "r1" || page.Items[0].Health != "PULLABLE" || page.Items[0].AuthMethod != "SSH" {
		t.Fatalf("List() = %#v, %v", page, err)
	}

	detail, err := service.Get(t.Context(), "r1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !detail.Decrypt || detail.URL != "git@github.com:acme/infra.git" || detail.Health != "PULLABLE" {
		t.Fatalf("detail = %#v", detail)
	}
}

func TestListPages(t *testing.T) {
	health := gqlclient.GitHealthPullable
	edges := make([]*gqlclient.GitRepositoryEdgeFragment, 0, 3)
	for _, id := range []string{"r1", "r2", "r3"} {
		edges = append(edges, &gqlclient.GitRepositoryEdgeFragment{
			Node: &gqlclient.GitRepositoryFragment{ID: id, URL: "git@github.com:acme/" + id + ".git", Health: &health},
		})
	}
	api := &fakeAPI{
		repos: &gqlclient.ListGitRepositories{GitRepositories: &gqlclient.ListGitRepositories_GitRepositories{Edges: edges}},
	}
	service := &Service{
		resolve:   fakeResolver{url: "https://console.example.com", token: "token"},
		newClient: func(string, string) (API, error) { return api, nil },
		pageSize:  2,
	}

	first, err := service.List(t.Context(), nil, "")
	if err != nil || len(first.Items) != 2 || !first.HasNext || first.EndCursor != "r2" {
		t.Fatalf("first page = %#v, %v", first, err)
	}
	after := first.EndCursor
	second, err := service.List(t.Context(), &after, "")
	if err != nil || len(second.Items) != 1 || second.HasNext || second.Items[0].ID != "r3" {
		t.Fatalf("second page = %#v, %v", second, err)
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
