.PHONY: tidy test build up down logs

# Download and sync dependencies (run this first after cloning)
tidy:
	go mod tidy

# Run all tests with race detector
test:
	go test ./... -race -v

# Build the binary locally
build:
	go build -o url-shortener .

# Start everything with Docker Compose
up:
	docker-compose up --build

# Stop containers
down:
	docker-compose down

# Follow logs
logs:
	docker-compose logs -f app

# Clean up everything including volumes (WARNING: deletes database data)
clean:
	docker-compose down -v
