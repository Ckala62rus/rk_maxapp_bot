# Makefile для миграций
include .env
export

.PHONY: migrate-create migrate-up migrate-down migrate-force migrate-version

# URL базы данных
DB_URL ?= postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable

# Создать новую миграцию
migrate-create:
	@read -p "Please enter new migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

# Применить все миграции
migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

# Откатить последнюю миграцию
migrate-down:
	migrate -path migrations -database "$(DB_URL)" down 1

# Откатить на N миграций
migrate-down-n:
	@read -p "How much N migration need to rollback?: " steps; \
	migrate -path migrations -database "$(DB_URL)" down $$steps

# Применить N миграций
migrate-up-n:
	@read -p "Apply N migration(s): " steps; \
	migrate -path migrations -database "$(DB_URL)" up $$steps

# Показать текущую версию
migrate-version:
	migrate -path migrations -database "$(DB_URL)" version

# Принудительно установить версию (опасно!)
migrate-force:
	@read -p "Enter version for force migration install: " version; \
	migrate -path migrations -database "$(DB_URL)" force $$version

# Создать все таблицы (для разработки)
setup-db: migrate-up

# Сбросить базу (удалить все таблицы)
reset-db:
	migrate -path migrations -database "$(DB_URL)" down -all

# --- Миграции из Docker (postgres = имя сервиса в docker-compose) ---
DB_URL_DOCKER = postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@postgres:5432/$(POSTGRES_DB)?sslmode=disable

migrate-up-docker:
	docker compose --profile tools run --rm migrate -path /migrations -database "$(DB_URL_DOCKER)" up

migrate-down-docker:
	docker compose --profile tools run --rm migrate -path /migrations -database "$(DB_URL_DOCKER)" down 1

migrate-version-docker:
	docker compose --profile tools run --rm migrate -path /migrations -database "$(DB_URL_DOCKER)" version