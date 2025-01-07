FROM ubuntu:kinetic-20230624 AS user

# Create a nonroot user for final image
RUN useradd -u 10001 nonroot

FROM golang:1.23-alpine3.19 AS builder

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
ARG APP_VSN
ARG APP_COMMIT
ARG APP_DATE
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build -ldflags '-s -w \
    -X "github.com/pluralsh/plural-cli/pkg/common.Version=${APP_VSN}" \
    -X "github.com/pluralsh/plural-cli/pkg/common.Commit=${APP_COMMIT}" \
    -X "github.com/pluralsh/plural-cli/pkg/common.Date=${APP_DATE}"' \
    -o plural ./cmd/plural

FROM golang:1.23.4-alpine3.20 AS final

WORKDIR /

RUN apk update && apk add --no-cache git build-base

# Copy nonroot user and switch to it
COPY --from=user /etc/passwd /etc/passwd
USER nonroot

COPY --chown=nonroot --from=builder /workspace/plural /go/bin/
RUN chmod a+x /go/bin/plural

ENTRYPOINT ["/go/bin/plural"]
