#!/bin/bash

cd examples
for EXAMPLE in *; do
    if [ -d "$EXAMPLE" ]; then
        rm $EXAMPLE.zip || true
        zip -r $EXAMPLE.zip $EXAMPLE
    fi
done
cd -

gh release delete examples-release || true
gh release create examples-release ./examples/*.zip --title "Examples" --notes "To download from the website"
