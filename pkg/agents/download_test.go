package agents

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/pluralsh/console/go/client"
)

func TestHTTPArchiveStoreDownloadStoresUnderPluralHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	store := NewHTTPArchiveStore(&http.Client{Transport: roundTripper(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString("archive")),
		}, nil
	})})
	download, err := store.Download(context.Background(), "https://example.com/session.tar.gz", "run-1", client.AgentRuntimeTypeCodex)
	if err != nil {
		t.Fatalf("Download returned error: %v", err)
	}

	expected := filepath.Join(home, ".plural", "ai", "agents", "sessions", string(client.AgentRuntimeTypeCodex), "run-1", SessionTarName)
	if download.Path != expected {
		t.Fatalf("expected path %q, got %q", expected, download.Path)
	}
	content, err := os.ReadFile(expected)
	if err != nil {
		t.Fatalf("read downloaded archive: %v", err)
	}
	if string(content) != "archive" {
		t.Fatalf("unexpected content %q", string(content))
	}
}

type roundTripper func(*http.Request) (*http.Response, error)

func (r roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return r(req)
}
