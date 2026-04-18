package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"

	"google.golang.org/grpc"

	"github.com/jethikaviduruwan/sentinel/hq/internal/api"
	"github.com/jethikaviduruwan/sentinel/hq/internal/db"
	"github.com/jethikaviduruwan/sentinel/hq/internal/server"
	pb "github.com/jethikaviduruwan/sentinel/proto/gen"
)

func main() {
	ctx := context.Background()

	// Connect to Postgres
	connStr := os.Getenv("DB_CONN")
	if connStr == "" {
		connStr = "postgres://sentinel:sentinel123@localhost:5432/sentinel"
	}
	database, err := db.New(ctx, connStr)
	if err != nil {
		log.Fatalf("[HQ] failed to connect to database: %v", err)
	}
	defer database.Close()
	log.Println("[HQ] connected to PostgreSQL")

	// Start gRPC server in background
	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("[HQ] failed to listen: %v", err)
		}
		grpcServer := grpc.NewServer()
		pb.RegisterMetricsServiceServer(grpcServer, &server.MetricsServer{DB: database})
		log.Println("[HQ] gRPC server listening on :50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("[HQ] failed to serve gRPC: %v", err)
		}
	}()

	// Start REST API
	router := api.NewRouter(database)
	log.Println("[HQ] REST API listening on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("[HQ] failed to serve REST: %v", err)
	}
}