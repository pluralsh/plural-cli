package services

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"

	servicesbridge "github.com/pluralsh/plural-cli/pkg/bridge/services"
	"github.com/pluralsh/plural-cli/tui/theme"
)

func TestServicesGoldens(t *testing.T) {
	clusters := New(t.Context(), nil, theme.New(colorprofile.ASCII))
	clusters.loading = false
	clusters.mode = modeClusters
	clusters.clusters = []servicesbridge.Cluster{
		{ID: "c1", Name: "production", Handle: "prod-eu"},
		{ID: "c2", Name: "staging", Handle: "staging"},
		{ID: "c3", Name: "edge"},
	}

	list := clusters
	list.mode = modeList
	list.cluster = clusters.clusters[0]
	list.page = servicesbridge.Page{Items: []servicesbridge.Summary{
		{ID: "1", Name: "api", Namespace: "default", Status: "HEALTHY", GitRef: "main", GitFolder: "services/api"},
		{ID: "2", Name: "worker", Namespace: "jobs", Status: "FAILED", GitRef: "main", GitFolder: "services/worker"},
		{ID: "3", Name: "canary", Namespace: "default", Status: "SYNCED", GitRef: "release", GitFolder: "services/canary"},
	}, HasNext: true, EndCursor: "cursor-1"}

	detail := list
	detail.mode = modeDetail
	detail.detail = servicesbridge.Detail{
		Summary:       servicesbridge.Summary{ID: "1", Name: "api", Namespace: "default", Status: "HEALTHY", GitRef: "main", GitFolder: "services/api"},
		ClusterHandle: "prod-eu",
		ClusterName:   "production",
		RevisionSHA:   "91ca21f0deadbeef",
		RevisionRef:   "main",
		Components:    14,
		Synced:        13,
		Errors:        []servicesbridge.ServiceError{{Source: "Deployment/api", Message: "exceeded rollout deadline"}},
	}

	for _, tc := range []struct {
		name   string
		model  Model
		width  int
		height int
	}{
		{"clusters-80", clusters, 80, 24},
		{"clusters-120", clusters, 120, 30},
		{"list-80", list, 80, 24},
		{"list-120", list, 120, 30},
		{"detail-80", detail, 80, 24},
		{"detail-120", detail, 120, 30},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeGoldenView(tc.model.View(tc.width, tc.height))
			golden := filepath.Join("testdata", "services-"+tc.name+".golden")
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("read golden: %v\nactual:\n%s", err, got)
			}
			if got != strings.TrimSuffix(string(want), "\n") {
				t.Fatalf("view changed\nwant:\n%s\n\ngot:\n%s", want, got)
			}
			assertGoldenDimensions(t, got, tc.width, tc.height)
		})
	}
}

func normalizeGoldenView(view string) string {
	lines := strings.Split(ansi.Strip(view), "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " ")
	}
	return strings.Join(lines, "\n")
}

func assertGoldenDimensions(t *testing.T, view string, width, height int) {
	t.Helper()
	lines := strings.Split(view, "\n")
	if len(lines) != height {
		t.Fatalf("view height = %d, want %d", len(lines), height)
	}
	for _, line := range lines {
		if got := lipgloss.Width(line); got > width {
			t.Fatalf("line width %d exceeds %d: %q", got, width, line)
		}
	}
}

func TestWriteServicesGoldens(t *testing.T) {
	if os.Getenv("UPDATE_GOLDEN") == "" {
		t.Skip("set UPDATE_GOLDEN=1 to refresh fixtures")
	}
	clusters := New(t.Context(), nil, theme.New(colorprofile.ASCII))
	clusters.loading = false
	clusters.mode = modeClusters
	clusters.clusters = []servicesbridge.Cluster{
		{ID: "c1", Name: "production", Handle: "prod-eu"},
		{ID: "c2", Name: "staging", Handle: "staging"},
		{ID: "c3", Name: "edge"},
	}
	list := clusters
	list.mode = modeList
	list.cluster = clusters.clusters[0]
	list.page = servicesbridge.Page{Items: []servicesbridge.Summary{
		{ID: "1", Name: "api", Namespace: "default", Status: "HEALTHY", GitRef: "main", GitFolder: "services/api"},
		{ID: "2", Name: "worker", Namespace: "jobs", Status: "FAILED", GitRef: "main", GitFolder: "services/worker"},
		{ID: "3", Name: "canary", Namespace: "default", Status: "SYNCED", GitRef: "release", GitFolder: "services/canary"},
	}, HasNext: true, EndCursor: "cursor-1"}
	detail := list
	detail.mode = modeDetail
	detail.detail = servicesbridge.Detail{
		Summary:       servicesbridge.Summary{ID: "1", Name: "api", Namespace: "default", Status: "HEALTHY", GitRef: "main", GitFolder: "services/api"},
		ClusterHandle: "prod-eu",
		ClusterName:   "production",
		RevisionSHA:   "91ca21f0deadbeef",
		RevisionRef:   "main",
		Components:    14,
		Synced:        13,
		Errors:        []servicesbridge.ServiceError{{Source: "Deployment/api", Message: "exceeded rollout deadline"}},
	}
	_ = os.MkdirAll("testdata", 0o755)
	for _, tc := range []struct {
		name   string
		model  Model
		width  int
		height int
	}{
		{"clusters-80", clusters, 80, 24},
		{"clusters-120", clusters, 120, 30},
		{"list-80", list, 80, 24},
		{"list-120", list, 120, 30},
		{"detail-80", detail, 80, 24},
		{"detail-120", detail, 120, 30},
	} {
		got := normalizeGoldenView(tc.model.View(tc.width, tc.height)) + "\n"
		if err := os.WriteFile(filepath.Join("testdata", "services-"+tc.name+".golden"), []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
