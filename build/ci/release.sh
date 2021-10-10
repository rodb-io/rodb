#!/bin/bash

GIT_TAG="$(./build/ci/git-tag.sh)"
DOCKER_TAGS="$(./build/ci/docker-tags.sh)"
DESCRIPTION="$(./build/ci/git-tag-description.sh)"

cd examples
for EXAMPLE in *; do
    if [ -d "$EXAMPLE" ]; then
        rm $EXAMPLE.zip || true
        sed -i "s/:master/:$GIT_TAG/" $EXAMPLE/docker-compose.yaml
        zip -r $EXAMPLE.zip $EXAMPLE
    fi
done
cd -

NOTES="$(echo -e 'Available docker images:\n```')"
for DOCKER_TAG in $DOCKER_TAGS; do
    NOTES="$NOTES$(echo -e ""\\n$DOCKER_TAG"")"
done
NOTES="$NOTES$(echo -e '\n```')"
NOTES="$NOTES$(echo -e '\nChangelog:\n```')"
NOTES="$NOTES$(echo -e ""\\n$DESCRIPTION"")"
NOTES="$NOTES$(echo -e '\n```')"

echo "$NOTES"

gh release delete $GIT_TAG || true
gh release create $GIT_TAG \
    ./examples/*.zip \
    --title "$GIT_TAG" \
    --notes "$NOTES"
