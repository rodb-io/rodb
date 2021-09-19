#!/bin/sh

inotifywait -q -m -r /schema/yaml -e create,delete,modify |
while read events; do
    generate-schema-doc \
        --config-file=/var/jsfh/config.yaml \
        /schema/yaml/config.yaml \
        /schema/html/schema.html
done
