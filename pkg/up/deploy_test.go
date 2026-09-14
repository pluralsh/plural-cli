package up

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/pluralsh/plural-cli/pkg/manifest"
)

func TestRunCheckpointAdvancesAndFlushes(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	man := &manifest.ProjectManifest{Cluster: "test"}
	if err := man.Write(filepath.Join(dir, "workspace.yaml")); err != nil {
		t.Fatal(err)
	}

	ctx := &Context{Manifest: man}
	if err := ctx.runCheckpoint("", "init", func() error { return nil }); err != nil {
		t.Fatal(err)
	}

	if man.Checkpoint != "init" {
		t.Fatalf("Checkpoint = %q, want init", man.Checkpoint)
	}

	loaded, err := manifest.ReadProject(filepath.Join(dir, "workspace.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Checkpoint != "init" {
		t.Fatalf("flushed Checkpoint = %q, want init", loaded.Checkpoint)
	}
}

func TestRunCheckpointDoesNotAdvanceOnError(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	man := &manifest.ProjectManifest{Cluster: "test"}
	if err := man.Write(filepath.Join(dir, "workspace.yaml")); err != nil {
		t.Fatal(err)
	}

	ctx := &Context{Manifest: man}
	err := ctx.runCheckpoint("", "init", func() error {
		return errors.New("terraform failed")
	})
	if err == nil {
		t.Fatal("expected error")
	}

	if man.Checkpoint != "" {
		t.Fatalf("Checkpoint = %q, want empty", man.Checkpoint)
	}

	loaded, err := manifest.ReadProject(filepath.Join(dir, "workspace.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Checkpoint != "" {
		t.Fatalf("flushed Checkpoint = %q, want empty", loaded.Checkpoint)
	}
}

func TestRunCheckpointSkipsCompleted(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	man := &manifest.ProjectManifest{Cluster: "test", Checkpoint: "init"}
	if err := man.Write(filepath.Join(dir, "workspace.yaml")); err != nil {
		t.Fatal(err)
	}

	called := false
	ctx := &Context{Manifest: man}
	if err := ctx.runCheckpoint(man.Checkpoint, "init", func() error {
		called = true
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	if called {
		t.Fatal("expected completed checkpoint to be skipped")
	}
	if man.Checkpoint != "init" {
		t.Fatalf("Checkpoint = %q, want init", man.Checkpoint)
	}
}

func TestRunCheckpointRunsLaterPhase(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	man := &manifest.ProjectManifest{Cluster: "test", Checkpoint: "init"}
	if err := man.Write(filepath.Join(dir, "workspace.yaml")); err != nil {
		t.Fatal(err)
	}

	ctx := &Context{Manifest: man}
	if err := ctx.runCheckpoint(man.Checkpoint, "commit", func() error { return nil }); err != nil {
		t.Fatal(err)
	}

	if man.Checkpoint != "commit" {
		t.Fatalf("Checkpoint = %q, want commit", man.Checkpoint)
	}

	loaded, err := manifest.ReadProject(filepath.Join(dir, "workspace.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Checkpoint != "commit" {
		t.Fatalf("flushed Checkpoint = %q, want commit", loaded.Checkpoint)
	}
}
