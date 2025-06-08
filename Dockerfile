FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod .
RUN go mod download

COPY wol_server.go ./
RUN go build -o wol-server

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/wol-server .

EXPOSE 8330

CMD ["./wol-server"]
