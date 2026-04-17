package sender

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/jethikaviduruwan/sentinel/proto/gen"
)

// Sender handles the gRPC connection to HQ
type Sender struct {
	conn   *grpc.ClientConn
	client pb.MetricsServiceClient
}

// New creates a new Sender with reconnect logic
func New(hqAddress string) (*Sender, error) {
	var conn *grpc.ClientConn
	var err error

	// Retry connecting up to 5 times with backoff
	for i := 1; i <= 5; i++ {
		conn, err = grpc.Dial(hqAddress,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err == nil {
			break
		}
		log.Printf("[Agent] connection attempt %d failed: %v — retrying in %ds", i, err, i*2)
		time.Sleep(time.Duration(i*2) * time.Second)
	}

	if err != nil {
		return nil, err
	}

	return &Sender{
		conn:   conn,
		client: pb.NewMetricsServiceClient(conn),
	}, nil
}

// Stream opens a stream and sends one payload
func (s *Sender) Send(ctx context.Context, payload *pb.MetricPayload) error {
	stream, err := s.client.StreamMetrics(ctx)
	if err != nil {
		return err
	}

	if err := stream.Send(payload); err != nil {
		return err
	}

	_, err = stream.CloseAndRecv()
	return err
}

// Close cleans up the connection
func (s *Sender) Close() {
	s.conn.Close()
}