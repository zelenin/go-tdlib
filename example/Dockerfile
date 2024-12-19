FROM ghcr.io/zelenin/tdlib-docker:22d49d5-alpine AS tdlib

FROM golang:1.23.3-alpine3.20 AS go-builder

ENV LANG=en_US.UTF-8
ENV TZ=UTC

RUN set -eux && \
    apk update && \
    apk upgrade && \
    apk add \
        bash \
        build-base \
        ca-certificates \
        curl \
        git \
        linux-headers \
        openssl-dev \
        zlib-dev

WORKDIR /src

COPY --from=tdlib /usr/local/include/td /usr/local/include/td/
COPY --from=tdlib /usr/local/lib/libtd* /usr/local/lib/
COPY . /src

RUN go build \
    -a \
    -trimpath \
    -ldflags "-s -w" \
    -o app \
    "./demo.go" && \
    ls -lah


FROM alpine:3.20.3

ENV LANG=en_US.UTF-8
ENV TZ=UTC

RUN apk upgrade --no-cache && \
    apk add --no-cache \
            ca-certificates \
            libstdc++

WORKDIR /app

COPY --from=go-builder /src/app .

CMD ["./app"]
