# Variables
BINARY_NAME=lsweb
VERSION=$(shell git describe --tags --always --dirty)
BUILD_DIR=dist

# Default target
all: deps build

# Fetch all dependencies
deps:
	@echo "Fetching dependencies..."
	go mod tidy
	go mod download

# Build the project using goreleaser
build: deps
	@echo "Building $(BINARY_NAME) version $(VERSION) using goreleaser..."
	goreleaser build --snapshot --rm-dist

# Clean up
clean:
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR)
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
	@echo "  all (default)   - Fetch dependencies and build the project"
	@echo "  deps            - Fetch all dependencies"
	@echo "  build           - Build the project using goreleaser"
	@echo "  clean           - Clean up"
	@echo "  install-goreleaser - Install goreleaser if not installed"
	@echo "  help            - Display this help"

.PHONY: all deps build clean install-goreleaser help
