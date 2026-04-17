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