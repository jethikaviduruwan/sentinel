package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/jethikaviduruwan/sentinel/hq/internal/server"
	pb "github.com/jethikaviduruwan/sentinel/proto/gen"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMetricsServiceServer(grpcServer, &server.MetricsServer{})

	log.Println("[HQ] gRPC server listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}