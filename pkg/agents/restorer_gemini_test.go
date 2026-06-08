package agents

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestGeminiRestorerPrepareCopiesChatsToGeminiHomeRepoTmpDir(t *testing.T) {
	geminiHome := t.TempDir()
	t.Setenv("GEMINI_CLI_HOME", geminiHome)

	archivePath := writeArchive(t, map[string]string{
		"chats/session.json":        "session",
		"chats/nested/message.json": "message",
	})
	workDir := t.TempDir()
	repoPath := filepath.Join(t.TempDir(), "plural")

	restorer := &GeminiRestorer{baseRestorer: baseRestorer{archive: TarGzipArchiveReader{}}}
	prepared, err := restorer.Prepare(context.Background(), RestoreOptions{
		RepoPath:    repoPath,
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

	assertFileContent(t, filepath.Join(workDir, "chats", "session.json"), "session")
	assertFileContent(t, filepath.Join(geminiHome, "tmp", "plural", "chats", "session.json"), "session")
	assertFileContent(t, filepath.Join(geminiHome, "tmp", "plural", "chats", "nested", "message.json"), "message")
}

func TestGeminiRestorerPrepareMergesChatsWithOverwritePrompt(t *testing.T) {
	geminiHome := t.TempDir()
	t.Setenv("GEMINI_CLI_HOME", geminiHome)

	archivePath := writeArchive(t, map[string]string{
		"chats/existing.json": "restored",
		"chats/new.json":      "new",
	})
	workDir := t.TempDir()
	repoPath := filepath.Join(t.TempDir(), "plural")
	existing := filepath.Join(geminiHome, "tmp", "plural", "chats", "existing.json")
	if err := os.MkdirAll(filepath.Dir(existing), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(existing, []byte("existing"), 0644); err != nil {
		t.Fatal(err)
	}

	var prompted []string
	restorer := &GeminiRestorer{baseRestorer: baseRestorer{archive: TarGzipArchiveReader{}}}
	if _, err := restorer.Prepare(context.Background(), RestoreOptions{
		RepoPath:    repoPath,
		ArchivePath: archivePath,
		WorkDir:     workDir,
		Manifest:    &SessionManifest{Session: SessionMetadata{ID: "session-id"}},
		ConfirmOverwrite: func(path string) (bool, error) {
			prompted = append(prompted, path)
			return false, nil
		},
	}); err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}

	assertFileContent(t, existing, "existing")
	assertFileContent(t, filepath.Join(geminiHome, "tmp", "plural", "chats", "new.json"), "new")
	if len(prompted) != 1 || prompted[0] != existing {
		t.Fatalf("expected prompt for %q, got %v", existing, prompted)
	}
}

func TestGeminiRestorerPrepareRemovesOlderTimestampedFilesForSameSession(t *testing.T) {
	geminiHome := t.TempDir()
	t.Setenv("GEMINI_CLI_HOME", geminiHome)

	archivePath := writeArchive(t, map[string]string{
		"chats/session-2026-06-02T10-00-00-session.json": `{"sessionId":"session-id","messages":["new"]}`,
	})
	workDir := t.TempDir()
	repoPath := filepath.Join(t.TempDir(), "plural")
	localChatsDir := filepath.Join(geminiHome, "tmp", "plural", "chats")
	oldSession := filepath.Join(localChatsDir, "session-2026-06-01T10-00-00-session.json")
	olderSession := filepath.Join(localChatsDir, "session-2026-05-31T10-00-00-session.json")
	otherSession := filepath.Join(localChatsDir, "session-2026-06-01T10-00-00-other.json")
	for path, content := range map[string]string{
		oldSession:   `{"sessionId":"session-id","messages":["old"]}`,
		olderSession: `{"session_id":"session-id","messages":["older"]}`,
		otherSession: `{"sessionId":"other-session","messages":["other"]}`,
	} {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	var prompted []string
	restorer := &GeminiRestorer{baseRestorer: baseRestorer{archive: TarGzipArchiveReader{}}}
	if _, err := restorer.Prepare(context.Background(), RestoreOptions{
		RepoPath:    repoPath,
		ArchivePath: archivePath,
		WorkDir:     workDir,
		Manifest:    &SessionManifest{Session: SessionMetadata{ID: "session-id"}},
		ConfirmOverwrite: func(path string) (bool, error) {
			prompted = append(prompted, path)
			return true, nil
		},
	}); err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}

	assertNotExists(t, oldSession)
	assertNotExists(t, olderSession)
	assertFileContent(t, otherSession, `{"sessionId":"other-session","messages":["other"]}`)
	assertFileContent(t, filepath.Join(localChatsDir, "session-2026-06-02T10-00-00-session.json"), `{"sessionId":"session-id","messages":["new"]}`)
	if len(prompted) != 1 {
		t.Fatalf("expected one prompt, got %v", prompted)
	}
}

func TestGeminiRestorerPrepareUsesExistingSessionWhenOverwriteDenied(t *testing.T) {
	geminiHome := t.TempDir()
	t.Setenv("GEMINI_CLI_HOME", geminiHome)

	archivePath := writeArchive(t, map[string]string{
		"chats/session-2026-06-02T10-00-00-session.json": `{"sessionId":"session-id","messages":["new"]}`,
	})
	workDir := t.TempDir()
	repoPath := filepath.Join(t.TempDir(), "plural")
	localChatsDir := filepath.Join(geminiHome, "tmp", "plural", "chats")
	existing := filepath.Join(localChatsDir, "session-2026-06-01T10-00-00-session.json")
	if err := os.MkdirAll(filepath.Dir(existing), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(existing, []byte(`{"sessionId":"session-id","messages":["old"]}`), 0644); err != nil {
		t.Fatal(err)
	}

	restorer := &GeminiRestorer{baseRestorer: baseRestorer{archive: TarGzipArchiveReader{}}}
	prepared, err := restorer.Prepare(context.Background(), RestoreOptions{
		RepoPath:    repoPath,
		ArchivePath: archivePath,
		WorkDir:     workDir,
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

	assertFileContent(t, existing, `{"sessionId":"session-id","messages":["old"]}`)
	assertNotExists(t, filepath.Join(localChatsDir, "session-2026-06-02T10-00-00-session.json"))
}

func assertFileContent(t *testing.T, path, expected string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if string(content) != expected {
		t.Fatalf("expected %s content %q, got %q", path, expected, string(content))
	}
}

func assertNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected %s not to exist", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat %s: %v", path, err)
	}
}
