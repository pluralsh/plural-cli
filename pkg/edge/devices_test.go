package edge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListFlashDevicesFindsUSBAndSkipsSystemDisk(t *testing.T) {
	root := t.TempDir()
	sysBlock := filepath.Join(root, "sys", "block")
	if err := os.MkdirAll(sysBlock, 0o755); err != nil {
		t.Fatal(err)
	}

	internal := filepath.Join(root, "sys", "devices", "pci0000:00", "ata1", "block", "sda")
	usb := filepath.Join(root, "sys", "devices", "pci0000:00", "usb2", "2-3", "block", "sdb")
	part := filepath.Join(root, "sys", "devices", "pci0000:00", "usb2", "2-3", "block", "sdb1")
	loop := filepath.Join(root, "sys", "devices", "virtual", "block", "loop0")
	for _, dir := range []string{internal, usb, part, loop} {
		if err := os.MkdirAll(filepath.Join(dir, "device"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	write := func(path, body string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(internal, "size"), "1953525168\n")
	write(filepath.Join(internal, "device", "model"), "Samsung SSD\n")
	write(filepath.Join(usb, "size"), "31266816\n")
	write(filepath.Join(usb, "device", "model"), "SanDisk Ultra\n")
	write(filepath.Join(part, "size"), "31266816\n")
	write(filepath.Join(loop, "size"), "2048\n")

	link := func(name, target string) {
		t.Helper()
		if err := os.Symlink(target, filepath.Join(sysBlock, name)); err != nil {
			t.Fatal(err)
		}
	}
	link("sda", internal)
	link("sdb", usb)
	link("sdb1", part)
	link("loop0", loop)

	mounts := filepath.Join(root, "proc", "mounts")
	if err := os.MkdirAll(filepath.Dir(mounts), 0o755); err != nil {
		t.Fatal(err)
	}
	write(mounts, "/dev/sda2 / ext4 rw 0 0\n/dev/sda1 /boot/efi vfat rw 0 0\n")

	got, err := deviceScan{sysBlock: sysBlock, mounts: mounts}.list()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Path != "/dev/sdb" || got[0].Model != "SanDisk Ultra" {
		t.Fatalf("devices = %#v", got)
	}
	if got[0].Size != 31266816*512 {
		t.Fatalf("size = %d", got[0].Size)
	}
	label := got[0].Label()
	if !strings.Contains(label, "/dev/sdb") || !strings.Contains(label, "SanDisk Ultra") || !strings.Contains(label, "14.9 GB") {
		t.Fatalf("label = %q", label)
	}
}

func TestDiskNameStripsPartitions(t *testing.T) {
	cases := map[string]string{
		"/dev/sda1":        "sda",
		"/dev/nvme0n1p2":   "nvme0n1",
		"/dev/mmcblk0p1":   "mmcblk0",
		"/dev/mapper/root": "",
		"/dev/sdb":         "sdb",
	}
	for in, want := range cases {
		if got := diskName(in); got != want {
			t.Fatalf("%s -> %q, want %q", in, got, want)
		}
	}
}

func TestIsFlashCandidateSkipsVirtualAndPartitions(t *testing.T) {
	skip := []string{"loop0", "sda1", "nvme0n1p1", "mmcblk0p1", "zram0"}
	for _, name := range skip {
		if isFlashCandidate(name) {
			t.Fatalf("expected skip %s", name)
		}
	}
	keep := []string{"sda", "sdb", "nvme0n1", "mmcblk0"}
	for _, name := range keep {
		if !isFlashCandidate(name) {
			t.Fatalf("expected keep %s", name)
		}
	}
}
