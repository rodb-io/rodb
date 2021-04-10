FROM alpine:3.13 as data

COPY ./JMdict.gz /

WORKDIR /
RUN apk add gzip
RUN gzip -d JMdict.gz

FROM localhost/rods:latest

COPY ./config.yaml /
COPY --from=data /JMdict /JMdict.xml

ENTRYPOINT ["/rods", "--config=/config.yaml", "--verbose"]
