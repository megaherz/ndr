.PHONY: dev build test test-integration test-coverage lint migrate-up migrate-down migrate-create migrate-force proto-gen clean

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=ndr-server

# Database parameters
DATABASE_URL ?= postgres://ndr:ndr@localhost:5432/ndr?sslmode=disable
MIGRATIONS_PATH = internal/storage/postgres/migrations

# Development with live reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
	air

# Build the binary
build:
	$(GOBUILD) -o bin/$(BINARY_NAME) -v ./cmd/server

# Run unit tests
test:
	$(GOTEST) -v ./...

# Run integration tests (requires Docker)
test-integration:
	$(GOTEST) -v -tags=integration ./tests/integration/...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Database migrations (requires golang-migrate)
migrate-up:
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" down 1

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $$name

migrate-force:
	@read -p "Enter version to force: " version; \
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" force $$version

# Generate protobuf code
proto-gen:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/**/*.proto

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f coverage.out coverage.html

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Install development tools
install-tools:
	$(GOGET) github.com/cosmtrek/air@latest
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOGET) -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest