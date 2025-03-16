FROM ghcr.io/zelenin/tdlib-docker:b498497-alpine AS tdlib

FROM golang:1.24.1-alpine3.21 AS go-builder

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


FROM alpine:3.21.3

ENV LANG=en_US.UTF-8
ENV TZ=UTC

RUN apk upgrade --no-cache && \
    apk add --no-cache \
            ca-certificates \
            libstdc++

WORKDIR /app

COPY --from=go-builder /src/app .

CMD ["./app"]
