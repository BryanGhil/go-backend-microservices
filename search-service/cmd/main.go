package main

import (
	"context"
	"ecommerce/pb"
	"ecommerce/search-service/internal/delivery"
	"ecommerce/search-service/internal/repository"
	"ecommerce/search-service/internal/worker"
	"ecommerce/search-service/pkg/tracing"
	"log"
	"net"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

func main() {
	// 1. Connect to Elasticsearch
	esClient, err := elasticsearch.NewDefaultClient() // Defaults to localhost:9200
	if err != nil {
		log.Fatalf("Error creating elastic client: %v", err)
	}

	// 2. Connect to Kafka (Consumer)
	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "product-events",
		GroupID: "search-service-group", // Important: Keeps track of what messages this group has read
	})
	defer kafkaReader.Close()

	// 3. Wire Architecture
	repo := repository.NewElasticsearchRepo(esClient)
	consumer := worker.NewKafkaConsumer(kafkaReader, repo)

	// 4. Start the infinite consumer loop
	go consumer.Start(context.Background())

	shutdown := tracing.InitTracer("search-service")
	defer shutdown(context.Background())

	lis, _ := net.Listen("tcp", ":9003") // Port 9003 for Search
	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	// 1. Create the handler using the public constructor
	searchHandler := delivery.NewSearchGrpcHandler(repo)

	// 2. Register the handler
	pb.RegisterSearchServiceServer(grpcServer, searchHandler)
	grpcServer.Serve(lis)
}
