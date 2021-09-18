#!/bin/sh

inotifywait -q -m -r /rodb/configs -e create,delete,modify |
while read events; do
    generate-schema-doc \
        --config-file=/var/jsfh/config.yaml \
        /rodb/configs/schema.yaml \
        /rodb/docs/schema/schema.html
done
