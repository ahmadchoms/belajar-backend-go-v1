include .env
export

run:
	docker compose up --build

stop:
	docker compose down

MIGRATE_DB_HOST=localhost
MIGRATE_URL=postgresql://$(DB_USER):$(DB_PASSWORD)@$(MIGRATE_DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

migrate-create:
	@echo "Creating migration files..."
	migrate create -ext sql -dir db/migrations -seq $(name)

migrate-up:
	@echo "Running migrations up..."
	migrate -path db/migrations -database "$(MIGRATE_URL)" -verbose up

migrate-down:
	@echo "Rolling back migrations..."
	migrate -path db/migrations -database "$(MIGRATE_URL)" -verbose down

clean:
	rm -f main
	docker compose down --volumes --remove-orphans

.PHONY: migrate-create migrate-up migrate-down run stop clean