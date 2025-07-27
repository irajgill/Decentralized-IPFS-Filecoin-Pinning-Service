.PHONY: build test clean migrate seed docker-build docker-push deploy

# Build variables
BINARY_NAME=pinning-service
DOCKER_IMAGE=pinning-service
DOCKER_TAG=latest
GO_VERSION=1.21

# Build the application
build:
	go build -o bin/server ./cmd/server
	go build -o bin/worker ./cmd/worker

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run database migrations
migrate:
	./scripts/migrate.sh up

# Seed database with test data
seed:
	./scripts/seed.sh

# Run the server locally
run-server:
	go run ./cmd/server

# Run the worker locally
run-worker:
	go run ./cmd/worker

# Docker build
docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# Docker push
docker-push:
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

# Deploy to Kubernetes
deploy:
	./scripts/deploy.sh

# Start local development environment
dev-up:
	docker-compose up -d

# Stop local development environment
dev-down:
	docker-compose down

# View logs
logs:
	docker-compose logs -f

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Generate mocks for testing
mocks:
	mockgen -source=internal/storage/repositories.go -destination=internal/mocks/repositories_mock.go

# Hot reload for development
dev:
	air
