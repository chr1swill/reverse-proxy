#!/bin/sh

set -xe

BIN=bin

if [ ! -d "$BIN" ]; then
  mkdir -p "$BIN"
fi

go build -o ${BIN}/main main.go
