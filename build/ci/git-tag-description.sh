#!/bin/bash

TAG=$(./build/ci/git-tag.sh)

echo "$(git tag -l --format='%(subject)' $TAG)"
