.PHONY: dev build run test migrate-up migrate-down templ tailwind clean seed-admin

# Development with live reload
dev:
	air

# Build the binary
build:
	go build -o bin/cairnpost ./cmd/server

# Run the binary
run: build
	./bin/cairnpost

# Run tests
test:
	go test ./...

# Generate templ files
templ:
	templ generate

# Build Tailwind CSS
tailwind:
	npx @tailwindcss/cli -i web/static/css/input.css -o web/static/css/output.css

# Watch Tailwind CSS
tailwind-watch:
	npx @tailwindcss/cli -i web/static/css/input.css -o web/static/css/output.css --watch

# Database migrations (requires migrate CLI)
migrate-up:
	migrate -path internal/database/migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path internal/database/migrations -database "$$DATABASE_URL" down 1

# Start PostgreSQL via Docker Compose
db:
	docker compose up -d db

# Seed admin user (requires NAME, EMAIL, PASSWORD env vars)
seed-admin:
	go run ./cmd/seed --name "$(NAME)" --email "$(EMAIL)" --password "$(PASSWORD)"

# Seed full demo dataset (requires EMAIL, PASSWORD env vars)
seed-demo:
	go run ./cmd/seed --demo --email "$(EMAIL)" --password "$(PASSWORD)"

# Clean build artifacts
clean:
	rm -rf bin/ tmp/
