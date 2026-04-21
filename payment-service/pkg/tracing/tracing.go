package tracing

import (
	"context"
	"log"
	// "time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// InitTracer connects to Jaeger and sets up the global OpenTelemetry provider
func InitTracer(serviceName string) func(context.Context) error {
	ctx := context.Background()

	// 1. Connect to Jaeger's OTLP port (4317)
	conn, err := grpc.DialContext(ctx, "localhost:4317", 
		grpc.WithTransportCredentials(insecure.NewCredentials()), 
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("failed to create gRPC connection to collector: %v", err)
	}

	// 2. Create the Exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		log.Fatalf("failed to create trace exporter: %v", err)
	}

	// 3. Define the Service Resource (Tells Jaeger the name of our app)
	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(serviceName)),
	)
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
	}

	// 4. Register the Tracer Provider globally
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // Capture 100% of requests for now
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	// 5. Setup Propagation (How the TraceID is passed in HTTP/gRPC headers)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Return a function to cleanly shut down the tracer when the app stops
	return tracerProvider.Shutdown
}