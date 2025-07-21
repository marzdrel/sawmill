.PHONY: build test clean release

# Binary name
BINARY_NAME=sawmill

# Get version from git tag (semantic versioning), fallback to "dev"
VERSION := $(shell git describe --tags --exact-match 2>/dev/null | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+' || echo "dev")

# Build the binary
build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY_NAME) .

# Run tests
test:
	gotestsum --format testdox
	# go test -v ./...

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

# Create a semantic version git tag for release
release:
	@echo "Current version: $(VERSION)"
	@read -p "Enter new semantic version (e.g., v1.0.0): " NEW_VERSION; \
	if echo $$NEW_VERSION | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+$$'; then \
		git tag -a $$NEW_VERSION -m "Release $$NEW_VERSION"; \
		echo "Created tag $$NEW_VERSION. Push with: git push origin $$NEW_VERSION"; \
	else \
		echo "Error: Version must follow semantic versioning format (v1.2.3)"; \
		exit 1; \
	fi

# Default target
all: deps test build
