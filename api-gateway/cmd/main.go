package main

import (
	"context"
	"ecommerce/api-gateway/internal/handler"
	"ecommerce/api-gateway/internal/middleware"
	"ecommerce/pb"
	"log"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "ecommerce/api-gateway/docs"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"

	"ecommerce/api-gateway/pkg/tracing"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	"github.com/prometheus/client_golang/prometheus/promhttp"
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

// --- GLOBAL SWAGGER CONFIGURATION ---
// @title E-Commerce Microservices API
// @version 1.0
// @description API Gateway for the Go gRPC E-Commerce system.
// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	shutdown := tracing.InitTracer("api-gateway")
	defer shutdown(context.Background())
	
	r := gin.Default()

	r.Use(middleware.MetricsMiddleware())

	r.Use(otelgin.Middleware("api-gateway"))

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Your Docker Redis port
	})
	
	// Create the Redis-backed Rate Limiter
	limiter := redis_rate.NewLimiter(redisClient)

	// 1. Establish gRPC Connections
	connUser := dial("localhost:9002")
	connProduct := dial("localhost:9001")
	connSearch := dial("localhost:9003")
	connOrder := dial("localhost:9004")
	connInventory := dial("localhost:9005")
	connPayment := dial("localhost:9006")

	// 2. Initialize Handlers
	userH := handler.NewUserHandler(pb.NewUserServiceClient(connUser))
	productH := handler.NewProductHandler(pb.NewProductServiceClient(connProduct))
	searchH := handler.NewSearchHandler(pb.NewSearchServiceClient(connSearch))
	orderH := handler.NewOrderHandler(pb.NewOrderServiceClient(connOrder))
	inventoryH := handler.NewInventoryHandler(pb.NewInventoryServiceClient(connInventory))
	paymentH := handler.NewPaymentHandler(pb.NewPaymentServiceClient(connPayment))

	// 3. Define Route Groups
	api := r.Group("/api")
	api.Use(middleware.RateLimitMiddleware(limiter))
	
	// Protected Group (Requires Auth)
	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware())

	// 4. Register All Routes
	userH.RegisterRoutes(api, protected)
	productH.RegisterRoutes(api, protected)
	searchH.RegisterRoutes(api)
	orderH.RegisterRoutes(api)
	inventoryH.RegisterRoutes(api)
	paymentH.RegisterRoutes(api)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	log.Println("API Gateway starting on :8080")
	log.Println("Swagger UI available at http://localhost:8080/swagger/index.html")
	r.Run(":8080")
}