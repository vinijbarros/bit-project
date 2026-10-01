-include .env
export

GOCACHE ?= $(CURDIR)/.cache/go-build
export GOCACHE

.PHONY: help db-up db-down api migrate-up migrate-status migrate-down seed-demo seed-reset-passwords frontend-install frontend-dev fmt check build

help:
	@echo "Alvos disponíveis:"
	@echo "  db-up            inicia somente o PostgreSQL"
	@echo "  db-down          encerra o PostgreSQL"
	@echo "  api              inicia a API Go"
	@echo "  migrate-up       aplica migrations Goose"
	@echo "  migrate-status   mostra o estado das migrations"
	@echo "  migrate-down     reverte a última migration (exige CONFIRM_DOWN=yes)"
	@echo "  seed-demo        cria dados locais (exige DEMO_SEED_ENABLED=true)"
	@echo "  seed-reset-passwords redefine senhas demo (exige confirmação)"
	@echo "  frontend-install instala dependências com npm ci"
	@echo "  frontend-dev     inicia o Vite em modo desenvolvimento"
	@echo "  fmt              formata o código Go"
	@echo "  check            executa testes Go, lint e tipos do frontend"
	@echo "  build            compila API/comandos e frontend"

db-up:
	docker compose up -d postgres

db-down:
	docker compose down

api:
	cd backend && go run ./cmd/api

migrate-up:
	cd backend && go run ./cmd/migrate up

migrate-status:
	cd backend && go run ./cmd/migrate status

migrate-down:
	@test "$(CONFIRM_DOWN)" = "yes" || (echo "Use CONFIRM_DOWN=yes somente em banco descartável ou rollback aprovado" && exit 1)
	cd backend && go run ./cmd/migrate down

seed-demo:
	cd backend && go run ./cmd/seed

seed-reset-passwords:
	@test "$(CONFIRM_SEED_PASSWORD_RESET)" = "yes" || (echo "Use CONFIRM_SEED_PASSWORD_RESET=yes para confirmar o reset" && exit 1)
	cd backend && go run ./cmd/seed --reset-passwords

frontend-install:
	cd frontend && npm ci

frontend-dev:
	cd frontend && npm run dev

fmt:
	cd backend && gofmt -w $$(find . -name '*.go' -type f)

check:
	cd backend && test -z "$$(gofmt -l .)" && go vet ./... && go test ./...
	cd frontend && npm run lint && npm run typecheck

build:
	cd backend && go build ./...
	cd frontend && npm run build
