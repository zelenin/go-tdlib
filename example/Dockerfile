FROM alpine:3.17 as tdlib-builder

ENV LANG en_US.UTF-8
ENV TZ UTC

ARG TD_TAG

RUN apk update && \
    apk upgrade && \
    apk add --update \
        build-base \
        ca-certificates \
        ccache \
        cmake \
        git \
        gperf \
        linux-headers \
        openssl-dev \
        php \
        php-ctype \
        readline-dev \
        zlib-dev && \
    git clone -b "${TD_TAG}" "https://github.com/tdlib/td.git" /src && \
    mkdir /src/build && \
    cd /src/build && \
    cmake \
        -DCMAKE_BUILD_TYPE=Release \
        -DCMAKE_INSTALL_PREFIX:PATH=/usr/local \
        .. && \
    cmake --build . --target prepare_cross_compiling && \
    cd .. && \
    php SplitSource.php && \
    cd build && \
    cmake --build . --target install && \
    ls -lah /usr/local


FROM golang:1.20.2-alpine3.17 as go-builder

ENV LANG en_US.UTF-8
ENV TZ UTC

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

COPY --from=tdlib-builder /usr/local/include/td /usr/local/include/td/
COPY --from=tdlib-builder /usr/local/lib/libtd* /usr/local/lib/
COPY . /src

RUN go build \
    -a \
    -trimpath \
    -ldflags "-s -w" \
    -o app \
    "./demo.go" && \
    ls -lah


FROM alpine:3.17

ENV LANG en_US.UTF-8
ENV TZ UTC

RUN apk upgrade --no-cache && \
    apk add --no-cache \
            ca-certificates \
            libstdc++

WORKDIR /app

COPY --from=go-builder /src/app .

CMD ["./app"]
