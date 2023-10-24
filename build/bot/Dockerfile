FROM golang:1.21 AS builder

RUN mkdir /app

WORKDIR /app

COPY . .

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bot ./cmd/bot

FROM alpine:3.18.4

RUN mkdir /app

COPY --from=builder /app/bot /app/

WORKDIR /app

CMD ["./bot"]