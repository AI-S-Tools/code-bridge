.PHONY: build install test clean run help

BINARY_NAME=code-bridge
INSTALL_PATH=/usr/local/bin

build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) ./cmd/code-bridge

install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	@sudo cp $(BINARY_NAME) $(INSTALL_PATH)/
	@echo "✓ Installed successfully"

uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "✓ Uninstalled successfully"

test:
	@echo "Running tests..."
	@go test -v ./...

clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -rf .code-bridge/
	@echo "✓ Cleaned"

run: build
	@./$(BINARY_NAME)

# Build for all platforms
release:
	@echo "Building releases..."
	@mkdir -p dist
	@GOOS=linux GOARCH=amd64 go build -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/code-bridge
	@GOOS=linux GOARCH=arm64 go build -o dist/$(BINARY_NAME)-linux-arm64 ./cmd/code-bridge
	@GOOS=darwin GOARCH=amd64 go build -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/code-bridge
	@GOOS=darwin GOARCH=arm64 go build -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/code-bridge
	@GOOS=windows GOARCH=amd64 go build -o dist/$(BINARY_NAME)-windows-amd64.exe ./cmd/code-bridge
	@echo "✓ Release binaries built in dist/"

# Development
dev: build
	@./$(BINARY_NAME) init
	@./$(BINARY_NAME) index

fmt:
	@echo "Formatting code..."
	@go fmt ./...

lint:
	@echo "Running linter..."
	@go vet ./...

help:
	@echo "Code-Bridge Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build      - Build the binary"
	@echo "  make install    - Install to $(INSTALL_PATH)"
	@echo "  make uninstall  - Remove from $(INSTALL_PATH)"
	@echo "  make test       - Run tests"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make run        - Build and run"
	@echo "  make release    - Build for all platforms"
	@echo "  make dev        - Build, init and index"
	@echo "  make fmt        - Format code"
	@echo "  make lint       - Run linter"
