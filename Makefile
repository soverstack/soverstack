# Makefile for Soverstack Launcher
# Provides convenient commands for building, testing, and developing

.PHONY: help build build-all test clean install run fmt vet deps

# Default target
help:
	@echo "Soverstack Launcher - Make targets"
	@echo ""
	@echo "Usage:"
	@echo "  make build       Build for current platform"
	@echo "  make build-all   Build for all platforms (Windows, Linux, macOS)"
	@echo "  make test        Run tests"
	@echo "  make fmt         Format code"
	@echo "  make vet         Run go vet"
	@echo "  make clean       Remove build artifacts"
	@echo "  make deps        Download dependencies"
	@echo "  make install     Build and install to GOPATH/bin"
	@echo "  make run         Build and run with example args"

# Build for current platform
build:
	@echo "Building for current platform..."
	go build -ldflags="-s -w -X main.Version=dev" -o soverstack .
	@echo "✓ Built: ./soverstack"

# Build for all platforms
build-all:
	@echo "Building for all platforms..."
	@chmod +x build/build.sh
	@./build/build.sh dev

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf dist/
	rm -f soverstack soverstack.exe
	@echo "✓ Clean complete"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod verify
	@echo "✓ Dependencies ready"

# Install to GOPATH/bin
install:
	@echo "Installing to GOPATH/bin..."
	go install -ldflags="-s -w -X main.Version=dev" .
	@echo "✓ Installed: $(shell go env GOPATH)/bin/soverstack"

# Build and run with example arguments
run: build
	@echo "Running launcher..."
	./soverstack --version
