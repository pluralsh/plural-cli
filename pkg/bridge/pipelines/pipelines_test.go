package pipelines

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
	pipelines *gqlclient.GetPipelines
	listErr   error
	detail    *gqlclient.PipelineFragment
	getErr    error
}

func (f *fakeAPI) ListPipelines() (*gqlclient.GetPipelines, error) { return f.pipelines, f.listErr }
func (f *fakeAPI) GetPipeline(string) (*gqlclient.PipelineFragment, error) {
	return f.detail, f.getErr
}

func TestListAndGet(t *testing.T) {
	api := &fakeAPI{
		pipelines: &gqlclient.GetPipelines{Pipelines: &gqlclient.GetPipelines_Pipelines{
			Edges: []*gqlclient.PipelineEdgeFragment{
				{Node: &gqlclient.PipelineFragment{
					ID: "p1", Name: "deploy-prod",
					Project: &gqlclient.TinyProjectFragment{Name: "acme"},
					Stages:  []*gqlclient.PipelineStageFragment{{Name: "dev"}, {Name: "prod"}},
				}},
				{Node: &gqlclient.PipelineFragment{ID: "p2", Name: "canary"}},
			},
		}},
		detail: &gqlclient.PipelineFragment{
			ID: "p1", Name: "deploy-prod",
			Project: &gqlclient.TinyProjectFragment{Name: "acme"},
			Stages: []*gqlclient.PipelineStageFragment{
				{Name: "dev", Services: []*gqlclient.PipelineStageFragment_Services{
					{Service: &gqlclient.ServiceDeploymentBaseFragment{Name: "api", Namespace: "default"}},
				}},
				{Name: "prod"},
			},
			Edges: []*gqlclient.PipelineStageEdgeFragment{
				{From: gqlclient.PipelineStageFragment{Name: "dev"}, To: gqlclient.PipelineStageFragment{Name: "prod"}},
			},
		},
	}
	service := &Service{
		resolve:   fakeResolver{url: "https://console.example.com", token: "token"},
		newClient: func(string, string) (API, error) { return api, nil },
		pageSize:  10,
	}

	page, err := service.List(t.Context(), nil, "deploy")
	if err != nil || len(page.Items) != 1 || page.Items[0].Name != "deploy-prod" || page.Items[0].StageCount != 2 {
		t.Fatalf("List() = %#v, %v", page, err)
	}

	detail, err := service.Get(t.Context(), "p1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if detail.Project != "acme" || len(detail.Stages) != 2 || len(detail.Edges) != 1 || detail.Edges[0].From != "dev" {
		t.Fatalf("detail = %#v", detail)
	}
	if len(detail.Stages[0].Services) != 1 || detail.Stages[0].Services[0] != "default/api" {
		t.Fatalf("stage services = %#v", detail.Stages[0].Services)
	}
}

func TestListPages(t *testing.T) {
	edges := make([]*gqlclient.PipelineEdgeFragment, 0, 3)
	for _, id := range []string{"p1", "p2", "p3"} {
		edges = append(edges, &gqlclient.PipelineEdgeFragment{
			Node: &gqlclient.PipelineFragment{ID: id, Name: id},
		})
	}
	api := &fakeAPI{pipelines: &gqlclient.GetPipelines{Pipelines: &gqlclient.GetPipelines_Pipelines{Edges: edges}}}
	service := &Service{
		resolve:   fakeResolver{url: "https://console.example.com", token: "token"},
		newClient: func(string, string) (API, error) { return api, nil },
		pageSize:  2,
	}
	first, err := service.List(t.Context(), nil, "")
	if err != nil || len(first.Items) != 2 || !first.HasNext || first.EndCursor != "p2" {
		t.Fatalf("first = %#v, %v", first, err)
	}
	after := first.EndCursor
	second, err := service.List(t.Context(), &after, "")
	if err != nil || len(second.Items) != 1 || second.HasNext || second.Items[0].ID != "p3" {
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
