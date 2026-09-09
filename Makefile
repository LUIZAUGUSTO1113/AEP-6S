.PHONY: swagger run test test-html test-unit test-db-up test-db-down docker-up docker-down docker-all

# Regenerate Swagger documentation
swagger:
	swag init -g main.go -d cmd/api,internal/sample --parseInternal

# Fast local development: regenerate Swagger and run native Go
run: swagger
	go run cmd/api/main.go

# Start only the local MongoDB database for development
docker-up:
	docker compose up -d mongodb

# Stop Docker containers
docker-down:
	docker compose down

# Start the full application (Database + API) in Docker
docker-all:
	docker compose up --build

# Run automated tests and display the coverage report
test:
	go test -v "-coverprofile=coverage.out" ./...
	go tool cover "-func=coverage.out"

# Generate the visual coverage report in an HTML page
test-html: test
	go tool cover "-html=coverage.out" -o coverage.html

# Fast unit/HTTP tests, with integration tests explicitly skipped.
test-unit:
	go test -short -v ./...

test-db-up:
	docker compose -f docker-compose.test.yml up -d --wait

test-db-down:
	docker compose -f docker-compose.test.yml down
