package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"

	"github.com/pluralsh/plural-cli/pkg/bridge"
	servicesbridge "github.com/pluralsh/plural-cli/pkg/bridge/services"
	"github.com/pluralsh/plural-cli/tui/navigation"
	"github.com/pluralsh/plural-cli/tui/theme"
)

type fakeLoader struct {
	clusters []servicesbridge.Cluster
	page     servicesbridge.Page
	detail   servicesbridge.Detail
	err      error
	listedID string
}

func (f *fakeLoader) ListClusters(context.Context, string) ([]servicesbridge.Cluster, error) {
	return f.clusters, f.err
}
func (f *fakeLoader) List(_ context.Context, clusterID string, _ *string, _ string) (servicesbridge.Page, error) {
	f.listedID = clusterID
	return f.page, f.err
}
func (f *fakeLoader) Get(context.Context, string) (servicesbridge.Detail, error) {
	return f.detail, f.err
}

func loadClusters(t *testing.T, model Model) Model {
	t.Helper()
	cmd := model.Init()
	model, cmd = model.Update(cmd())
	if cmd == nil {
		t.Fatal("expected clusters command")
	}
	model, _ = model.Update(cmd())
	return model
}

func TestSelectClusterThenOpenService(t *testing.T) {
	loader := &fakeLoader{
		clusters: []servicesbridge.Cluster{
			{ID: "c1", Name: "production", Handle: "prod-eu"},
			{ID: "c2", Name: "staging", Handle: "staging"},
		},
		page: servicesbridge.Page{Items: []servicesbridge.Summary{
			{ID: "1", Name: "api", Namespace: "default", Status: "HEALTHY"},
		}},
		detail: servicesbridge.Detail{
			Summary:       servicesbridge.Summary{ID: "1", Name: "api", Namespace: "default", Status: "HEALTHY"},
			ClusterHandle: "prod-eu",
		},
	}
	model := loadClusters(t, New(t.Context(), loader, theme.New(colorprofile.ASCII)))
	if model.mode != modeClusters || len(model.clusters) != 2 {
		t.Fatalf("clusters state = mode=%d count=%d", model.mode, len(model.clusters))
	}
	view := model.View(80, 24)
	if !strings.Contains(view, "@prod-eu") || !strings.Contains(view, "Choose a cluster") {
		t.Fatalf("cluster view missing handle:\n%s", view)
	}

	model, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter did not load services")
	}
	model, _ = model.Update(cmd())
	if model.mode != modeList || loader.listedID != "c1" || len(model.page.Items) != 1 {
		t.Fatalf("list state = mode=%d id=%q items=%d", model.mode, loader.listedID, len(model.page.Items))
	}
	if !strings.Contains(model.View(80, 24), "@prod-eu") {
		t.Fatalf("service list missing cluster label:\n%s", model.View(80, 24))
	}

	model, cmd = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(cmd())
	if model.mode != modeDetail || model.detail.ClusterHandle != "prod-eu" {
		t.Fatalf("detail state = %#v", model.detail)
	}

	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if model.mode != modeList {
		t.Fatalf("mode after detail esc = %d", model.mode)
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if model.mode != modeClusters {
		t.Fatalf("mode after list esc = %d", model.mode)
	}
}

func TestNoConsoleNavigatesToAccess(t *testing.T) {
	loader := &fakeLoader{err: &bridge.Error{Code: bridge.ErrorUnauthenticated, Err: errors.New("connect")}}
	model := loadClusters(t, New(t.Context(), loader, theme.New(colorprofile.ASCII)))
	if !model.needsAuth {
		t.Fatal("expected needsAuth")
	}
	_, cmd := model.Update(tea.KeyPressMsg{Code: 'c'})
	if cmd == nil {
		t.Fatal("expected access navigation")
	}
	if msg := cmd(); msg != (navigation.NavigateMsg{Route: navigation.Access}) {
		t.Fatalf("msg = %#v", msg)
	}
}

func TestBackFromClustersReturnsDeployments(t *testing.T) {
	model := loadClusters(t, New(t.Context(), &fakeLoader{}, theme.New(colorprofile.ASCII)))
	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected deployments navigation")
	}
	if msg := cmd(); msg != (navigation.NavigateMsg{Route: navigation.Deployments}) {
		t.Fatalf("msg = %#v", msg)
	}
}
