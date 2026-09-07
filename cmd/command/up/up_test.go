package up

import (
	"path/filepath"
	"testing"

	"github.com/pluralsh/plural-cli/pkg/manifest"
)

func TestAppDomainAlreadyConfigured(t *testing.T) {
	tests := []struct {
		name    string
		project *manifest.ProjectManifest
		want    bool
	}{
		{
			name:    "nil project",
			project: nil,
			want:    false,
		},
		{
			name:    "never asked",
			project: &manifest.ProjectManifest{},
			want:    false,
		},
		{
			name:    "legacy manifest with app domain",
			project: &manifest.ProjectManifest{AppDomain: "apps.example.com"},
			want:    true,
		},
		{
			name:    "explicit skip",
			project: &manifest.ProjectManifest{AppDomainConfigured: true},
			want:    true,
		},
		{
			name: "configured with domain",
			project: &manifest.ProjectManifest{
				AppDomain:           "apps.example.com",
				AppDomainConfigured: true,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := appDomainAlreadyConfigured(tt.project); got != tt.want {
				t.Errorf("appDomainAlreadyConfigured() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessAppDomainPersistsSkip(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	project := &manifest.ProjectManifest{Cluster: "test"}
	if err := project.Write(filepath.Join(dir, "workspace.yaml")); err != nil {
		t.Fatal(err)
	}

	if err := processAppDomain("", project); err != nil {
		t.Fatal(err)
	}

	if !project.AppDomainConfigured {
		t.Fatal("expected AppDomainConfigured after skipping domain")
	}
	if project.AppDomain != "" {
		t.Fatalf("expected empty AppDomain, got %q", project.AppDomain)
	}

	loaded, err := manifest.ReadProject(filepath.Join(dir, "workspace.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !loaded.AppDomainConfigured {
		t.Fatal("expected AppDomainConfigured to be flushed to workspace.yaml")
	}
	if loaded.AppDomain != "" {
		t.Fatalf("expected flushed AppDomain to be empty, got %q", loaded.AppDomain)
	}
}

func TestProcessAppDomainPersistsDomain(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	project := &manifest.ProjectManifest{Cluster: "test", Provider: "aws"}
	if err := project.Write(filepath.Join(dir, "workspace.yaml")); err != nil {
		t.Fatal(err)
	}

	if err := processAppDomain("apps.example.com", project); err != nil {
		t.Fatal(err)
	}

	if !project.AppDomainConfigured {
		t.Fatal("expected AppDomainConfigured after setting domain")
	}
	if project.AppDomain != "apps.example.com" {
		t.Fatalf("AppDomain = %q, want apps.example.com", project.AppDomain)
	}

	loaded, err := manifest.ReadProject(filepath.Join(dir, "workspace.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !loaded.AppDomainConfigured {
		t.Fatal("expected AppDomainConfigured to be flushed to workspace.yaml")
	}
	if loaded.AppDomain != "apps.example.com" {
		t.Fatalf("flushed AppDomain = %q, want apps.example.com", loaded.AppDomain)
	}
}
