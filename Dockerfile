FROM golang:1-alpine AS build

ADD . /src

WORKDIR /src

RUN apk add --no-cache git && CGO_ENABLED=0 go build

FROM debian:testing-slim

RUN apt-get update && \
    apt-get install -y \
        pngquant \
        jpegoptim \
        kicad \
        kicad-libraries \
        kicad-footprints \
        kicad-packages3d \
    && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

COPY --from=build /src/website /bin/website

COPY docker-entrypoint.sh /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]
