FROM localhost/rodb:latest

COPY ./config.yaml /etc/rodb.yaml
COPY ./*.csv /var

CMD ["--config=/etc/rodb.yaml", "--verbose"]
