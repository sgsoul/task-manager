FROM golang:latest AS BUILDER

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download
COPY main.go main.go

RUN go build -a -o tg-bot .

FROM ubuntu:22.04

RUN apt-get update && apt-get install -y ca-certificates

WORKDIR /app
COPY --from=BUILDER /app/tg-bot .

ENTRYPOINT [ "/app/tg-bot" ]
