package oplog

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteIncludesErrorAndStripsANSI(t *testing.T) {
	dir := t.TempDir()
	path, err := Write(dir, "down", []string{"\x1b[31mError: sts 403\x1b[0m", "ok"}, errors.New("exit status 1"))
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(path) != dir {
		t.Fatalf("path=%s", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if !strings.Contains(got, "Error: sts 403") || strings.Contains(got, "\x1b[") {
		t.Fatalf("expected stripped log, got:\n%s", got)
	}
	if !strings.Contains(got, "# error: exit status 1") {
		t.Fatalf("missing error header:\n%s", got)
	}
}
