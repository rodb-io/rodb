ARG BASE_IMAGE

FROM alpine:3.13 as data

COPY ./JMdict.gz /

WORKDIR /
RUN apk add gzip
RUN gzip -d JMdict.gz

RUN mkdir -p /etc/ssl/certs
RUN apk add openssl
RUN openssl req \
    -x509 \
    -nodes \
    -newkey rsa:2048 \
    -keyout /etc/ssl/certs/rodb.key \
    -out /etc/ssl/certs/rodb.crt \
    -subj "/C=JP/ST=Tokyo/L=Tokyo/O=RODB/OU=RODB/CN=localhost"
RUN chmod a+r-w -R /etc/ssl/certs/rodb.*

FROM $BASE_IMAGE

COPY ./config.yaml /etc/rodb.yaml
COPY --from=data /JMdict /srv/JMdict.xml
COPY --from=data /etc/ssl/certs/rodb.key /etc
COPY --from=data /etc/ssl/certs/rodb.crt /etc

CMD ["--config=/etc/rodb.yaml", "--loglevel=debug"]
