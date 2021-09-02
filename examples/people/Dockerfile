ARG BASE_IMAGE

FROM golang:1.16-alpine as generator

ENV GOPATH=/go
ENV GOCACHE=/gocache
ENV GOROOT=/usr/local/go
ENV CGO_ENABLED=1
ENV GP_PATH=$GOROOT/src/generate_people

COPY ./generate_people.go ${GP_PATH}/generate_people.go
COPY ./go.mod ${GP_PATH}/go.mod
COPY ./go.sum ${GP_PATH}/go.sum

WORKDIR ${GP_PATH}

RUN go mod download

RUN go run generate_people.go

FROM $BASE_IMAGE

COPY ./config.yaml /etc/rodb.yaml
COPY --from=generator /people.json /srv/people.json

CMD ["--config=/etc/rodb.yaml", "--verbose"]
