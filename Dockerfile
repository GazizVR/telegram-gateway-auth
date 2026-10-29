# 1. Build
FROM golang:1.27.1 AS builder
WORKDIR /app

COPY cmd/server ./cmd/server
COPY internal ./internal
COPY otpservice ./otpservice
COPY go.mod .
COPY go.sum .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" ./cmd/server

# 2. Run
FROM alpine:3.24.3

COPY --from=builder /app/server /server

ENTRYPOINT [ "./server" ]