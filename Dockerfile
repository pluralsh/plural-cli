FROM ubuntu:22.10 AS user

# Create a nonroot user for final image
RUN useradd -u 10001 nonroot

# golang:1.26.6-alpine3.24, pinned to the published multi-platform index.
FROM golang@sha256:3889b425f035be855a72fb4755265311293b6d414521f0a519d819df32222d83 AS builder

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
    go build -ldflags "-s -w \
    -X \"github.com/pluralsh/plural-cli/pkg/common.Version=${APP_VSN}\" \
    -X \"github.com/pluralsh/plural-cli/pkg/common.Commit=${APP_COMMIT}\" \
    -X \"github.com/pluralsh/plural-cli/pkg/common.Date=${APP_DATE}\"" \
    -o plural ./cmd/plural

FROM golang@sha256:3889b425f035be855a72fb4755265311293b6d414521f0a519d819df32222d83 AS final

WORKDIR /

# Fetch immutable, version-addressed APK artifacts directly instead of
# resolving against the mutable APKINDEX for the v3.24 release branch.
# Keep the existing runtime tools and pin their complete APK closure.
ARG TARGETARCH
RUN case "${TARGETARCH}" in \
        amd64) apk_arch=x86_64 ;; \
        arm64) apk_arch=aarch64 ;; \
        *) echo "unsupported TARGETARCH: ${TARGETARCH}" >&2; exit 1 ;; \
    esac \
    && set -- \
        binutils=2.45.1-r1 brotli-libs=1.2.0-r1 build-base=0.5-r4 \
        c-ares=1.34.8-r0 file=5.47-r2 fortify-headers=3.0.1-r2 \
        g++=15.2.0-r5 gcc=15.2.0-r5 git=2.54.0-r0 \
        git-init-template=2.54.0-r0 gmp=6.3.0-r4 isl26=0.26-r2 \
        jansson=2.15.0-r0 libatomic=15.2.0-r5 libcrypto3=3.5.8-r0 \
        libcurl=8.22.0-r0 libexpat=2.8.4-r0 libgcc=15.2.0-r5 \
        libgcc-static=15.2.0-r5 libgomp=15.2.0-r5 libidn2=2.3.8-r0 \
        libmagic=5.47-r2 libpsl=0.21.5-r3 libssl3=3.5.8-r0 \
        libstdc++=15.2.0-r5 libstdc++-dev=15.2.0-r5 libunistring=1.4.2-r0 \
        make=4.4.1-r4 mpc1=1.3.1-r1 mpfr4=4.2.2-r0 musl-dev=1.2.6-r2 \
        nghttp2-libs=1.69.0-r0 patch=2.8-r0 pcre2=10.48-r0 zstd-libs=1.5.7-r2 \
    && mkdir -p /tmp/apks \
    && for package; do \
        name=${package%%=*}; version=${package#*=}; \
        wget -q -P /tmp/apks "https://dl-cdn.alpinelinux.org/alpine/v3.24/main/${apk_arch}/${name}-${version}.apk"; \
    done \
    && apk add --no-network --no-cache /tmp/apks/*.apk \
    && rm -rf /tmp/apks \
    && apk info -e 'libssl3=3.5.8-r0' \
    && apk info -e 'libcrypto3=3.5.8-r0'

# Copy nonroot user and switch to it
COPY --from=user /etc/passwd /etc/passwd
USER nonroot

COPY --chown=nonroot --from=builder /workspace/plural /go/bin/
RUN chmod a+x /go/bin/plural

ENTRYPOINT ["/go/bin/plural"]
