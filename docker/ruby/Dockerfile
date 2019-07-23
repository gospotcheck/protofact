FROM alpine:3.9.2

ARG PROTOFACT_VERSION
ARG ARTIFACTORY_USERNAME
ARG ARTIFACTORY_PASSWORD

ENV NAME protofact
ENV LANGUAGE ruby
ENV VERSION $PROTOFACT_VERSION

RUN apk add --update bash curl git

COPY docker/ruby/install_ruby.sh .

RUN ./install_ruby.sh

# TODO: put in retrieval of Github release binary

RUN touch config.yaml

RUN chmod +x protofact-$VERSION-linux-x86_64

CMD ./protofact-$VERSION-linux-x86_64
