#!/bin/sh

set -eu

docker run \
    -v ${PWD}:/usr/local/go/src/rods \
    -w /usr/local/go/src/rods \
    -it \
    --rm \
    golang:1.15-alpine \
    go get ./...
