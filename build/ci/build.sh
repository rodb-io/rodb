#!/bin/bash

DOCKER_TAG=$(git symbolic-ref -q --short HEAD || git describe --tags --exact-match)
DOCKER_IMAGE="ghcr.io/rodb-io/rodb:$DOCKER_TAG"

echo "$GITHUB_TOKEN" | docker login ghcr.io -u $GITHUB_USER --password-stdin

DOCKER_BUILDKIT=1 docker build \
    -t $DOCKER_IMAGE \
    -f ./build/docker/service.Dockerfile \
    .

docker push $DOCKER_IMAGE
