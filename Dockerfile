FROM golang:1-alpine AS build

ARG PAGEFIND_VERSION=1.4.0

ADD . /src

WORKDIR /src

RUN apk add --no-cache \
        git \
        wget \
        tar \
    && \
    CGO_ENABLED=0 go build && \
    wget \
        --output-document - \
        --quiet \
        "https://github.com/Pagefind/pagefind/releases/download/v${PAGEFIND_VERSION}/pagefind-v${PAGEFIND_VERSION}-x86_64-unknown-linux-musl.tar.gz" \
    | \
    tar \
        --extract \
        --gzip \
        --verbose \
        --file -

FROM alpine:edge

RUN apk add --no-cache coreutils pngquant jpegoptim

COPY --from=build /src/pagefind /bin/pagefind
COPY --from=build /src/website /bin/website

COPY docker-entrypoint.sh /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]
