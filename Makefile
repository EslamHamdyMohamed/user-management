.PHONY: help build run test migrate seed clean lint

help:
	@echo "Available commands:"
	@echo "  build    - Build the application"
	@echo "  run      - Run the application locally"
	@echo "  test     - Run tests"
	@echo "  lint     - Run linter"
	@echo "  migrate  - Run database migrations"
	@echo "  seed     - Seed the database"
	@echo "  docker   - Build and run with docker-compose"
	@echo "  clean    - Clean up build artifacts"

build:
	@echo "Building..."
	go build -o bin/server ./cmd/server

run: build
	@echo "Starting server..."
	./bin/server

test:
	@echo "Running tests..."
	go test ./... -v -cover

lint:
	@echo "Running linter..."
	golangci-lint run

migrate:
	@echo "Running migrations..."
	go run scripts/migrate.go

seed:
	@echo "Seeding database..."
	go run scripts/seed.go

docker:
	@echo "Starting with Docker..."
	docker-compose up --build

docker-prod:
	@echo "Starting production environment..."
	docker-compose -f docker-compose.prod.yml up --build

clean:
	@echo "Cleaning..."
	rm -rf bin/
	go clean