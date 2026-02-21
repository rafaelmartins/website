FROM golang:1-alpine AS build

ADD . /src

WORKDIR /src

RUN apk add --no-cache git && CGO_ENABLED=0 go build

FROM ghcr.io/rafaelmartins/website:hardware-base-latest

COPY --from=build /src/website /bin/website

COPY hardware/docker-entrypoint.sh /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]
