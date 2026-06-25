FROM golang:1.26-alpine AS builder

ADD . /src/
WORKDIR /src

ARG OS=linux
ARG ARCH=amd64

RUN CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH go build

FROM alpine:3.24

COPY --from=builder /src/version-fmt /version-fmt

WORKDIR /
ENTRYPOINT ["/version-fmt"]
