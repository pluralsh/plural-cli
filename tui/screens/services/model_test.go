package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"

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
	kicked   string
	deleted  string
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
func (f *fakeLoader) Kick(_ context.Context, id string) (servicesbridge.Detail, error) {
	f.kicked = id
	return f.detail, f.err
}
func (f *fakeLoader) Delete(_ context.Context, id string) error {
	f.deleted = id
	return f.err
}
func (f *fakeLoader) Create(context.Context, servicesbridge.CreateInput) (servicesbridge.Detail, error) {
	return f.detail, f.err
}
func (f *fakeLoader) Update(context.Context, servicesbridge.UpdateInput) (servicesbridge.Detail, error) {
	return f.detail, f.err
}
func (f *fakeLoader) Clone(context.Context, servicesbridge.CloneInput) (servicesbridge.Detail, error) {
	return f.detail, f.err
}
func (f *fakeLoader) DownloadTarball(context.Context, string, string) (string, error) {
	return "/tmp/tarball", f.err
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

func TestNextPrevPage(t *testing.T) {
	loader := &fakeLoader{
		clusters: []servicesbridge.Cluster{{ID: "c1", Handle: "prod-eu"}},
		page: servicesbridge.Page{
			Items:     []servicesbridge.Summary{{ID: "1", Name: "api", Namespace: "default", Status: "HEALTHY"}},
			EndCursor: "1",
			HasNext:   true,
		},
	}
	model := loadClusters(t, New(t.Context(), loader, theme.New(colorprofile.ASCII)))
	model, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(cmd())
	if model.mode != modeList {
		t.Fatalf("mode = %d", model.mode)
	}
	if !strings.Contains(model.View(80, 24), "n next") {
		t.Fatalf("missing next pager:\n%s", model.View(80, 24))
	}

	model, cmd = model.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
	if cmd == nil {
		t.Fatal("expected next-page list command")
	}
	loader.page = servicesbridge.Page{
		Items: []servicesbridge.Summary{{ID: "2", Name: "worker", Namespace: "jobs", Status: "FAILED"}},
	}
	model, _ = model.Update(cmd())
	if model.after == nil || *model.after != "1" || len(model.prevCursors) != 1 {
		t.Fatalf("after page turn after=%v prev=%v", model.after, model.prevCursors)
	}
	if !strings.Contains(model.View(80, 24), "worker") || !strings.Contains(model.View(80, 24), "p prev") {
		t.Fatalf("second page view:\n%s", model.View(80, 24))
	}

	model, cmd = model.Update(tea.KeyPressMsg{Code: 'p', Text: "p"})
	if cmd == nil {
		t.Fatal("expected prev-page list command")
	}
	loader.page = servicesbridge.Page{
		Items:     []servicesbridge.Summary{{ID: "1", Name: "api", Namespace: "default", Status: "HEALTHY"}},
		EndCursor: "1",
		HasNext:   true,
	}
	model, _ = model.Update(cmd())
	if model.after != nil || len(model.prevCursors) != 0 {
		t.Fatalf("after prev after=%v prev=%v", model.after, model.prevCursors)
	}
}

func TestKickFromDetail(t *testing.T) {
	loader := &fakeLoader{
		clusters: []servicesbridge.Cluster{{ID: "c1", Name: "production", Handle: "prod-eu"}},
		page:     servicesbridge.Page{Items: []servicesbridge.Summary{{ID: "1", Name: "api", Namespace: "default", Status: "HEALTHY"}}},
		detail:   servicesbridge.Detail{Summary: servicesbridge.Summary{ID: "1", Name: "api", Namespace: "default", Status: "HEALTHY"}, ClusterHandle: "prod-eu"},
	}
	model := loadClusters(t, New(t.Context(), loader, theme.New(colorprofile.ASCII)))
	model, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(cmd())
	model, cmd = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(cmd())
	if model.mode != modeDetail {
		t.Fatalf("mode = %d", model.mode)
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: 'k', Text: "k"})
	if model.mode != modeReview || model.pending.kind != actionKick {
		t.Fatalf("review = mode=%d kind=%d", model.mode, model.pending.kind)
	}
	model, cmd = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected kick command")
	}
	model, _ = model.Update(cmd())
	if model.mode != modeResult || loader.kicked != "1" || model.result != "ok" {
		t.Fatalf("result mode=%d kicked=%q result=%q", model.mode, loader.kicked, model.result)
	}
}

func TestDeleteRequiresTypedName(t *testing.T) {
	loader := &fakeLoader{
		clusters: []servicesbridge.Cluster{{ID: "c1", Handle: "prod-eu"}},
		page:     servicesbridge.Page{Items: []servicesbridge.Summary{{ID: "1", Name: "api"}}},
		detail:   servicesbridge.Detail{Summary: servicesbridge.Summary{ID: "1", Name: "api"}},
	}
	model := loadClusters(t, New(t.Context(), loader, theme.New(colorprofile.ASCII)))
	model, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(cmd())
	model, cmd = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(cmd())
	model, _ = model.Update(tea.KeyPressMsg{Code: 'd', Text: "d"})
	if model.mode != modeDeleteConfirm {
		t.Fatalf("mode = %d", model.mode)
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if model.err == nil {
		t.Fatal("expected name mismatch error")
	}
	model.formInput.SetValue("api")
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if model.mode != modeReview || model.pending.kind != actionDelete {
		t.Fatalf("expected delete review, mode=%d kind=%d", model.mode, model.pending.kind)
	}
}

func TestClonePicksDestinationCluster(t *testing.T) {
	loader := &fakeLoader{
		clusters: []servicesbridge.Cluster{
			{ID: "c1", Name: "production", Handle: "prod-eu"},
			{ID: "c2", Name: "staging", Handle: "staging"},
		},
		page: servicesbridge.Page{Items: []servicesbridge.Summary{{ID: "1", Name: "api", Namespace: "default"}}},
		detail: servicesbridge.Detail{
			Summary:   servicesbridge.Summary{ID: "1", Name: "api", Namespace: "default"},
			ClusterID: "c1", ClusterHandle: "prod-eu", ClusterName: "production",
		},
	}
	model := loadClusters(t, New(t.Context(), loader, theme.New(colorprofile.ASCII)))
	model, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(cmd())
	model, cmd = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(cmd())

	model, cmd = model.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
	if cmd == nil {
		t.Fatal("expected cluster reload for clone")
	}
	model, _ = model.Update(cmd())
	if model.mode != modeCloneCluster || !model.pickingCloneDest {
		t.Fatalf("clone cluster mode = %d picking=%v", model.mode, model.pickingCloneDest)
	}
	if !strings.Contains(model.View(80, 24), "Choose destination cluster") {
		t.Fatalf("missing destination picker:\n%s", model.View(80, 24))
	}
	if !strings.Contains(ansi.Strip(model.View(80, 24)), "(source)") {
		t.Fatalf("source cluster not marked:\n%s", ansi.Strip(model.View(80, 24)))
	}

	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if model.mode != modeClone || model.cloneDest.ID != "c2" {
		t.Fatalf("clone form dest = %#v mode=%d", model.cloneDest, model.mode)
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter}) // accept name
	model.formInput.SetValue("default")
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter}) // accept namespace → review
	if model.mode != modeReview || model.pending.clone == nil || model.pending.clone.DestClusterID != "c2" {
		t.Fatalf("review = mode=%d pending=%#v", model.mode, model.pending)
	}
	if model.pending.clone.Name != "api-clone" {
		t.Fatalf("clone name = %q", model.pending.clone.Name)
	}
	if !strings.Contains(strings.Join(model.pending.lines, "\n"), "@staging") {
		t.Fatalf("review missing dest label: %#v", model.pending.lines)
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

func TestListScrollKeepsCursorVisible(t *testing.T) {
	items := make([]servicesbridge.Summary, 0, 20)
	for i := 0; i < 20; i++ {
		items = append(items, servicesbridge.Summary{
			ID: string(rune('a'+i)), Name: "svc-" + string(rune('a'+i)), Namespace: "default", Status: "HEALTHY",
		})
	}
	loader := &fakeLoader{
		clusters: []servicesbridge.Cluster{{ID: "c1", Handle: "prod-eu"}},
		page:     servicesbridge.Page{Items: items},
	}
	model := loadClusters(t, New(t.Context(), loader, theme.New(colorprofile.ASCII)))
	model, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(cmd())
	if model.mode != modeList {
		t.Fatalf("mode = %d", model.mode)
	}
	for i := 0; i < 12; i++ {
		model, _ = model.Update(tea.KeyPressMsg{Code: 'j'})
	}
	view := ansi.Strip(model.View(80, 24))
	if !strings.Contains(view, "svc-m") {
		t.Fatalf("cursor row not visible after scroll:\n%s", view)
	}
	if !strings.Contains(view, "…") {
		t.Fatalf("expected window indicator:\n%s", view)
	}
}
