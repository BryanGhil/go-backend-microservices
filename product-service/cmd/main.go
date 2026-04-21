package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"time"

	"ecommerce/pb"
	"ecommerce/product-service/internal/delivery"
	"ecommerce/product-service/internal/repository"
	"ecommerce/product-service/internal/usecase"
	"ecommerce/product-service/pkg/tracing"

	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

func main() {
	// 1. PostgreSQL Connection String
	// Format: "host=localhost port=5432 user=postgres password=yourpassword dbname=ecommerce sslmode=disable"
	dsn := "host=localhost port=5433 user=postgres password=postgres dbname=ecommerce_db sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer db.Close()

	// Verify connection
	if err := db.Ping(); err != nil {
		log.Fatalf("postgres ping failed: %v", err)
	}

	kafkaWriter := &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "product-events",
		Balancer: &kafka.LeastBytes{},

		Async: true,
		BatchTimeout: 10 * time.Millisecond,
	}
	defer kafkaWriter.Close()

	// 2. Wire up Clean Architecture (Remains the same!)
	repo := repository.NewPostgresProductRepo(db)
	publisher := repository.NewKafkaPublisher(kafkaWriter)
	uc := usecase.NewProductUseCase(repo, publisher)
	handler := delivery.NewProductGrpcHandler(uc)

	// 3. Start gRPC Server
	lis, err := net.Listen("tcp", ":9001")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	shutdown := tracing.InitTracer("product-service")
	defer shutdown(context.Background())

	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	pb.RegisterProductServiceServer(grpcServer, handler)

	log.Println("Product Service (Postgres) running on port :9001")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
