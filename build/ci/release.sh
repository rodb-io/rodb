#!/bin/bash

cd examples
for EXAMPLE in *; do
    if [ -d "$EXAMPLE" ]; then
        rm $EXAMPLE.zip || true
        zip -r $EXAMPLE.zip $EXAMPLE
    fi
done
cd -

GIT_TAG="$(./build/ci/git-tag.sh)"
DOCKER_TAGS="$(./build/ci/docker-tags.sh)"

NOTES="$(echo -e 'Available docker images:\n```')"
for DOCKER_TAG in $DOCKER_TAGS; do
    NOTES="$NOTES$(echo -e ""\\n$DOCKER_TAG"")"
done
NOTES="$NOTES$(echo -e '\n```')"

gh release delete $GIT_TAG || true
gh release create $GIT_TAG \
    ./examples/*.zip \
    --title "$GIT_TAG" \
    --notes "$NOTES"
