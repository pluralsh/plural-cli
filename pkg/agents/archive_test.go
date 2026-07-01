package agents

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestTarGzipArchiveReaderReadsManifestAndExtractsSubtree(t *testing.T) {
	archivePath := writeArchive(t, map[string]string{
		"manifest.json":         `{"version":1,"agentRunId":"run-1","provider":"codex","repository":"git@github.com:pluralsh/plural.git","session":{"archivePath":"sessions"},"resume":{"commands":[["codex","resume"]]}}`,
		"sessions/thread.jsonl": "session",
	})

	reader := TarGzipArchiveReader{}
	manifest, err := reader.ReadManifest(archivePath)
	if err != nil {
		t.Fatalf("ReadManifest returned error: %v", err)
	}
	if manifest.AgentRunID != "run-1" || manifest.Provider != "CODEX" {
		t.Fatalf("unexpected manifest: %#v", manifest)
	}

	dst := filepath.Join(t.TempDir(), "sessions")
	if err := reader.ExtractSubtree(archivePath, "sessions", dst); err != nil {
		t.Fatalf("ExtractSubtree returned error: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(dst, "thread.jsonl"))
	if err != nil {
		t.Fatalf("read extracted file: %v", err)
	}
	if string(content) != "session" {
		t.Fatalf("unexpected extracted content %q", string(content))
	}
}

func TestTarGzipArchiveReaderOverwritesExistingFileByDefault(t *testing.T) {
	archivePath := writeArchive(t, map[string]string{
		"sessions/thread.jsonl": "restored",
	})
	dst := filepath.Join(t.TempDir(), "sessions")
	if err := os.MkdirAll(dst, 0755); err != nil {
		t.Fatal(err)
	}
	existing := filepath.Join(dst, "thread.jsonl")
	if err := os.WriteFile(existing, []byte("existing"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := (TarGzipArchiveReader{}).ExtractSubtree(archivePath, "sessions", dst); err != nil {
		t.Fatalf("ExtractSubtree returned error: %v", err)
	}
	content, err := os.ReadFile(existing)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "restored" {
		t.Fatalf("expected existing file to be overwritten, got %q", string(content))
	}
}

func TestTarGzipArchiveReaderRejectsLinks(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent-session.tar.gz")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	gzw := gzip.NewWriter(file)
	tw := tar.NewWriter(gzw)
	if err := tw.WriteHeader(&tar.Header{Name: "projects/link", Typeflag: tar.TypeSymlink, Linkname: "/tmp/target"}); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	err = TarGzipArchiveReader{}.ExtractSubtree(path, "projects", filepath.Join(t.TempDir(), "projects"))
	if err == nil {
		t.Fatalf("expected link rejection")
	}
}

func writeArchive(t *testing.T, entries map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "agent-session.tar.gz")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	gzw := gzip.NewWriter(file)
	tw := tar.NewWriter(gzw)
	for name, content := range entries {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0644, Size: int64(len(content)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}
