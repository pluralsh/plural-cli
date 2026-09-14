package pullrequests

import (
	"context"
	"strings"
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
	automations *gqlclient.ListPrAutomations
	listErr     error
	detail      *gqlclient.PrAutomationFragment
	getErr      error
	created     *gqlclient.PullRequestFragment
	createErr   error
}

func (f *fakeAPI) ListPrAutomations() (*gqlclient.ListPrAutomations, error) {
	return f.automations, f.listErr
}
func (f *fakeAPI) GetPrAutomation(string) (*gqlclient.PrAutomationFragment, error) {
	return f.detail, f.getErr
}
func (f *fakeAPI) CreatePullRequest(string, *string, *string) (*gqlclient.PullRequestFragment, error) {
	return f.created, f.createErr
}

func TestListAndGet(t *testing.T) {
	api := &fakeAPI{
		automations: &gqlclient.ListPrAutomations{PrAutomations: &gqlclient.ListPrAutomations_PrAutomations{
			Edges: []*gqlclient.ListPrAutomations_PrAutomations_Edges{
				{Node: &gqlclient.PrAutomationFragment{
					ID: "pra1", Name: "cluster-create", Title: lo.ToPtr("Create cluster"),
					Addon: lo.ToPtr("cluster"), Identifier: lo.ToPtr("ops/cluster-create"),
				}},
				{Node: &gqlclient.PrAutomationFragment{ID: "pra2", Name: "service-bump"}},
			},
		}},
		detail: &gqlclient.PrAutomationFragment{
			ID: "pra1", Name: "cluster-create", Title: lo.ToPtr("Create cluster"),
			Addon: lo.ToPtr("cluster"), Identifier: lo.ToPtr("ops/cluster-create"),
			Message: lo.ToPtr("Opens a PR to provision a new cluster"), InsertedAt: lo.ToPtr("2026-01-01T00:00:00Z"),
		},
	}
	service := &Service{
		resolve:   fakeResolver{url: "https://console.example.com", token: "token"},
		newClient: func(string, string) (API, error) { return api, nil },
		pageSize:  10,
	}

	page, err := service.List(t.Context(), nil, "cluster")
	if err != nil || len(page.Items) != 1 || page.Items[0].Name != "cluster-create" {
		t.Fatalf("List() = %#v, %v", page, err)
	}
	if page.Items[0].Title != "Create cluster" || page.Items[0].Identifier != "ops/cluster-create" {
		t.Fatalf("summary = %#v", page.Items[0])
	}

	detail, err := service.Get(t.Context(), "pra1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if detail.Message == "" || !strings.Contains(detail.Message, "provision") {
		t.Fatalf("detail = %#v", detail)
	}
}

func TestCreateAndTrigger(t *testing.T) {
	api := &fakeAPI{
		created: &gqlclient.PullRequestFragment{
			ID: "pr1", URL: "https://github.com/acme/fleet/pull/1",
			Title: lo.ToPtr("Create cluster"), Status: lo.ToPtr(gqlclient.PrStatusOpen),
		},
	}
	service := &Service{
		resolve:   fakeResolver{url: "https://console.example.com", token: "token"},
		newClient: func(string, string) (API, error) { return api, nil },
	}

	created, err := service.CreatePR(t.Context(), CreatePRInput{AutomationID: "pra1", Branch: "feat/cluster"})
	if err != nil || created.ID != "pr1" || created.URL == "" {
		t.Fatalf("CreatePR() = %#v, %v", created, err)
	}

	triggered, err := service.TriggerPR(t.Context(), TriggerPRInput{
		AutomationID: "pra1", Name: "cluster-create",
		Configuration: map[string]string{"cluster": "demo"},
	})
	if err != nil || triggered.ID != "pr1" {
		t.Fatalf("TriggerPR() = %#v, %v", triggered, err)
	}
}

func TestListPages(t *testing.T) {
	edges := make([]*gqlclient.ListPrAutomations_PrAutomations_Edges, 0, 3)
	for _, id := range []string{"pra1", "pra2", "pra3"} {
		edges = append(edges, &gqlclient.ListPrAutomations_PrAutomations_Edges{
			Node: &gqlclient.PrAutomationFragment{ID: id, Name: id},
		})
	}
	api := &fakeAPI{automations: &gqlclient.ListPrAutomations{PrAutomations: &gqlclient.ListPrAutomations_PrAutomations{Edges: edges}}}
	service := &Service{
		resolve:   fakeResolver{url: "https://console.example.com", token: "token"},
		newClient: func(string, string) (API, error) { return api, nil },
		pageSize:  2,
	}
	first, err := service.List(t.Context(), nil, "")
	if err != nil || len(first.Items) != 2 || !first.HasNext || first.EndCursor != "pra2" {
		t.Fatalf("first = %#v, %v", first, err)
	}
	after := first.EndCursor
	second, err := service.List(t.Context(), &after, "")
	if err != nil || len(second.Items) != 1 || second.HasNext || second.Items[0].ID != "pra3" {
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
