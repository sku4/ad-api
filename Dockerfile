FROM golang:1.20.5-alpine3.18 AS builder

RUN go version

COPY . /ad-api/
WORKDIR /ad-api/

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

RUN go mod download
RUN go build -o ./.bin/ad-api -tags=go_tarantool_ssl_disable ./cmd/api/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /ad-api/.bin/ad-api .
COPY --from=builder /ad-api/configs/config.yml configs/config.yml
COPY --from=builder /ad-api/web web/
COPY --from=builder /ad-api/templates templates/
RUN touch .env

EXPOSE 8000

CMD /app/ad-api
