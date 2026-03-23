include .env
export

COMPOSE     := docker compose -f docker-compose.yml
COMPOSE_ETH := docker compose -f docker-compose.ethemeral.yml
DB_URL      := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable
DB_ETH      := postgres://$(ETHEMERAL_USER):$(ETHEMERAL_PASSWORD)@$(ETHEMERAL_HOST):$(ETHEMERAL_PORT)/$(ETHEMERAL_DB)?sslmode=disable
ATLAS       := atlas
SQLC        := sqlc
SCHEMA      := ./db/schema.sql
SQLC_CONFIG := ./db/sqlc.yml

.PHONY: up down eth-up eth-down schema-apply

up:
	$(COMPOSE) up -d

down:
	$(COMPOSE) down

eth-up:
	$(COMPOSE_ETH) up -d

eth-down:
	$(COMPOSE_ETH) down

schema-apply:
	$(ATLAS) schema apply --url $(DB_URL) --to file://$(SCHEMA) --dev-url $(DB_ETH)

sqlc-generate:
	$(SQLC) generate -f $(SQLC_CONFIG)
