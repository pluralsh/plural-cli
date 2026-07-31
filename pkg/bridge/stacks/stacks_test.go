package stacks

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
	stacks  *gqlclient.ListInfrastructureStacks
	listErr error
	detail  *gqlclient.InfrastructureStackFragment
	getErr  error
}

func (f *fakeAPI) ListStacks() (*gqlclient.ListInfrastructureStacks, error) {
	return f.stacks, f.listErr
}
func (f *fakeAPI) GetStack(string) (*gqlclient.InfrastructureStackFragment, error) {
	return f.detail, f.getErr
}

func TestListAndGet(t *testing.T) {
	api := &fakeAPI{
		stacks: &gqlclient.ListInfrastructureStacks{InfrastructureStacks: &gqlclient.ListInfrastructureStacks_InfrastructureStacks{
			Edges: []*gqlclient.InfrastructureStackEdgeFragment{
				{Node: &gqlclient.InfrastructureStackFragment{
					ID: lo.ToPtr("s1"), Name: "gke-demo", Type: gqlclient.StackTypeTerraform,
					Approval:   lo.ToPtr(true),
					Project:    &gqlclient.TinyProjectFragment{Name: "acme"},
					Cluster:    &gqlclient.TinyClusterFragment{Name: "mgmt"},
					Repository: &gqlclient.GitRepositoryFragment{URL: "https://github.com/acme/fleet"},
					Git:        gqlclient.GitRefFragment{Ref: "main", Folder: "terraform"},
				}},
				{Node: &gqlclient.InfrastructureStackFragment{
					ID: lo.ToPtr("s2"), Name: "ansible-edge", Type: gqlclient.StackTypeAnsible,
				}},
			},
		}},
		detail: &gqlclient.InfrastructureStackFragment{
			ID: lo.ToPtr("s1"), Name: "gke-demo", Type: gqlclient.StackTypeTerraform,
			Approval:    lo.ToPtr(true),
			Workdir:     lo.ToPtr("gke-cluster"),
			ManageState: lo.ToPtr(true),
			Project:     &gqlclient.TinyProjectFragment{Name: "acme"},
			Cluster:     &gqlclient.TinyClusterFragment{Name: "mgmt"},
			Repository:  &gqlclient.GitRepositoryFragment{URL: "https://github.com/acme/fleet"},
			Git:         gqlclient.GitRefFragment{Ref: "main", Folder: "terraform"},
			Configuration: gqlclient.StackConfigurationFragment{
				Version: lo.ToPtr("1.8.2"),
			},
			Environment: []*gqlclient.StackEnvironmentFragment{
				{Name: "TF_VAR_cluster", Value: "secret-value", Secret: lo.ToPtr(true)},
			},
			Output: []*gqlclient.StackOutputFragment{
				{Name: "cluster_name", Value: "gke-demo"},
				{Name: "token", Value: "x", Secret: lo.ToPtr(true)},
			},
		},
	}
	service := &Service{
		resolve:   fakeResolver{url: "https://console.example.com", token: "token"},
		newClient: func(string, string) (API, error) { return api, nil },
		pageSize:  10,
	}

	page, err := service.List(t.Context(), nil, "gke")
	if err != nil || len(page.Items) != 1 || page.Items[0].Name != "gke-demo" || page.Items[0].Type != "TERRAFORM" {
		t.Fatalf("List() = %#v, %v", page, err)
	}
	if page.Items[0].Cluster != "mgmt" || page.Items[0].Approval != "true" {
		t.Fatalf("summary = %#v", page.Items[0])
	}

	detail, err := service.Get(t.Context(), "s1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if detail.Workdir != "gke-cluster" || detail.GitRef != "main" || detail.ConfigVersion != "1.8.2" {
		t.Fatalf("detail = %#v", detail)
	}
	if len(detail.EnvNames) != 1 || detail.EnvNames[0] != "TF_VAR_cluster" {
		t.Fatalf("env names leaked values? %#v", detail.EnvNames)
	}
	if len(detail.OutputNames) != 2 || detail.OutputNames[1] != "token (secret)" {
		t.Fatalf("outputs = %#v", detail.OutputNames)
	}
}

func TestListPages(t *testing.T) {
	edges := make([]*gqlclient.InfrastructureStackEdgeFragment, 0, 3)
	for _, id := range []string{"s1", "s2", "s3"} {
		edges = append(edges, &gqlclient.InfrastructureStackEdgeFragment{
			Node: &gqlclient.InfrastructureStackFragment{ID: lo.ToPtr(id), Name: id, Type: gqlclient.StackTypeTerraform},
		})
	}
	api := &fakeAPI{stacks: &gqlclient.ListInfrastructureStacks{InfrastructureStacks: &gqlclient.ListInfrastructureStacks_InfrastructureStacks{Edges: edges}}}
	service := &Service{
		resolve:   fakeResolver{url: "https://console.example.com", token: "token"},
		newClient: func(string, string) (API, error) { return api, nil },
		pageSize:  2,
	}
	first, err := service.List(t.Context(), nil, "")
	if err != nil || len(first.Items) != 2 || !first.HasNext || first.EndCursor != "s2" {
		t.Fatalf("first = %#v, %v", first, err)
	}
	after := first.EndCursor
	second, err := service.List(t.Context(), &after, "")
	if err != nil || len(second.Items) != 1 || second.HasNext || second.Items[0].ID != "s3" {
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
