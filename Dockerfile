FROM golang:1.27.1
WORKDIR /app
COPY cmd/server ./cmd/server
COPY internal ./internal
COPY go.mod .
COPY go.sum .
RUN go build -o server ./cmd/server
ENTRYPOINT [ "./server" ]