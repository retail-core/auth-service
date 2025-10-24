
DB_URL=postgres://postgres:postgres@localhost:5432/auth_db?sslmode=disable

# Run API
run:
	@go run cmd/server/main.go

# Run SQLC generation
sqlc:
	@echo "🧠 Generating SQLC code..."
	@sqlc generate

# Apply all up migrations
migrate-up:
	@echo "📦 Applying migrations..."
	@migrate -path internal/db/migrations -database "$(DB_URL)" up

# Rollback last migration
migrate-down:
	@echo "⏪ Rolling back last migration..."
	@migrate -path internal/db/migrations -database "$(DB_URL)" down 1

# Drop all migrations
migrate-drop:
	@echo "🔥 Dropping all migrations..."
	@migrate -path internal/db/migrations -database "$(DB_URL)" drop

# Create new migration
migration:
	@if [ -z "$(name)" ]; then echo "❌ Usage: make migration name=your_migration_name"; exit 1; fi
	@migrate create -ext sql -dir internal/db/migrations -seq $(name)

# Run Go tests
test:
	@echo "🧪 Running tests..."
	@go test ./... -v

# Tidy modules
tidy:
	@echo "🧹 Cleaning up modules..."
	@go mod tidy
