FROM golang:1.16-alpine

ENV GOPATH=/go
ENV GOCACHE=/gocache
ENV GOROOT=/usr/local/go
ENV CGO_ENABLED=0
ENV E2E_PATH=$GOROOT/src/e2e

COPY ./examples /tmp/examples

RUN mkdir -p $E2E_PATH \
    && cd /tmp/examples \
    && (for d in */; do mv /tmp/examples/$d/e2e $E2E_PATH/$d; done)

WORKDIR ${E2E_PATH}

ENTRYPOINT ["go", "test", "./...", "-test.timeout", "0"]
