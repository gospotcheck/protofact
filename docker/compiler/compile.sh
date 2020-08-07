#! /usr/bin/env bash

set -e

GOARCH=amd64
GOOS=linux

packr2

go build -o protofact_linux-amd64

chmod "0777" protofact_linux-amd64
