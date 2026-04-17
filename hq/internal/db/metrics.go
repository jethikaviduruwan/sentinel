package db

import (
	"context"
	"fmt"

	pb "github.com/jethikaviduruwan/sentinel/proto/gen"
)

// UpsertServer inserts or updates the server record
func (d *DB) UpsertServer(ctx context.Context, serverID string, timestamp int64) error {
	_, err := d.Pool.Exec(ctx, `
		INSERT INTO servers (id, last_seen, online)
		VALUES ($1, $2, true)
		ON CONFLICT (id) DO UPDATE
		SET last_seen = $2, online = true
	`, serverID, timestamp)
	if err != nil {
		return fmt.Errorf("upsert server: %w", err)
	}
	return nil
}

// SaveSystemMetrics saves a system metric row
func (d *DB) SaveSystemMetrics(ctx context.Context, s *pb.SystemMetrics) error {
	_, err := d.Pool.Exec(ctx, `
		INSERT INTO system_metrics
			(server_id, timestamp, cpu_percent, mem_total, mem_used, mem_free, disk_total, disk_used, disk_free)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`,
		s.ServerId, s.Timestamp, s.CpuPercent,
		s.MemTotal, s.MemUsed, s.MemFree,
		s.DiskTotal, s.DiskUsed, s.DiskFree,
	)
	if err != nil {
		return fmt.Errorf("save system metrics: %w", err)
	}
	return nil
}

// SaveServiceMetrics saves all service metric rows for one payload
func (d *DB) SaveServiceMetrics(ctx context.Context, services []*pb.ServiceMetric) error {
	for _, svc := range services {
		_, err := d.Pool.Exec(ctx, `
			INSERT INTO service_metrics
				(server_id, timestamp, name, running, cpu_percent, mem_rss)
			VALUES ($1,$2,$3,$4,$5,$6)
		`,
			svc.ServerId, svc.Timestamp, svc.Name,
			svc.Running, svc.CpuPercent, svc.MemRss,
		)
		if err != nil {
			return fmt.Errorf("save service metric [%s]: %w", svc.Name, err)
		}
	}
	return nil
}

// GetAllServers returns all servers with their online status
func (d *DB) GetAllServers(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := d.Pool.Query(ctx, `
		SELECT id, last_seen, online FROM servers ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("get all servers: %w", err)
	}
	defer rows.Close()

	var servers []map[string]interface{}
	for rows.Next() {
		var id string
		var lastSeen int64
		var online bool
		if err := rows.Scan(&id, &lastSeen, &online); err != nil {
			return nil, err
		}
		servers = append(servers, map[string]interface{}{
			"id":        id,
			"last_seen": lastSeen,
			"online":    online,
		})
	}
	return servers, nil
}

// GetLatestSystemMetrics returns the latest system metrics for a server
func (d *DB) GetLatestSystemMetrics(ctx context.Context, serverID string) (map[string]interface{}, error) {
	row := d.Pool.QueryRow(ctx, `
		SELECT server_id, timestamp, cpu_percent, mem_total, mem_used, mem_free,
		       disk_total, disk_used, disk_free
		FROM system_metrics
		WHERE server_id = $1
		ORDER BY timestamp DESC
		LIMIT 1
	`, serverID)

	var serverID2 string
	var timestamp int64
	var cpuPercent float64
	var memTotal, memUsed, memFree int64
	var diskTotal, diskUsed, diskFree int64

	err := row.Scan(&serverID2, &timestamp, &cpuPercent,
		&memTotal, &memUsed, &memFree,
		&diskTotal, &diskUsed, &diskFree)
	if err != nil {
		return nil, fmt.Errorf("get latest system metrics: %w", err)
	}

	return map[string]interface{}{
		"server_id":   serverID2,
		"timestamp":   timestamp,
		"cpu_percent": cpuPercent,
		"memory": map[string]interface{}{
			"total_mb": memTotal / 1024 / 1024,
			"used_mb":  memUsed / 1024 / 1024,
			"free_mb":  memFree / 1024 / 1024,
		},
		"disk": map[string]interface{}{
			"total_gb": diskTotal / 1024 / 1024 / 1024,
			"used_gb":  diskUsed / 1024 / 1024 / 1024,
			"free_gb":  diskFree / 1024 / 1024 / 1024,
		},
	}, nil
}

// GetLatestServiceMetrics returns the latest status of each service for a server
func (d *DB) GetLatestServiceMetrics(ctx context.Context, serverID string) ([]map[string]interface{}, error) {
	rows, err := d.Pool.Query(ctx, `
		SELECT DISTINCT ON (name)
			server_id, name, running, cpu_percent, mem_rss, timestamp
		FROM service_metrics
		WHERE server_id = $1
		ORDER BY name, timestamp DESC
	`, serverID)
	if err != nil {
		return nil, fmt.Errorf("get latest service metrics: %w", err)
	}
	defer rows.Close()

	var services []map[string]interface{}
	for rows.Next() {
		var serverID2, name string
		var running bool
		var cpuPercent float64
		var memRSS, timestamp int64

		if err := rows.Scan(&serverID2, &name, &running, &cpuPercent, &memRSS, &timestamp); err != nil {
			return nil, err
		}

		status := "down"
		if running {
			status = "up"
		}

		services = append(services, map[string]interface{}{
			"name":        name,
			"status":      status,
			"cpu_percent": cpuPercent,
			"mem_rss_kb":  memRSS / 1024,
			"timestamp":   timestamp,
		})
	}
	return services, nil
}