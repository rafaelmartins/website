FROM golang:1-alpine AS build

ADD . /src

WORKDIR /src

RUN apk add --no-cache git && CGO_ENABLED=0 go build

FROM ghcr.io/rafaelmartins/website:kicad-base-latest

COPY --from=build /src/website /bin/website

COPY kicad/docker-entrypoint.sh /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]
