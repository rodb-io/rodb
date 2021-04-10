FROM localhost/rods:latest

COPY ./config.yaml /
COPY ./*.csv /

ENTRYPOINT ["/rods", "--config=/config.yaml", "--verbose"]
