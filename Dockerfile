FROM golang:1.14-alpine3.11 AS builder

RUN apk update && apk add build-base make git

WORKDIR /go/src/paxos
COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download && go mod verify

COPY . .

RUN make build

FROM alpine:3.11

COPY --from=builder /go/src/paxos/paxosc /usr/bin/paxosc
COPY --from=builder /go/src/paxos/paxosd /usr/bin/paxosd

USER 10001:10001
