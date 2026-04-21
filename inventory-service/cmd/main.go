package main

import (
	"context"
	"database/sql"
	"log"
	"net"

	"ecommerce/inventory-service/internal/delivery"
	"ecommerce/inventory-service/internal/repository"
	"ecommerce/inventory-service/internal/usecase"
	"ecommerce/inventory-service/internal/worker"
	"ecommerce/inventory-service/tracing"
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

	// CREATE TABLE IF NOT EXISTS inventory (product_id INT PRIMARY KEY, stock INT);

	// 2. Kafka Publisher
	kw := &kafka.Writer{
		Addr:     kafka.TCP("127.0.0.1:9092"),
		Balancer: &kafka.LeastBytes{},
		Async:    true,
	}

	// 3. Kafka Consumer (Listening to Product, Order, and Payment events!)
	kr := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{"127.0.0.1:9092"},
		GroupTopics: []string{"product-events", "order-events", "payment-events"},
		GroupID:     "inventory-worker-group",
	})

	// 4. Wiring
	repo := repository.NewPostgresInventoryRepo(db)
	pub := repository.NewKafkaPublisher(kw)
	uc := usecase.NewInventoryUseCase(repo) // UC for the gRPC handlers

	// 5. Start Worker in background
	consumer := worker.NewInventoryConsumer(kr, repo, pub)
	go consumer.Start(context.Background())

	// 6. Start gRPC Server
	grpcHandler := delivery.NewInventoryGrpcHandler(uc) 
	lis, _ := net.Listen("tcp", ":9005") // Port 9005

	shutdown := tracing.InitTracer("inventory-service")
	defer shutdown(context.Background())

	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	pb.RegisterInventoryServiceServer(grpcServer, grpcHandler)
	
	log.Println("Inventory Service running on port :9005")
	grpcServer.Serve(lis)
}