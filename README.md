# Plural CLI

Deploying your services using the Plural CLI.

## Installation

The Plural CLI is available on homebrew, a single line install can be done with:

```bash
brew install pluralsh/plural/plural
```

If you are using a machine that is not compatible with homebrew,
we recommend simply downloading a pre-built release on github and installing it onto your machines path. The releases can be found here: https://github.com/pluralsh/plural-cli/releases.

## Requirements

Plural does require a few other CLI's to be installed, namely:
* `helm`
* `terraform`
* `kubectl`
* cloud provider CLI for the infrastructure you're deploying to, like `aws`, `az`, `gcloud` etc.

## Quickstart

Detailed instructions can be found at https://docs.plural.sh/deployments/cli-quickstart.

## Reaching Out

If you have any issues with your plural installations, or just want to show us some love, feel free to drop into our discord [here](https://discord.gg/bEBAMXV64s)
## Security remediation note

This repository already pins the Go toolchain and Go module versions requested for the console-mapped remediation:

- Go toolchain: `1.26.4` in `go.mod`
- Container build/test images: `golang:1.26.4` in `Dockerfile`, `test.Dockerfile`, and hack scripts
- `github.com/go-git/go-git/v5`: `v5.19.1`
- `github.com/containerd/containerd`: `v1.7.32`
- `github.com/aws/aws-sdk-go-v2/service/s3`: `v1.97.3`
- `github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream`: `v1.7.8`

No steampipe postgres plugin sources or build definitions were found in this repository, so the affected `steampipe_postgres_{aws,azure,gcp}.so` artifacts appear to be imported from elsewhere. This repo's nearest remediation path is to keep the embedded CLI build inputs pinned to the fixed Go/toolchain and module versions above and document that the plugin artifacts must be rebuilt in their owning source repository or image pipeline.
