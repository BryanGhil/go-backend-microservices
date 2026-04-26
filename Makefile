.PHONY: run-infra build-services run-services stop-services start status clean generate-proto

# Starts your databases, Kafka, Jaeger, etc.
run-infra:
	docker-compose up -d

# Compiles all services into a bin/ folder
build-services:
	@echo "Building microservices..."
	@mkdir -p bin
	@go build -o bin/api-gateway ./api-gateway/cmd/main.go
	@go build -o bin/product-service ./product-service/cmd/main.go
	@go build -o bin/user-service ./user-service/cmd/main.go
	@go build -o bin/search-service ./search-service/cmd/main.go
	@go build -o bin/order-service ./order-service/cmd/main.go
	@go build -o bin/inventory-service ./inventory-service/cmd/main.go
	@go build -o bin/payment-service ./payment-service/cmd/main.go
	@echo "All services built successfully!"

# Runs the compiled binaries in the background
run-services: build-services
	@echo "Starting all microservices..."
	@mkdir -p logs
	@./bin/api-gateway > logs/api-gateway.log 2>&1 &
	@./bin/product-service > logs/product-service.log 2>&1 &
	@./bin/user-service > logs/user-service.log 2>&1 &
	@./bin/search-service > logs/search-service.log 2>&1 &
	@./bin/order-service > logs/order-service.log 2>&1 &
	@./bin/inventory-service > logs/inventory-service.log 2>&1 &
	@./bin/payment-service > logs/payment-service.log 2>&1 &
	@echo "All services running! Logs are being written to the /logs directory."
	@wait

# Kills all running custom binaries cleanly without killing VS Code
stop-services:
	@echo "Stopping all microservices..."
	@pkill -f "[.]/bin/" || true
	@echo "All services stopped."

# The ultimate command: starts infra, waits, then builds and starts services
start: run-infra
	@echo "Waiting for infrastructure to boot..."
	@sleep 5
	@make run-services

# Checks if the binaries are currently running
status:
	@echo "Checking status of microservices..."
	@ps aux | grep "[.]/bin/" || echo "No services are currently running."

# Cleans up the compiled files and logs
clean: stop-services
	@echo "Cleaning up..."
	@rm -rf bin/ logs/
	@echo "Clean complete!"

generate-proto:
	protoc --proto_path=pb \
	--go_out=pb --go_opt=paths=source_relative \
	--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	pb/*.proto