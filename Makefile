.PHONY: build test clean

# Binary name
BINARY_NAME=sawmill

# Build the binary
build:
	go build -o $(BINARY_NAME) .

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)

# Install dependencies (if any)
deps:
	go mod tidy

# Run tests with coverage
test-coverage:
	go test -v -cover ./...

# Build and run
run: build
	./$(BINARY_NAME)

# Default target
all: deps test build
