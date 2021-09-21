#!/bin/sh

generate () {
    generate-schema-doc \
        --config-file=/var/jsfh/config.yaml \
        /schema/yaml/config.yaml \
        /schema/html/schema.html
}

# Building a first time at startup
generate

# Watching and rebuilding on change
inotifywait -q -m -r /schema/yaml -e create,delete,modify |
while read events; do
    generate
done
