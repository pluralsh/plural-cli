FROM ubuntu:22.10 as user

# Create a nonroot user for final image
RUN useradd -u 10001 nonroot

FROM gcr.io/pluralsh/golang:1.18.2-alpine3.15 AS builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/ cmd/
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 OUTFILE=plural make build-cli

FROM gcr.io/pluralsh/golang:1.18.2-alpine3.15

WORKDIR /

RUN apk update && apk add --no-cache git build-base

COPY --from=builder /workspace/plural /go/bin/

# Copy nonroot user and switch to it
COPY --from=user /etc/passwd /etc/passwd
USER nonroot

RUN /go/bin/plural --help
