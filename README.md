# Portal de Solicitações Internas

Base executável do desafio técnico da bit Soluções. Nesta etapa existem a infraestrutura inicial da API, probes de saúde/prontidão, conexão PostgreSQL, executor de migrations e uma página React de confirmação. Autenticação, dashboard e CRUD de solicitações ainda não foram implementados.

## Pré-requisitos fixados

- Go 1.22.2.
- Node.js 24.14.0 e npm 11.9.0.
- PostgreSQL 17; o Compose usa a imagem `postgres:17.6-alpine`.
- Docker com Compose para iniciar o banco pelo alvo `db-up`.
- Make 4.3 é opcional; há comandos equivalentes abaixo.

As dependências exatas ficam em `backend/go.mod`, `backend/go.sum`, `frontend/package.json` e `frontend/package-lock.json`. Atualizações devem ser deliberadas, compatíveis com as versões acima e registradas em `docs/DECISOES.md`.

## Portas e origens locais

| Serviço | Endereço local |
| --- | --- |
| Frontend Vite | `http://127.0.0.1:5173` |
| API Go | `http://127.0.0.1:8080` |
| PostgreSQL | `127.0.0.1:5432` |

O frontend usa URLs relativas sob `/api/v1`. Em desenvolvimento, o Vite encaminha `/api` para `http://127.0.0.1:8080`, evitando CORS e preparando o uso futuro de cookies no mesmo site.

## Configuração

Copie os exemplos e mantenha os arquivos reais fora do versionamento:

```sh
cp .env.example .env
cp frontend/.env.example frontend/.env
```

Os valores fornecidos são apenas de demonstração local. `DATABASE_URL` é obrigatória para a API. As demais variáveis possuem padrões seguros para desenvolvimento e são validadas quando informadas. A API encerra com mensagem de configuração quando uma variável obrigatória está ausente ou inválida; banco temporariamente inacessível não impede o processo de servir `/healthz`, mas mantém `/readyz` em `503`.

Para carregar `.env` sem Make:

```sh
set -a
. ./.env
set +a
```

## Execução com Make

```sh
make db-up
make migrate-up
make api
```

Em outro terminal:

```sh
make frontend-install
make frontend-dev
```

A composição desta etapa inicia somente PostgreSQL. API e frontend serão adicionados ao Compose completo na etapa final de infraestrutura.

## Comandos equivalentes sem Make

```sh
docker compose up -d postgres

cd backend
go run -buildvcs=false ./cmd/migrate up
go run -buildvcs=false ./cmd/api
```

Em outro terminal, partindo da raiz:

```sh
cd frontend
npm ci
npm run dev
```

## Verificações

```sh
make fmt
make check
make build
```

Equivalentes sem Make:

```sh
cd backend
test -z "$(gofmt -l .)"
go vet ./...
go test ./...
go build -buildvcs=false ./...

cd ../frontend
npm ci
npm run lint
npm run typecheck
npm run build
```

Probes da API:

```sh
curl -i http://127.0.0.1:8080/healthz
curl -i http://127.0.0.1:8080/readyz
```

- `/healthz` retorna `200` enquanto o processo HTTP está vivo.
- `/readyz` retorna `200` apenas com conexão válida ao banco e `503` quando o banco está indisponível.
- Nenhum probe expõe URL, credencial ou outra configuração sensível.

O workspace recebido contém um diretório `.git` sem metadados válidos. Por isso os comandos locais de `run` e `build` usam `-buildvcs=false`; isso desativa somente a incorporação automática de metadados VCS no binário e poderá ser removido quando o repositório Git for inicializado corretamente.

## Migrations e seed

O comando `backend/cmd/migrate` suporta `up`, `down`, `status` e `version` usando Goose. Ainda não há migrations de tabelas porque o schema funcional será introduzido junto à etapa correspondente.

Não existe executável de seed nesta etapa. Ele será criado com os usuários de demonstração na implementação da autenticação, evitando um comando vazio ou dados fictícios apresentados como funcionais.
