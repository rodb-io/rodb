FROM python:3.9.7-alpine3.14

RUN pip install json-schema-for-humans==0.34.2

COPY ./build/docker/schema-doc/config.yaml /var/jsfh/config.yaml

ENTRYPOINT ["/usr/local/bin/generate-schema-doc", "--config-file=/var/jsfh/config.yaml"]
