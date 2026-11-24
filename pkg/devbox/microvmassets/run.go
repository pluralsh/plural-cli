// Package microvmassets stages the sandbox binaries of Agent Substrate's microvm class (cloud-hypervisor, virtiofsd,
// the Kata guest kernel and rootfs) into an object store at microvm-assets/<arch>/<sha256>, so enterprises can
// pre-stage them from a workstation that has egress to GitHub. It mirrors the microvm-assets command shipped in the
// deployment-operator image, which the agent-substrate chart runs in-cluster.
package microvmassets

import (
	"context"
	"fmt"
	"strings"

	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/s3blob"
)

// Options configures Run.
type Options struct {
	// BucketURL is a gocloud.dev blob URL, e.g. s3://bucket?region=us-east-1, gs://bucket or azblob://container.
	BucketURL string
	// SandboxConfig optionally points at a microvm SandboxConfig whose default version supplies the pinned digests;
	// PinnedDigests is used when empty.
	SandboxConfig string
	Arches        []string
	WorkDir       string
	KataVersion   string
	CHVersion     string
	// ReleaseBaseURL replaces https://github.com, for mirrors with the same path layout.
	ReleaseBaseURL string
	// Logf receives progress output; nil discards it.
	Logf func(format string, args ...any)
}

var logf = func(string, ...any) {}

// Run stages the assets for each arch. It is idempotent: objects that already exist are skipped.
func Run(ctx context.Context, o Options) error {
	if o.BucketURL == "" {
		return fmt.Errorf("a bucket URL is required")
	}
	if len(o.Arches) == 0 {
		o.Arches = []string{"amd64"}
	}
	if o.KataVersion == "" {
		o.KataVersion = DefaultKataVersion
	}
	if o.CHVersion == "" {
		o.CHVersion = DefaultCloudHypervisorVersion
	}
	if o.ReleaseBaseURL == "" {
		o.ReleaseBaseURL = DefaultReleaseBaseURL
	}
	if o.Logf != nil {
		logf = o.Logf
	}

	digests := PinnedDigests
	if o.SandboxConfig != "" {
		var err error
		if digests, err = loadDigests(o.SandboxConfig); err != nil {
			return fmt.Errorf("loading pinned digests: %w", err)
		}
	}

	bucket, err := openBucket(ctx, o.BucketURL)
	if err != nil {
		base, _, _ := strings.Cut(o.BucketURL, "?")
		return fmt.Errorf("opening %s: %w", base, err)
	}
	defer bucket.Close()

	workDir := o.WorkDir
	s := &stager{
		bucket:       bucket,
		workDir:      workDir,
		kataVersion:  o.KataVersion,
		chVersion:    o.CHVersion,
		releasesBase: o.ReleaseBaseURL,
	}
	for _, arch := range o.Arches {
		arch = strings.TrimSpace(arch)
		if arch == "" {
			continue
		}
		if err := s.stage(ctx, arch, digests[arch]); err != nil {
			return fmt.Errorf("staging %s assets: %w", arch, err)
		}
	}
	return nil
}
