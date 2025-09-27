FROM golang:1-alpine AS build

ADD . /src

WORKDIR /src

RUN apk add --no-cache git && CGO_ENABLED=0 go build

FROM debian:testing-slim

ARG INTERACTIVE_HTML_BOM_VERSION=2.10.0

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        wget \
        pngquant \
        jpegoptim \
        kicad \
        kicad-libraries \
        kicad-footprints \
        kicad-packages3d \
    && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    mkdir -p /opt/ibom && \
    wget \
        --output-document - \
        --quiet \
        "https://github.com/openscopeproject/InteractiveHtmlBom/archive/refs/tags/v${INTERACTIVE_HTML_BOM_VERSION}.tar.gz" \
    | \
    tar \
        --extract \
        --gzip \
        --verbose \
        --file - \
        --strip-components 1 \
        --directory /opt/ibom/

COPY --from=build /src/website /bin/website

COPY kicad/docker-entrypoint.sh /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]
