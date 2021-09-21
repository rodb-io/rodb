#!/bin/sh

generate () {
    generate-schema-doc \
        --config-file=/var/jsfh/config.yaml \
        /schema/yaml/config.yaml \
        /schema/html/schema.html.tmp

    tidy \
        --show-warnings no \
        -quiet \
        -indent \
        --wrap-attributes no \
        --indent-attributes \
        /schema/html/schema.html.tmp \
    > /schema/html/schema.html

    rm /schema/html/schema.html.tmp
}

# Building a first time at startup
generate

# Watching and rebuilding on change
inotifywait -q -m -r /schema/yaml -e create,delete,modify |
while read events; do
    generate
done
