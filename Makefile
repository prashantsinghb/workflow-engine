# Makefile for Service
# Update SERVICE_NAME variable with your service name

SERVICE_NAME := service
BINARY_NAME := $(SERVICE_NAME)
IMAGE_NAME := $(SERVICE_NAME)
VERSION ?= latest

# Build the service binary
build:
	go build -o $(BINARY_NAME) ./cmd/$(SERVICE_NAME)/main.go

# Run the service locally
run: build
	./$(BINARY_NAME) -config default.yml

# Generate proto code
generate:
	cd api/$(SERVICE_NAME) && go generate

# Run tests
test:
	go test ./...

# Build Docker image
docker-build:
	docker build -t $(IMAGE_NAME):$(VERSION) -f build/$(SERVICE_NAME)/Dockerfile .

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	go clean

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run ./...

# Install dependencies
deps:
	go mod download
	go mod tidy

.PHONY: build run generate test docker-build clean fmt lint deps

