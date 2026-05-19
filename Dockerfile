# syntax=docker/dockerfile:1.7

FROM golang:1.26-alpine AS build

WORKDIR /src

RUN apk add --no-cache \
    ca-certificates

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /tubehive \
    ./cmd/server


FROM scratch

WORKDIR /app

COPY --from=build /tubehive /app/tubehive
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY templates /app/templates
COPY static /app/static

ENV ADDR=:8080
ENV DB_PATH=/data/tubehive.db

EXPOSE 8080

VOLUME ["/data"]

ENTRYPOINT ["/app/tubehive"]