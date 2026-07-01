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
