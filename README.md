# Sentinel — Distributed System Monitor

A distributed monitoring system built in Go, consisting of a lightweight **Agent** daemon and a central **HQ** backend service.

## Architecture
```
┌─────────────────────┐         gRPC Stream        ┌─────────────────────┐
│       Agent         │  ─────────────────────────► │        HQ           │
│                     │                             │                     │
│  - CPU metrics      │                             │  - gRPC server      │
│  - Memory metrics   │                             │  - REST API         │
│  - Disk metrics     │                             │  - PostgreSQL       │
│  - Service monitor  │                             │                     │
└─────────────────────┘                             └─────────────────────┘
                                                              │
                                                              ▼
                                                    ┌─────────────────────┐
                                                    │     PostgreSQL      │
                                                    │                     │
                                                    │  - servers          │
                                                    │  - system_metrics   │
                                                    │  - service_metrics  │
                                                    └─────────────────────┘
```

## Project Structure
```
sentinel/
├── agent/                        # Agent daemon
│   ├── cmd/main.go               # Entry point
│   ├── config.yaml               # Agent configuration
│   └── internal/
│       ├── collector/            # Concurrent metric collectors
│       ├── config/               # Config loader (viper)
│       └── sender/               # gRPC streaming client
├── hq/                           # HQ central server
│   ├── cmd/main.go               # Entry point
│   ├── init.sql                  # Database schema
│   └── internal/
│       ├── api/                  # REST API (gin)
│       ├── db/                   # PostgreSQL queries (pgx)
│       └── server/               # gRPC server
├── proto/                        # Protobuf definitions
│   ├── metrics.proto             # gRPC service + message definitions
│   └── gen/                      # Generated Go code
├── docker-compose.yml            # Full stack orchestration
└── postman_collection.json       # Postman API collection
```

## Tech Stack

| Component     | Technology                          |
|---------------|-------------------------------------|
| Language      | Go 1.24                             |
| Communication | gRPC (client-side streaming)        |
| Database      | PostgreSQL 16                       |
| REST API      | Gin                                 |
| Config        | Viper (YAML + env vars)             |
| DB Driver     | pgx/v5                              |
| Metrics       | gopsutil/v3                         |
| Container     | Docker + docker-compose             |

## Database Schema

Three tables are used to store time-series monitoring data:

**`servers`** — tracks each connected agent and its online status.
```sql
CREATE TABLE servers (
    id        TEXT PRIMARY KEY,
    last_seen BIGINT NOT NULL,
    online    BOOLEAN NOT NULL DEFAULT false
);
```

**`system_metrics`** — stores host-level metrics (CPU, memory, disk) per agent per timestamp.
```sql
CREATE TABLE system_metrics (
    id          BIGSERIAL PRIMARY KEY,
    server_id   TEXT NOT NULL REFERENCES servers(id),
    timestamp   BIGINT NOT NULL,
    cpu_percent DOUBLE PRECISION NOT NULL,
    mem_total   BIGINT NOT NULL,
    mem_used    BIGINT NOT NULL,
    mem_free    BIGINT NOT NULL,
    disk_total  BIGINT NOT NULL,
    disk_used   BIGINT NOT NULL,
    disk_free   BIGINT NOT NULL
);
```

**`service_metrics`** — stores per-process metrics (running status, CPU, memory) for each monitored service.
```sql
CREATE TABLE service_metrics (
    id          BIGSERIAL PRIMARY KEY,
    server_id   TEXT NOT NULL REFERENCES servers(id),
    timestamp   BIGINT NOT NULL,
    name        TEXT NOT NULL,
    running     BOOLEAN NOT NULL,
    cpu_percent DOUBLE PRECISION NOT NULL,
    mem_rss     BIGINT NOT NULL
);
```

Indexes on `(server_id, timestamp DESC)` are added to both metric tables for fast latest-value queries.

## How to Run

### Prerequisites
- Docker Desktop
- Docker Compose

### Start the full stack

```bash
git clone https://github.com/jethikaviduruwan/sentinel.git
cd sentinel
docker-compose up --build
```

This starts three containers:
- `sentinel-db` — PostgreSQL with schema auto-initialized
- `sentinel-hq` — gRPC server on `:50051` + REST API on `:8080`
- `sentinel-agent` — begins streaming metrics to HQ immediately

### Test the REST API

```bash
# List all connected servers
curl http://localhost:8080/servers

# Get latest system stats for a server
curl http://localhost:8080/servers/Jethika/stats

# Get monitored services for a server
curl http://localhost:8080/servers/Jethika/services
```

### Configure the Agent

Edit `agent/config.yaml` to change the monitored services or interval:

```yaml
server_id: "Jethika"
hq_address: "localhost:50051"
interval_seconds: 5
services:
  - nginx
  - postgres
  - redis
  - bash
```

Environment variables override config values:
- `HQ_ADDRESS` — overrides `hq_address`
- `DB_CONN` — overrides the HQ database connection string

## REST API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/servers` | List all servers and their online/offline status |
| GET | `/servers/:id/stats` | Latest CPU, memory, and disk metrics for a server |
| GET | `/servers/:id/services` | Monitored services and their current status |

## Running Tests

```bash
# Agent tests
cd agent && go test ./... -v

# HQ tests
cd hq && go test ./... -v
```

## Key Design Decisions

**gRPC client-side streaming** — the Agent holds a persistent stream open to HQ and pushes metrics at a configurable interval. This is more efficient than repeated unary calls and enables real-time updates.

**Concurrent collectors** — the Agent uses goroutines + `sync.WaitGroup` to collect system and service metrics in parallel, minimizing the time it takes to assemble each payload.

**Upsert pattern for servers** — HQ uses `INSERT ... ON CONFLICT DO UPDATE` so servers are automatically registered on first contact without needing a separate registration step.

**Reconnect with backoff** — the Agent retries the gRPC connection up to 5 times with increasing delays, so it recovers gracefully if HQ restarts.

**Environment variable overrides** — both Agent and HQ read config from YAML files but allow environment variables to override key values, making them Docker and Kubernetes friendly.
