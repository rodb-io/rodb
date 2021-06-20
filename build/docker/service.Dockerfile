# syntax=docker/dockerfile:experimental
FROM golang:1.16-alpine as builder

ENV GOPATH=/go
ENV GOCACHE=/gocache
ENV GOROOT=/usr/local/go
ENV CGO_ENABLED=0
ENV RODB_PACKAGE_NAME=rodb.io
ENV RODB_PATH=$GOROOT/src/${RODB_PACKAGE_NAME}

COPY ./go.mod ${RODB_PATH}/go.mod
COPY ./go.sum ${RODB_PATH}/go.sum

WORKDIR ${RODB_PATH}

RUN go mod download

COPY ./cmd ${RODB_PATH}/cmd
COPY ./pkg ${RODB_PATH}/pkg

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/gocache \
    go mod vendor \
    && go build -v -o /rodb ./cmd/main.go

RUN if [ "$(go fmt ./... | wc -l)" -gt 0 ]; then echo "Invalid code-style. Please run 'go fmt ./...'" && exit 1; fi

RUN go test -timeout 3s ./...

FROM scratch

WORKDIR /

COPY --from=builder /rodb /rodb
COPY ./configs/default.yaml /rodb.yaml

STOPSIGNAL SIGINT

ENTRYPOINT ["/rodb"]
