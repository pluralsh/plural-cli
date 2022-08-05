#!/usr/bin/env bash

set -euo pipefail

cd $(dirname $0)/..

source hack/lib.sh

CONTAINERIZE_IMAGE=golang:1.18.4 containerize  ./hack/run-linter.sh

go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run
