# syntax = docker/dockerfile:1.3
FROM golang:1.16-alpine as builder

ENV GOPATH=/go
ENV GOCACHE=/gocache
ENV GOROOT=/usr/local/go
ENV CGO_ENABLED=1
ENV RODB_PACKAGE_NAME=github.com/rodb-io/rodb
ENV RODB_PATH=$GOROOT/src/${RODB_PACKAGE_NAME}

RUN apk add --no-cache gcc musl-dev

RUN addgroup -S rodb \
    && adduser -S rodb -G rodb \
    && mkdir -p /scratchfs/bin /scratchfs/etc /scratchfs/srv /scratchfs/var \
    && chmod -R 755 /scratchfs \
    && chown -R root:root /scratchfs \
    && chown -R rodb:rodb /scratchfs/var

COPY ./go.mod ${RODB_PATH}/go.mod
COPY ./go.sum ${RODB_PATH}/go.sum

WORKDIR ${RODB_PATH}

RUN go mod download

COPY ./cmd ${RODB_PATH}/cmd
COPY ./pkg ${RODB_PATH}/pkg

ENV BUILD_TAGS='sqlite_fts5'
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/gocache \
    go mod vendor \
    && go build -v \
        -o /scratchfs/bin/rodb \
        -ldflags '-linkmode external -extldflags "-static"' \
        --tags "$BUILD_TAGS" \
        ./cmd/main.go \
    && chmod 755 /scratchfs/bin/rodb \
    && chown root:root /scratchfs/bin/rodb

RUN if [ "$(go fmt ./... | wc -l)" -gt 0 ]; then echo "Invalid code-style. Please run 'go fmt ./...'" && exit 1; fi

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/gocache \
    go test --tags "$BUILD_TAGS" -timeout 3s ./...

FROM scratch
LABEL org.label-schema.name="RODB"
LABEL org.label-schema.url="https://rodb-io.github.io/rodb/"
LABEL org.label-schema.vcs-url="github.com:rodb-io/rodb.git"

WORKDIR /

COPY --from=builder /scratchfs/bin /bin
COPY --from=builder /scratchfs/etc /etc
COPY --from=builder /scratchfs/srv /srv
COPY --from=builder /scratchfs/var /var
COPY --from=builder /etc/group /etc/group
COPY --from=builder /etc/passwd /etc/passwd

USER rodb

STOPSIGNAL SIGINT

ENTRYPOINT ["/bin/rodb"]
