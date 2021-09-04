#!/bin/bash

IMAGE="$(./build/ci/docker-image.sh)"
TAGS="$(./build/ci/docker-tags.sh)"

TAG_ARGUMENTS=""
for TAG in $TAGS; do
    TAG_ARGUMENTS="$TAG_ARGUMENTS -t $TAG"
done

echo "$GITHUB_TOKEN" | docker login ghcr.io -u $GITHUB_USER --password-stdin

DOCKER_BUILDKIT=1 docker build \
    $TAG_ARGUMENTS \
    -f ./build/docker/service.Dockerfile \
    .

docker image push --all-tags $IMAGE
