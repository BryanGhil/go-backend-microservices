package main

import (
	"context"
	"log"
	"net"

	"ecommerce/cart-service/internal/delivery"
	"ecommerce/cart-service/internal/repository"
	"ecommerce/cart-service/internal/usecase"
	"ecommerce/cart-service/pkg/tracing" // Ensure you have your tracing package here
	"ecommerce/pb"

	"github.com/go-redis/redis/v8" // FIX 1: Matched the v8 import from your repository
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Helper function to connect to other microservices
func dial(addr string) *grpc.ClientConn {
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		log.Fatalf("could not connect to %s: %v", addr, err)
	}
	return conn
}

func main() {
	// 1. CONNECT TO REDIS
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set by default in docker
		DB:       0,  // use default DB
	})
	defer redisClient.Close()

	// Verify Redis connection
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	log.Println("Connected to Redis successfully!")

	// 2. CONNECT TO PRODUCT SERVICE (Client)
	// Assuming Product Service is running on port 9001
	connProduct := dial("localhost:9001")
	defer connProduct.Close()
	productClient := pb.NewProductServiceClient(connProduct)

	// 3. WIRE UP CLEAN ARCHITECTURE
	repo := repository.NewRedisCartRepository(redisClient)
	
	// FIX 2: Pass the Product gRPC Client into the UseCase!
	uc := usecase.NewCartUseCase(repo, productClient)
	
	handler := delivery.NewCartGrpcHandler(uc)

	// 4. START CART GRPC SERVER
	lis, err := net.Listen("tcp", ":9007")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Initialize Jaeger Tracing
	shutdown := tracing.InitTracer("cart-service")
	defer shutdown(context.Background())

	// Create gRPC server with tracing interceptor
	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	
	// Register the handler with the gRPC server
	pb.RegisterCartServiceServer(grpcServer, handler)

	log.Println("Cart Service (Redis) running on port :9007")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}