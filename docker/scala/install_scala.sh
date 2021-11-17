#! /usr/bin/env bash

echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories
apk update && apk add openjdk8 gcc curl git wget ca-certificates

# Manually installing sbt b/c apk fails with missing java-jdk constraint error
wget -c "https://github.com/sbt/sbt/releases/download/v1.5.5/sbt-1.5.5.tgz" -O - | tar -xz -C /usr/share/
