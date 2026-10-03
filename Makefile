-include .env
export

GOCACHE ?= $(CURDIR)/.cache/go-build
export GOCACHE

.PHONY: help compose-up compose-seed compose-down db-up db-down db-test-up db-test-down api migrate-up migrate-status migrate-down seed-demo seed-reset-passwords test-integration frontend-install frontend-dev fmt check build

help:
	@echo "Alvos disponíveis:"
	@echo "  compose-up        compila e inicia a aplicação completa"
	@echo "  compose-seed      executa explicitamente o seed demo"
	@echo "  compose-down      encerra a aplicação preservando o volume"
	@echo "  db-up            inicia somente o PostgreSQL"
	@echo "  db-down          encerra o PostgreSQL"
	@echo "  db-test-up       inicia PostgreSQL descartável de integração"
	@echo "  db-test-down     remove PostgreSQL descartável de integração"
	@echo "  api              inicia a API Go"
	@echo "  migrate-up       aplica migrations Goose"
	@echo "  migrate-status   mostra o estado das migrations"
	@echo "  migrate-down     reverte a última migration (exige CONFIRM_DOWN=yes)"
	@echo "  seed-demo        cria dados locais (exige DEMO_SEED_ENABLED=true)"
	@echo "  seed-reset-passwords redefine senhas demo (exige confirmação)"
	@echo "  test-integration executa testes PostgreSQL (exige TEST_DATABASE_URL e confirmação)"
	@echo "  frontend-install instala dependências com npm ci"
	@echo "  frontend-dev     inicia o Vite em modo desenvolvimento"
	@echo "  fmt              formata o código Go"
	@echo "  check            executa testes Go, lint e tipos do frontend"
	@echo "  build            compila API/comandos e frontend"

compose-up:
	docker compose up -d --build --wait

compose-seed:
	DEMO_SEED_ENABLED=true docker compose --profile demo run --rm seed

compose-down:
	docker compose down

db-up:
	docker compose up -d postgres

db-down:
	docker compose down

db-test-up:
	docker compose -f compose.test.yaml -p bit-project-integration up -d --wait postgres-test

db-test-down:
	docker compose -f compose.test.yaml -p bit-project-integration down --remove-orphans

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

test-integration:
	@test -n "$(TEST_DATABASE_URL)" || (echo "Defina TEST_DATABASE_URL para um banco descartável com nome terminado em _test" && exit 1)
	@test "$(TEST_DATABASE_ALLOW_RESET)" = "yes" || (echo "Defina TEST_DATABASE_ALLOW_RESET=yes para autorizar a limpeza do banco de teste" && exit 1)
	cd backend && go test -count=1 ./internal/integration

frontend-install:
	cd frontend && npm ci

frontend-dev:
	cd frontend && npm run dev

fmt:
	cd backend && gofmt -w $$(find . -name '*.go' -type f)

check:
	cd backend && test -z "$$(gofmt -l .)" && go vet ./... && go test ./...
	cd frontend && npm run test && npm run lint && npm run typecheck

build:
	cd backend && go build ./...
	cd frontend && npm run build
