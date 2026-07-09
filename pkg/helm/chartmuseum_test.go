package helm

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/pluralsh/plural-cli/pkg/config"
)

func TestChartMuseumGetDownloadsChart(t *testing.T) {
	t.Helper()

	var gotAuth string
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/repo/charts/test.tgz" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = io.WriteString(w, "chart-bytes")
	}))
	defer server.Close()

	oldClient := http.DefaultClient
	http.DefaultClient = server.Client()
	defer func() { http.DefaultClient = oldClient }()

	setTestConfig(t, "test-token")

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	buf, err := (&ChartMuseum{}).Get(fmt.Sprintf("cm://%s/repo/charts/test.tgz", u.Host))
	if err != nil {
		t.Fatal(err)
	}

	if gotAuth != "Bearer test-token" {
		t.Fatalf("expected bearer token auth, got %q", gotAuth)
	}
	if buf.String() != "chart-bytes" {
		t.Fatalf("unexpected body: %q", buf.String())
	}
}

func TestChartMuseumGetReturnsHTTPErrorBody(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "missing chart", http.StatusNotFound)
	}))
	defer server.Close()

	oldClient := http.DefaultClient
	http.DefaultClient = server.Client()
	defer func() { http.DefaultClient = oldClient }()

	setTestConfig(t, "")

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	_, err = (&ChartMuseum{}).Get(fmt.Sprintf("cm://%s/repo/charts/test.tgz", u.Host))
	if err == nil {
		t.Fatal("expected error")
	}

	if got := err.Error(); got != "failed to download chart cm://"+u.Host+"/repo/charts/test.tgz: status 404 Not Found: missing chart" {
		t.Fatalf("unexpected error: %s", got)
	}
}

func setTestConfig(t *testing.T, token string) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	config.SetConfig(nil)
	config.ProfileFile = filepath.Join(home, ".plural", config.ConfigName)
	if err := os.MkdirAll(filepath.Dir(config.ProfileFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config.ProfileFile, []byte(fmt.Sprintf("apiVersion: platform.plural.sh/v1alpha1\nkind: Config\nspec:\n  token: %s\n", token)), 0o644); err != nil {
		t.Fatal(err)
	}
}
