ifneq (,$(wildcard .env))
    include .env
    export
endif

.PHONY: swagger migration-create migration-up migration-down \
        docker-up docker-down docker-build docker-logs

swagger:
	swag init -g ./cmd/api/main.go -o ./internal/docs

migration-create:
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migration-create name=<migration_name>"; \
		exit 1; \
	fi
	migrate create -ext sql -dir migrations -seq $(name)

migration-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migration-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

docker-build:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f app