FROM golang:1-alpine AS build

ADD . /src

WORKDIR /src

RUN apk add --no-cache git && CGO_ENABLED=0 go build ./cmd/website

FROM alpine:edge

RUN apk add --no-cache coreutils pngquant jpegoptim

COPY --from=build /src/website /bin/website

COPY cmd/website/docker-entrypoint.sh /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]
