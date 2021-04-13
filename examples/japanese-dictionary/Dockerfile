FROM alpine:3.13 as data

COPY ./JMdict.gz /

WORKDIR /
RUN apk add gzip
RUN gzip -d JMdict.gz

FROM localhost/rodb:latest

COPY ./config.yaml /
COPY --from=data /JMdict /JMdict.xml

ENTRYPOINT ["/rodb", "--config=/config.yaml", "--verbose"]
