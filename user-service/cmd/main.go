package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"

	"ecommerce/pb"
	"ecommerce/user-service/internal/delivery"
	"ecommerce/user-service/internal/repository"
	"ecommerce/user-service/internal/usecase"
	email "ecommerce/user-service/pkg/sender"
	"ecommerce/user-service/pkg/tracing"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
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
		"file://user-service/db/migrations",
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
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found. Falling back to system environment variables.")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	gmailUser := os.Getenv("GMAIL_USER")
	gmailPass := os.Getenv("GMAIL_APP_PASSWORD")

	emailSender := email.NewGmailSender(gmailUser, gmailPass)

	// --- 1. CONNECT TO POSTGRES ---
	dsn := "host=localhost port=5433 user=postgres password=postgres dbname=user_db sslmode=disable"
	pgDB, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pgDB.Close()

	runDBMigrations(pgDB)

	// --- 2. CONNECT TO REDIS ---
	// Assuming your Docker Redis is running on the default port 6379
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set by default in docker
		DB:       0,  // use default DB
	})
	defer redisClient.Close()

	// --- 3. WIRE UP CLEAN ARCHITECTURE ---

	// A. Initialize Repositories (One for Postgres, One for Redis)
	userRepo := repository.NewPostgresUserRepo(pgDB)           // Ensure you have this struct built in your repo folder!
	sessionRepo := repository.NewRedisSessionRepo(redisClient) // Ensure you have this struct built in your repo folder!

	// B. Initialize UseCase (Pass BOTH repos into the constructor)
	userUC := usecase.NewUserUseCase(userRepo, sessionRepo, jwtSecret, googleClientID, emailSender)

	// C. Initialize Delivery Handler (Pass the UseCase into the gRPC handler)
	grpcHandler := delivery.NewUserGrpcHandler(userUC)

	// --- 4. START GRPC SERVER ---
	lis, err := net.Listen("tcp", ":9002") // Note: Port 9002!
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	shutdown := tracing.InitTracer("user-service")
	defer shutdown(context.Background())

	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	pb.RegisterUserServiceServer(grpcServer, grpcHandler)

	log.Println("User Service (gRPC) running on port :9002")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
