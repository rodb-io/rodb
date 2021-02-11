# syntax=docker/dockerfile:experimental
FROM golang:1.15-alpine as builder

ENV GOPATH=/go
ENV GOCACHE=/gocache
ENV GOROOT=/usr/local/go
ENV CGO_ENABLED=0
ENV RODS_PACKAGE_NAME=rods
ENV RODS_PATH=$GOROOT/src/${RODS_PACKAGE_NAME}

COPY ./go.mod ${RODS_PATH}/go.mod
COPY ./go.sum ${RODS_PATH}/go.sum

WORKDIR ${RODS_PATH}

RUN go mod download

COPY ./scripts ${RODS_PATH}/scripts
COPY ./cmd ${RODS_PATH}/cmd
COPY ./internal ${RODS_PATH}/internal
COPY ./pkg ${RODS_PATH}/pkg

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/gocache \
    go mod vendor \
    && go build -v -o /rods ./cmd/main.go

RUN if [ "$(go fmt ./... | wc -l)" -gt 0 ]; then echo "Invalid code-style. Please run 'go fmt ./...'" && exit 1; fi

RUN go test ./...

FROM scratch

WORKDIR /

COPY --from=builder /rods /rods
COPY ./configs/default.yaml /rods.yaml

ENTRYPOINT ["/rods"]
