FROM alpine:3.12.0

ARG PROTOFACT_VERSION

ENV NAME protofact
ENV LANGUAGE ruby
ENV VERSION $PROTOFACT_VERSION

RUN apk add --update bash curl git

COPY docker/ruby/install_ruby.sh .

RUN ./install_ruby.sh

# TODO: put in retrieval of Github release binary
RUN wget https://github.com/gospotcheck/protofact/releases/download/${VERSION}/protofact_linux-amd64.tar.gz && \
  tar -xvf protofact_linux-amd64 && \
  rm protofact_linux-amd64.tar.gz

RUN apk del .build-dependencies wget curl

RUN touch config.yaml

RUN chmod +x protofact_linux-amd64

CMD ./protofact_linux-amd64
