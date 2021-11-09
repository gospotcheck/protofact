#! /usr/bin/env bash

echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories

apk update && apk add sbt openjdk8 gcc curl git wget ca-certificates

SCALA_HOME=/usr/share/scala
CROSS_SCALA_VERSIONS="${CURRENT_SCALA_VERSION} ${LEGACY_SCALA_VERSION}"

for sparkVersion in $CROSS_SCALA_VERSIONS
do
  apk add --no-cache --virtual=.build-dependencies wget ca-certificates && \
    apk add --no-cache bash && \
    cd "/tmp" && \
    wget "https://downloads.typesafe.com/scala/${sparkVersion}/scala-${sparkVersion}.tgz" && \
    tar xzf "scala-${sparkVersion}.tgz" && \
    mkdir "${SCALA_HOME}" && \
    rm "/tmp/scala-${sparkVersion}/bin/"*.bat && \
    mv "/tmp/scala-${sparkVersion}/bin" "/tmp/scala-${sparkVersion}/lib" "${SCALA_HOME}" && \
    ln -s "${SCALA_HOME}/bin/"* "/usr/bin/" && \
    rm -rf "/tmp/"*
done
