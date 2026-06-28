.PHONY: clean all init generate generate_mocks build test test-unit test-api run docker-up docker-down help

# Variables
BINARY_NAME=api
BINARY_PATH=bin/$(BINARY_NAME)
MAIN_PATH=./cmd/main.go
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

all: build

help:
	@echo "Available commands:"
	@echo "  init         Initialize the project (install dependencies, generate code)"
	@echo "  build        Build the binary"
	@echo "  test         Run all tests"
	@echo "  test-unit    Run unit tests with coverage"
	@echo "  test-api     Run API tests"
	@echo "  run          Run the application locally"
	@echo "  docker-up    Start Docker containers"
	@echo "  docker-down  Stop Docker containers"
	@echo "  clean        Clean build artifacts"

build: generate
	@echo "Building application..."
	mkdir -p bin
	go build -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_PATH)"

clean:
	@echo "Cleaning up..."
	rm -rf bin/
	rm -rf generated/
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	rm -f internal/repository/*.mock.gen.go
	@echo "Clean complete"


init: clean generate
	@echo "Initializing project..."
	rm -rf vendor
	rm -f go.sum
	go clean -cache
	go clean -testcache
	go clean -modcache
	go mod tidy
	go mod vendor
	@echo "Project initialized"


run:
	@echo "Running application..."
	go run $(MAIN_PATH)

test:
	go clean -testcache
	go test -short -coverprofile $(COVERAGE_FILE) -v ./...

test-unit:
	go clean -testcache
	go test -v -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./internal/...
	go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report: $(COVERAGE_HTML)"

test_api:
	go clean -testcache
	go test -v ./tests/...
	
generate: generated generate_mocks

generated: api.yml
	@echo "Generating files..."
	@mkdir -p generated
	@if command -v oapi-codegen >/dev/null 2>&1; then \
		oapi-codegen --package generated -generate types,server,spec $< > generated/api.gen.go; \
	else \
		echo "Installing oapi-codegen v2.3.0..."; \
		go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.3.0; \
		$$HOME/go/bin/oapi-codegen --package generated -generate types,server,spec $< > generated/api.gen.go; \
	fi
	@echo "OpenAPI code generated"

REPOSITORY_GO_FILES := $(shell find internal/repository -name "*repo.go")

MOCK_DIR := mock

REPOSITORY_MOCK_FILES := $(patsubst internal/repository/%.go,$(MOCK_DIR)/%_mock.go,$(REPOSITORY_GO_FILES))

generate_mocks: $(MOCK_DIR) $(REPOSITORY_MOCK_FILES)

$(MOCK_DIR):
	@mkdir -p $(MOCK_DIR)

$(MOCK_DIR)/%_mock.go: internal/repository/%.go
	@echo "Generating mock $@"
	@mockgen -source=$< -destination=$@ -package=mock


docker-up:
	@echo "Starting Docker containers..."
	docker-compose up -d --build
	@echo "Waiting for API to be ready..."
	sleep 10
	@echo "Docker containers started"
	@echo "API: http://localhost:8080"

docker-down:
	@echo "Stopping Docker containers..."
	docker-compose down --volumes
	@echo "Docker containers stopped"

docker-logs:
	docker-compose logs -f

docker-restart: docker-down docker-up