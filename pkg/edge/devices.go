package edge

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

var (
	scsiPartition = regexp.MustCompile(`^(sd|vd|hd|xvd)[a-z]+[0-9]+$`)
	nvmePartition = regexp.MustCompile(`p[0-9]+$`)
)

// FlashDevice is a whole-disk USB (or USB-attached) block device suitable for flash.
type FlashDevice struct {
	Path      string
	Name      string
	Model     string
	Size      uint64
	NeedsRoot bool
}

// Label is the TUI/CLI summary: path, size, and model.
func (d FlashDevice) Label() string {
	parts := []string{d.Path}
	if d.Size > 0 {
		parts = append(parts, formatBytes(d.Size))
	}
	if model := strings.TrimSpace(d.Model); model != "" {
		parts = append(parts, model)
	}
	return strings.Join(parts, "  ")
}

// ListFlashDevices returns USB whole disks, excluding the system disk.
func ListFlashDevices() ([]FlashDevice, error) {
	if runtime.GOOS != "linux" {
		return nil, nil
	}
	return deviceScan{}.list()
}

type deviceScan struct {
	sysBlock string
	mounts   string
}

func (s deviceScan) list() ([]FlashDevice, error) {
	sysBlock := s.sysBlock
	if sysBlock == "" {
		sysBlock = "/sys/block"
	}
	entries, err := os.ReadDir(sysBlock)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	system := s.systemDisks()
	var devices []FlashDevice
	for _, entry := range entries {
		name := entry.Name()
		if !isFlashCandidate(name) || system[name] {
			continue
		}
		blockDir := filepath.Join(sysBlock, name)
		if !isUSBBlock(blockDir) {
			continue
		}
		path := "/dev/" + name
		devices = append(devices, FlashDevice{
			Path:      path,
			Name:      name,
			Model:     strings.TrimSpace(readSysFile(filepath.Join(blockDir, "device", "model"))),
			Size:      sysSize(blockDir),
			NeedsRoot: s.sysBlock == "" && !DeviceWritable(path),
		})
	}
	return devices, nil
}

func isFlashCandidate(name string) bool {
	prefixes := []string{"loop", "ram", "sr", "fd", "dm-", "zram", "md", "nbd"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(name, prefix) {
			return false
		}
	}
	return !isPartitionName(name)
}

func isPartitionName(name string) bool {
	if strings.HasPrefix(name, "nvme") || strings.HasPrefix(name, "mmcblk") {
		return nvmePartition.MatchString(name)
	}
	return scsiPartition.MatchString(name)
}

func isUSBBlock(blockDir string) bool {
	resolved, err := filepath.EvalSymlinks(blockDir)
	if err != nil {
		resolved = blockDir
	}
	return strings.Contains(strings.ToLower(filepath.ToSlash(resolved)), "/usb")
}

func sysSize(blockDir string) uint64 {
	raw := strings.TrimSpace(readSysFile(filepath.Join(blockDir, "size")))
	sectors, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0
	}
	return sectors * 512
}

func readSysFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func (s deviceScan) systemDisks() map[string]bool {
	path := s.mounts
	if path == "" {
		path = "/proc/mounts"
	}
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	out := map[string]bool{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		source, mount := fields[0], fields[1]
		if mount != "/" && mount != "/boot" && mount != "/boot/efi" {
			continue
		}
		if disk := diskName(source); disk != "" {
			out[disk] = true
		}
	}
	return out
}

func diskName(source string) string {
	if !strings.HasPrefix(source, "/dev/") {
		return ""
	}
	name := strings.TrimPrefix(source, "/dev/")
	if strings.Contains(name, "/") {
		return ""
	}
	if !isPartitionName(name) {
		return name
	}
	if strings.HasPrefix(name, "nvme") || strings.HasPrefix(name, "mmcblk") {
		if i := strings.LastIndex(name, "p"); i > 0 {
			return name[:i]
		}
	}
	i := len(name)
	for i > 0 && name[i-1] >= '0' && name[i-1] <= '9' {
		i--
	}
	return name[:i]
}

func formatBytes(n uint64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := uint64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}
