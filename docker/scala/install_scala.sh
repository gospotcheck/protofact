#! /usr/bin/env bash

echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories

apk update && apk add sbt openjdk8 gcc curl git wget ca-certificates

SCALA_HOME=/usr/share/scala

apk add --no-cache --virtual=.build-dependencies wget ca-certificates && \
  apk add --no-cache bash && \
  cd "/tmp" && \
  wget "https://downloads.typesafe.com/scala/${CURRENT_SCALA_VERSION}/scala-${CURRENT_SCALA_VERSION}.tgz" && \
  tar xzf "scala-${CURRENT_SCALA_VERSION}.tgz" && \
  mkdir "${SCALA_HOME}" && \
  rm "/tmp/scala-${CURRENT_SCALA_VERSION}/bin/"*.bat && \
  mv "/tmp/scala-${CURRENT_SCALA_VERSION}/bin" "/tmp/scala-${CURRENT_SCALA_VERSION}/lib" "${SCALA_HOME}" && \
  ln -s "${SCALA_HOME}/bin/"* "/usr/bin/" && \
  rm -rf "/tmp/"*
