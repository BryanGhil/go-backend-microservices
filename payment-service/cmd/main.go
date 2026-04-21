package main

import (
	"context"
	"database/sql"
	"log"
	"net"

	"ecommerce/payment-service/internal/delivery"
	"ecommerce/payment-service/internal/repository"
	"ecommerce/payment-service/internal/usecase"
	"ecommerce/payment-service/internal/worker"
	"ecommerce/payment-service/pkg/tracing"
	"ecommerce/pb"

	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

func main() {
	// 1. Postgres
	dsn := "host=localhost port=5433 user=postgres password=postgres dbname=ecommerce_db sslmode=disable"
	db, _ := sql.Open("postgres", dsn)

	// CREATE TABLE IF NOT EXISTS payments (id SERIAL PRIMARY KEY, order_id INT, amount NUMERIC, status TEXT);

	// 2. Kafka Publisher (Shouting out Success/Decline)
	kw := &kafka.Writer{
		Addr:     kafka.TCP("127.0.0.1:9092"),
		Balancer: &kafka.LeastBytes{},
		Async:    true,
	}

	// 3. Kafka Consumer (Listening to Inventory ONLY)
	kr := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"127.0.0.1:9092"},
		Topic:   "inventory-events", // We only care about inventory success
		GroupID: "payment-worker-group",
	})

	// 4. Wiring
	repo := repository.NewPostgresPaymentRepo(db)
	pub := repository.NewKafkaPublisher(kw)
	uc := usecase.NewPaymentUseCase(repo)

	// 5. Start Worker
	consumer := worker.NewPaymentConsumer(kr, uc, pub)
	go consumer.Start(context.Background())

	// 6. Start gRPC Server
	grpcHandler := delivery.NewPaymentGrpcHandler(uc)
	lis, _ := net.Listen("tcp", ":9006") // Port 9006

	shutdown := tracing.InitTracer("payment-service")
	defer shutdown(context.Background())

	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	pb.RegisterPaymentServiceServer(grpcServer, grpcHandler)

	log.Println("Payment Service running on port :9006")
	grpcServer.Serve(lis)
}