# Makefile für GitLab-to-Todoist Exporter

# Variables
BINARY_NAME=gitlab-todoist-exporter
BINARY_UNIX=$(BINARY_NAME)_unix
BINARY_WINDOWS=$(BINARY_NAME).exe
BINARY_DARWIN=$(BINARY_NAME)_darwin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Build flags
BUILD_FLAGS=-ldflags="-s -w"
BUILD_DIR=./bin

# Default target
.PHONY: all
all: clean test build

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) -v

# Build for multiple platforms
.PHONY: build-all
build-all: build-linux build-windows build-darwin

.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_UNIX) -v

.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_WINDOWS) -v

.PHONY: build-darwin
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_DARWIN) -v

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@rm -f $(BINARY_UNIX)
	@rm -f $(BINARY_WINDOWS)
	@rm -f $(BINARY_DARWIN)
	@rm -f *.csv
	@echo "Clean completed"

# Test
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

# Lint code
.PHONY: lint
lint:
	@echo "Running linter..."
	$(GOLINT) run

# Download dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Update dependencies
.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	$(GOMOD) tidy
	$(GOGET) -u ./...

# Install binary
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)

# Run the application
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BUILD_DIR)/$(BINARY_NAME)

# Run with example parameters
.PHONY: run-example
run-example: build
	@echo "Running example..."
	@$(BUILD_DIR)/$(BINARY_NAME) -project "youruser/yourproject" -milestone "v1.0" -output "example_export.csv"

# Development helpers
.PHONY: dev
dev: fmt lint test build

# Create release archives
.PHONY: release
release: build-all
	@echo "Creating release archives..."
	@mkdir -p $(BUILD_DIR)/releases
	@tar -czf $(BUILD_DIR)/releases/$(BINARY_NAME)-linux-amd64.tar.gz -C $(BUILD_DIR) $(BINARY_UNIX)
	@tar -czf $(BUILD_DIR)/releases/$(BINARY_NAME)-darwin-amd64.tar.gz -C $(BUILD_DIR) $(BINARY_DARWIN)
	@zip -j $(BUILD_DIR)/releases/$(BINARY_NAME)-windows-amd64.zip $(BUILD_DIR)/$(BINARY_WINDOWS)
	@echo "Release archives created in $(BUILD_DIR)/releases/"

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build        - Build the binary for current platform"
	@echo "  build-all    - Build binaries for all platforms"
	@echo "  build-linux  - Build binary for Linux"
	@echo "  build-windows- Build binary for Windows"
	@echo "  build-darwin - Build binary for macOS"
	@echo "  clean        - Remove build artifacts and output files"
	@echo "  test         - Run tests"
	@echo "  fmt          - Format source code"
	@echo "  lint         - Run linter"
	@echo "  deps         - Download dependencies"
	@echo "  deps-update  - Update dependencies"
	@echo "  install      - Install binary to GOPATH/bin"
	@echo "  run          - Build and run the application"
	@echo "  run-example  - Run with example parameters"
	@echo "  dev          - Format, lint, test, and build"
	@echo "  release      - Create release archives for all platforms"
	@echo "  help         - Show this help message"

# Print build info
.PHONY: info
info:
	@echo "Build Information:"
	@echo "  Binary Name: $(BINARY_NAME)"
	@echo "  Build Dir:   $(BUILD_DIR)"
	@echo "  Go Version:  $(shell $(GOCMD) version)"
	@echo "  OS/Arch:     $(shell $(GOCMD) env GOOS)/$(shell $(GOCMD) env GOARCH)"

