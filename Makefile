.PHONY: up
up:
	docker compose up -d --no-cache

.PHONY: down
down:
	docker compose down