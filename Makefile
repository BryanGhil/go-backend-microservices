.PHONY: run-infra run-services stop-services start

# Starts your databases, Kafka, Jaeger, etc.
run-infra:
	docker-compose up -d

# Runs all your Go services in parallel
run-services:
	@echo "Starting all microservices..."
	@cd api-gateway && go run cmd/main.go &
	@cd product-service && go run cmd/main.go &
	@cd user-service && go run cmd/main.go &
	@cd search-service && go run cmd/main.go &
	@cd order-service && go run cmd/main.go &
	@cd inventory-service && go run cmd/main.go &
	@cd payment-service && go run cmd/main.go &
	@wait

# Kills all running Go processes
stop-services:
	@echo "Stopping all microservices..."
	@pkill -f "go run cmd/main.go" || true
	@echo "All services stopped."

# The ultimate command: starts infra, waits 5 seconds, then starts services
start: run-infra
	@echo "Waiting for infrastructure to boot..."
	@sleep 5
	@make run-services