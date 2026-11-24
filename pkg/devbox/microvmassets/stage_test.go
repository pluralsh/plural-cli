package microvmassets

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/klauspost/compress/zstd"
	"gocloud.dev/blob"
	"gocloud.dev/blob/memblob"
)

func sum(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func kataArchive(t *testing.T) []byte {
	t.Helper()
	var tb bytes.Buffer
	tw := tar.NewWriter(&tb)
	add := func(h tar.Header, body string) {
		h.Size = int64(len(body))
		if err := tw.WriteHeader(&h); err != nil {
			t.Fatal(err)
		}
		_, _ = tw.Write([]byte(body))
	}
	add(tar.Header{Name: "./opt/kata/share/kata-containers/vmlinux-6.1", Typeflag: tar.TypeReg, Mode: 0o600}, "kernel")
	add(tar.Header{Name: "./opt/kata/share/kata-containers/kata-image.img", Typeflag: tar.TypeReg, Mode: 0o600}, "image")
	add(tar.Header{Name: "./opt/kata/share/kata-containers/vmlinux.container", Typeflag: tar.TypeSymlink, Linkname: "vmlinux-6.1"}, "")
	add(tar.Header{Name: "./opt/kata/share/kata-containers/kata-containers.img", Typeflag: tar.TypeSymlink, Linkname: "kata-image.img"}, "")
	add(tar.Header{Name: "./opt/kata/libexec/virtiofsd", Typeflag: tar.TypeReg, Mode: 0o600}, "virtiofsd")
	add(tar.Header{Name: "./opt/kata/bin/ignored", Typeflag: tar.TypeReg, Mode: 0o600}, "ignored")
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	var zb bytes.Buffer
	zw, _ := zstd.NewWriter(&zb)
	_, _ = zw.Write(tb.Bytes())
	_ = zw.Close()
	return zb.Bytes()
}

// releaseServer serves a fake GitHub release layout for amd64.
func releaseServer(t *testing.T) string {
	t.Helper()
	archive := kataArchive(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/kata-containers/kata-containers/releases/download/", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(archive) })
	mux.HandleFunc("/cloud-hypervisor/cloud-hypervisor/releases/download/", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ch")) })
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv.URL
}

func goodDigests() map[string]string {
	return map[string]string{
		"cloud-hypervisor": sum("ch"),
		"virtiofsd":        sum("virtiofsd"),
		"kata-kernel":      sum("kernel"),
		"kata-image":       sum("image"),
	}
}

func newStager(t *testing.T, base string) (*stager, *blob.Bucket) {
	t.Helper()
	b := memblob.OpenBucket(nil)
	t.Cleanup(func() { b.Close() })
	return &stager{bucket: b, workDir: t.TempDir(), kataVersion: "4.3.1", chVersion: "v53.0", releasesBase: base}, b
}

func objectCount(t *testing.T, b *blob.Bucket) int {
	t.Helper()
	it := b.List(nil)
	n := 0
	for {
		if _, err := it.Next(context.Background()); err != nil {
			return n
		}
		n++
	}
}

func TestStageUploadsVerifiedAssets(t *testing.T) {
	s, b := newStager(t, releaseServer(t))
	digests := goodDigests()
	if err := s.stage(context.Background(), "amd64", digests); err != nil {
		t.Fatal(err)
	}
	for key, d := range digests {
		got, err := b.ReadAll(context.Background(), objectKey("amd64", d))
		if err != nil {
			t.Fatalf("%s not uploaded: %v", key, err)
		}
		if sum(string(got)) != d {
			t.Errorf("%s uploaded with the wrong content", key)
		}
	}
}

func TestStageRejectsHashMismatchWithoutUploading(t *testing.T) {
	s, b := newStager(t, releaseServer(t))
	digests := goodDigests()
	digests["kata-kernel"] = sum("something else")
	err := s.stage(context.Background(), "amd64", digests)
	if err == nil || !strings.Contains(err.Error(), "kata-kernel") {
		t.Fatalf("err = %v, want a kata-kernel sha256 mismatch", err)
	}
	if n := objectCount(t, b); n != 0 {
		t.Errorf("%d objects uploaded despite the mismatch", n)
	}
}

func TestStageSkipsWhenAlreadyStaged(t *testing.T) {
	// An unreachable release server proves nothing is downloaded.
	s, b := newStager(t, "http://127.0.0.1:1")
	digests := goodDigests()
	for _, d := range digests {
		if err := b.WriteAll(context.Background(), objectKey("amd64", d), []byte("x"), nil); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.stage(context.Background(), "amd64", digests); err != nil {
		t.Fatalf("expected a no-op, got %v", err)
	}
}

func TestStageRequiresAllPinnedDigests(t *testing.T) {
	s, _ := newStager(t, "http://127.0.0.1:1")
	digests := goodDigests()
	delete(digests, "virtiofsd")
	if err := s.stage(context.Background(), "amd64", digests); err == nil {
		t.Error("expected an error for a missing pinned digest")
	}
}

func TestLoadDigestsReadsDefaultVersion(t *testing.T) {
	cfg := `apiVersion: ate.dev/v1alpha1
kind: SandboxConfig
spec:
  defaultVersion: v2
  versions:
    - name: v1
      assets:
        amd64:
          virtiofsd: {sha256: old}
    - name: v2
      assets:
        amd64:
          virtiofsd: {sha256: new}
          kata-kernel: {sha256: kern}
`
	p := filepath.Join(t.TempDir(), "microvm.yaml")
	if err := os.WriteFile(p, []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := loadDigests(p)
	if err != nil {
		t.Fatal(err)
	}
	if got["amd64"]["virtiofsd"] != "new" || got["amd64"]["kata-kernel"] != "kern" {
		t.Errorf("digests = %v, want the v2 values", got)
	}
}
