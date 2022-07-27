#!/usr/bin/env bash

set -euo pipefail

cd $(dirname $0)/..

source hack/lib.sh

CONTAINERIZE_IMAGE=golang:1.18.4 containerize ./hack/gen-client-mocks.sh

go run github.com/vektra/mockery/v2@latest  --dir=pkg/api/ --name=Client --output=pkg/test/mocks
go run github.com/vektra/mockery/v2@latest  --dir=pkg/kubernetes --name=Kube --output=pkg/test/mocks
