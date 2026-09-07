package up

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/manifest"
)

func TestHasWorkspace(t *testing.T) {
	dir := t.TempDir()
	prev, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })

	if HasWorkspace() {
		t.Fatal("expected no workspace")
	}
	pm := &manifest.ProjectManifest{Cluster: "demo", Provider: api.ProviderAWS, Region: "us-east-2"}
	if err := pm.Write(filepath.Join(dir, "workspace.yaml")); err != nil {
		t.Fatal(err)
	}
	if !HasWorkspace() {
		t.Fatal("expected workspace present")
	}
	ws, err := LoadExistingWorkspace()
	if err != nil {
		t.Fatal(err)
	}
	if ws.Cluster != "demo" || ws.ProviderID != api.ProviderAWS {
		t.Fatalf("ws = %#v", ws)
	}
}

func TestLoadExistingWorkspaceAppDomain(t *testing.T) {
	dir := t.TempDir()
	prev, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })

	pm := &manifest.ProjectManifest{
		Cluster:             "demo",
		Provider:            api.ProviderAWS,
		Region:              "us-east-2",
		AppDomain:           "apps.example.com",
		AppDomainConfigured: true,
	}
	if err := pm.Write(filepath.Join(dir, "workspace.yaml")); err != nil {
		t.Fatal(err)
	}
	ws, err := LoadExistingWorkspace()
	if err != nil {
		t.Fatal(err)
	}
	if ws.AppDomain != "apps.example.com" || !ws.AppDomainConfigured {
		t.Fatalf("ws = %#v", ws)
	}
}

func TestPersistAppDomain(t *testing.T) {
	dir := t.TempDir()
	prev, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })

	pm := &manifest.ProjectManifest{Cluster: "demo", Provider: api.ProviderAWS, Region: "us-east-2"}
	if err := pm.Write(filepath.Join(dir, "workspace.yaml")); err != nil {
		t.Fatal(err)
	}
	if err := PersistAppDomain(""); err != nil {
		t.Fatal(err)
	}
	loaded, err := manifest.ReadProject(filepath.Join(dir, "workspace.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !loaded.AppDomainConfigured {
		t.Fatal("expected AppDomainConfigured")
	}
}
