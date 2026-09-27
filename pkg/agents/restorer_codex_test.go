package agents

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCodexRestorerPrepareCopiesDatePartitionedSessionToCodexHome(t *testing.T) {
	codexHome := t.TempDir()
	t.Setenv("CODEX_HOME", codexHome)

	sessionContent := `{"type":"session_meta","payload":{"id":"session-id","timestamp":"2026-06-02T10:00:00Z"}}`
	archivePath := writeArchive(t, map[string]string{
		"sessions/2026/06/02/session.jsonl": sessionContent,
	})
	workDir := t.TempDir()

	restorer := &CodexRestorer{baseRestorer: baseRestorer{archive: TarGzipArchiveReader{}}}
	prepared, err := restorer.Prepare(context.Background(), RestoreOptions{
		RepoPath:    filepath.Join(t.TempDir(), "plural"),
		ArchivePath: archivePath,
		WorkDir:     workDir,
		Manifest:    &SessionManifest{Session: SessionMetadata{ID: "session-id"}},
	})
	if err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}
	if prepared.SessionID != "session-id" {
		t.Fatalf("expected session id to be preserved, got %q", prepared.SessionID)
	}

	assertFileContent(t, filepath.Join(workDir, ".codex", "sessions", "2026", "06", "02", "session.jsonl"), sessionContent)
	assertFileContent(t, filepath.Join(codexHome, "sessions", "2026", "06", "02", "session.jsonl"), sessionContent)
}

func TestCodexRestorerPreparePlacesFlatSessionUnderMetadataDate(t *testing.T) {
	codexHome := t.TempDir()
	t.Setenv("CODEX_HOME", codexHome)

	sessionContent := `{"type":"session_meta","payload":{"id":"session-id","timestamp":"2026-06-02T10:00:00Z"}}`
	archivePath := writeArchive(t, map[string]string{
		"sessions/session.jsonl": sessionContent,
	})

	restorer := &CodexRestorer{baseRestorer: baseRestorer{archive: TarGzipArchiveReader{}}}
	if _, err := restorer.Prepare(context.Background(), RestoreOptions{
		RepoPath:    filepath.Join(t.TempDir(), "plural"),
		ArchivePath: archivePath,
		WorkDir:     t.TempDir(),
		Manifest:    &SessionManifest{Session: SessionMetadata{ID: "session-id"}},
	}); err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}

	assertFileContent(t, filepath.Join(codexHome, "sessions", "2026", "06", "02", "session.jsonl"), sessionContent)
}

func TestCodexRestorerPrepareRemovesExistingFilesForSameSession(t *testing.T) {
	codexHome := t.TempDir()
	t.Setenv("CODEX_HOME", codexHome)

	sessionContent := `{"type":"session_meta","payload":{"id":"session-id","timestamp":"2026-06-02T10:00:00Z"}}`
	archivePath := writeArchive(t, map[string]string{
		"sessions/2026/06/02/session.jsonl": sessionContent,
	})
	oldSession := filepath.Join(codexHome, "sessions", "2026", "06", "01", "old.jsonl")
	otherSession := filepath.Join(codexHome, "sessions", "2026", "06", "01", "other.jsonl")
	for path, content := range map[string]string{
		oldSession:   `{"type":"session_meta","payload":{"id":"session-id","timestamp":"2026-06-01T10:00:00Z"}}`,
		otherSession: `{"type":"session_meta","payload":{"id":"other-session","timestamp":"2026-06-01T10:00:00Z"}}`,
	} {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	var prompted []string
	restorer := &CodexRestorer{baseRestorer: baseRestorer{archive: TarGzipArchiveReader{}}}
	if _, err := restorer.Prepare(context.Background(), RestoreOptions{
		RepoPath:    filepath.Join(t.TempDir(), "plural"),
		ArchivePath: archivePath,
		WorkDir:     t.TempDir(),
		Manifest:    &SessionManifest{Session: SessionMetadata{ID: "session-id"}},
		ConfirmOverwrite: func(path string) (bool, error) {
			prompted = append(prompted, path)
			return true, nil
		},
	}); err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}

	assertNotExists(t, oldSession)
	assertFileContent(t, otherSession, `{"type":"session_meta","payload":{"id":"other-session","timestamp":"2026-06-01T10:00:00Z"}}`)
	assertFileContent(t, filepath.Join(codexHome, "sessions", "2026", "06", "02", "session.jsonl"), sessionContent)
	if len(prompted) != 1 {
		t.Fatalf("expected one prompt, got %v", prompted)
	}
}

func TestCodexRestorerPrepareUsesExistingSessionWhenOverwriteDenied(t *testing.T) {
	codexHome := t.TempDir()
	t.Setenv("CODEX_HOME", codexHome)

	sessionContent := `{"type":"session_meta","payload":{"id":"session-id","timestamp":"2026-06-02T10:00:00Z"}}`
	archivePath := writeArchive(t, map[string]string{
		"sessions/2026/06/02/session.jsonl": sessionContent,
	})
	existing := filepath.Join(codexHome, "sessions", "2026", "06", "01", "old.jsonl")
	if err := os.MkdirAll(filepath.Dir(existing), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(existing, []byte(`{"type":"session_meta","payload":{"id":"session-id","timestamp":"2026-06-01T10:00:00Z"}}`), 0644); err != nil {
		t.Fatal(err)
	}

	restorer := &CodexRestorer{baseRestorer: baseRestorer{archive: TarGzipArchiveReader{}}}
	prepared, err := restorer.Prepare(context.Background(), RestoreOptions{
		RepoPath:    filepath.Join(t.TempDir(), "plural"),
		ArchivePath: archivePath,
		WorkDir:     t.TempDir(),
		Manifest:    &SessionManifest{Session: SessionMetadata{ID: "session-id"}},
		ConfirmOverwrite: func(_ string) (bool, error) {
			return false, nil
		},
	})
	if err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}
	if prepared.SessionID != "session-id" {
		t.Fatalf("expected session id to be preserved, got %q", prepared.SessionID)
	}

	assertFileContent(t, existing, `{"type":"session_meta","payload":{"id":"session-id","timestamp":"2026-06-01T10:00:00Z"}}`)
	assertNotExists(t, filepath.Join(codexHome, "sessions", "2026", "06", "02", "session.jsonl"))
}

func TestCodexResumeInvocationDisablesCwdFilterAndSetsHome(t *testing.T) {
	codexHome := t.TempDir()
	t.Setenv("CODEX_HOME", codexHome)
	repo := t.TempDir()
	abs, err := filepath.Abs(repo)
	if err != nil {
		t.Fatal(err)
	}

	home, env, args, err := (&CodexRestorer{}).resumeInvocation(&PreparedSession{
		RepoPath:  repo,
		SessionID: "01a08034-e9fb-7a00-baba-df7d0660e69f",
	})
	if err != nil {
		t.Fatalf("resumeInvocation returned error: %v", err)
	}
	if home != codexHome {
		t.Fatalf("expected home %q, got %q", codexHome, home)
	}
	if len(env) != 1 || env[0] != "CODEX_HOME="+codexHome {
		t.Fatalf("expected CODEX_HOME env, got %v", env)
	}
	want := []string{"resume", "--all", "01a08034-e9fb-7a00-baba-df7d0660e69f", "-C", abs}
	if len(args) != len(want) {
		t.Fatalf("expected args %v, got %v", want, args)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf("expected args %v, got %v", want, args)
		}
	}
}

func TestCodexConfigDirUsesSnapHome(t *testing.T) {
	userHome := t.TempDir()
	t.Setenv("HOME", userHome)
	t.Setenv("CODEX_HOME", "")
	snapHome := filepath.Join(userHome, "snap", "codex", "current")
	if err := os.MkdirAll(snapHome, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(snapHome, "config.toml"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	orig := lookPath
	t.Cleanup(func() { lookPath = orig })
	lookPath = func(string) (string, error) {
		return "/snap/bin/codex", nil
	}

	got, err := (&CodexRestorer{}).configDir()
	if err != nil {
		t.Fatalf("configDir returned error: %v", err)
	}
	if got != snapHome {
		t.Fatalf("expected snap home %q, got %q", snapHome, got)
	}
}

func TestCodexConfigDirFallsBackToDotCodex(t *testing.T) {
	userHome := t.TempDir()
	t.Setenv("HOME", userHome)
	t.Setenv("CODEX_HOME", "")

	orig := lookPath
	t.Cleanup(func() { lookPath = orig })
	lookPath = func(string) (string, error) {
		return "/usr/local/bin/codex", nil
	}

	got, err := (&CodexRestorer{}).configDir()
	if err != nil {
		t.Fatalf("configDir returned error: %v", err)
	}
	want := filepath.Join(userHome, ".codex")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestSnapCodexHome(t *testing.T) {
	userHome := t.TempDir()
	if got := snapCodexHome("/usr/bin/codex", userHome); got != "" {
		t.Fatalf("expected empty for non-snap binary, got %q", got)
	}

	current := filepath.Join(userHome, "snap", "codex", "current")
	if err := os.MkdirAll(current, 0755); err != nil {
		t.Fatal(err)
	}
	got := snapCodexHome("/snap/bin/codex", userHome)
	if got != current {
		t.Fatalf("expected %q when snap dir exists, got %q", current, got)
	}

	if err := os.WriteFile(filepath.Join(current, "config.toml"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	got = snapCodexHome("/snap/bin/codex", userHome)
	if got != current {
		t.Fatalf("expected %q when config exists, got %q", current, got)
	}
}
