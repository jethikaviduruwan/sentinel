FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY proto /proto
COPY agent/go.mod agent/go.sum ./
RUN go mod download

COPY agent/ .
RUN go build -o agent ./cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/agent .
COPY agent/config.yaml .

EXPOSE 50051
CMD ["./agent"]