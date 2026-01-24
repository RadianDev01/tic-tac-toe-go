.PHONY: build run clean build-all docker-build docker-run

# Binary name
BINARY=tictactoe
SOURCE=tictactoe.go

# Default target: build and run
all: build run

# Build for current platform
build:
	@echo "Building $(BINARY)..."
	go build -o $(BINARY) $(SOURCE)
	@echo "Build complete!"

# Run the game
run:
	@echo "Starting Tic Tac Toe..."
	./$(BINARY)

# Build for all major platforms
build-all:
	@echo "Building for all platforms..."
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -o dist/$(BINARY)-linux-amd64 $(SOURCE)
	GOOS=darwin GOARCH=amd64 go build -o dist/$(BINARY)-darwin-amd64 $(SOURCE)
	GOOS=darwin GOARCH=arm64 go build -o dist/$(BINARY)-darwin-arm64 $(SOURCE)
	GOOS=windows GOARCH=amd64 go build -o dist/$(BINARY)-windows-amd64.exe $(SOURCE)
	@echo "All builds complete! Check the dist/ directory"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY)
	rm -f $(BINARY).exe
	rm -rf dist/
	@echo "Clean complete!"

# Docker operations
docker-build:
	@echo "Building Docker image..."
	docker build -t tictactoe-go .
	@echo "Docker image built!"

docker-run:
	@echo "Running in Docker..."
	docker run -it tictactoe-go

# Quick test run without building binary
test-run:
	go run $(SOURCE)
