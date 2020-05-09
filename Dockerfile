ARG GO_VERSION=1.14
ARG ALPINE_VERSION=3.11

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

RUN apk update && apk add build-base make git

WORKDIR /go/src/paxos
COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download && go mod verify

COPY . .

RUN make build

FROM alpine:${ALPINE_VERSION}

COPY --from=builder /go/src/paxos/paxosc /usr/bin/paxosc
COPY --from=builder /go/src/paxos/paxosd /usr/bin/paxosd

USER 10001:10001
