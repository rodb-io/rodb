FROM localhost/rodb:latest

COPY ./config.yaml /
COPY ./*.csv /

ENTRYPOINT ["/rodb", "--config=/config.yaml", "--verbose"]
