# Binary name
BINARY_NAME=opennet

# Build the application
build:
	@echo "Building..."
	@go build -o bin/$(BINARY_NAME) cmd/server/main.go

# Run the application
run: build
	@echo "Running..."
	@./bin/$(BINARY_NAME)

dev: build
	@echo "Running development mode"
	@. ./scripts/load_env.sh && ./bin/$(BINARY_NAME)

# Clean up binary
clean:
	@echo "Cleaning..."
	@rm -f bin/$(BINARY_NAME)

# Run tests
test:
	@echo "Testing..."
	@go test ./... -v

loadenv:
	@echo "Loading environment variables..."
	source ./scripts/load_env.sh

# Build and run
all: build run