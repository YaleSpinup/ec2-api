.PHONY: all build run test clean lint dev docker-build docker-run help

# Go related variables
BINARY_NAME=ec2apigo
MAIN_FILE=main.go
BUILD_DIR=./tmp

# Docker related variables
DOCKER_IMAGE=ec2apigo
DOCKER_TAG=latest

# Go build flags
LDFLAGS=-ldflags "-w -s"

# Default target
all: build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(LDFLAGS) $(MAIN_FILE)

# Run the application
run:
	@go run $(MAIN_FILE)

# Run with live reload using air
dev:
	@echo "Starting development server with air..."
	@air

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -cover ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "Cleaned build cache"

# Run go fmt
fmt:
	@echo "Running go fmt..."
	@go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint is not installed. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) -f docker/Dockerfile .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	@docker run -p 8080:8080 $(DOCKER_IMAGE):$(DOCKER_TAG)

# Install development tools
tools:
	@echo "Installing development tools..."
	@go install github.com/air-verse/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Display help information
help:
	@echo "Available targets:"
	@echo "  all            - Build the application (default)"
	@echo "  build          - Build the application"
	@echo "  run           - Run the application"
	@echo "  dev           - Run the application with live reload using air"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  clean         - Clean build artifacts"
	@echo "  fmt           - Run go fmt"
	@echo "  vet           - Run go vet"
	@echo "  deps          - Install dependencies"
	@echo "  lint          - Run linter"
	@echo "  tools         - Install development tools"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  help          - Display this help message"