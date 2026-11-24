package microvmassets

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gocloud.dev/blob"
	"sigs.k8s.io/yaml"
)

const assetsPrefix = "microvm-assets"

// assetFiles maps each asset key in the SandboxConfig to the file assembled locally.
var assetFiles = map[string]string{
	"cloud-hypervisor": "cloud-hypervisor",
	"virtiofsd":        "virtiofsd",
	"kata-kernel":      "vmlinux",
	"kata-image":       "rootfs.img",
}

// loadDigests reads arch -> asset key -> sha256 from the default version of a SandboxConfig.
func loadDigests(path string) (map[string]map[string]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg struct {
		Spec struct {
			DefaultVersion string `json:"defaultVersion"`
			Versions       []struct {
				Name   string `json:"name"`
				Assets map[string]map[string]struct {
					SHA256 string `json:"sha256"`
				} `json:"assets"`
			} `json:"versions"`
		} `json:"spec"`
	}
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	for _, v := range cfg.Spec.Versions {
		if v.Name != cfg.Spec.DefaultVersion {
			continue
		}
		out := map[string]map[string]string{}
		for arch, assets := range v.Assets {
			out[arch] = map[string]string{}
			for key, a := range assets {
				out[arch][key] = a.SHA256
			}
		}
		return out, nil
	}
	return nil, fmt.Errorf("%s has no version %q", path, cfg.Spec.DefaultVersion)
}

type stager struct {
	bucket       *blob.Bucket
	workDir      string
	kataVersion  string
	chVersion    string
	releasesBase string
}

func objectKey(arch, sha string) string {
	return fmt.Sprintf("%s/%s/%s", assetsPrefix, arch, sha)
}

func (s *stager) stage(ctx context.Context, arch string, digests map[string]string) error {
	if len(digests) == 0 {
		return fmt.Errorf("no pinned digests for %s in the SandboxConfig", arch)
	}
	for key := range assetFiles {
		if digests[key] == "" {
			return fmt.Errorf("no pinned digest for %s/%s", arch, key)
		}
	}

	missing := false
	for key := range assetFiles {
		ok, err := s.bucket.Exists(ctx, objectKey(arch, digests[key]))
		if err != nil {
			return fmt.Errorf("checking %s/%s: %w", arch, key, err)
		}
		if !ok {
			missing = true
		}
	}
	if !missing {
		logf("%s assets already staged; skipping", arch)
		return nil
	}

	if s.workDir == "" {
		s.workDir = os.TempDir()
	}
	if err := os.MkdirAll(s.workDir, 0o755); err != nil {
		return err
	}
	dir, err := os.MkdirTemp(s.workDir, "microvm-assets-"+arch+"-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	if err := assemble(ctx, s.releasesBase, dir, arch, s.kataVersion, s.chVersion); err != nil {
		return err
	}
	for key, file := range assetFiles {
		got, err := fileSHA256(filepath.Join(dir, file))
		if err != nil {
			return err
		}
		if got != digests[key] {
			return fmt.Errorf("%s/%s: sha256 is %s, the SandboxConfig pins %s; the release or --kata-version/--cloud-hypervisor-version do not match the chart", arch, key, got, digests[key])
		}
	}
	for key, file := range assetFiles {
		if err := s.upload(ctx, objectKey(arch, digests[key]), filepath.Join(dir, file)); err != nil {
			return fmt.Errorf("uploading %s/%s: %w", arch, key, err)
		}
	}
	return nil
}

func (s *stager) upload(ctx context.Context, key, path string) error {
	if ok, err := s.bucket.Exists(ctx, key); err != nil {
		return err
	} else if ok {
		logf("%s already exists; skipping", key)
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	logf("uploading %s", key)
	w, err := s.bucket.NewWriter(ctx, key, &blob.WriterOptions{ContentType: "application/octet-stream"})
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, f); err != nil {
		_ = w.Close()
		return err
	}
	return w.Close()
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
