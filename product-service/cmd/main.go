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
	"ecommerce/product-service/worker"

	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

)

func dial(addr string) *grpc.ClientConn {
	conn, err := grpc.Dial(addr, 
					grpc.WithTransportCredentials(insecure.NewCredentials()), 
					grpc.WithStatsHandler(otelgrpc.NewClientHandler()),)
	if err != nil {
		log.Fatalf("could not connect to %s: %v", addr, err)
	}
	return conn
}

func runDBMigrations(db *sql.DB) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Could not create postgres driver for migration: %v", err)
	}

	// Tell it to look in the "db/migrations" folder
	m, err := migrate.NewWithDatabaseInstance(
		"file://product-service/db/migrations",
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
	// 1. PostgreSQL Connection String
	// Format: "host=localhost port=5432 user=postgres password=yourpassword dbname=ecommerce sslmode=disable"
	dsn := "host=localhost port=5433 user=postgres password=postgres dbname=product_db sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer db.Close()

	runDBMigrations(db)

	// Verify connection
	if err := db.Ping(); err != nil {
		log.Fatalf("postgres ping failed: %v", err)
	}

	kafkaWriter := &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "product-events",
		Balancer: &kafka.LeastBytes{},

		Async:        true,
		BatchTimeout: 10 * time.Millisecond,
	}
	defer kafkaWriter.Close()

	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"localhost:9092"},
		Topic:     "user-events",           // Listening to the User Service!
		GroupID:   "product-service-group", // Tracks which messages this service has read
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
	})
	defer kafkaReader.Close()

	connUser := dial("localhost:9002")
	connInventory := dial("localhost:9005")

	// 2. Wire up Clean Architecture (Remains the same!)
	repo := repository.NewPostgresProductRepo(db)
	publisher := repository.NewKafkaPublisher(kafkaWriter)
	uc := usecase.NewProductUseCase(repo, publisher, pb.NewUserServiceClient(connUser), pb.NewInventoryServiceClient(connInventory))
	handler := delivery.NewProductGrpcHandler(uc)

	userEventConsumer := worker.NewUserEventConsumer(kafkaReader, repo)
	go userEventConsumer.Start(context.Background())

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
