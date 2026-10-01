-include .env
export

GOCACHE ?= $(CURDIR)/.cache/go-build
export GOCACHE

.PHONY: help db-up db-down api migrate-up migrate-status frontend-install frontend-dev fmt check build

help:
	@echo "Alvos disponíveis:"
	@echo "  db-up            inicia somente o PostgreSQL"
	@echo "  db-down          encerra o PostgreSQL"
	@echo "  api              inicia a API Go"
	@echo "  migrate-up       aplica migrations Goose"
	@echo "  migrate-status   mostra o estado das migrations"
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
	cd backend && go run -buildvcs=false ./cmd/api

migrate-up:
	cd backend && go run -buildvcs=false ./cmd/migrate up

migrate-status:
	cd backend && go run -buildvcs=false ./cmd/migrate status

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
	cd backend && go build -buildvcs=false ./...
	cd frontend && npm run build
