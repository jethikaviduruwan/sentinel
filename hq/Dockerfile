FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY proto /proto
COPY hq/go.mod hq/go.sum ./
RUN go mod download

COPY hq/ .
RUN go build -o hq ./cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/hq .

EXPOSE 50051 8080
CMD ["./hq"]