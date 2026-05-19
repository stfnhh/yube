# syntax=docker/dockerfile:1.7

FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

WORKDIR /src

RUN apk add --no-cache \
    binutils \
    ca-certificates \
    upx

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 \
    GOOS=${TARGETOS:-linux} \
    GOARCH=${TARGETARCH:-amd64} \
    GOARM=${TARGETVARIANT#v} \
    go build \
    -buildvcs=false \
    -trimpath \
    -tags="netgo osusergo" \
    -ldflags="-s -w -buildid=" \
    -o /yube \
    ./cmd/server \
    && strip /yube \
    && upx --best --lzma /yube


FROM scratch

WORKDIR /app

COPY --from=build /yube /app/yube
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY templates /app/templates
COPY static /app/static

ENV ADDR=:8080
ENV DB_PATH=/data/yube.db

EXPOSE 8080

VOLUME ["/data"]

ENTRYPOINT ["/app/yube"]
