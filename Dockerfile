FROM golang:1.16-alpine as build

ENV GOFLAGS="-mod=vendor"
ENV CGO_ENABLED=0

ARG version=unknown

ADD . /build
WORKDIR /build

RUN go build -o /build/plugin-ip2location -ldflags "-X main.version=${version} -s -w"


FROM ghcr.io/umputun/baseimage/app:v1.6.1 as base

FROM scratch

COPY --from=build /build/plugin-ip2location /srv/plugin-ip2location

WORKDIR /srv
ENTRYPOINT ["/srv/plugin-ip2location"]