package server

import (
	"fmt"
	"io"
	"log"

	"github.com/jethikaviduruwan/sentinel/hq/internal/db"
	pb "github.com/jethikaviduruwan/sentinel/proto/gen"
)

type MetricsServer struct {
	pb.UnimplementedMetricsServiceServer
	DB *db.DB
}

func (s *MetricsServer) StreamMetrics(stream pb.MetricsService_StreamMetricsServer) error {
	ctx := stream.Context()

	for {
		payload, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.Ack{Ok: true})
		}
		if err != nil {
			return fmt.Errorf("error receiving metrics: %w", err)
		}

		sys := payload.System

		// 1. Upsert server record
		if err := s.DB.UpsertServer(ctx, sys.ServerId, sys.Timestamp); err != nil {
			log.Printf("[HQ] db error (upsert server): %v", err)
			continue
		}

		// 2. Save system metrics
		if err := s.DB.SaveSystemMetrics(ctx, sys); err != nil {
			log.Printf("[HQ] db error (system metrics): %v", err)
			continue
		}

		// 3. Save service metrics
		if err := s.DB.SaveServiceMetrics(ctx, payload.Services); err != nil {
			log.Printf("[HQ] db error (service metrics): %v", err)
			continue
		}

		log.Printf("[HQ] saved | server: %s | CPU: %.2f%% | MEM: %dMB",
			sys.ServerId,
			sys.CpuPercent,
			sys.MemUsed/1024/1024,
		)
	}
}