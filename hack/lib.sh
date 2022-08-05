#!/usr/bin/env bash

echodate() {
  # do not use -Is to keep this compatible with macOS
  echo "[$(date +%Y-%m-%dT%H:%M:%S%:z)]" "$@"
}

containerize() {
  local cmd="$1"
  local image="${CONTAINERIZE_IMAGE:-golang:1.18.4}"
  local gocache="${CONTAINERIZE_GOCACHE:-/tmp/.gocache}"
  local gomodcache="${CONTAINERIZE_GOMODCACHE:-/tmp/.gomodcache}"
  local skip="${NO_CONTAINERIZE:-}"

  # short-circuit containerize when in some cases it needs to be avoided
  [ -n "$skip" ] && return

  if ! [ -f /.dockerenv ]; then
    echodate "Running $cmd in a Docker container using $image..."
    mkdir -p "$gocache"
    mkdir -p "$gomodcache"

    exec docker run \
      -v "$PWD":/go/src/pluralsh/gqlclient \
      -v "$gocache":"$gocache" \
      -v "$gomodcache":"$gomodcache" \
      -w /go/src/pluralsh/gqlclient \
      -e "GOCACHE=$gocache" \
      -e "GOMODCACHE=$gomodcache" \
      -u "$(id -u):$(id -g)" \
      --entrypoint="$cmd" \
      --rm \
      -it \
      $image $@

    exit $?
  fi
}