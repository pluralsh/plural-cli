// Derived from cmd/ate-setup/internal/steps/microvm.go in github.com/agent-substrate/substrate (v1.0.0),
// Copyright 2026 The Agent Substrate Authors, licensed under the Apache License, Version 2.0.
// The download and extraction are kept as upstream has them; only the staging target differs.

package microvmassets

import (
	"archive/tar"
	"bufio"
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"
)

const (
	// Keep at Kata 4.1.0 or later: earlier virtiofsd builds hang Cloud Hypervisor's snapshot restore.
	DefaultKataVersion            = "4.3.1"
	DefaultCloudHypervisorVersion = "v53.0"
	DefaultReleaseBaseURL         = "https://github.com"

	kataSharePrefix      = "opt/kata/share/kata-containers/"
	kataKernelSymlink    = "opt/kata/share/kata-containers/vmlinux.container"
	kataRootfsSymlink    = "opt/kata/share/kata-containers/kata-containers.img"
	kataLibexecVirtiofsd = "opt/kata/libexec/virtiofsd"

	maxAssetBytes int64 = 8 << 30
)

func cloudHypervisorAssetName(arch string) (string, error) {
	switch arch {
	case "arm64":
		return "cloud-hypervisor-static-aarch64", nil
	case "amd64":
		return "cloud-hypervisor-static", nil
	default:
		return "", fmt.Errorf("unsupported arch %q", arch)
	}
}

// assemble downloads the Kata static release and Cloud Hypervisor for arch into outDir, leaving the four files of
// assetFiles there.
func assemble(ctx context.Context, releaseBaseURL, outDir, arch, kataVer, chVer string) error {
	chAsset, err := cloudHypervisorAssetName(arch)
	if err != nil {
		return err
	}
	workDir, err := os.MkdirTemp(outDir, ".assemble-*")
	if err != nil {
		return fmt.Errorf("creating temporary directory in %s: %w", outDir, err)
	}
	defer os.RemoveAll(workDir)

	baseURL := strings.TrimSuffix(releaseBaseURL, "/")
	kataURL := fmt.Sprintf("%s/kata-containers/kata-containers/releases/download/%s/kata-static-%s-%s.tar.zst",
		baseURL, kataVer, kataVer, arch)
	logf("downloading kata-static %s (%s)", kataVer, arch)
	if err := downloadAndExtractKataAssets(ctx, kataURL, workDir, outDir); err != nil {
		return err
	}

	chURL := fmt.Sprintf("%s/cloud-hypervisor/cloud-hypervisor/releases/download/%s/%s", baseURL, chVer, chAsset)
	logf("downloading cloud-hypervisor %s (%s)", chVer, chAsset)
	chTmp := filepath.Join(workDir, "cloud-hypervisor")
	if err := downloadToFile(ctx, chURL, chTmp, 0o755); err != nil {
		return err
	}
	if err := os.Rename(chTmp, filepath.Join(outDir, "cloud-hypervisor")); err != nil {
		return err
	}

	for _, f := range assetFiles {
		if fi, err := os.Stat(filepath.Join(outDir, f)); err != nil || !fi.Mode().IsRegular() {
			return fmt.Errorf("missing asset %s/%s", outDir, f)
		}
	}
	return nil
}

func downloadAndExtractKataAssets(ctx context.Context, kataURL, workDir, outDir string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, kataURL, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("downloading %s: %w", kataURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading %s: unexpected HTTP status %s", kataURL, resp.Status)
	}

	links, err := extractKataTarZst(ctx, resp.Body, workDir)
	if err != nil {
		return fmt.Errorf("extracting %s: %w", kataURL, err)
	}

	for _, item := range []struct {
		archivePath string
		outName     string
		mode        fs.FileMode
	}{
		{archivePath: kataKernelSymlink, outName: "vmlinux", mode: 0o644},
		// rootfs.img is attached over virtio-pmem and mounted with DAX, so it must be a multiple of 2 MiB with a
		// page-aligned partition and filesystem block size. kata-containers.img is built for this.
		{archivePath: kataRootfsSymlink, outName: "rootfs.img", mode: 0o644},
		{archivePath: kataLibexecVirtiofsd, outName: "virtiofsd", mode: 0o755},
	} {
		resolved := item.archivePath
		for range 8 {
			target, ok := links[resolved]
			if !ok {
				break
			}
			resolved = target
		}
		src := filepath.Join(workDir, filepath.FromSlash(resolved))
		if fi, err := os.Stat(src); err != nil || !fi.Mode().IsRegular() {
			return fmt.Errorf("resolving %s in kata archive: missing regular file %s", item.archivePath, resolved)
		}
		if err := os.Chmod(src, item.mode); err != nil {
			return err
		}
		if err := os.Rename(src, filepath.Join(outDir, item.outName)); err != nil {
			return err
		}
	}
	return nil
}

// extractKataTarZst writes the regular files under opt/kata/share/kata-containers/ and opt/kata/libexec/virtiofsd
// of a zstd-compressed tar into workDir, and returns the archive's symlinks as path -> cleaned target.
func extractKataTarZst(ctx context.Context, r io.Reader, workDir string) (map[string]string, error) {
	zr, err := zstd.NewReader(bufio.NewReader(r))
	if err != nil {
		return nil, fmt.Errorf("creating zstd reader: %w", err)
	}
	defer zr.Close()

	links := map[string]string{}
	tr := tar.NewReader(zr)
	var totalBytes int64
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		hdr, err := tr.Next()
		if err == io.EOF {
			return links, nil
		}
		if err != nil {
			return nil, err
		}

		raw := strings.TrimPrefix(strings.TrimPrefix(hdr.Name, "./"), "/")
		if !strings.HasPrefix(raw, kataSharePrefix) && raw != kataLibexecVirtiofsd {
			continue
		}
		name := strings.TrimPrefix(path.Clean(raw), "/")
		if !filepath.IsLocal(filepath.FromSlash(name)) {
			return nil, fmt.Errorf("tar entry %q escapes the extraction directory", hdr.Name)
		}
		if !strings.HasPrefix(name, kataSharePrefix) && name != kataLibexecVirtiofsd {
			continue
		}

		switch hdr.Typeflag {
		case tar.TypeReg:
			totalBytes += hdr.Size
			if totalBytes > maxAssetBytes {
				return nil, fmt.Errorf("kata archive exceeds %d-byte limit", maxAssetBytes)
			}
			if err := writeStreamToFile(filepath.Join(workDir, filepath.FromSlash(name)), 0o600, tr); err != nil {
				return nil, err
			}
		case tar.TypeSymlink:
			if path.IsAbs(hdr.Linkname) {
				return nil, fmt.Errorf("tar symlink %q has absolute target %q", hdr.Name, hdr.Linkname)
			}
			target := path.Clean(path.Join(path.Dir(name), hdr.Linkname))
			if !filepath.IsLocal(filepath.FromSlash(target)) {
				return nil, fmt.Errorf("tar symlink %q target %q escapes the extraction directory", hdr.Name, hdr.Linkname)
			}
			links[name] = target
		}
	}
}

func writeStreamToFile(dstPath string, mode fs.FileMode, r io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	n, err := io.Copy(f, io.LimitReader(r, maxAssetBytes+1))
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if n > maxAssetBytes {
		return fmt.Errorf("extracted file exceeds %d-byte limit", maxAssetBytes)
	}
	return closeErr
}

func downloadToFile(ctx context.Context, url, dstPath string, mode fs.FileMode) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("downloading %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading %s: unexpected HTTP status %s", url, resp.Status)
	}
	return writeStreamToFile(dstPath, mode, resp.Body)
}
