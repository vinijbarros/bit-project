# Portal de Solicitações Internas

Portal em desenvolvimento para o desafio técnico da bit Soluções. O backend já oferece autenticação persistente, CRUD, status, listagem/filtros e dashboard reais. O frontend possui login completo, recuperação da sessão por cookie, rotas protegidas, logout e shell responsivo; as telas de solicitações e dashboard ainda são placeholders explícitos. Metadata continua planejada no backend.

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

O frontend usa URLs relativas sob `/api/v1`. Em desenvolvimento, o Vite encaminha `/api` para `http://127.0.0.1:8080`, evitando CORS e mantendo o cookie de sessão no fluxo same-origin. `TRUSTED_ORIGINS` recebe uma lista de origens HTTP(S) exatas, separadas por vírgula; o padrão local é `http://127.0.0.1:5173`.

## Configuração

Copie os exemplos e mantenha os arquivos reais fora do versionamento:

```sh
cp .env.example .env
cp frontend/.env.example frontend/.env
```

Os valores fornecidos são apenas de demonstração local. `DATABASE_URL` é obrigatória para a API. As demais variáveis possuem padrões seguros para desenvolvimento e são validadas quando informadas. A API encerra com mensagem de configuração quando uma variável obrigatória está ausente ou inválida; banco temporariamente inacessível não impede o processo de servir `/healthz`, mas mantém `/readyz` em `503`.

## Contrato HTTP

O contrato está em `docs/openapi.yaml`, com explicações em `docs/API.md`. Probes, login/logout/me, CRUD, status, listagem/filtros e dashboard são reais. Metadata permanece planejada; sem sessão retorna `401` e, autenticado, `404 route_not_found` até receber um handler real.

A infraestrutura oferece JSON estrito limitado a 1 MiB, erros estruturados, validação de IDs/paginação, request ID, logs seguros, recuperação de panic e proteção de origem. Login/logout e POST/PATCH/DELETE de solicitações exigem `Origin`; futuras mutações seguirão a mesma regra. Não existe sucesso simulado.

## Autenticação local

Com banco, migrations, seed e API ativos:

```sh
curl -i -c /tmp/portal-cookies.txt \
  -H 'Content-Type: application/json' \
  -H 'Origin: http://127.0.0.1:5173' \
  --data '{"username":"colaborador1","password":"DemoLocal-Colaborador1"}' \
  http://127.0.0.1:8080/api/v1/auth/login

curl -i -b /tmp/portal-cookies.txt \
  http://127.0.0.1:8080/api/v1/auth/me

curl -i -b /tmp/portal-cookies.txt \
  -H 'Content-Type: application/json' \
  -H 'Origin: http://127.0.0.1:5173' \
  --data '{"title":"Acesso ao sistema","description":"Solicito acesso ao ambiente interno.","category":"ti"}' \
  http://127.0.0.1:8080/api/v1/requests

curl -i -b /tmp/portal-cookies.txt \
  http://127.0.0.1:8080/api/v1/requests/1

curl -i -b /tmp/portal-cookies.txt --get \
  --data-urlencode 'date_from=2026-10-01' \
  --data-urlencode 'date_to=2026-10-31' \
  --data-urlencode 'category=ti' \
  --data-urlencode 'status=aberto' \
  --data-urlencode 'q=acesso' \
  --data-urlencode 'page=1' \
  --data-urlencode 'page_size=20' \
  http://127.0.0.1:8080/api/v1/requests

curl -i -b /tmp/portal-cookies.txt \
  -X PATCH \
  -H 'Content-Type: application/json' \
  -H 'Origin: http://127.0.0.1:5173' \
  --data '{"status":"em_atendimento"}' \
  http://127.0.0.1:8080/api/v1/requests/1/status

curl -i -b /tmp/portal-cookies.txt \
  -H 'Origin: http://127.0.0.1:5173' \
  -X POST http://127.0.0.1:8080/api/v1/auth/logout
```

A sessão tem expiração absoluta de 8 horas por padrão e não é renovada. O token original fica somente no cookie `HttpOnly`; o PostgreSQL armazena SHA-256. `SESSION_COOKIE_SECURE=false` é exclusivo do HTTP local e produção exige `true`. O limite padrão de login é 5 falhas por username/IP em 15 minutos; por ser em memória, não é compartilhado entre instâncias.

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
go run ./cmd/migrate up
go run ./cmd/api
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
go build ./...

cd ../frontend
npm ci
npm run test
npm run lint
npm run typecheck
npm run build
```

### Integração PostgreSQL

Os testes de integração usam migrations reais e nunca reutilizam `DATABASE_URL`. O banco precisa ser indicado em `TEST_DATABASE_URL`, ter nome terminado em `_test` e receber a confirmação separada `TEST_DATABASE_ALLOW_RESET=yes`, pois a suíte executa `TRUNCATE` entre cenários.

```sh
make db-test-up
TEST_DATABASE_URL='postgres://portal_test:portal_test@127.0.0.1:55435/portal_test?sslmode=disable' \
  TEST_DATABASE_ALLOW_RESET=yes \
  make test-integration
make db-test-down
```

Sem Make, execute o mesmo `docker compose -f compose.test.yaml -p bit-project-integration ...` e, dentro de `backend/`, `go test -count=1 ./internal/integration`. A estratégia e a matriz detalhada estão em `docs/TESTES.md`.

Probes da API:

```sh
curl -i http://127.0.0.1:8080/healthz
curl -i http://127.0.0.1:8080/readyz
```

- `/healthz` retorna `200` enquanto o processo HTTP está vivo.
- `/readyz` retorna `200` apenas com conexão válida ao banco e `503` quando o banco está indisponível.
- Nenhum probe expõe URL, credencial ou outra configuração sensível.

## Migrations e seed

O comando `backend/cmd/migrate` suporta `up`, `status`, `down` e `version` usando Goose. As migrations criam `users`, `sessions` e `requests`, seus relacionamentos, constraints, índices e a identificação interna dos exemplos do seed.

```sh
make migrate-up
make migrate-status
```

`down` reverte somente a última migration, é bloqueado em `APP_ENV=production` e deve ser reservado a banco descartável ou rollback aprovado:

```sh
CONFIRM_DOWN=yes make migrate-down
```

### Dados de demonstração

O seed é um comando separado da API, transacional, repetível e bloqueado em produção. Ele só executa com habilitação explícita:

```sh
# após copiar .env.example para .env e executar as migrations
make seed-demo DEMO_SEED_ENABLED=true

# equivalente sem Make, partindo da raiz e com o ambiente carregado
cd backend
go run ./cmd/seed
```

As credenciais abaixo são **públicas e exclusivamente locais/de teste**. Elas vêm de variáveis, portanto podem ser trocadas antes da primeira execução:

| Usuário | Senha padrão do exemplo |
| --- | --- |
| `colaborador1` | `DemoLocal-Colaborador1` |
| `colaborador2` | `DemoLocal-Colaborador2` |

As senhas são persistidas como bcrypt, nunca em texto puro. Uma nova execução compara a senha configurada com o hash existente e não duplica usuários nem solicitações. Se um username já existir com outra senha, o comando falha e reverte toda a transação; ele não troca a senha silenciosamente. Quando o reset for realmente desejado:

```sh
make seed-reset-passwords DEMO_SEED_ENABLED=true CONFIRM_SEED_PASSWORD_RESET=yes
```

O seed cria cinco solicitações sintéticas, com as cinco categorias, os três status e datas anteriores ao momento da execução para demonstrar filtros de período:

| Exemplo | Autor que pode editar/excluir enquanto estiver `aberto` |
| --- | --- |
| Notebook não inicializa | `colaborador1` |
| Atualização de dados cadastrais | `colaborador2` |
| Reposição de materiais de escritório | `colaborador1` |
| Dúvida sobre reembolso de viagem | `colaborador2` |
| Ajuste de iluminação da sala | `colaborador1` |

Essas datas retroativas são uma capacidade exclusiva do seed. A criação normal pela API define `created_at` no backend e rejeita esse campo quando enviado pelo cliente. Registros alheios ao seed são preservados, e solicitações reais podem ter títulos repetidos. Uma base sem solicitações continua sendo um estado válido e o frontend não depende do seed. Os hashes são usados pelo login HTTP implementado.
