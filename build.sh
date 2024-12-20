#!/bin/sh

set -xe

BIN=bin
FILE=reverse-proxy

if [ ! -d "$BIN" ]; then
  rm -rf "$BIN"
fi

mkdir -p "$BIN"

go build -o ${BIN}/${FILE} ${FILE}.go
