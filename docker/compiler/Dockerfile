FROM golang:1.14.7-alpine3.12

RUN apk add --update git bash

# Install the packr2 CLI at the same version as the packr/v2 module in go.mod.
# The version must be pinned and resolved in module mode: an unpinned `go get -u`
# runs in GOPATH mode, ignores go.mod, and resolves packr's default branch, which
# now requires a far newer Go than this image provides.
RUN mkdir -p /tmp/packr && cd /tmp/packr && \
  go mod init packrinstall && \
  GO111MODULE=on go get github.com/gobuffalo/packr/v2/packr2@v2.8.0 && \
  rm -rf /tmp/packr
