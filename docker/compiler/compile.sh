#! /usr/bin/env bash

set -e

# These must be exported. Without `export` they are shell-local, `go build` never
# sees them, and the binary is built for the container's own architecture despite
# the amd64 filename. Exporting them lets the container run natively on any host
# and cross-compile to the linux/amd64 target the release images expect.
export GOARCH=amd64
export GOOS=linux
export CGO_ENABLED=0

packr2

go build -o protofact_linux-amd64

chmod "0777" protofact_linux-amd64
