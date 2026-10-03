# Portal de Solicitações Internas

Aplicação full stack desenvolvida para o desafio técnico da bit Soluções. Colaboradores autenticados registram demandas internas, consultam a visão global e acompanham cada solicitação até a conclusão. Os dados são persistidos em PostgreSQL; não há arrays locais simulando as funcionalidades.

A interface identifica o portal pela marca **b1t**, uma referência visual ao bit binário representado por zero ou um.

## Sumário

- [Funcionalidades e regras](#funcionalidades-e-regras)
- [Stack e arquitetura](#stack-e-arquitetura)
- [Decisões técnicas](#principais-decisões-técnicas)
- [Pré-requisitos e configuração](#pré-requisitos-e-configuração)
- [Execução com Docker](#execução-com-docker-compose)
- [Execução sem Docker](#execução-sem-docker)
- [Banco, migrations e seed](#banco-migrations-e-seed)
- [Testes e CI](#testes-e-ci)
- [API](#api)
- [Deploy](#deploy)
- [Troubleshooting](#troubleshooting)
- [Evidências](#evidências)
- [Limitações e melhorias](#limitações-e-melhorias-futuras)
- [Uso de inteligência artificial](#uso-de-inteligência-artificial)

## Funcionalidades e regras

Funcionalidades entregues:

- login, sessão persistida em PostgreSQL e logout com revogação;
- criação com autor, data e status inicial definidos pelo backend;
- edição e exclusão física somente pelo autor enquanto a solicitação estiver aberta;
- listagem global com código, título, categoria, solicitante, abertura e status;
- detalhes, alteração livre entre os três status e reabertura;
- filtros combináveis por período, categoria, status e substring do título;
- dashboard global com total, abertas, em atendimento e concluídas;
- escolha entre modo claro e modo noturno, persistida somente no navegador;
- estados de carga, vazio, validação, erro e sessão expirada em interface responsiva.

### Regras de negócio e permissões

O PDF exige autenticação, cadastro, edição e exclusão de solicitações abertas, acompanhamento por três status, filtros e dashboard. Ele não define autoria, papéis, fluxo obrigatório de status, paginação nem vários limites operacionais. Para preencher essas lacunas, a implementação adota explicitamente as regras abaixo; elas são decisões deste projeto, não exigências atribuídas à bit Soluções.

- Todo usuário autenticado consulta a listagem global, os detalhes e o dashboard global.
- Somente o autor edita ou exclui sua solicitação, e apenas enquanto ela estiver em `Aberto`.
- Qualquer usuário autenticado pode mudar o status de qualquer solicitação para `Aberto`, `Em Atendimento` ou `Concluído`, inclusive reabrir. Repetir o status atual é idempotente.
- A exclusão é física. Não há lixeira, histórico de status, cadastro público, recuperação de senha ou papéis administrativos nesta versão.
- O backend define solicitante, `created_at`, `updated_at` e status inicial `aberto`; o cliente envia somente título, descrição e categoria ao criar.
- Após remoção de espaços externos, o título aceita 3–150 pontos de código Unicode e a descrição 10–5000. As categorias são TI, RH, Compras, Financeiro e Infraestrutura.
- Instantes são persistidos com fuso e tratados em UTC. Os filtros `YYYY-MM-DD` representam dias civis em `America/Sao_Paulo`.
- A pesquisa textual procura somente uma substring do título, sem diferenciar maiúsculas de minúsculas; os filtros são combinados por `AND`.
- A listagem usa paginação com 20 itens por padrão e máximo de 100. O código visual `SOL-000001` é calculado a partir do ID.

## Stack e arquitetura

- Backend: Go 1.22.2, `net/http`, `database/sql`, pgx 5.7.1, Goose 3.24.0 e bcrypt.
- Banco: PostgreSQL 17.6.
- Frontend: React 19.3.0, TypeScript 6.0.3, Vite 8.3.2, React Router 7.18.4 e CSS próprio.
- Distribuição: Docker/Compose, Nginx 1.27.4 e imagens multi-stage.
- Qualidade: testes Go/`httptest`/PostgreSQL, Vitest, Testing Library, Playwright, ESLint, OpenAPI/Redocly e GitHub Actions.

### Visão geral

A solução é uma aplicação única organizada em módulos, com frontend e backend separados. O navegador carrega a SPA React e chama caminhos relativos `/api/v1`; Vite encaminha essas chamadas no desenvolvimento e Nginx faz o mesmo na imagem final. A API Go aplica as regras e persiste no PostgreSQL.

```text
Navegador
  └─ React + React Router
       └─ JSON/HTTP + cookie de sessão HttpOnly
            └─ Vite (desenvolvimento) ou Nginx (Compose)
                 └─ handler → service → repository → PostgreSQL

Goose → migrations versionadas
seed  → usuários e solicitações sintéticas, somente quando habilitado
```

No backend, handlers traduzem HTTP/JSON, services concentram validação, autenticação e autorização, e repositories executam SQL parametrizado e transações. No frontend, páginas compõem telas, componentes tratam elementos reutilizáveis, features reúnem lógica de autenticação/formulários/filtros, e services usam o cliente Fetch centralizado. A autorização continua obrigatória no backend mesmo quando a interface oculta uma ação.

### Estrutura principal

```text
backend/
  cmd/api/               composição e servidor HTTP
  cmd/migrate/           comando Goose
  cmd/seed/              carga demonstrativa explícita
  db/migrations/         schema SQL versionado
  internal/domain/       tipos de domínio
  internal/http/         handlers, middleware e contrato HTTP
  internal/service/      regras de negócio e autorização
  internal/repository/   database/sql, consultas e transações
frontend/
  e2e/                   fluxos Playwright
  evidence/              roteiro automatizado das capturas
  src/api/               cliente HTTP central
  src/app/routes/        rotas públicas e protegidas
  src/components/        layout, formulários e feedback
  src/features/          lógica compartilhada por domínio
  src/pages/             telas da aplicação
  src/services/          operações da API
  src/styles/            CSS responsivo
  src/types/             contratos TypeScript
docs/                    documentação principal, OpenAPI e evidências
```

## Principais decisões técnicas

| Decisão adotada | Motivo principal | Benefício neste projeto |
| --- | --- | --- |
| Go com `net/http` | Manter o protocolo HTTP e as dependências explícitos em uma API pequena. | Código direto, binário simples e poucos componentes externos. |
| `database/sql` + pgx, sem ORM | Controlar SQL, transações e condições atômicas das regras de autoria/status. | Persistência auditável e aderente ao PostgreSQL. |
| PostgreSQL + Goose | Recriar e versionar estrutura, constraints e índices. | Instalação reproduzível e evolução rastreável do schema. |
| Sessão opaca persistida | Permitir revogação efetiva no logout sem guardar o token original. | Sessão sobrevive ao reinício da API e pode ser invalidada no servidor. |
| React + TypeScript + Vite | Construir uma SPA separada com estados e contratos tipados sem framework full stack. | Navegação e formulários organizados, com build e proxy simples. |
| URLs relativas e proxy same-origin | Usar o cookie HttpOnly sem expor token ao JavaScript nem exigir CORS no fluxo normal. | Menor superfície de configuração entre navegador e API. |
| CSS próprio e componentes focados | Evitar uma biblioteca visual grande para o escopo. | Interface responsiva com dependências menores. |
| Docker Compose com seed opcional | Reproduzir banco, migration, API e frontend sem criar credenciais silenciosamente. | Avaliação local previsível e preservação do volume por padrão. |
| Testes em camadas + CI | Separar regra unitária, contrato HTTP, PostgreSQL real, componentes e E2E. | Falhas ficam mais localizáveis sem chamar mocks de integração. |

As justificativas, alternativas e compromissos completos estão no [Memorial Técnico de Desenvolvimento](docs/MEMORIAL_TECNICO_DE_DESENVOLVIMENTO.md).

## Pré-requisitos e configuração

### Pré-requisitos

- Go 1.22.2.
- Node.js 24.14.0 e npm 11.9.0.
- PostgreSQL 17; o Compose usa a imagem `postgres:17.6-alpine`.
- Docker com Compose v2 para iniciar a aplicação completa.
- Chromium e suas bibliotecas para executar E2E localmente; o Playwright pode instalá-los pelo comando documentado abaixo.
- Make 4.3 é opcional; há comandos equivalentes abaixo.

As dependências exatas ficam em `backend/go.mod`, `backend/go.sum`, `frontend/package.json` e `frontend/package-lock.json`. Os lockfiles devem ser preservados para instalação reproduzível.

### Portas e origens locais

| Serviço | Endereço local |
| --- | --- |
| Frontend Docker/Nginx | `http://127.0.0.1:8080` |
| Frontend Vite | `http://127.0.0.1:5173` |
| API Go sem Docker | `http://127.0.0.1:8080` |
| PostgreSQL | `127.0.0.1:5432` |

O frontend usa URLs relativas sob `/api/v1`. No Compose, Nginx encaminha `/api` para a API interna e o navegador usa somente `http://127.0.0.1:8080`. No desenvolvimento sem Docker, o Vite encaminha `/api` para `http://127.0.0.1:8080`. `TRUSTED_ORIGINS`/`COMPOSE_TRUSTED_ORIGINS` recebem origens HTTP(S) exatas, sem curingas.

### Variáveis de ambiente

Copie os exemplos e mantenha os arquivos reais fora do versionamento:

```sh
cp .env.example .env
cp frontend/.env.example frontend/.env
```

Os valores fornecidos são públicos e exclusivamente de demonstração local. `DATABASE_URL` é obrigatória para os comandos Go. A API encerra com erro de configuração quando uma variável obrigatória está ausente ou inválida; banco temporariamente inacessível não impede `/healthz`, mas mantém `/readyz` em `503`.

| Variável | Significado | Exemplo/padrão local |
| --- | --- | --- |
| `APP_ENV` | Ambiente `development`, `test` ou `production` | `development` |
| `HTTP_ADDR` | Endereço da API sem Compose | `:8080` |
| `DATABASE_URL` | URL PostgreSQL obrigatória para execução no host | URL `127.0.0.1` de `.env.example` |
| `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_PORT` | Inicialização e publicação local do PostgreSQL Compose | `portal`, `portal`, `portal`, `5432` |
| `FRONTEND_BIND_ADDRESS`, `FRONTEND_PORT` | Endereço/porta pública do Nginx | `127.0.0.1`, `8080` |
| `COMPOSE_DATABASE_URL` | URL interna; dentro dos containers o host é `postgres` | exemplo em `.env.example` |
| `TRUSTED_ORIGINS` | Origens exatas aceitas pela API sem Docker | `http://127.0.0.1:5173` |
| `COMPOSE_TRUSTED_ORIGINS` | Origem pública exata do Compose | `http://127.0.0.1:8080` |
| `BACKEND_IMAGE`, `FRONTEND_IMAGE` | Nome/tag local das imagens | tags `:local` |
| `DATABASE_CONNECT_TIMEOUT` | Timeout de conexão/readiness | `3s` |
| `DATABASE_MAX_OPEN_CONNS`, `DATABASE_MAX_IDLE_CONNS` | Limites do pool SQL | `10`, `5` |
| `DATABASE_CONN_MAX_LIFETIME`, `DATABASE_CONN_MAX_IDLE_TIME` | Reciclagem de conexões | `30m`, `5m` |
| `HTTP_READ_HEADER_TIMEOUT`, `HTTP_READ_TIMEOUT`, `HTTP_WRITE_TIMEOUT`, `HTTP_IDLE_TIMEOUT` | Timeouts HTTP | `5s`, `10s`, `15s`, `60s` |
| `SHUTDOWN_TIMEOUT` | Prazo de encerramento gracioso | `10s` |
| `MIGRATIONS_DIR` | Diretório das migrations a partir de `backend/` | `db/migrations` |
| `SESSION_DURATION`, `SESSION_TOKEN_BYTES` | Validade absoluta e bytes aleatórios do token | `8h`, `32` |
| `SESSION_COOKIE_SECURE` | Exige HTTPS; obrigatório em produção | `false` somente no HTTP local |
| `LOGIN_RATE_LIMIT_MAX_ATTEMPTS`, `LOGIN_RATE_LIMIT_WINDOW`, `LOGIN_RATE_LIMIT_MAX_ENTRIES` | Limite de falhas por usuário/IP e teto em memória | `5`, `15m`, `10000` |
| `DEMO_SEED_ENABLED` | Habilitação explícita do seed | `false`; informe `true` ao executar |
| `DEMO_USER{1,2}_USERNAME`, `DISPLAY_NAME`, `PASSWORD` | Dois usuários públicos de demonstração | valores abaixo e em `.env.example` |
| `TEST_DATABASE_URL`, `TEST_DATABASE_ALLOW_RESET`, `TEST_POSTGRES_PORT` | Banco descartável e trava da integração | banco `_test`, `no`, `55435` |
| `VITE_API_BASE_PATH` | Única variável pública do frontend; caminho relativo da API | `/api/v1` |

Para carregar `.env` sem Make:

```sh
set -a
. ./.env
set +a
```

## Execução com Docker Compose

```sh
cp .env.example .env
docker compose up -d --build --wait
DEMO_SEED_ENABLED=true docker compose --profile demo run --rm seed
```

Abra `http://127.0.0.1:8080`. O PostgreSQL fica persistido em volume, migrations são aplicadas por uma tarefa que termina antes da API e o seed nunca roda silenciosamente. Para encerrar preservando os dados:

```sh
docker compose down
```

O volume `postgres_data` preserva o banco. `docker compose down --volumes` apaga esse volume e deve ser usado somente quando houver intenção explícita de reiniciar uma base local de demonstração; ele não faz parte do fluxo normal.

### Como usar a aplicação

1. Entre com um dos usuários demonstrativos indicados na seção de seed.
2. Consulte os quatro indicadores globais no Dashboard.
3. Abra **Solicitações** para ver a lista, combinar filtros e navegar entre páginas.
4. Use **Nova solicitação** para informar título, descrição e categoria; autor, datas e status inicial são automáticos.
5. Abra um item para consultar todos os detalhes ou alterar seu status.
6. Enquanto a solicitação estiver aberta, seu autor verá as ações de edição e exclusão. Outro usuário ainda poderá consultar e mudar o status, mas não editar ou excluir.
7. Use **Sair** para revogar a sessão no servidor.

## Execução sem Docker

Esta alternativa pressupõe que o PostgreSQL 17 esteja instalado, iniciado e acessível conforme `DATABASE_URL`. Em uma instalação local nova, crie o usuário e o banco demonstrativos antes de aplicar as migrations:

```sh
psql -h 127.0.0.1 -U postgres -c "CREATE USER portal WITH PASSWORD 'portal';"
psql -h 127.0.0.1 -U postgres -c "CREATE DATABASE portal OWNER portal;"
```

Esses comandos são exclusivamente para uma base local nova. Se o usuário ou o banco já existirem, não os recrie. Depois, carregue as variáveis como indicado em [Variáveis de ambiente](#variáveis-de-ambiente) e use uma das sequências abaixo.

### Com Make

```sh
make migrate-up
make seed-demo DEMO_SEED_ENABLED=true
make api
```

Em outro terminal:

```sh
make frontend-install
make frontend-dev
```

Esses alvos iniciam os processos de desenvolvimento fora dos containers. Para a composição completa, use `make compose-up`, `make compose-seed` e `make compose-down`.

### Sem Make

```sh
cd backend
go run ./cmd/migrate up
DEMO_SEED_ENABLED=true go run ./cmd/seed
go run ./cmd/api
```

Em outro terminal, partindo da raiz:

```sh
cd frontend
npm ci
npm run dev
```

## Banco, migrations e seed

O comando `backend/cmd/migrate` suporta `up`, `status`, `down` e `version` usando Goose. As migrations versionadas criam `users`, `sessions` e `requests`, com chaves, constraints e índices. O schema completo está no [Dicionário de Dados](docs/DICIONARIO_DE_DADOS.md).

```sh
make migrate-up
make migrate-status
```

`down` reverte a última migration, é bloqueado em produção e deve ser reservado a banco descartável ou rollback aprovado:

```sh
CONFIRM_DOWN=yes make migrate-down
```

### Dados de demonstração

O seed é separado da API, transacional, repetível, bloqueado em produção e exige habilitação explícita:

```sh
make seed-demo DEMO_SEED_ENABLED=true
```

Credenciais **públicas e exclusivamente locais/de teste**:

| Usuário | Senha demonstrativa |
| --- | --- |
| `colaborador1` | `DemoLocal-Colaborador1` |
| `colaborador2` | `DemoLocal-Colaborador2` |

As senhas são armazenadas como bcrypt. O seed cria cinco solicitações sintéticas que cobrem as cinco categorias, os três status, os dois autores e diferentes datas. Uma repetição não duplica seus registros nem altera silenciosamente a senha existente. O frontend também funciona com uma base sem solicitações.

Reset deliberado das senhas demonstrativas:

```sh
make seed-reset-passwords DEMO_SEED_ENABLED=true CONFIRM_SEED_PASSWORD_RESET=yes
```

## Testes e CI

### Verificações gerais

```sh
make fmt
make check
make build
make openapi-check
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

cd ..
npx --yes @redocly/cli@2.57.0 lint docs/openapi.yaml
```

### E2E no navegador

Os E2E exigem frontend, API Go e PostgreSQL reais em um ambiente isolado, com migrations e seed de teste já aplicados. Instale o Chromium compatível e execute apontando para a origem pública:

```sh
cd frontend
npx playwright install --with-deps chromium
E2E_BASE_URL='http://127.0.0.1:8080' npm run test:e2e
```

A suíte usa somente credenciais públicas de demonstração/teste, controla seus próprios registros e retém trace, screenshot e vídeo em caso de falha. O job `e2e` da CI prepara banco, migrations, seed, API e frontend antes de executar o navegador.

As evidências finais usam um roteiro separado e devem ser geradas somente contra ambiente demonstrativo controlado:

```sh
E2E_BASE_URL='http://127.0.0.1:8080' npm run evidence
```

### Integração contínua

`.github/workflows/ci.yml` executa, em push para `main` e pull requests, jobs separados de backend, frontend, E2E com PostgreSQL real e build Docker. O workflow não publica imagens nem faz deploy: CD externo permanece melhoria futura. A configuração foi reproduzida localmente; a primeira execução no GitHub ainda precisa ocorrer após o arquivo ser versionado.

Resultados confiáveis registrados durante o desenvolvimento:

- backend: testes unitários/HTTP, integração PostgreSQL, detector de corrida, `go vet` e build aprovados;
- frontend atual: 9 arquivos/54 testes, typecheck, lint e build aprovados;
- Playwright: 3 fluxos E2E no Chromium aprovados antes da inclusão do tema noturno;
- Compose: instalação limpa, migrations, seed repetível, fluxo funcional e persistência após reinício verificados;
- OpenAPI considerado válido pelo Redocly, com avisos de estilo não bloqueantes.

Esta reorganização documental não repetiu essas verificações. O seletor de tema possui teste de componente, mas não recebeu nova captura ou repetição E2E.

### Integração PostgreSQL

Os testes de integração usam migrations reais e nunca reutilizam `DATABASE_URL`. O banco precisa ser indicado em `TEST_DATABASE_URL`, ter nome terminado em `_test` e receber a confirmação separada `TEST_DATABASE_ALLOW_RESET=yes`, pois a suíte executa `TRUNCATE` entre cenários.

```sh
make db-test-up
TEST_DATABASE_URL='postgres://portal_test:portal_test@127.0.0.1:55435/portal_test?sslmode=disable' \
  TEST_DATABASE_ALLOW_RESET=yes \
  make test-integration
make db-test-down
```

Sem Make, use `docker compose -f compose.test.yaml -p bit-project-integration` para controlar o banco e, dentro de `backend/`, execute `go test -count=1 ./internal/integration` com as mesmas variáveis de segurança.

Probes da API:

```sh
curl -i http://127.0.0.1:8080/healthz
curl -i http://127.0.0.1:8080/readyz
```

- `/healthz` retorna `200` enquanto o processo HTTP está vivo.
- `/readyz` retorna `200` apenas com conexão válida ao banco e `503` quando o banco está indisponível.
- Nenhum probe expõe URL, credencial ou outra configuração sensível.

## API

A API usa o prefixo `/api/v1`; somente `/healthz` e `/readyz` ficam na raiz e são públicos. Login cria uma sessão opaca no cookie `portal_session` (`HttpOnly`, `SameSite=Lax`, `Path=/`), enquanto o banco armazena apenas SHA-256 do token. A duração padrão é de oito horas, sem renovação deslizante. Login, logout e mutações exigem origem exata autorizada.

O detalhamento de endpoints, parâmetros, payloads, exemplos, filtros, paginação, permissões e erros está em [API.md](docs/API.md). A especificação legível por ferramentas permanece em [OpenAPI 3.0.3](docs/openapi.yaml).

## Deploy

Para servidor, mantenha o Nginx/frontend como única origem pública, termine TLS em proxy reverso, use `APP_ENV=production`, `SESSION_COOKIE_SECURE=true`, origem HTTPS exata, PostgreSQL em rede privada e segredos fornecidos pelo ambiente. Migrations devem rodar como tarefa anterior à API; nunca automatize `down` em produção. Preserve o banco/volume entre recriações e estabeleça backup, restauração, observabilidade e rotação de segredos na infraestrutura escolhida.

O repositório não implementa publicação automática, TLS, backup ou infraestrutura cloud. O Compose é uma execução reproduzível local e uma base para servidor único, não uma declaração de prontidão para produção. O workflow implementa CI; CD e deploy externo permanecem melhorias futuras.

## Troubleshooting

- **`DATABASE_URL is required`**: copie `.env.example`, carregue-o com `set -a; . ./.env; set +a` e execute comandos Go a partir de `backend/`.
- **Banco indisponível ou `/readyz` em 503**: confirme `docker compose ps`, credenciais, host e porta. No host use `127.0.0.1`; dentro do Compose use `postgres` por `COMPOSE_DATABASE_URL`.
- **Porta 8080, 5173 ou 5432 ocupada**: encerre o processo conflitante ou altere `FRONTEND_PORT`/`POSTGRES_PORT`. Na execução sem Docker, `HTTP_ADDR` e o target do proxy Vite precisam continuar coerentes.
- **Login falha após trocar senha demo**: o seed não sobrescreve hash silenciosamente. Em banco exclusivamente demonstrativo, use o reset explícito; em produção, não use o seed.
- **Cookie não é enviado no HTTP local**: use `APP_ENV=development`, `SESSION_COOKIE_SECURE=false` e acesse exatamente a origem configurada. Em HTTPS/produção, `Secure=true` é obrigatório.
- **`403 origin_not_allowed`**: configure esquema, host e porta exatos, sem caminho ou barra final. Vite usa `http://127.0.0.1:5173`; Compose padrão usa `http://127.0.0.1:8080`.
- **Rota React retorna 404 ao recarregar**: use o Nginx fornecido, que possui fallback para `index.html`.
- **`npm ci` falha em volume Windows/WSL**: remova somente o `frontend/node_modules` local e repita em filesystem Linux ou pelo build Docker. Não remova `package-lock.json`.
- **Migration `down` recusada**: ela é bloqueada em produção e o Make exige `CONFIRM_DOWN=yes`; use somente em banco descartável ou rollback aprovado.

## Evidências

As nove capturas em [docs/evidencias](docs/evidencias/README.md) foram geradas em Chromium contra frontend, API e PostgreSQL reais. Elas cobrem login sem credencial exposta, dashboard, listagem, filtros, criação, detalhe, estados Em Atendimento/Concluído e viewport móvel.

As imagens antecedem a marca `b1t` e o tema noturno. Continuam válidas como evidência dos fluxos, mas não representam a identidade visual mais recente nem comprovam o modo escuro.

### Documentação principal

- [Memorial Técnico](docs/MEMORIAL_TECNICO_DE_DESENVOLVIMENTO.md)
- [Dicionário de dados](docs/DICIONARIO_DE_DADOS.md)
- [Contrato narrativo da API](docs/API.md)
- [Especificação OpenAPI](docs/openapi.yaml)

## Limitações e melhorias futuras

- O limitador de login vive em memória, perde contadores no reinício e não coordena múltiplas instâncias.
- Não há auditoria de alterações, histórico de status, exclusão lógica, papéis, gestão de usuários ou recuperação de senha.
- Duas edições concorrentes válidas seguem “última gravação vence”; não há versão, `ETag` ou merge.
- A busca por substring usa `ILIKE`, adequada ao volume do desafio, mas sem otimização específica como `pg_trgm`.
- A suíte E2E usa Chromium; não foi executada uma matriz completa de navegadores, teste de carga, pentest ou restauração de backup.
- A CI está configurada e seus comandos equivalentes têm resultados locais registrados, mas a execução no GitHub ainda não foi registrada. Não há CD.
- Para produção seriam necessários HTTPS, gestão de segredos, backup/restauração testados, observabilidade, auditoria, política de identidade e rate limit distribuído.

Esses itens são propostas ou limitações, não funcionalidades existentes.

## Uso de inteligência artificial

Durante o desenvolvimento, ChatGPT e Codex foram utilizados como apoio à tomada de decisões técnicas, à elaboração e revisão da documentação e ao design do frontend. Esse apoio contribuiu para analisar alternativas, organizar justificativas e propor a estrutura das telas, componentes, navegação e apresentação visual. O [Memorial Técnico](docs/MEMORIAL_TECNICO_DE_DESENVOLVIMENTO.md#10-uso-de-inteligência-artificial) detalha esse uso; os arquivos, testes e evidências registrados no repositório continuam sendo a referência verificável da solução.
