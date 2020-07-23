#! /usr/bin/env bash

apk update && apk upgrade && apk --update add \
    gcc g++ ruby ruby-bigdecimal ruby-json \
    libstdc++ tzdata ca-certificates \
    &&  echo 'gem: --no-document' > /etc/gemrc

mkdir ~/.gem
touch ~/.gem/credentials
chmod 600 ~/.gem/credentials
