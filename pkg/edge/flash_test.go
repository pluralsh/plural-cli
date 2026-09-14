package edge

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFlashCopiesImageOntoDevice(t *testing.T) {
	dir := t.TempDir()
	image := filepath.Join(dir, "kairos.img")
	device := filepath.Join(dir, "mmcblk0")
	if err := os.WriteFile(image, []byte("edge-image"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(device, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Flash(FlashOptions{Image: image, Device: device}); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(device)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "edge-image" {
		t.Fatalf("device = %q", got)
	}
}

func TestFlashRequiresPaths(t *testing.T) {
	if err := Flash(FlashOptions{Device: "/dev/sda"}); err == nil {
		t.Fatal("expected image path error")
	}
	if err := Flash(FlashOptions{Image: "kairos.img"}); err == nil {
		t.Fatal("expected device path error")
	}
}

func TestDefaultFlashImageFindsKairosInImageDir(t *testing.T) {
	dir := t.TempDir()
	if got := DefaultFlashImage(dir); got != "" {
		t.Fatalf("empty dir = %q", got)
	}
	nested := filepath.Join(dir, "image")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(nested, "kairos.img")
	if err := os.WriteFile(path, []byte("img"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := DefaultFlashImage(dir)
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != abs {
		t.Fatalf("got %q want %q", got, abs)
	}
}

func TestFlashElevatesOnPermissionDenied(t *testing.T) {
	origOpen, origElevate := openFlashDevice, elevateFlash
	t.Cleanup(func() {
		openFlashDevice, elevateFlash = origOpen, origElevate
	})
	openFlashDevice = func(string) (*os.File, error) { return nil, os.ErrPermission }
	called := false
	elevateFlash = func(options FlashOptions, err error) error {
		called = true
		if options.Device != "/dev/sdb" || !isPermissionErr(err) {
			t.Fatalf("elevate options=%#v err=%v", options, err)
		}
		return nil
	}
	if err := Flash(FlashOptions{Image: "kairos.img", Device: "/dev/sdb"}); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("expected sudo/pkexec retry")
	}
}

func TestFlashPermissionHintWhenElevateFails(t *testing.T) {
	origOpen, origElevate := openFlashDevice, elevateFlash
	t.Cleanup(func() {
		openFlashDevice, elevateFlash = origOpen, origElevate
	})
	openFlashDevice = func(string) (*os.File, error) { return nil, os.ErrPermission }
	elevateFlash = func(options FlashOptions, err error) error {
		return fmt.Errorf("could not open device: %w\nneed root to write %s — retry with: sudo plural tui", err, options.Device)
	}
	err := Flash(FlashOptions{Image: "kairos.img", Device: "/dev/sdb"})
	if err == nil || !strings.Contains(err.Error(), "sudo plural tui") {
		t.Fatalf("error = %v", err)
	}
}

func TestParseDDProgress(t *testing.T) {
	n, ok := parseDDProgress("6606028800 bytes (6,6 GB, 6,2 GiB) copied, 3 s, 2,2 GB/s")
	if !ok || n != 6606028800 {
		t.Fatalf("got %d ok=%v", n, ok)
	}
	n, ok = parseDDProgress(" 7780433920 bytes (7,8 GB, 7,2 GiB) copied, 64 s, 121 MB/s")
	if !ok || n != 7780433920 {
		t.Fatalf("got %d ok=%v", n, ok)
	}
	if _, ok := parseDDProgress("sudo: a password is required"); ok {
		t.Fatal("log line should not parse as progress")
	}
}

func TestFlashReportsProgress(t *testing.T) {
	dir := t.TempDir()
	image := filepath.Join(dir, "kairos.img")
	device := filepath.Join(dir, "mmcblk0")
	payload := bytes.Repeat([]byte("x"), 2<<20)
	if err := os.WriteFile(image, payload, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(device, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	var last int64
	if err := Flash(FlashOptions{
		Image:  image,
		Device: device,
		OnProgress: func(written, total int64) {
			last = written
			if total != int64(len(payload)) {
				t.Fatalf("total = %d", total)
			}
		},
	}); err != nil {
		t.Fatal(err)
	}
	if last != int64(len(payload)) {
		t.Fatalf("written = %d", last)
	}
}

func TestDDProgressKeepsLogsAndParsesBytes(t *testing.T) {
	var logs strings.Builder
	var got int64
	parser := &ddProgress{
		total: 10 << 30,
		log:   &logs,
		fn:    func(written, _ int64) { got = written },
	}
	_, _ = parser.Write([]byte("sudo: a password is required\n6606028800 bytes (6,6 GB, 6,2 GiB) copied, 3 s, 2,2 GB/s\r"))
	parser.Close()
	if got != 6606028800 {
		t.Fatalf("written = %d", got)
	}
	if !strings.Contains(logs.String(), "sudo: a password is required") {
		t.Fatalf("logs = %q", logs.String())
	}
	if strings.Contains(logs.String(), "copied") {
		t.Fatalf("progress leaked into logs: %q", logs.String())
	}
}
