FROM alpine:3.12.0

ARG CURRENT_SCALA_VERSION=2.12.10
ARG LEGACY_SCALA_VERSION=2.11.12
ARG PROTOFACT_VERSION

ENV NAME protofact
ENV LANGUAGE scala

RUN apk add --update bash

COPY docker/scala/install_scala.sh .
RUN ./install_scala.sh
RUN rm ./install_scala.sh

ENV JAVA_HOME /usr/lib/jvm/java-1.8-openjdk
ENV SCALA_HOME /usr/share/scala
ENV SBT_HOME /usr/share/sbt

ENV PATH "$JAVA_HOME/bin:$SBT_HOME/bin:$PATH"

RUN wget https://github.com/gospotcheck/protofact/releases/download/${PROTOFACT_VERSION}/protofact_linux-amd64.tar.gz && \
  tar -xvf protofact_linux-amd64.tar.gz && \
  rm protofact_linux-amd64.tar.gz

RUN apk del wget curl

RUN touch config.yaml

RUN chmod +x protofact_linux-amd64 && mv protofact_linux-amd64 /usr/bin/protofact

CMD ["protofact"]
