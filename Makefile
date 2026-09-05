.PHONY: swagger run test test-html docker-up docker-down docker-all

# Regenerate Swagger documentation
swagger:
	swag init -g cmd/api/main.go --parseDependency --parseInternal

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
	go test -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# Generate the visual coverage report in an HTML page
test-html: test
	go tool cover -html=coverage.out -o coverage.html