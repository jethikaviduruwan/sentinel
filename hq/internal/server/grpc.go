package server

import (
	"fmt"
	"io"
	"log"

	pb "github.com/jethikaviduruwan/sentinel/proto/gen"
)

// MetricsServer implements the gRPC MetricsService
type MetricsServer struct {
	pb.UnimplementedMetricsServiceServer
}

// StreamMetrics receives a stream of metric payloads from an Agent
func (s *MetricsServer) StreamMetrics(stream pb.MetricsService_StreamMetricsServer) error {
	for {
		payload, err := stream.Recv()
		if err == io.EOF {
			// Agent closed the stream — send acknowledgement
			return stream.SendAndClose(&pb.Ack{Ok: true})
		}
		if err != nil {
			return fmt.Errorf("error receiving metrics: %w", err)
		}

		// For now just log what we receive
		sys := payload.System
		log.Printf("[HQ] Server: %s | CPU: %.2f%% | MEM: %d MB | DISK: %d GB",
			sys.ServerId,
			sys.CpuPercent,
			sys.MemUsed/1024/1024,
			sys.DiskUsed/1024/1024/1024,
		)

		for _, svc := range payload.Services {
			status := "DOWN"
			if svc.Running {
				status = "UP"
			}
			log.Printf("[HQ]   service: %-12s %s", svc.Name, status)
		}
	}
}