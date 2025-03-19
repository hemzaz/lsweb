# Variables
BINARY_NAME=lsweb
VERSION=$(shell git describe --tags --always --dirty)
BUILD_DIR=dist
COVERAGE_DIR=coverage

# Default target
all: deps test build

# Fetch all dependencies
deps:
	@echo "Fetching dependencies..."
	go mod tidy
	go mod download

# Build the project using goreleaser
build: deps
	@echo "Building $(BINARY_NAME) version $(VERSION) using goreleaser..."
	goreleaser build --snapshot --clean

# Format all Go files
fmt:
	@echo "Formatting all Go files..."
	go fmt ./...

# Run tests
test: fmt
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
cover: fmt
	@echo "Running tests with coverage..."
	mkdir -p $(COVERAGE_DIR)
	go test -v -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report generated in $(COVERAGE_DIR)/coverage.html"

# Lint code
lint: fmt
	@echo "Linting code..."
	go vet ./...

# Clean up
clean:
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR) $(COVERAGE_DIR)
	go clean

# Install goreleaser if not installed
install-goreleaser:
ifeq (, $(shell which goreleaser))
	@echo "Installing goreleaser..."
	curl -sfL https://install.goreleaser.com/github.com/goreleaser/goreleaser.sh | sh
else
	@echo "goreleaser is already installed"
endif

# Help
help:
	@echo "Available targets:"
	@echo "  all (default)   - Fetch dependencies, run tests, and build the project"
	@echo "  deps            - Fetch all dependencies"
	@echo "  build           - Build the project using goreleaser"
	@echo "  fmt             - Format all Go files"
	@echo "  test            - Run tests"
	@echo "  cover           - Run tests with coverage"
	@echo "  lint            - Lint code with go vet"
	@echo "  clean           - Clean up"
	@echo "  install-goreleaser - Install goreleaser if not installed"
	@echo "  help            - Display this help"

.PHONY: all deps build fmt test cover lint clean install-goreleaser help
