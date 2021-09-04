#!/bin/bash

IMAGE=$(./build/ci/docker-image.sh)

# Getting the tag or branch name
TAG=$(./build/ci/git-tag.sh)

TAGS=""

# Printing all image tags using the version number variants
while true; do
    TAGS="$TAGS$IMAGE:$TAG "

    if [[ "$TAG" == *"."* ]]; then
        TAG="${TAG%.*}"
    else
        # No more dots to shorten the tag
        break
    fi
done

# Tag using the commit hash
TAGS="$TAGS$IMAGE:$(git rev-parse HEAD)"

echo $TAGS
