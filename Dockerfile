FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o qbrdt ./cmd/qbrdt
RUN apk update && apk add dos2unix && dos2unix entrypoint.sh

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/qbrdt /app/qbrdt
COPY --from=builder /app/entrypoint.sh /app/entrypoint.sh

RUN chmod +x /app/entrypoint.sh
RUN mkdir /config

EXPOSE 8080

ENTRYPOINT ["./entrypoint.sh"]