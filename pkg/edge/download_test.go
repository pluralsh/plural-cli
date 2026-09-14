package edge

import (
	"io"
	"strings"
	"testing"
)

func TestDownloadRequiresOCIURL(t *testing.T) {
	err := Download(DownloadOptions{To: t.TempDir()}, nil)
	if err == nil || !strings.Contains(err.Error(), "oci-url") {
		t.Fatalf("error = %v", err)
	}
}

func TestDownloadFetchesAndUntars(t *testing.T) {
	dir := t.TempDir()
	var fetched, dest string
	err := Download(DownloadOptions{
		OCIURL: "ghcr.io/pluralsh/edge:latest",
		To:     dir,
		Fetch: func(url string) (io.ReadCloser, error) {
			fetched = url
			return io.NopCloser(strings.NewReader("image-bytes")), nil
		},
		Untar: func(d string, r io.Reader) error {
			dest = d
			b, err := io.ReadAll(r)
			if err != nil {
				return err
			}
			if string(b) != "image-bytes" {
				t.Fatalf("untar bytes = %q", b)
			}
			return nil
		},
	}, func(string) {})
	if err != nil {
		t.Fatal(err)
	}
	if fetched != "ghcr.io/pluralsh/edge:latest" || dest != dir {
		t.Fatalf("fetched=%q dest=%q", fetched, dest)
	}
}
