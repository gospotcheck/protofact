FROM alpine:3.12.0

ARG PROTOFACT_VERSION

ENV NAME protofact
ENV LANGUAGE ruby

RUN apk add --update bash curl git

COPY docker/ruby/install_ruby.sh .

RUN ./install_ruby.sh

RUN wget https://github.com/gospotcheck/protofact/releases/download/${PROTOFACT_VERSION}/protofact_linux-amd64.tar.gz && \
  tar -xvf protofact_linux-amd64.tar.gz && \
  rm protofact_linux-amd64.tar.gz

RUN apk del wget curl

RUN touch config.yaml

RUN chmod +x protofact_linux-amd64 && mv protofact_linux-amd64 /usr/bin/protofact

CMD ["protofact"]
