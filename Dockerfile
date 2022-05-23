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
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o plural -ldflags '-s -w' ./cmd/plural/

FROM gcr.io/pluralsh/golang:1.18.2-alpine3.15
RUN apk update && apk add --no-cache git build-base
WORKDIR /
COPY --from=builder /workspace/plural /go/bin/
RUN /go/bin/plural --help
