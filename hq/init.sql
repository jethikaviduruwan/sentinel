CREATE TABLE IF NOT EXISTS servers (
    id          TEXT PRIMARY KEY,
    last_seen   BIGINT NOT NULL,
    online      BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS system_metrics (
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

CREATE TABLE IF NOT EXISTS service_metrics (
    id          BIGSERIAL PRIMARY KEY,
    server_id   TEXT NOT NULL REFERENCES servers(id),
    timestamp   BIGINT NOT NULL,
    name        TEXT NOT NULL,
    running     BOOLEAN NOT NULL,
    cpu_percent DOUBLE PRECISION NOT NULL,
    mem_rss     BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_system_metrics_server_time ON system_metrics(server_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_service_metrics_server_time ON service_metrics(server_id, timestamp DESC);