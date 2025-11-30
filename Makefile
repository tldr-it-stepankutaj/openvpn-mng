.PHONY: build run clean test swagger deps

# Application name
APP_NAME=openvpn-mng
VERSION=1.0.0

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Directories
BUILD_DIR=./bin
CMD_DIR=./cmd/server

# Default target
all: deps build

# Install dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Build the application
build:
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) $(CMD_DIR)

# Run the application
run:
	$(GOCMD) run $(CMD_DIR)/main.go

# Run with custom config
run-config:
	$(GOCMD) run $(CMD_DIR)/main.go -config=$(CONFIG)

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Generate Swagger documentation
swagger:
	swag init -g cmd/server/main.go -o docs

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Install development tools
tools:
	go install github.com/swaggo/swag/cmd/swag@latest

# Docker build
docker-build:
	docker build -t $(APP_NAME):$(VERSION) .

# Docker run
docker-run:
	docker run -p 8080:8080 $(APP_NAME):$(VERSION)

# Help
help:
	@echo "Available targets:"
	@echo "  all          - Install dependencies and build"
	@echo "  deps         - Download and tidy dependencies"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application"
	@echo "  test         - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  swagger      - Generate Swagger documentation"
	@echo "  clean        - Clean build artifacts"
	@echo "  tools        - Install development tools"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
