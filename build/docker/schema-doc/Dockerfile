FROM python:3.9.7-alpine3.14

RUN pip install json-schema-for-humans==0.34.2

RUN apk --no-cache add inotify-tools

COPY ./build/docker/schema-doc/templates /usr/local/lib/python3.9/site-packages/json_schema_for_humans/templates
COPY ./build/docker/schema-doc/config.yaml /var/jsfh/config.yaml
COPY ./build/docker/schema-doc/watch.sh /watch.sh

ENTRYPOINT ["/watch.sh"]
