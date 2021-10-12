ARG BASE_IMAGE

FROM $BASE_IMAGE

COPY ./config.yaml /etc/rodb.yaml
COPY ./*.csv /var

CMD ["--config=/etc/rodb.yaml", "--loglevel=debug"]
