.PHONY: all build run run-config clean test test-race test-services test-handlers test-middleware test-integration test-coverage test-coverage-html swagger deps tools docker-build docker-run lint release release-snapshot package-local help

# Application name
APP_NAME=openvpn-mng
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"

# Directories
BUILD_DIR=./bin
DIST_DIR=./dist
CMD_DIR=./cmd/server

# Default target
all: deps swagger build

# Install dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Build the application
build:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) $(CMD_DIR)

# Build for all platforms
build-all: swagger
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(CMD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 $(CMD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 $(CMD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 $(CMD_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe $(CMD_DIR)

# Run the application
run:
	$(GOCMD) run $(CMD_DIR)/main.go

# Run with custom config
run-config:
	$(GOCMD) run $(CMD_DIR)/main.go -config=$(CONFIG)

# Run tests
test:
	$(GOTEST) -v ./test/...

# Run tests with race detection
test-race:
	$(GOTEST) -v -race ./test/...

# Run specific test package
test-services:
	$(GOTEST) -v ./test/services/...

test-handlers:
	$(GOTEST) -v ./test/handlers/...

test-middleware:
	$(GOTEST) -v ./test/middleware/...

test-integration:
	$(GOTEST) -v ./test/integration/...

# Run tests with coverage (tests in test/ folder)
test-coverage:
	$(GOTEST) -v ./test/... -coverprofile=coverage.out
	$(GOCMD) tool cover -func=coverage.out

# Run tests with coverage HTML report
test-coverage-html:
	$(GOTEST) -v ./test/... -coverprofile=coverage.out
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Generate Swagger documentation
swagger:
	swag init -g cmd/server/main.go -o docs

# Lint code
lint:
	golangci-lint run --timeout=5m

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR) $(DIST_DIR)
	rm -f coverage.out coverage.html

# Install development tools
tools:
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/goreleaser/goreleaser/v2@latest

# Docker build
docker-build:
	docker build -t $(APP_NAME):$(VERSION) .

# Docker run
docker-run:
	docker run -p 8080:8080 $(APP_NAME):$(VERSION)

# Release with GoReleaser (requires tag)
release:
	goreleaser release --clean

# Release snapshot (for testing, no tag required)
release-snapshot:
	goreleaser release --snapshot --clean

# Build packages locally (for testing)
package-local: swagger
	goreleaser build --snapshot --clean
	@echo ""
	@echo "Build artifacts in ./dist/"
	@ls -la $(DIST_DIR)/

# Generate checksums for dist files
checksums:
	@cd $(DIST_DIR) && sha256sum *.tar.gz *.zip *.deb *.rpm 2>/dev/null > checksums.txt || true
	@echo "Checksums generated: $(DIST_DIR)/checksums.txt"

# Help
help:
	@echo "OpenVPN Manager - Build System"
	@echo ""
	@echo "Development:"
	@echo "  make deps             - Download and tidy dependencies"
	@echo "  make build            - Build the application"
	@echo "  make build-all        - Build for all platforms"
	@echo "  make run              - Run the application"
	@echo "  make swagger          - Generate Swagger documentation"
	@echo "  make lint             - Run linter"
	@echo "  make clean            - Clean build artifacts"
	@echo "  make tools            - Install development tools"
	@echo ""
	@echo "Testing:"
	@echo "  make test             - Run all tests"
	@echo "  make test-race        - Run tests with race detection"
	@echo "  make test-services    - Run service tests only"
	@echo "  make test-handlers    - Run handler tests only"
	@echo "  make test-middleware  - Run middleware tests only"
	@echo "  make test-integration - Run integration tests only"
	@echo "  make test-coverage    - Run tests with coverage report"
	@echo "  make test-coverage-html - Run tests with HTML coverage report"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build     - Build Docker image"
	@echo "  make docker-run       - Run Docker container"
	@echo ""
	@echo "Release:"
	@echo "  make release          - Create release with GoReleaser (requires tag)"
	@echo "  make release-snapshot - Create snapshot release (for testing)"
	@echo "  make package-local    - Build packages locally"
	@echo "  make checksums        - Generate checksums for dist files"
	@echo ""
	@echo "Current version: $(VERSION)"
