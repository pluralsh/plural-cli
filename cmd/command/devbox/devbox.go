package devbox

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/urfave/cli"

	"github.com/pluralsh/plural-cli/pkg/devbox/microvmassets"
	"github.com/pluralsh/plural-cli/pkg/utils"
)

func Command() cli.Command {
	return cli.Command{
		Name:        "devbox",
		Usage:       "tooling for plural devboxes",
		Subcommands: commands(),
		Category:    "Workspace",
	}
}

func commands() []cli.Command {
	return []cli.Command{
		{
			Name:  "stage",
			Usage: "stages the microvm sandbox binaries into an object store",
			Description: "Downloads cloud-hypervisor, virtiofsd and the Kata guest kernel and rootfs, verifies them against the " +
				"digests pinned by the agent-substrate chart, and uploads them to microvm-assets/<arch>/<sha256> in the bucket. " +
				"Run it from a workstation with egress to GitHub and write access to the bucket; it is idempotent.\n\n" +
				"Auth uses each cloud's standard credential chain: for s3:// the AWS SDK default chain (env vars, shared config and " +
				"profiles, SSO, assume-role, instance roles), for gs:// application default credentials, falling back to your active `gcloud auth login`, for azblob:// " +
				"DefaultAzureCredential. Bucket URLs accept provider options, e.g. s3://bucket?region=us-east-2&profile=prod.\n\n" +
				"Example: plural devbox stage --bucket-url 's3://my-assets?region=us-east-2'",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "bucket-url",
					Usage: "blob URL of the asset bucket, e.g. s3://bucket?region=us-east-1, gs://bucket or azblob://container",
				},
				cli.StringSliceFlag{
					Name:  "arch",
					Usage: "architecture to stage (amd64, arm64); repeatable (default: amd64)",
				},
				cli.StringFlag{
					Name:  "sandboxconfig",
					Usage: "microvm SandboxConfig `FILE` to read pinned digests from, instead of the built-in ones",
				},
				cli.StringFlag{
					Name:  "work-dir",
					Usage: "scratch directory, needs ~1GiB free",
				},
				cli.StringFlag{
					Name:  "kata-version",
					Usage: "Kata Containers release to take the guest kernel, rootfs and virtiofsd from",
					Value: microvmassets.DefaultKataVersion,
				},
				cli.StringFlag{
					Name:  "cloud-hypervisor-version",
					Usage: "Cloud Hypervisor release",
					Value: microvmassets.DefaultCloudHypervisorVersion,
				},
				cli.StringFlag{
					Name:  "release-base-url",
					Usage: "where GitHub releases are downloaded from, for internal mirrors with the same path layout",
					Value: microvmassets.DefaultReleaseBaseURL,
				},
				cli.DurationFlag{
					Name:  "timeout",
					Usage: "overall deadline",
					Value: 30 * time.Minute,
				},
			},
			Action: handleStageMicrovmAssets,
		},
	}
}

func handleStageMicrovmAssets(c *cli.Context) error {
	bucketURL := c.String("bucket-url")
	if bucketURL == "" {
		return fmt.Errorf("--bucket-url is required")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	ctx, cancelTimeout := context.WithTimeout(ctx, c.Duration("timeout"))
	defer cancelTimeout()

	arches := c.StringSlice("arch")
	if len(arches) == 0 {
		arches = []string{"amd64"}
	}

	err := microvmassets.Run(ctx, microvmassets.Options{
		BucketURL:      bucketURL,
		SandboxConfig:  c.String("sandboxconfig"),
		Arches:         arches,
		WorkDir:        c.String("work-dir"),
		KataVersion:    c.String("kata-version"),
		CHVersion:      c.String("cloud-hypervisor-version"),
		ReleaseBaseURL: c.String("release-base-url"),
		Logf:           func(format string, args ...any) { utils.Highlight(strings.TrimSuffix(format, "\n")+"\n", args...) },
	})
	if err != nil {
		return err
	}
	utils.Success("microvm assets are staged\n")
	return nil
}
