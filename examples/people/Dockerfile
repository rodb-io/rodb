ARG BASE_IMAGE

FROM alpine:3.13 as data

COPY ./people.gz /

WORKDIR /
RUN apk add gzip
RUN gzip -d people.gz

FROM $BASE_IMAGE

COPY ./config.yaml /etc/rodb.yaml
COPY --from=data /people /srv/people.json

CMD ["--config=/etc/rodb.yaml", "--loglevel=debug"]
