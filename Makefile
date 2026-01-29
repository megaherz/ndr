.PHONY: dev stop clean proto-gen test

# Start full local development environment
dev:
	@echo "Starting local development environment..."
	docker-compose -f deployments/docker-compose.yml up -d
	@echo "Waiting for services to be ready..."
	@sleep 10
	@echo "Services started. Run 'make dev-backend' and 'make dev-frontend' in separate terminals."

# Start backend development server
dev-backend:
	cd backend && make dev

# Start frontend development server  
dev-frontend:
	cd frontend && npm run dev

# Stop all Docker services
stop:
	docker-compose -f deployments/docker-compose.yml down

# Clean build artifacts and Docker volumes
clean:
	docker-compose -f deployments/docker-compose.yml down -v
	cd backend && make clean
	cd frontend && rm -rf dist node_modules

# Generate Protobuf code
proto-gen:
	cd backend && make proto-gen

# Run all tests
test:
	cd backend && make test
	cd frontend && npm run test

# Install dependencies
deps:
	cd backend && make deps
	cd frontend && npm install

# Install development tools
install-tools:
	cd backend && make install-tools