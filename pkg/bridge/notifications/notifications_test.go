package notifications

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
	sinks   *gqlclient.ListNotificationSinks_NotificationSinks
	listErr error
	detail  *gqlclient.NotificationSinkFragment
	getErr  error
}

func (f *fakeAPI) ListNotificationSinks(*string, *int64) (*gqlclient.ListNotificationSinks_NotificationSinks, error) {
	return f.sinks, f.listErr
}
func (f *fakeAPI) GetNotificationSink(string) (*gqlclient.NotificationSinkFragment, error) {
	return f.detail, f.getErr
}

func TestListAndGet(t *testing.T) {
	api := &fakeAPI{
		sinks: &gqlclient.ListNotificationSinks_NotificationSinks{
			Edges: []*gqlclient.NotificationSinkEdgeFragment{
				{Node: &gqlclient.NotificationSinkFragment{
					ID: "s1", Name: "ops-slack", Type: gqlclient.SinkTypeSLACk,
					Configuration: gqlclient.SinkConfigurationFragment{
						Slack: &gqlclient.URLSinkConfigurationFragment{URL: "https://hooks.slack.com/x"},
					},
				}},
				{Node: &gqlclient.NotificationSinkFragment{
					ID: "s2", Name: "ops-teams", Type: gqlclient.SinkTypeTeams,
					Configuration: gqlclient.SinkConfigurationFragment{
						Teams: &gqlclient.URLSinkConfigurationFragment{URL: "https://teams.example/hook"},
					},
				}},
			},
		},
		detail: &gqlclient.NotificationSinkFragment{
			ID: "s1", Name: "ops-slack", Type: gqlclient.SinkTypeSLACk,
			Configuration: gqlclient.SinkConfigurationFragment{
				Slack: &gqlclient.URLSinkConfigurationFragment{URL: "https://hooks.slack.com/x"},
			},
			NotificationBindings: []*gqlclient.PolicyBindingFragment{
				{User: &gqlclient.UserFragment{Email: "ops@acme.io", Name: "ops"}},
				{Group: &gqlclient.GroupFragment{Name: "platform"}},
			},
		},
	}
	service := &Service{
		resolve:   fakeResolver{url: "https://console.example.com", token: "token"},
		newClient: func(string, string) (API, error) { return api, nil },
		pageSize:  10,
	}

	page, err := service.List(t.Context(), nil, "slack")
	if err != nil || len(page.Items) != 1 || page.Items[0].Name != "ops-slack" || page.Items[0].Type != "SLACK" {
		t.Fatalf("List() = %#v, %v", page, err)
	}

	detail, err := service.Get(t.Context(), "s1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if detail.URL != "https://hooks.slack.com/x" || len(detail.Bindings) != 2 {
		t.Fatalf("detail = %#v", detail)
	}
	if detail.Bindings[0].Kind != "user" || detail.Bindings[0].Name != "ops@acme.io" {
		t.Fatalf("user binding = %#v", detail.Bindings[0])
	}
	if detail.Bindings[1].Kind != "group" || detail.Bindings[1].Name != "platform" {
		t.Fatalf("group binding = %#v", detail.Bindings[1])
	}
}

func TestListPages(t *testing.T) {
	edges := make([]*gqlclient.NotificationSinkEdgeFragment, 0, 3)
	for _, id := range []string{"s1", "s2", "s3"} {
		edges = append(edges, &gqlclient.NotificationSinkEdgeFragment{
			Node: &gqlclient.NotificationSinkFragment{ID: id, Name: id, Type: gqlclient.SinkTypeSLACk},
		})
	}
	api := &fakeAPI{sinks: &gqlclient.ListNotificationSinks_NotificationSinks{Edges: edges}}
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
