package main

import (
	"context"
	"database/sql"
	"log"
	"net"

	"ecommerce/order-service/internal/delivery"
	"ecommerce/order-service/internal/repository"
	"ecommerce/order-service/internal/usecase"
	"ecommerce/order-service/internal/worker"
	"ecommerce/order-service/pkg/tracing"
	"ecommerce/pb"

	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"

	// --- NEW MIGRATION IMPORTS ---
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func runDBMigrations(db *sql.DB) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Could not create postgres driver for migration: %v", err)
	}

	// Tell it to look in the "db/migrations" folder
	m, err := migrate.NewWithDatabaseInstance(
		"file://order-service/db/migrations",
		"postgres", driver)
	if err != nil {
		log.Fatalf("Could not initialize migrate instance: %v", err)
	}

	// Run the UP migrations
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Could not run up migrations: %v", err)
	}

	log.Println("Database migrations applied successfully!")
}


func main() {
	// 1. Postgres
	dsn := "host=localhost port=5433 user=postgres password=postgres dbname=order_db sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer db.Close()

	runDBMigrations(db)

	// 2. Kafka Publisher (Async)
	kw := &kafka.Writer{
		Addr:         kafka.TCP("127.0.0.1:9092"),
		Balancer:     &kafka.LeastBytes{},
		Async:        true,
	}

	// 3. Kafka Consumer (Listening to multiple topics using GroupID)
	kr := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{"127.0.0.1:9092"},
		GroupTopics: []string{"payment-events", "inventory-events"}, // Listen to both!
		GroupID:     "order-saga-coordinator",
	})

	// 4. Wiring
	repo := repository.NewPostgresOrderRepo(db)
	pub := repository.NewKafkaPublisher(kw)
	uc := usecase.NewOrderUseCase(repo, pub)
	
	// 5. Start Kafka Consumer in background
	consumer := worker.NewSagaConsumer(kr, uc)
	go consumer.Start(context.Background())

	// 1. Initialize Tracing
	shutdown := tracing.InitTracer("order-service") // Change name for each service!
	defer shutdown(context.Background())

	// 6. Start gRPC Server (for API Gateway)
	// You will need to create the delivery/grpc_handler.go file just like the other services!
	grpcHandler := delivery.NewOrderGrpcHandler(uc) 
	
	lis, _ := net.Listen("tcp", ":9004")
	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	pb.RegisterOrderServiceServer(grpcServer, grpcHandler)
	
	log.Println("Order Service running on port :9004")
	grpcServer.Serve(lis)
}