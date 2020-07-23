FROM alpine:3.12.0

ARG SCALA_VERSION=2.11.12
ARG PROTOFACT_VERSION
ARG ARTIFACTORY_USERNAME
ARG ARTIFACTORY_PASSWORD

ENV NAME protofact
ENV LANGUAGE scala
ENV VERSION $PROTOFACT_VERSION

RUN apk add --update bash

COPY docker/scala/install_scala.sh .

RUN ./install_scala.sh

ENV JAVA_HOME /usr/lib/jvm/java-1.8-openjdk

ENV PATH "$JAVA_HOME/bin:$PATH"

# TODO: put in retrieval of Github release binary

RUN apk del .build-dependencies wget curl

RUN touch config.yaml

RUN chmod +x protofact-$VERSION-linux-x86_64

CMD ./protofact-$VERSION-linux-x86_64
