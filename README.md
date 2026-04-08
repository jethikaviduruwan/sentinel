# Sentinel — Distributed System Monitor

A distributed monitoring system built in Go, consisting of:
- **Agent**: Lightweight daemon that collects system/service metrics and streams to HQ via gRPC
- **HQ**: Central backend that aggregates metrics, stores in PostgreSQL, and exposes a REST API

## Stack
- Language: Go 1.22
- Communication: gRPC (streaming)
- Database: PostgreSQL
- Containerization: Docker + docker-compose

## Status
🚧 In development
