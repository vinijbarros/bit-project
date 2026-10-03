# Progresso do projeto

Este arquivo é um registro factual. **Implementado** significa que o artefato foi criado; **validado** significa que a verificação indicada foi realmente executada e passou.

## Estado geral

| Etapa | Escopo | Estado |
| --- | --- | --- |
| E1 | Escopo e organização | Concluída |
| E2 | Fundação e autenticação | Implementada e validada |
| E3 | API de solicitações | Implementada e validada, incluindo metadata |
| E4 | Frontend base | Base, autenticação e componentes compartilhados implementados e validados |
| E5 | Fluxos completos | Concluída — dashboard, lista, detalhe, criação, edição, status e exclusão implementados e validados |
| E6 | Qualidade e entrega | Em andamento — testes backend/frontend existentes; entrega final pendente |

## Registro de etapas

### E1 — Escopo e organização — 01/10/2026

**Implementado**

- Inventário inicial do workspace.
- Leitura integral do PDF de referência (12 páginas).
- Matriz de requisitos com origem, critérios de aceite, etapa, estado e campo de evidência.
- Registro das decisões de negócio fornecidas, dúvidas, arquitetura, alternativas, modelo inicial, telas, endpoints e fases.
- Orientações persistentes do projeto em `AGENTS.md`.

**Arquivos**

- `AGENTS.md` — orientações para próximas etapas.
- `docs/REQUISITOS.md` — escopo e rastreabilidade.
- `docs/DECISOES.md` — arquitetura e justificativas iniciais.
- `docs/PROGRESSO.md` — este registro.
- `docs/referencia/Seleção DEV Jr. 09_2026 - mini-projeto Full Stack.pdf` — somente lido, não alterado.

**Verificações executadas**

- `find . -maxdepth 3 -type f -print | sort` — confirmou que inicialmente havia somente o PDF e, ao final, os quatro documentos criados além da referência.
- `git status --short --branch` — não executou a consulta de estado porque `.git` não é um repositório válido; condição registrada abaixo.
- `file docs/referencia/<arquivo>.pdf` — confirmou um PDF 1.4; a ferramenta reportou 8 páginas físicas.
- Extração com `pypdf` (`PdfReader` e `extract_text`) — leu 12 páginas lógicas e recuperou o enunciado completo, do contexto aos diferenciais.
- Checagem shell com `test`, `rg`, `sed`, `sort`, `uniq` e `awk` — passou: quatro arquivos não vazios, 39 requisitos com IDs únicos, todos pendentes, tabela com número correto de colunas e seções/orientações obrigatórias presentes.

**Bloqueios e observações**

- Não havia `AGENTS.md`; ele foi criado conforme solicitado.
- O workspace continha apenas o PDF e diretórios internos vazios na inspeção inicial.
- O diretório `.git` existe, mas não contém metadados de um repositório Git válido; por isso `git status` não está disponível nesta etapa. Isso não impede a documentação, mas deverá ser resolvido pelo responsável antes de depender de histórico/CI.
- `pdfinfo` e `pdftotext` não estavam instalados. O PDF foi lido com `pypdf` 6.19.0 instalado temporariamente em `/tmp/bit_project_pdf_reader`; nenhuma dependência foi adicionada ao projeto.
- Nenhuma funcionalidade, migration, teste automatizado ou execução da aplicação foi implementada nesta etapa.

### E2 — Base executável — 01/10/2026

**Implementado**

- Módulo local Go `portal-solicitacoes`, sem endereço GitHub, com pgx/database/sql e executor Goose.
- Configuração por ambiente com validação de URL, ambiente, endereço, durações e limites do pool.
- Pool PostgreSQL limitado, timeout de ping, timeouts HTTP e encerramento gracioso.
- `GET /healthz` independente do banco e `GET /readyz` baseado em `PingContext`, com erro público seguro.
- Compose inicial apenas para PostgreSQL 17.6 Alpine, com volume e healthcheck; composição completa reservada à etapa 18.
- Frontend React/TypeScript/Vite em modo strict, React Router, ESLint, página inicial honesta e fallback 404.
- Proxy Vite de `/api` para a API e variável pública relativa `/api/v1`.
- `.gitignore`, exemplos de ambiente sem segredos, Makefile e README com comandos equivalentes.
- Testes unitários da configuração e dos probes.

**Validado**

- `go mod tidy` — gerou `go.sum` com pgx 5.7.1 e Goose 3.24.0 compatíveis com Go 1.22.2.
- `test -z "$(gofmt -l .)"` — nenhum arquivo Go fora do formato.
- `GOCACHE=/tmp/portal-solicitacoes-go-cache go vet ./...` — passou.
- `GOCACHE=/tmp/portal-solicitacoes-go-cache go test ./...` — passou; pacotes `internal/config` e `internal/http` testados.
- `GOCACHE=/tmp/portal-solicitacoes-go-cache go build -buildvcs=false ./...` — passou.
- Execução da API com `DATABASE_URL` ausente e com esquema inválido — ambas encerraram com código 1 e mensagem específica de configuração, sem revelar valor sensível.
- Execução real da API em `127.0.0.1:18080` com banco inacessível — processo iniciou, `GET /healthz` retornou `200`, `GET /readyz` retornou `503`, `POST /healthz` retornou `405` e `SIGINT` encerrou graciosamente com código 0.
- `npm install --package-lock-only --ignore-scripts` — lockfile gerado após resolver compatibilidade; auditoria reportou 0 vulnerabilidades.
- `npm ci` — instalou 120 pacotes a partir do lockfile; auditoria reportou 0 vulnerabilidades.
- `npm run lint` — passou sem avisos.
- `npm run typecheck` — passou em TypeScript strict.
- `npm run build` — passou com Vite 8.3.2; 26 módulos transformados.
- `npm run dev` e `curl http://127.0.0.1:5173/` — Vite iniciou e a página respondeu `200`.
- `GOCACHE=/tmp/portal-solicitacoes-go-cache make check build` — alvo agregado passou, cobrindo formatação, vet, testes, lint, tipos e builds.
- `npm ls --depth=0` e leitura do lockfile — dependências diretas correspondem às versões fixadas e o lockfile v3 é válido.
- Parse de `compose.yaml` com PyYAML — sintaxe válida e healthcheck do serviço PostgreSQL presente; isso não substitui execução pelo Docker.
- Checagem estrutural com `rg`, `sed`, `sort`, `uniq` e `awk` — 39 requisitos com IDs únicos e tabela íntegra.
- Checagem final com `test`/`rg` — nenhum `.env` real presente, exclusões obrigatórias registradas, variável Vite limitada a `/api/v1` e nenhum marcador `TODO`/`FIXME`/`PLACEHOLDER` encontrado nos fontes.

**Arquivos**

- `.gitignore`, `.env.example`, `frontend/.env.example` — exclusões e configuração de demonstração.
- `Makefile`, `compose.yaml`, `README.md` — automação local, PostgreSQL inicial e instruções.
- `backend/go.mod`, `backend/go.sum` — módulo e dependências Go fixadas.
- `backend/cmd/api/main.go` — ciclo da API e encerramento gracioso.
- `backend/cmd/migrate/main.go` — ações Goose `up`, `down`, `status` e `version`.
- `backend/cmd/seed/README.md`, `backend/db/migrations/README.md` — reservas explícitas, sem executável/migration vazios.
- `backend/internal/config/*` — configuração e testes.
- `backend/internal/http/*` — rotas/probes e testes.
- `backend/internal/repository/database/database.go` — criação e limites do pool.
- `frontend/package.json`, `frontend/package-lock.json` — scripts e versões fixadas.
- `frontend/index.html`, `frontend/tsconfig.json`, `frontend/eslint.config.js`, `frontend/vite.config.ts` — ferramenta e build.
- `frontend/src/*` — bootstrap, router, páginas iniciais, tipos Vite e CSS.
- `docs/DECISOES.md`, `docs/REQUISITOS.md`, `docs/PROGRESSO.md` — versões, decisões, rastreabilidade e evidências atualizadas.

**Bloqueios e pendências**

- Docker e PostgreSQL não estão instalados/disponíveis nesta distribuição WSL. O Compose, migrations e `/readyz` com banco real não puderam ser executados; localmente, habilitar Docker Desktop/WSL e rodar `make db-up`, `make migrate-up` e consultar `/readyz`.
- O primeiro `go build ./...` falhou apenas no VCS stamping porque `.git` continua inválido. A compilação passou com `-buildvcs=false`, registrado no Makefile/README; nenhum metadado existente foi apagado.
- O cache Go padrão tentou gravar em `/home/vinijbarros/.cache/go-build`, que é somente leitura neste ambiente. As verificações usaram `/tmp/portal-solicitacoes-go-cache`; o Makefile define cache local ignorado em `.cache/go-build` para execução reproduzível.
- TypeScript 7.0.2, embora disponível, conflitou com o peer range `<6.1.0` do `typescript-eslint` 8.71.0. Foi fixado TypeScript 6.0.3, a versão mais recente compatível consultada, sem `--force` ou `--legacy-peer-deps`.
- Não foram implementados autenticação, seed de usuário, schema/tabelas, CRUD, dashboard ou consumo funcional da API.

### E2 — Modelagem SQL e migrations — 01/10/2026

**Implementado**

- Migration Goose `00001_create_core_tables.sql` com Up/Down transacionais para `users`, `sessions` e `requests`.
- Chaves primárias, identidade numérica, FKs com `ON DELETE RESTRICT`, `NOT NULL`, `TIMESTAMPTZ`, defaults e CHECKs nomeadas.
- Categorias/status internos estáveis, limites de título/descrição, username normalizado e hash de sessão SHA-256 hexadecimal.
- Índices para solicitante, status, categoria, abertura, usuário da sessão e expiração.
- Comando de migration com `up`, `status`, `down` e `version`; `down` bloqueado em `APP_ENV=production`.
- Alvo `migrate-down` exige `CONFIRM_DOWN=yes`.
- Teste SQL transacional de constraints em `backend/db/migrations/testdata/constraints.sql`, sempre finalizado com `ROLLBACK`.
- Dicionário completo em `docs/DICIONARIO_DE_DADOS.md`, incluindo relações, rótulos, código `SOL-000001` e decisão sobre `ILIKE`.
- Nenhum seed, usuário ou solicitação de demonstração foi criado.

**Validado**

- `go test ./...`, `go vet ./...` e `go build ./...` — passaram após a alteração do comando.
- Compose isolado `bit-project-migration-test`, PostgreSQL `17.6-alpine`, porta `55432` e volume exclusivo — healthcheck saudável.
- Primeiro `go run ./cmd/migrate up` — aplicou `00001_create_core_tables.sql` e chegou à versão 1.
- `go run ./cmd/migrate status` e `version` — mostraram a migration aplicada e versão 1.
- Segundo `go run ./cmd/migrate up` — informou `no migrations to run`, sem reaplicação.
- `testdata/constraints.sql` — casos válidos e rejeições de username, unicidade, token, expiração, título, descrição, categoria, status, data e exclusão de usuário passaram; transação terminou em `ROLLBACK`.
- Contagem após o teste — `users=0`, `sessions=0`, `requests=0`, confirmando ausência de dados residuais.
- Catálogo PostgreSQL — 12 CHECKs, 2 FKs, 3 PKs e 1 constraint de unicidade nas três tabelas.
- `APP_ENV=production go run ./cmd/migrate down` — bloqueado antes da conexão e encerrou com status 1.
- `down` no banco descartável — removeu `users`, `sessions` e `requests`; consulta com `to_regclass` confirmou ausência.
- Novo `up` após o `down` — recriou o schema e os testes de constraints passaram novamente.
- `docker compose ... down -v --remove-orphans` — removeu exclusivamente container, rede e volume do teste.
- `go run ./cmd/migrate redo` — rejeitou ação não suportada com mensagem clara e status diferente de zero.
- `make migrate-down` sem `CONFIRM_DOWN=yes` — bloqueado com orientação e status diferente de zero.
- `GOCACHE=/tmp/portal-solicitacoes-go-cache make check build` — passou após todas as mudanças: gofmt, vet, testes, builds, lint e typecheck.

**Arquivos**

- `backend/db/migrations/00001_create_core_tables.sql` — schema versionado e reversível.
- `backend/db/migrations/testdata/constraints.sql` — validação transacional do schema.
- `backend/db/migrations/README.md` — uso seguro e inventário.
- `backend/cmd/migrate/main.go`, `main_test.go` — comandos, proteção de produção e testes.
- `Makefile`, `README.md` — alvos e instruções de migration.
- `docs/DICIONARIO_DE_DADOS.md` — dicionário do schema.
- `docs/DECISOES.md`, `docs/REQUISITOS.md`, `docs/PROGRESSO.md` — decisões e rastreabilidade atualizadas.

**Pendências**

- Normalização de username, hashing do token e definição explícita de status/timestamps ainda serão implementados no backend junto à autenticação/CRUD.
- Formatação visual `SOL-%06d` será adicionada aos DTOs; não existe coluna duplicada no banco.
- Busca por título com `ILIKE` será implementada no repository. Nenhum índice B-tree foi prometido para substring.
- O bloqueio anterior de Docker/PostgreSQL foi resolvido neste ambiente; a modelagem passou em banco real descartável.

### E2 — Seed de demonstração — 01/10/2026

**Implementado**

- Comando independente `backend/cmd/seed`, habilitado somente por `DEMO_SEED_ENABLED=true` e bloqueado em produção.
- Dois usuários configuráveis por ambiente, com validação de username, nome e senha de 8–72 bytes.
- Hash bcrypt no custo padrão, com comparação imediata do hash gerado e sem registro de senha, hash ou token em logs.
- Recusa de troca silenciosa de senha; `--reset-passwords` e alvo Make com confirmação fornecem reset deliberado.
- Cinco solicitações sintéticas cobrindo cinco categorias, três status, dois autores e cinco datas retroativas distintas.
- Transação serializável e repetição por `requests.seed_key` anulável/única, sem depender do título e sem remover dados externos.
- Migration `00002_add_request_seed_key.sql`, constraint de formato, índice único parcial e casos SQL de formato/unicidade.
- Variáveis públicas locais em `.env.example`, instruções Make/sem Make, credenciais, autoria, datas retroativas e limites documentados.
- Dependência `golang.org/x/crypto v0.31.0` promovida a direta em `go.mod`; versão já presente no lock de módulos da fundação.

**Validado**

- `go test ./internal/seed ./cmd/seed` — passou; validou configuração explícita, limite de 72 bytes, bcrypt, opções e cobertura dos exemplos.
- Compose isolado `bit-project-seed-test`, PostgreSQL `17.6-alpine`, banco `portal_seed_test` e porta `55433` — healthcheck saudável.
- Primeiro `go run ./cmd/migrate up` — aplicou `00001` e `00002`, chegando à versão 2; segundo `up` informou `no migrations to run`.
- Em segundo banco vazio isolado, `down` da versão 2 removeu `seed_key`/índice e preservou `requests`; novo `up` reaplicou `00002` e retornou à versão 2.
- Primeira execução do seed — `users_created=2` e `requests_created=5`; segunda execução — `users_existing=2`, `requests_existing=5` e nenhuma criação.
- Consultas SQL — 2 usuários demo com formato bcrypt; 5 solicitações, 5 categorias, 3 status, 2 autores, datas variadas e todas anteriores ao instante da consulta.
- A segunda execução comparou as duas senhas públicas com seus hashes e passou, confirmando hashes prontos para a futura autenticação; login HTTP não foi afirmado nem testado.
- Execução com senha divergente e sem reset — falhou com status 1 e orientação segura, sem exibir credencial/hash.
- `--reset-passwords` — alterou exatamente a senha divergente; restauração e alvo `make seed-reset-passwords ... CONFIRM_SEED_PASSWORD_RESET=yes` também passaram.
- Registro externo ao seed inserido no banco isolado — após nova carga, totais permaneceram 3 usuários/6 solicitações, com 1 usuário e 1 solicitação externos preservados.
- `DEMO_SEED_ENABLED=false` e `APP_ENV=production` — ambos bloquearam o comando antes de qualquer alteração e retornaram status 1.
- `testdata/constraints.sql` — passou em transação com `ROLLBACK`, inclusive rejeições de formato e duplicidade de `seed_key`.
- `GOCACHE=/tmp/bit-project-go-build make check build` — passou: gofmt, vet, todos os testes Go, ESLint, TypeScript strict e builds Go/Vite; Vite transformou 26 módulos.
- `docker compose -p bit-project-seed-test down -v --remove-orphans` — removeu somente o container, a rede e o volume descartáveis da validação.
- `docker compose -p bit-project-seed-migration-test down -v --remove-orphans` — removeu também todos os recursos do teste isolado de reversibilidade.

**Arquivos**

- `backend/cmd/seed/main.go`, `main_test.go`, `README.md` — CLI, proteção, logs seguros, testes e orientação.
- `backend/internal/seed/seed.go`, `seed_test.go` — configuração, bcrypt, transação e dados sintéticos.
- `backend/db/migrations/00002_add_request_seed_key.sql`, `testdata/constraints.sql` — identificação repetível e validação do banco.
- `backend/go.mod`, `backend/go.sum`, `.env.example`, `Makefile`, `README.md` — dependência, configuração, comandos e credenciais públicas locais.
- `docs/DICIONARIO_DE_DADOS.md`, `docs/DECISOES.md`, `docs/REQUISITOS.md`, `docs/PROGRESSO.md` — modelagem, decisões e evidências.

**Bloqueios e pendências**

- Autenticação HTTP permanece para a etapa 6; nesta etapa foi validada apenas a geração/comparação bcrypt usada pelo futuro login.
- O frontend continua correto com banco vazio e ainda não consome o seed.
- A primeira tentativa de `go test` usou o cache global somente leitura do ambiente; a verificação foi repetida com `GOCACHE=/tmp/bit-project-go-build` e passou.

### E2 — Contrato e infraestrutura HTTP compartilhada — 01/10/2026

**Implementado**

- Contrato OpenAPI 3.0.3 com 10 caminhos e 13 operações; 2 probes marcados como implementados e 11 operações de domínio como planejadas.
- `docs/API.md` com execução real, erros, sessão/cookie, proteção de origem, filtros, paginação e responsabilidades de handlers/services/repositories.
- Envelope `error.code`, `error.message`, `error.fields` opcional e `request_id`; mensagens públicas em português e códigos estáveis.
- Leitura JSON limitada a 1 MiB, `application/json` obrigatório, campos desconhecidos e conteúdo extra rejeitados.
- Parse reutilizável de ID positivo e paginação 1/20 com `page_size` máximo 100 e rejeição de parâmetros repetidos.
- Middleware de request ID, log estruturado, duração/status e recuperação de panic sem expor causa; query string e corpo não entram nos campos de log.
- Guard reutilizável para exigir origem exata em métodos mutáveis e `TRUSTED_ORIGINS` validada como lista de origens HTTP(S) sem caminho.
- Fallback JSON para `/api/*`; rotas planejadas retornam `404 route_not_found`, sem handlers de sucesso simulados.
- Probes mantidos públicos; readiness diferencia indisponibilidade do banco por `503 database_unavailable`.

**Validado**

- `GOCACHE=/tmp/bit-project-go-build go test ./internal/config ./internal/http ./cmd/api` — passou.
- Testes `httptest` cobriram mídia, JSON válido/inválido, campos desconhecidos, conteúdo extra, corpo vazio/acima do limite, IDs, paginação, fields, request ID, logs, panic, causa interna e origem confiável/não confiável.
- PyYAML 6.0.1 carregou `docs/openapi.yaml`; script estrutural resolveu 100 referências internas e confirmou 10 caminhos, 13 operações, 2 implementadas e 11 planejadas.
- Nenhum validador OpenAPI dedicado (`swagger-cli`, Redocly, Spectral ou `yq`) estava instalado; portanto não foi afirmada validação semântica por essas ferramentas.
- `GOCACHE=/tmp/bit-project-go-build make check build` — passou: gofmt, vet, todos os testes Go, ESLint, TypeScript strict e builds Go/Vite.
- API real em `127.0.0.1:18081`, com banco propositalmente inacessível: `/healthz` retornou `200`, `/readyz` retornou `503` seguro e `POST /api/v1/auth/login` retornou `404 route_not_found`.
- Logs reais incluíram os request IDs, métodos, padrões, status e durações dos três pedidos; a senha enviada ao caminho planejado não apareceu. `SIGINT` encerrou graciosamente.

**Arquivos**

- `backend/internal/http/request.go`, `response.go` — entrada e saída HTTP compartilhadas.
- `backend/internal/http/middleware.go`, `origin.go` — request ID, logs, recuperação e proteção de origem.
- `backend/internal/http/*_test.go` — testes de infraestrutura e probes.
- `backend/internal/http/router.go`, `probes.go`, `backend/cmd/api/main.go` — composição do router e logger.
- `backend/internal/config/config.go`, `config_test.go`, `.env.example` — origens confiáveis.
- `docs/openapi.yaml`, `docs/API.md`, `docs/DECISOES.md`, `docs/REQUISITOS.md`, `docs/PROGRESSO.md`, `README.md` — contrato, decisões, execução e rastreabilidade.

**Bloqueios e pendências**

- Login, sessão, metadata, solicitações e dashboard continuam sem handlers e estão explicitamente marcados como planejados.
- O guard de origem será conectado às rotas mutáveis quando cada handler real for registrado; por enquanto foi implementado e testado isoladamente para não transformar rotas ausentes em respostas enganosas.
- Tipos de domínio, services e repositories funcionais serão criados com as primeiras regras/consultas reais; não foram adicionadas camadas vazias nesta etapa.

### E2 — Autenticação e sessões persistentes — 01/10/2026

**Implementado**

- `POST /api/v1/auth/login` com JSON estrito, normalização/validação do username, limite bcrypt de 72 bytes e erro genérico de credenciais.
- Busca real de usuário e comparação bcrypt; usuário inexistente executa comparação dummy para reduzir diferença de tempo observável.
- Token base64url gerado por `crypto/rand` com 32 bytes por padrão; somente SHA-256 hexadecimal é persistido em `sessions`.
- Rotação transacional: cookie anterior válido é revogado e sessão nova inserida na mesma transação; coincidência do novo token com o anterior é rejeitada.
- Cookie `portal_session` com `HttpOnly`, `SameSite=Lax`, `Path=/`, expiração absoluta configurável e `Secure=true` obrigatório em produção.
- Middleware de autenticação por sessão persistida, com identidade mínima no context e diferenciação entre sessão inválida (`401`) e infraestrutura (`500`).
- `GET /api/v1/auth/me` retorna somente `id`, `username` e `display_name`.
- `POST /api/v1/auth/logout` revoga no banco antes de limpar o cookie e é idempotente quando repetido sem cookie.
- Proteção estrita de `Origin` conectada a login/logout; lista de origens continua validada por configuração.
- Rate limit configurável por username normalizado + IP de `RemoteAddr`, sem confiar em headers encaminhados, com teto de entradas e eviction da mais antiga.
- Camadas reais `domain`, `service` e `repository`; handlers não contêm SQL e repository usa consultas parametrizadas/context.
- OpenAPI/API.md, README, decisões, ambiente, requisitos e progresso atualizados para as três rotas executáveis.

**Validado**

- `go test ./internal/service ./internal/http ./internal/config ./cmd/api` — passou com casos de bcrypt, hash, rotação, validação, cookie, `/me`, logout, origem, rate limit e infraestrutura.
- Login real no PostgreSQL com `colaborador1` — `200`; cookie trouxe `HttpOnly`, `SameSite=Lax`, `Path=/`, `Max-Age=28800` e não expôs token/hash no JSON.
- Senha errada — `401 invalid_credentials` genérico; acesso anônimo — `401 authentication_required`; origem hostil — `403 origin_not_allowed` antes do service.
- Consulta agregada ao banco — 1 sessão ativa, todos os hashes com 64 caracteres hexadecimais e expiração fixa de exatamente 8 horas.
- API encerrada e reiniciada — o mesmo cookie continuou autenticando `/auth/me` com `200`, comprovando persistência fora da memória do processo.
- Expiração forçada em banco — `/auth/me` retornou `401`; novo login rotacionou a sessão expirada.
- Logout — `204`, uso posterior do cookie retornou `401`; segundo logout retornou `204`; contagem final da tabela `sessions` foi zero.
- Rate limit real — cinco tentativas do mesmo username/IP chegaram a `401` e a sexta retornou `429`; teste automatizado confirmou que outro username não é bloqueado e `X-Forwarded-For` é ignorado.
- Segunda API com banco propositalmente inacessível — login e `/me` retornaram `500 internal_error`, nunca `401`, sem detalhe SQL/credencial.
- Logs reais continham request ID, método, padrão de rota, status e duração; corpos, cookies, senhas e tokens não apareceram.
- Os dois processos temporários de API foram encerrados graciosamente; PostgreSQL do projeto permaneceu ativo.

**Arquivos**

- `backend/internal/domain/auth.go` — identidade, credenciais internas e sessão.
- `backend/internal/repository/auth.go` — usuários, rotação, consulta e revogação com SQL parametrizado.
- `backend/internal/service/auth.go`, `auth_test.go` — regras de autenticação, tokens, hashes e testes.
- `backend/internal/http/auth.go`, `auth_test.go`, `rate_limit.go` — endpoints, middleware, cookies, CSRF e limiter.
- `backend/internal/config/config.go`, `config_test.go`, `.env.example` — duração, token, cookie e limites.
- `backend/cmd/api/main.go`, `backend/internal/http/router.go`, `middleware.go` — composição e rotas reais.
- `README.md`, `docs/API.md`, `docs/openapi.yaml`, `docs/DECISOES.md`, `docs/REQUISITOS.md`, `docs/PROGRESSO.md` — documentação e rastreabilidade.

**Limitações e pendências**

- O rate limit é por instância e reinicia com o processo; implantação horizontal futura deverá usar armazenamento compartilhado ou controle no gateway.
- `Secure=false` existe somente para HTTP local explícito; produção rejeita essa configuração. HTTPS não foi terminado nesta etapa.
- Frontend ainda não possui tela/estado de sessão. Metadata, solicitações e dashboard seguem sem handlers funcionais e permanecem protegidos pelo fallback autenticado.

### E3 — CRUD individual de solicitações — 02/10/2026

**Implementado**

- `POST /api/v1/requests` autenticado, com JSON estrito apenas para título, descrição e categoria; trim, limites Unicode, categoria fechada, autor da sessão, status `aberto` e datas explícitos no backend.
- `GET /api/v1/requests/{id}` acessível a qualquer usuário autenticado, com DTO seguro, código `SOL-%06d`, rótulos portugueses, timestamps UTC e flags de permissão.
- `PATCH /api/v1/requests/{id}` parcial, distinguindo ausente, `null` e vazio; somente autor/aberta, sem permitir status ou campos automáticos.
- `DELETE /api/v1/requests/{id}` físico, somente autor/aberta, com `204` sem corpo.
- Escritas condicionais atômicas por ID/autor/status e classificação posterior de `404`, `403`, `409` ou `400 no_changes`; não há janela entre autorização e mutação.
- Camadas `domain`, `service`, `repository` e handlers reais, SQL parametrizado/context, composição na API e padrões de rota nos logs.
- Testes de service/HTTP para Unicode, limites, trim, payload estrito, detalhe seguro, PATCH, autoria/estado, inexistência, infraestrutura, autenticação/origem e DELETE.
- API.md/OpenAPI, decisões, README e matriz atualizados. A ausência de controle de versão otimista e a política última gravação vence foram registradas.

**Validado**

- `GOCACHE=/tmp/bit-project-go-cache make fmt check build` — passou: gofmt, vet, todos os testes Go, ESLint, TypeScript strict, builds dos comandos Go e build Vite (26 módulos).
- Após acrescentar o caso explícito de autenticação/origem, `GOCACHE=/tmp/bit-project-go-cache go test ./internal/http ./internal/service` — passou.
- API recompilada e PostgreSQL real: criação Unicode com trim retornou `201`, `Location: /api/v1/requests/6`, autor da sessão e status aberto; outro usuário consultou o recurso com `200` e permissões falsas.
- PATCH parcial persistiu título/categoria e atualizou `updated_at`; `null`, título vazio, `status` no payload e valores inalterados retornaram `400`; outro autor recebeu `403`.
- Após mudança controlada do registro temporário para `concluido`, PATCH e DELETE retornaram `409`. Restaurado para aberto, DELETE por outro autor retornou `403`, pelo autor retornou `204`, e GET posterior retornou `404`.
- Resposta recompilada confirmou `created_at`/`updated_at` em UTC com sufixo `Z`; recurso inexistente retornou `404` e ID inválido `400`.
- Duas sessões temporárias foram revogadas por logout; consultas finais confirmaram `sessions=0` e ausência do registro temporário. Os processos de API foram encerrados graciosamente.
- `docker compose stop postgres` — encerrou somente o container iniciado para a validação, preservando o volume local.
- PyYAML 6.0.1 carregou `docs/openapi.yaml`, resolveu 102 referências e confirmou 10 caminhos/13 operações: 9 implementadas e 4 planejadas. `git diff --check` não apontou erro.

**Arquivos**

- `backend/internal/domain/request.go` — valores estáveis, rótulos e entidade segura.
- `backend/internal/repository/request.go` — criação, detalhe e mutações condicionais no PostgreSQL.
- `backend/internal/service/request.go`, `request_test.go` — validação, autorização por resultado e regras de negócio.
- `backend/internal/http/requests.go`, `requests_test.go`, `router.go`, `middleware.go` — handlers, DTOs, proteção, rotas e logs.
- `backend/cmd/api/main.go` — composição do repository/service de solicitações.
- `.env.example`, `README.md`, `docs/API.md`, `docs/openapi.yaml`, `docs/DECISOES.md`, `docs/REQUISITOS.md`, `docs/PROGRESSO.md` — configuração, uso e rastreabilidade.

**Limitações e pendências**

- `GET /api/v1/requests`, metadata e dashboard continuam planejados e não simulam sucesso; a mudança de status foi implementada na etapa registrada a seguir.
- Não há versão otimista: duas edições simultâneas do mesmo autor enquanto a solicitação segue aberta usam última gravação vence. Autoria/status permanecem protegidos atomicamente.
- O frontend ainda não consome o CRUD.

### E3 — Transições de status — 02/10/2026

**Implementado**

- `PATCH /api/v1/requests/{id}/status` autenticado e protegido por origem confiável, aceitando exclusivamente `aberto`, `em_atendimento` ou `concluido`.
- Payload estrito: campo ausente, `null`, tipo inválido, valor desconhecido e propriedades adicionais são rejeitados com `400`.
- Política explícita de transições livres: qualquer usuário autenticado pode alterar qualquer solicitação, inclusive reabrir, sem papéis ou sequência obrigatória.
- Repetição do status é idempotente e preserva `updated_at`; mudança efetiva atualiza apenas `status` e `updated_at`.
- Repository transacional com `SELECT ... FOR UPDATE`, coordenado com UPDATE/DELETE condicionais do CRUD; exclusão já concluída resulta em `404`.
- DTO reutilizado com rótulos `Aberto`, `Em Atendimento` e `Concluído` e permissões recalculadas após cada transição.
- Testes de service/HTTP cobrem matriz 3×3, usuário diferente, validação, sessão, origem, inexistência e permissões após fechamento/reabertura.
- Contrato, decisões, exemplos HTTP, README e matriz de requisitos atualizados.

**Validado**

- `GOCACHE=/tmp/bit-project-go-cache go test ./internal/service ./internal/http ./internal/repository ./cmd/api` — passou após a implementação inicial.
- `GOCACHE=/tmp/bit-project-go-cache make fmt check build` — passou ao final: gofmt, vet, toda a suíte Go, ESLint, TypeScript strict, builds Go e Vite (26 módulos).
- Após endurecer a rejeição do ano zero, `GOCACHE=/tmp/bit-project-go-cache go test ./internal/service ./internal/http ./internal/repository` e `git diff --check` — passaram.
- Após reforçar o teste stateful da matriz e da idempotência, `GOCACHE=/tmp/bit-project-go-cache go test ./internal/service ./internal/http` — passou.
- PostgreSQL real e API em `127.0.0.1:18085`: as nove combinações entre os três estados retornaram `200`, usando também `colaborador2` sobre solicitação de `colaborador1`.
- Repetições em `aberto`, `em_atendimento` e `concluido` preservaram exatamente o `updated_at`; mudanças efetivas produziram timestamp posterior e rótulo correto.
- Em `em_atendimento`, PATCH de conteúdo e DELETE pelo autor retornaram `409 request_not_open`; depois da reabertura, PATCH retornou `200` e DELETE retornou `204`.
- Campo ausente, `null`, número, `cancelado` e campo extra retornaram `400`; anônimo retornou `401`, origem hostil `403` e tentativa sobre o recurso já excluído `404`.
- As duas sessões de validação foram revogadas (`204`); banco confirmou `sessions=0` e ausência da solicitação temporária de ID 7.
- API recebeu `SIGINT` e encerrou graciosamente; `docker compose stop postgres` encerrou o container preservando o volume.

**Arquivos**

- `backend/internal/repository/request.go` — transação, lock e atualização idempotente/efetiva.
- `backend/internal/service/request.go`, `request_test.go` — validação, política e matriz de transições.
- `backend/internal/http/requests.go`, `requests_test.go`, `router.go`, `middleware.go` — handler, rota, segurança, respostas e log por padrão.
- `README.md`, `.env`, `.env.example`, `docs/API.md`, `docs/openapi.yaml`, `docs/DECISOES.md`, `docs/REQUISITOS.md`, `docs/PROGRESSO.md` — execução e rastreabilidade.

**Pendências**

- Interface de mudança de status no frontend ainda não foi implementada.
- Listagem/filtros, metadata e dashboard continuam planejados.

### E3 — Listagem e filtros de solicitações — 02/10/2026

**Implementado**

- `GET /api/v1/requests` autenticado, global e persistido no PostgreSQL, sem filtro implícito por solicitante.
- DTO resumido com ID/código, título, categoria/rótulo, solicitante legível, abertura RFC3339 UTC e status/rótulo; nenhum hash, sessão, descrição ou campo interno é carregado/exposto.
- Filtros opcionais `date_from`, `date_to`, `category`, `status` e `q`, combinados por `AND`; vazios equivalem a ausência.
- Dias civis em `America/Sao_Paulo`, com início inclusivo e início do dia seguinte exclusivo calculados pelo calendário, sem soma cega de 24 horas.
- Busca somente no título por `ILIKE`, case-insensitive, com trim/limite de 150 pontos de código e escape literal de `%`, `_` e barra invertida.
- Paginação 1/20, máximo 100, resposta `{items, pagination: {page, page_size, total_items, total_pages}}` e lista vazia além do fim.
- Ordem fixa `created_at DESC, id DESC`; parâmetros desconhecidos/repetidos e inteiros/datas/valores inválidos retornam `400 invalid_query_parameter`.
- Count e página compartilham a mesma cláusula SQL estática e argumentos parametrizados em transação somente leitura `REPEATABLE READ`.
- Testes de service, HTTP e repository para calendário/fuso, validações, DTO, paginação e escape de padrão.

**Validado**

- `GOCACHE=/tmp/bit-project-go-cache go test ./internal/service ./internal/http ./internal/repository` — passou durante o desenvolvimento.
- `GOCACHE=/tmp/bit-project-go-cache make fmt check build` — passou ao final: gofmt, vet, toda a suíte Go, ESLint, TypeScript strict, builds Go e Vite (26 módulos).
- Teste de calendário confirmou `2019-02-16` em São Paulo como intervalo UTC de 25 horas (`02:00Z` até `03:00Z` do dia seguinte), demonstrando limite civil em vez de `+24h`.
- API real em `127.0.0.1:18086` e PostgreSQL: filtro combinado de um dia/categoria/status retornou IDs 9 e 8 na ordem esperada; período de um dia para título duplicado incluiu ID 10 em `02:59:59Z` e excluiu ID 11 em `03:00:00Z`.
- Somente `date_from` retornou ID 11; somente `date_to` retornou ID 10. Categoria `compras` retornou IDs 13/12 e status `concluido` retornou ID 10.
- `q=RELATÓRIO` encontrou IDs 9/8 sem distinguir caixa; `q=100%_d'água` encontrou somente o título literal ID 8, não o candidato com caracteres nas posições dos curingas.
- Dois títulos `Ordem empatada` com o mesmo `created_at` foram preservados e ordenados como IDs 13/12; páginas 1/2 com tamanho 1 retornaram um ID cada, e página 99 retornou `items=[]` mantendo total 2/páginas 2.
- Filtros vazios retornaram defaults e total 11 durante o cenário. Data impossível, intervalo invertido, parâmetro repetido, inteiro inválido, limite 101, categoria/status desconhecidos e parâmetro `sort` retornaram `400`; anônimo retornou `401`.
- PyYAML 6.0.1 carregou o OpenAPI, resolveu 105 referências e confirmou 10 caminhos/13 operações: 11 implementadas e 2 planejadas.
- Logout retornou `204`; seis registros claramente marcados foram removidos e consultas finais confirmaram `sessions=0` e `temporary_requests=0`. API encerrou por `SIGINT` e PostgreSQL foi parado preservando o volume.

**Arquivos**

- `backend/internal/repository/request.go`, `request_test.go` — consulta/count parametrizados, snapshot, ordenação e escape do `ILIKE`.
- `backend/internal/service/request.go`, `request_test.go` — filtros, calendário de São Paulo, paginação e validações.
- `backend/internal/http/request.go`, `request_test.go`, `requests.go`, `requests_test.go`, `router.go`, `middleware.go` — query HTTP, DTO, resposta, autenticação e logs.
- `README.md`, `docs/API.md`, `docs/openapi.yaml`, `docs/DECISOES.md`, `docs/REQUISITOS.md`, `docs/PROGRESSO.md` — contrato, exemplos e rastreabilidade.

**Pendências**

- A interface de listagem/filtros no frontend ainda não foi implementada.
- Metadata e dashboard continuam planejados.
- `ILIKE` permanece sensível a acentos diferentes e não usa índice para substring; `pg_trgm` só será considerado com necessidade comprovada.

### E3 — Dashboard global — 02/10/2026

**Implementado**

- `GET /api/v1/dashboard` protegido pela sessão existente, com resposta estável `data.total`, `data.abertas`, `data.em_atendimento` e `data.concluidas`.
- Camadas próprias de handler, service, repository e domínio, conectadas na composição da API e no nome estruturado da rota para logs.
- Uma única instrução PostgreSQL calcula `COUNT(*)` e as três agregações `FILTER`, sem filtro de solicitante, filtros da listagem ou cache.
- O service verifica valores não negativos e `total = abertas + em_atendimento + concluidas`; falha ou inconsistência resulta em erro interno, nunca zeros falsos.
- Contrato e exemplos para base vazia/populada atualizados em API.md e OpenAPI. A futura tela continua explicitamente pendente.

**Validado**

- `GOCACHE=/tmp/bit-project-go-cache go test ./...` — toda a suíte Go passou. A primeira tentativa sem `GOCACHE` falhou porque o cache global do ambiente estava somente leitura; a repetição em `/tmp` resolveu sem alterar o projeto.
- `GOCACHE=/tmp/bit-project-go-cache make fmt check build` — passou: gofmt, vet, toda a suíte Go, ESLint, TypeScript strict, builds dos comandos Go e build Vite com 26 módulos.
- `python3 -c ...` com PyYAML — OpenAPI 3.0.3 carregado: 10 caminhos, 13 operações, 12 implementadas e 105 referências.
- PostgreSQL 17.6 isolado no projeto Compose `bit-project-dashboard-test`, porta `55434`: migrations `up` e seed criaram 2 usuários/5 solicitações; as cinco solicitações seed foram removidas para iniciar o cenário vazio.
- Sem cookie, o dashboard retornou `401 authentication_required`; com sessão válida e nenhuma solicitação, retornou `0/0/0/0`.
- Três criações reais produziram `3/3/0/0`; transições para `em_atendimento` e `concluido` produziram `3/1/1/1`; exclusão da aberta produziu `2/0/1/1`.
- Uma nova solicitação criada por `colaborador2` foi imediatamente vista pela sessão de `colaborador1`, produzindo `3/1/1/1` e confirmando escopo global.
- Com o PostgreSQL parado, a chamada autenticada retornou `500 internal_error` e não uma resposta zerada. Os testes de handler também confirmaram que o detalhe do erro de infraestrutura não vaza no JSON.
- API encerrou graciosamente por `SIGINT`; `docker compose -p bit-project-dashboard-test down -v --remove-orphans` removeu container, rede e volume exclusivos da validação.

**Arquivos**

- `backend/internal/domain/dashboard.go` — estrutura dos quatro contadores.
- `backend/internal/repository/dashboard.go` — agregação PostgreSQL em uma consulta.
- `backend/internal/service/dashboard.go`, `dashboard_test.go` — orquestração, invariável e propagação de falhas.
- `backend/internal/http/dashboard.go`, `dashboard_test.go`, `router.go`, `middleware.go` — DTO, handler, rota autenticada e cobertura HTTP.
- `backend/cmd/api/main.go` — composição das novas camadas.
- `README.md`, `docs/API.md`, `docs/openapi.yaml`, `docs/DECISOES.md`, `docs/REQUISITOS.md`, `docs/PROGRESSO.md` — contrato, decisões e evidências.

**Pendências**

- A tela do dashboard no frontend ainda não foi implementada; somente o endpoint backend está pronto e validado.
- `GET /api/v1/metadata` permanece planejado.

### E6 — Revisão e testes concentrados do backend — 02/10/2026

**Implementado**

- Suíte PostgreSQL reproduzível em `backend/internal/integration`, separada dos testes unitários de services e dos testes HTTP com `httptest` já existentes.
- Aplicação das migrations Goose reais e montagem de router/services/repositories reais; stubs permanecem restritos às regras isoladas que eles de fato verificam.
- Proteção destrutiva em três níveis: variável exclusiva `TEST_DATABASE_URL`, nome do banco terminado em `_test` e confirmação exata `TEST_DATABASE_ALLOW_RESET=yes`. A trava possui teste próprio.
- Compose dedicado `compose.test.yaml` com PostgreSQL 17.6, banco `portal_test`, porta 55435 e armazenamento `tmpfs`; alvos `db-test-up`, `test-integration` e `db-test-down` no Makefile.
- Integrações para autenticação, expiração/revogação/persistência, autorização com dois usuários, CRUD e campos automáticos, mass assignment, status/reabertura, filtros/LIKE/paginação, SQL tratado como texto, dashboard, erro de banco e disputa controlada entre status e edição.
- Estratégia, comandos, matriz de cenários e limitações registrados em `docs/TESTES.md`; requisitos receberam matriz nominal de evidências automatizadas.

**Revisado**

- `http.Server` mantém `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout` e `IdleTimeout`, com shutdown limitado por contexto.
- Comandos fecham `sql.DB`; repositories fecham/inspecionam rows, fazem rollback defensivo e verificam `RowsAffected` na exclusão.
- Escritas usam parâmetros. Edição/exclusão condicionam autor/status na própria instrução; mudança de status usa transação e `FOR UPDATE`.
- Logs usam padrão de rota e tipo do erro, sem corpo, query string, cookie, senha ou SQL detalhado enviado ao cliente.

**Validado**

- `make fmt` — formatou todos os fontes Go.
- `GOCACHE=/tmp/bit-project-go-cache go vet ./...` — passou sem diagnósticos.
- `GOCACHE=/tmp/bit-project-go-cache go test ./...` — passou; sem `TEST_DATABASE_URL`, somente cenários dependentes do PostgreSQL são pulados com causa explícita e a trava unitária continua ativa.
- `TEST_DATABASE_URL=postgres://portal_test:...@127.0.0.1:55435/portal_test?... TEST_DATABASE_ALLOW_RESET=yes go test -count=1 -v ./internal/integration` — passou: 5 testes principais, incluindo 4 casos da trava e migrations reais até a versão 2.
- `make test-integration` com as mesmas variáveis — passou em 1,620 s, confirmando o alvo documentado.
- A integração confirmou credencial inválida genérica, `/me` seguro, sessão após reconstrução do servidor, expiração e logout; dois usuários para autoria; reabertura; filtros combinados por dia de São Paulo; `%`, `_`, apóstrofo e aparência de SQL persistidos como texto; tabela de usuários intacta; contadores e erros de banco preservando semântica.
- `TestConditionalMutationLosesRaceToStatusChange` segurou lock de linha, iniciou edição concorrente, concluiu a solicitação e confirmou `repository.ErrConflict` com título preservado.
- Após adicionar reconexão por um novo `sql.DB`, a primeira execução completa apresentou um `500` isolado na reabertura. O cenário isolado passou em seguida; autenticação+CRUD foram repetidos 10 vezes e a suíte completa 5 vezes (15 repetições relacionadas, todas aprovadas). A ocorrência não foi ocultada e não voltou a ser reproduzida.
- O detector de corrida foi tentado em `internal/service` + `internal/http` e em `internal/integration`, cada um com `timeout 180s`; ambos terminaram com exit 124 ainda na compilação, sem resultado de testes ou diagnóstico de corrida.

**Arquivos**

- `backend/internal/integration/backend_integration_test.go` — suíte real e trava de banco.
- `compose.test.yaml`, `Makefile`, `.env.example`, `README.md` — ambiente e comandos reproduzíveis.
- `docs/TESTES.md`, `docs/DECISOES.md`, `docs/REQUISITOS.md`, `docs/PROGRESSO.md` — estratégia, decisões, matriz e evidências.

**Pendências**

- Reexecutar `go test -race` em CI ou Linux nativo com tempo de compilação adequado; não foi marcado como aprovado nesta revisão.
- Metadata e frontend funcional continuam fora desta etapa, portanto não há cobertura de fluxo para eles.

### E4 — Fundação definitiva do frontend — 02/10/2026

**Implementado**

- Estrutura simples em `app/routes`, `pages`, `components`, `features/auth`, `api`, `services`, `types` e `styles`, sem biblioteca visual, estado global externo ou cliente de consultas.
- React Router com `/login`, redirecionamento `/` → `/dashboard`, `/dashboard`, `/solicitacoes`, `/solicitacoes/nova`, `/solicitacoes/:id`, `/solicitacoes/:id/editar` e fallback 404.
- `AuthProvider` consulta `/auth/me`, cancela a leitura ao desmontar, protege rotas e reage centralmente a `401`. O shell mostra o nome seguro e executa logout real; token não é acessível nem persistido por JavaScript.
- Cliente `fetch` central usa `/api/v1` relativo, cookies same-origin, JSON apenas quando há corpo e nenhum retry automático. Trata `204`, envelope/campos, resposta não JSON, rede, cancelamento e status 401/403/404/409/415/429/500/503.
- Services tipados para autenticação, metadata, solicitações e dashboard; todas as operações aceitam `AbortSignal`. O endpoint metadata permanece sem chamada enquanto o backend estiver planejado.
- Tipos de `User`, `Request`, item/lista paginada, categoria, status, dashboard, metadata, comandos e erros seguem OpenAPI. Badges recebem rótulos da API e não duplicam mapas no frontend.
- Componentes reais de layout/navegação, cabeçalho, campo acessível, mensagens, loading/erro/vazio, badges e confirmação. CSS próprio fornece contraste, foco visível e adaptação em 760/480 px.
- Vitest 5.0.3 adicionado apenas para o cliente HTTP. `make check` agora inclui os testes frontend.

**Decisões**

- O navegador, por fetch same-origin, fornece `Origin` nas mutações; o código não tenta definir manualmente o cabeçalho protegido.
- Sem biblioteca de consultas nesta etapa. Páginas futuras manterão carga local cancelável e, após mutação bem-sucedida, recarregarão explicitamente ou navegarão usando a representação retornada.
- Placeholders declaram que não carregam dados nem executam fluxos e não contam como telas implementadas.

**Validado**

- `npm run test` — 1 arquivo e 8 casos Vitest passaram: JSON/cookie, `204`, campos/401, 403, 404, 409, erro não JSON/rede sem repetição e abort.
- `npm run typecheck` — TypeScript strict passou.
- `npm run lint` — ESLint passou sem avisos.
- `npm run build` — Vite 8.3.2 gerou produção com 44 módulos, CSS 5,68 kB e JavaScript 322,54 kB.
- `npm ci && npm run test` — reinstalação exata do lockfile adicionou 139 pacotes, auditou 140 sem vulnerabilidades e repetiu os 8 casos com sucesso.
- `GOCACHE=/tmp/bit-project-go-cache make check` — passou após incluir Vitest no alvo: gofmt/vet/suíte Go, 8 testes frontend, ESLint e TypeScript strict.

**Arquivos**

- `frontend/src/api`, `services`, `types` — transporte e contrato.
- `frontend/src/app/routes`, `features/auth` — roteamento, sessão e proteção.
- `frontend/src/components`, `pages`, `styles` — shell, elementos reutilizáveis, placeholders e responsividade base.
- `frontend/package.json`, `package-lock.json`, `Makefile`, `README.md` — teste, dependência fixada e execução.
- `docs/DECISOES.md`, `docs/REQUISITOS.md`, `docs/TESTES.md`, `docs/PROGRESSO.md` — decisões e estado real.

**Pendências**

- Cards do dashboard, lista/filtros, criação, detalhe/status, edição e exclusão ainda não foram implementados no frontend.
- `GET /api/v1/metadata` ainda não existe no backend; formulários não devem inventar suas opções enquanto essa dependência estiver pendente.
- Validação visual em navegador e evidências responsivas devem ocorrer quando as telas reais substituírem os placeholders.

### E4 — Autenticação completa do frontend — 02/10/2026

**Implementado**

- Página de login real com labels, `autocomplete=username/current-password`, validação de preenchimento, erros por campo/API, estado `Entrando…` e bloqueio de submissão duplicada. Senha permanece somente no estado transitório e é apagada após sucesso/desmontagem.
- Contexto com fases `checking`, `authenticated` e `unauthenticated`; `/auth/me` é concluído antes de renderizar ou redirecionar conteúdo protegido. Falha de rede/500 mantém erro de verificação e botão de retry.
- Login e `/me` desabilitam o handler global de `401`: credencial inválida aparece no formulário e ausência inicial de cookie não é chamada de expiração. `401` em outra operação limpa identidade e redireciona com mensagem de sessão expirada.
- Destino pós-login validado centralmente: somente `/dashboard` e `/solicitacoes...` locais são aceitos; URL absoluta, `//host` e rota externa caem no dashboard.
- Logout só limpa memória depois da resposta bem-sucedida. Falha mantém a área autenticada, mostra erro e permite repetição; sucesso protege histórico/URL direta e informa encerramento.
- Nome legível permanece no shell. Não foram adicionados cadastro, recuperação de senha, token em storage ou logs de credenciais.
- Corrigidas duas condições de corrida descobertas pelos testes: guard público agora também restaura o destino anterior, e confirmação de logout pertence ao contexto em vez de depender da ordem de navegação do layout.

**Testes com API simulada**

- Testing Library 16.3.3, user-event 14.6.7 e jsdom 29.1.1 adicionados como dependências de desenvolvimento. jsdom 30.1.1 foi rejeitado porque exige Node 24.15+, acima do 24.14 fixado.
- `npm run test` — 2 arquivos/20 casos passaram após as correções: cliente HTTP e login válido/inválido, preenchimento, duplicação, destino interno, sessão ao recarregar, acesso direto, logout e falha, expiração, indisponibilidade/retry e bloqueio de redirect externo.
- Uma rodada intermediária falhou em três asserções de destino/mensagem e revelou as duas corridas citadas; as falhas não foram ocultadas e a implementação foi ajustada antes da repetição aprovada.

**Verificação com API real**

- PostgreSQL descartável `portal_test` na porta 55436 recebeu migrations v2 e seed de 2 usuários/5 solicitações; API real executou em 8080 e Vite em 5173.
- Chamadas através do proxy Vite confirmaram: `/me` anônimo `401`; login válido `200`; `/me` com cookie `200` e usuário seguro; logout `204`; `/me` após logout `401`; senha incorreta `401 invalid_credentials`.
- Essa evidência valida proxy, cookie e contrato real, mas não é chamada de teste E2E visual. Roteamento/formulário foram exercitados em jsdom; navegador automatizado permanece melhoria da etapa final.
- API encerrou graciosamente, Vite foi interrompido, cookie temporário e ambiente Compose descartável foram removidos.

**Verificações**

- `npm run typecheck` — TypeScript strict passou.
- `npm run lint` — ESLint passou sem avisos.
- `npm run build` — Vite 8.3.2 passou com 46 módulos, CSS 5,77 kB e JavaScript 325,89 kB.
- A repetição de `npm ci` diretamente no volume WSL `/mnt/d` falhou com `ENOTEMPTY` durante operações internas em `node_modules`; a árvore parcial foi descartada como evidência. Uma cópia exata dos fontes, configurações, `package.json` e `package-lock.json` foi validada no filesystem Linux em `/tmp`: `npm ci --no-audit --no-fund` instalou 192 pacotes; `npm run test` passou com 2 arquivos/20 casos; tipos, lint e build passaram novamente com os mesmos 46 módulos e tamanhos acima.
- Nessa execução limpa, o primeiro teste de interação excedeu por 0,48 s o timeout padrão de 5 s, enquanto os outros 19 passaram. Foi aplicado timeout local de 10 s apenas ao cenário que simula digitação e duas submissões; a repetição completa passou em 10,56 s, sem alterar o timeout global.

**Arquivos**

- `frontend/src/pages/LoginPage.tsx`, `app/routes/destination.ts` — formulário e destino seguro.
- `frontend/src/features/auth/AuthContext.tsx`, `AuthFlow.test.tsx` — estado, expiração, login/logout e testes de fluxo.
- `frontend/src/app/routes/ProtectedRoute.tsx`, `PublicOnlyRoute.tsx` — guards durante verificação e navegação.
- `frontend/src/api/client.ts`, `client.test.ts`, `services/auth.ts` — distinção de `401` e transporte.
- `frontend/package.json`, `package-lock.json`, estilos e documentação — dependências fixadas, interface e rastreabilidade.

**Pendências**

- Teste E2E em navegador real ainda não foi adotado; a separação entre jsdom e HTTP real está documentada.
- Dashboard e solicitações continuam placeholders, e metadata continua pendente no backend.

### E5 — Listagem, filtros e metadata — 02/10/2026

**Implementado**

- `GET /api/v1/metadata` protegido, com cinco categorias e três status em ordem estável a partir dos valores/rótulos do domínio; sem consulta de banco ou dados sensíveis.
- Tela real de solicitações consumindo metadata e `GET /requests`, com código, título, categoria, solicitante legível, abertura em `America/Sao_Paulo`, status textual e link identificado para a rota de detalhes.
- Filtros explícitos por período de abertura, categoria, status e texto no título. Datas permanecem strings civis `YYYY-MM-DD`; intervalo invertido é rejeitado também na interface.
- URL como estado aplicado de filtros/página, preservando reload, histórico e compartilhamento. Aplicar/limpar retorna à página 1; digitação não dispara consultas.
- Paginação conforme os metadados da API, cancelamento com `AbortController` e guarda contra resposta antiga. Estados distintos para carga inicial/atualização, erro/retry, base vazia, filtro sem resultados e página além do fim.
- Tabela rolável por teclado em telas estreitas, sem remover colunas; badges mantêm texto além da cor e controles possuem labels/foco do sistema visual existente.

**Testes simulados do frontend**

- `npm run test` na cópia exata em `/tmp` — 5 arquivos/29 testes passaram. Os novos casos cobrem serialização de URL sem conversão de data, intervalo invertido, campos/rótulos, abertura `2026-10-02T02:30:00Z` exibida como `01/10/2026 23:30` em São Paulo, submissão/limpeza/página 1, voltar do navegador, cancelamento/ordem de resposta, estados vazio/erro/página excedente e detalhe/ID inválido.
- A primeira rodada teve 25 aprovações e 2 falhas: nome acessível do link concatenado e timeout do cenário composto. O link recebeu `aria-label` explícito; somente o teste de quatro estados recebeu timeout de 10 s. Após incluir detalhe e estados dos botões de paginação, a suíte final de 29 casos passou em 22,77 s.
- Esses testes usam services simulados para observar a UI e não são apresentados como integração real.

**Verificação com API e PostgreSQL reais**

- Compose descartável `bit-project-list-frontend`, PostgreSQL 17.6 em `portal_test`/porta 55437, migrations v2 e seed: 2 usuários e 5 solicitações.
- API temporária em `127.0.0.1:18081`: prontidão `200`; login dos dois usuários demo `200`; metadata `200` com 5 categorias/3 status.
- Ambos os usuários receberam 5 solicitações contendo 2 solicitantes distintos. Filtro combinado `2026-10-01..02 + infraestrutura + em_atendimento + iluminação` retornou somente `SOL-000005`; página 99 retornou 0 itens preservando `page=99`, `total_items=5`, `total_pages=1`.
- O caso de data próximo à meia-noite foi validado na apresentação pelo teste frontend e os limites civis/UTC permanecem cobertos pela integração PostgreSQL do backend; não foi chamado de navegador E2E.

**Verificações de qualidade**

- Primeira tentativa Go sem `GOCACHE` falhou antes dos testes porque o cache padrão é somente leitura; repetição com `GOCACHE=/tmp/bit-project-go-cache` passou.
- `go vet ./...` e `go test ./...` — passaram; HTTP incluiu os novos testes de metadata autenticada/anônima.
- `npm run typecheck` e `npm run lint` — passaram.
- `npm run build` — Vite 8.3.2 passou com 52 módulos, CSS 8,00 kB e JavaScript 338,04 kB.
- Uma invocação de `npm run test` feita por engano no workspace falhou imediatamente com `vitest: not found`, pois o `node_modules` parcial havia sido removido intencionalmente. O mesmo arquivo já sincronizado foi executado na cópia reproduzível em `/tmp`, onde os 29 casos passaram; não foi registrada como falha da aplicação.

**Arquivos**

- `backend/internal/domain/request.go`, `internal/http/metadata.go`, `router.go`, `middleware.go`, `probes_test.go` — metadata real, proteção, logging e testes.
- `frontend/src/pages/RequestsPage.tsx`, `RequestDetailPage.tsx`, `features/requests` e respectivos testes — lista/detalhe, URL, filtros, paginação, fuso e cobertura.
- `frontend/src/styles/base.css`, `components.css` — região rolável, formulário, paginação e utilitário acessível.
- `README.md`, `docs/API.md`, `openapi.yaml`, `DECISOES.md`, `REQUISITOS.md`, `TESTES.md`, `PROGRESSO.md` — contrato e evidências.

**Pendências**

- Alteração de status, criação, edição e exclusão continuam pendentes no frontend; nenhum botão simulado foi incluído no detalhe.
- Não houve automação em navegador real nem inspeção visual responsiva; comportamento React foi validado em jsdom e contrato/dados por HTTP real.

### E5 — Formulários de criação e edição — 03/10/2026

**Implementado**

- Formulário compartilhado para título, descrição e categoria, com labels, contadores, erros por campo, opções reais de metadata e informação explícita de que solicitante/data/status são automáticos.
- Validação compartilhada após trim: título 3–150 e descrição 10–5000 por pontos de código Unicode (`Array.from`), coerente com `utf8.RuneCountInString` do backend; categoria deve existir na metadata.
- Criação envia exclusivamente `title`, `description` e `category`, bloqueia submissão duplicada, preserva campos em falha e navega ao detalhe retornado com confirmação.
- Edição carrega recurso/metadata por ID, exige `permissions.can_edit`, diferencia outro autor de status fechado e monta PATCH apenas com campos alterados. O backend continua autoridade: `403`, `404` e `409` são tratados após o envio sem apagar o formulário.
- Conflito `409` informa mudança de status e oferece link para recarregar detalhes. `401` permanece centralizado no contexto de autenticação; falha de rede mantém os valores e permite nova tentativa.
- Cancelar não chama mutação, pede confirmação quando há alterações e `beforeunload` cobre fechamento/reload. Quebras de linha permanecem texto em textarea/detalhe, sem `dangerouslySetInnerHTML`.
- Detalhe mostra ação de edição somente quando autorizada, preserva o retorno à lista filtrada e apresenta feedback de criação/edição. Lista/detalhe são recarregados ao visitar porque não existe cache global; o dashboard futuro seguirá a mesma leitura atual da API.

**Testes frontend com API simulada**

- `requestForm.test.ts`: trim, Unicode/emoji, limites exatos 150/5000, excesso, vazios e categoria desconhecida.
- `RequestForms.test.tsx`: criação válida com payload mínimo, campos vazios, envio duplicado, rede com conteúdo preservado, cancelamento sem mutação, edição do autor com PATCH parcial, URL direta de outro usuário/status fechado e respostas de servidor `403/409` preservando a alteração.
- Suíte final: `npm run test` passou com 7 arquivos e 40 testes em 24,12 s.
- Esses testes verificam a experiência React, mas não são apresentados como persistência nem autorização real.

**Autoria e PostgreSQL**

- A tentativa de subir `bit-project-forms-frontend` na porta 55438 não iniciou nada: o WSL informou `docker: command not found` e recomendou habilitar a integração do Docker Desktop. `psql`/`pg_isready` também não estão disponíveis e nenhuma porta PostgreSQL local foi encontrada.
- `go test -v -run TestRequestsAuthorizationStateFiltersAndDashboard ./internal/integration` foi executado e marcou o caso como `SKIP` por `TEST_DATABASE_URL` ausente; isso não foi contado como nova validação real.
- A regra não é considerada validada por botão oculto. A evidência real anterior permanece a integração PostgreSQL já registrada: dois usuários produziram `403` para edição por outro autor e `409` para o autor após fechamento, além do teste concorrente do repository. Repetir o fluxo real desta etapa permanece pendente quando Docker/PostgreSQL voltar a estar disponível.

**Verificações de qualidade**

- `npm ci --no-audit --no-fund` em cópia exata no filesystem Linux `/tmp` — 192 pacotes instalados pelo lockfile em 20 s.
- A primeira checagem de tipos encontrou import não usado e estreitamento nulo em três acessos; ambos foram corrigidos. Uma invocação posterior feita no workspace retornou `tsc: not found` porque `node_modules` não é mantido ali; a execução correta em `/tmp` passou.
- `npm run typecheck` e `npm run lint` — passaram.
- `npm run build` — Vite 8.3.2 passou com 56 módulos, CSS 8,42 kB e JavaScript 346,09 kB.
- `GOCACHE=/tmp/bit-project-go-cache go test ./internal/http ./internal/service` — passou; nenhum código backend foi alterado nesta etapa.

**Arquivos**

- `frontend/src/features/requests/RequestForm.tsx`, `requestForm.ts`, `useRequestMetadata.ts`, `useUnsavedChangesWarning.ts` — campos, validação e hooks compartilhados.
- `frontend/src/pages/RequestCreatePage.tsx`, `RequestEditPage.tsx`, `RequestDetailPage.tsx` — fluxos reais, feedback e navegação.
- `frontend/src/features/requests/requestForm.test.ts`, `RequestForms.test.tsx` — validações e cenários de formulário.
- `frontend/src/components/feedback/ErrorState.tsx`, `styles/components.css` — ação contextual e layout responsivo.
- `README.md`, `docs/DECISOES.md`, `REQUISITOS.md`, `TESTES.md`, `PROGRESSO.md` — estado e evidências.

**Pendências**

- Repetir criação/edição/403/409 por HTTP contra PostgreSQL descartável quando Docker ou `TEST_DATABASE_URL` estiver disponível.
- Alteração de status, exclusão e dashboard continuam pendentes no frontend. Bloqueio geral de toda navegação SPA com mudanças não salvas não foi adicionado; Cancelar e unload estão protegidos.

### E5 — Detalhes, status e exclusão no frontend — 03/10/2026

**Implementado**

- Detalhe por URL consulta `GET /api/v1/requests/{id}` e apresenta código, título, descrição textual com quebras de linha, categoria, status, solicitante, criação e atualização em `America/Sao_Paulo`.
- Status usa as opções reais de metadata e `PATCH /api/v1/requests/{id}/status`; não há atualização otimista. A representação devolvida substitui a anterior e recalcula `can_edit`/`can_delete`, inclusive ao fechar e reabrir.
- Todo usuário autenticado recebe o controle de status. Editar/excluir aparecem somente quando as flags do backend permitem; o frontend não substitui a autorização do servidor.
- Exclusão usa `DELETE` real e diálogo modal nomeado, com foco inicial no cancelamento, Escape, retorno de foco e indicação de código/título. Cancelar não chama a API e sucesso navega para a lista preservada com uma confirmação.
- `403`/`409` de exclusão fecham o diálogo e recarregam o detalhe; `404` distingue recurso removido; falha de rede permanece no diálogo para repetição. Carga inicial também diferencia inexistência de infraestrutura, enquanto `401` continua tratado pelo fluxo central de autenticação.
- Layout de gerenciamento e modal foi adaptado a telas estreitas. A descrição continua renderizada por interpolação React, sem `dangerouslySetInnerHTML`.

**Testes e verificações efetivamente executados**

- `npm run typecheck` na cópia exata com dependências em `/tmp/bit-project-frontend-validation` — passou.
- `npm run lint` — passou sem warnings.
- `npm test` — 7 arquivos/47 testes passaram em 48,80 s. Os 10 casos de detalhe cobrem URL direta, datas, quebras/HTML semelhante a texto, fechamento/reabertura, outro usuário, bloqueio de envio simultâneo, modal/Cancelar/Escape/foco, exclusão confirmada, recarga em `403/409`, remoção concorrente `404`, infraestrutura e ID inválido.
- `npm run build` — Vite 8.3.2 passou com 57 módulos; CSS 9,42 kB e JavaScript 350,72 kB.
- `GOCACHE=/tmp/bit-project-go-cache go test ./internal/http ./internal/service` e `go vet ./...` — passaram.
- Testes backend direcionados, sem cache, para DELETE/status/permissões/transições passaram: pacote HTTP em 0,329 s e service em 0,195 s.
- A tentativa direcionada de integração PostgreSQL executou, mas `TestRequestsAuthorizationStateFiltersAndDashboard` e `TestConditionalMutationLosesRaceToStatusChange` foram marcados `SKIP` porque `TEST_DATABASE_URL` não está definido. Esse resultado não foi contado como nova validação PostgreSQL.

**Arquivos**

- `frontend/src/pages/RequestDetailPage.tsx`, `RequestsPage.tsx` — leitura, status, exclusão, conflitos, navegação e feedback.
- `frontend/src/components/confirmation/ConfirmDialog.tsx` — acessibilidade, foco, Escape, estado ocupado e erro recuperável.
- `frontend/src/features/requests/RequestDetailPage.test.tsx` — cenários observáveis de detalhe e mutações.
- `frontend/src/styles/components.css` — controles, diálogo e responsividade.
- `README.md`, `docs/DECISOES.md`, `REQUISITOS.md`, `TESTES.md`, `PROGRESSO.md` — estado e evidências atualizados.

**Bloqueios e pendências**

- A UI foi exercitada em jsdom com services simulados; não houve automação em navegador real nesta etapa.
- A integração PostgreSQL desta rodada ficou bloqueada por ausência de `TEST_DATABASE_URL`/Docker no ambiente. A evidência real anterior do backend continua válida, mas não foi apresentada como nova execução.
- Dashboard permanece a única tela de negócio ainda não implementada.

### E5 — Dashboard e revisão funcional/visual — 03/10/2026

**Implementado**

- Dashboard real com quatro indicadores globais vindos exclusivamente de `GET /api/v1/dashboard`: total, abertas, em atendimento e concluídas.
- Cada montagem da rota consulta novamente a API. Não há cache remoto a invalidar, portanto voltar ao dashboard após criar, excluir ou mudar status obtém novos contadores; logout desmonta a área protegida e descarta o estado local.
- Cards de total/status levam à listagem, usando o filtro `status` já suportado na URL. Base vazia mantém os quatro zeros reais e recebe uma explicação; erro não é convertido em zeros e oferece repetição.
- Revisão de navegação e responsividade corrigiu seção ativa em detalhe/edição, adicionou `aria-current`, atalho de teclado ao conteúdo, menu móvel com quebra de linha, espaçamento de ações, `min-width: 0` no conteúdo e overflow horizontal restrito à região da tabela.
- Cards usam quatro colunas no desktop, duas no tablet e uma no celular. Textos, badges, labels, foco visível e regiões de feedback existentes foram revisados sem adicionar animações, gráficos ou métricas extras.

**Fluxo manual com API e PostgreSQL reais**

- Ambiente descartável `bit-project-dashboard-frontend`, PostgreSQL 17.6 na porta 55439, migrations v2 e seed com 2 usuários/5 solicitações. API compilada temporariamente; nenhum segredo real foi usado.
- colaborador1: login `200`; dashboard inicial `total=5, abertas=2, em_atendimento=2, concluidas=1`; criação de `SOL-000006`; dashboard `6/3/2/1`; filtro combinado encontrou exatamente o registro; detalhe e edição retornaram os dados atualizados.
- Após Em Atendimento, flags `can_edit/can_delete=false`, dashboard `6/2/3/1` e tentativas do autor de editar/excluir retornaram `409 request_not_open`. Conclusão e reabertura devolveram os rótulos/permissões esperados.
- colaborador2: login `200`; detalhe mostrou autor colaborador1 e permissões falsas; edição/exclusão retornaram `403 request_forbidden`; mudança para Em Atendimento e reabertura retornaram `200`, confirmando a política global de status.
- colaborador1 excluiu a solicitação reaberta (`204`), GET posterior retornou `404` e dashboard final voltou a `5/2/2/1`. Os dois logouts retornaram `204` e os dois `/me` posteriores retornaram `401`.
- A API, os arquivos temporários e o Compose descartável foram encerrados/removidos. As portas de teste 55439 e 18082–18084 ficaram sem listeners.

**Defeitos encontrados e corrigidos**

- Página atual incorreta nas rotas filhas de solicitações.
- Menu estreito com rolagem horizontal desnecessária.
- Ausência de link para pular navegação por teclado.
- Risco de largura mínima/overflow escapar da região rolável da tabela.
- Falta de espaçamento entre cabeçalho empilhado e suas ações.

Não foram observados defeitos funcionais no fluxo HTTP real após as correções.

**Verificações efetivamente executadas**

- Testes direcionados de dashboard/autenticação: 2 arquivos/17 testes passaram em 19,14 s.
- Suíte frontend completa: 8 arquivos/53 testes passaram em 66,01 s.
- `npm run lint` e `npm run typecheck` — passaram.
- `npm run build` — Vite 8.3.2 passou com 57 módulos; CSS 11,25 kB e JavaScript 352,56 kB.
- Cálculo de contraste das combinações principais — razões de 5,63:1 a 8,32:1 para texto secundário, botão primário, três badges de status e alerta de erro.
- Integração PostgreSQL com `TEST_DATABASE_URL` e trava de reset — passou em 2,174 s.
- `gofmt` sem diferenças, `go vet ./...` e `go test -count=1 ./...` — passaram; o último confirmou todos os pacotes em execução sem cache.

**Arquivos**

- `frontend/src/pages/DashboardPage.tsx`, `services/dashboard.ts` — consulta e apresentação dos indicadores; o antigo `RoutePlaceholderPage.tsx`, sem outros usos, foi removido.
- `frontend/src/features/dashboard/DashboardPage.test.tsx` — carga, erro, zeros, links e atualização por navegação.
- `frontend/src/components/layout/AppLayout.tsx`, `features/auth/AuthFlow.test.tsx` — navegação atual e acesso por teclado.
- `frontend/src/styles/base.css`, `layout.css`, `components.css` — dashboard e correções responsivas.
- `README.md`, `docs/DECISOES.md`, `REQUISITOS.md`, `TESTES.md`, `PROGRESSO.md` — estado, defeitos e evidências.

**Pendências**

- Não foi adotado E2E automatizado em navegador. A revisão responsiva desta etapa foi estrutural em CSS/DOM; capturas visuais reais permanecem reservadas para a etapa 20, conforme solicitado.
- Compose completo, CI, memorial técnico e evidências finais continuam nas etapas de qualidade/entrega.

### E6 — Docker e execução reproduzível — 03/10/2026

**Implementado**

- Dockerfile multi-stage do backend gera `api`, `migrate` e `seed` com Go 1.22.2 e entrega runtime Alpine sem toolchain, usuário não privilegiado, migrations, certificados, fuso e healthcheck com ferramenta presente.
- Dockerfile multi-stage do frontend executa `npm ci`/build com Node 24.14.0 e entrega somente a SPA no Nginx 1.27.4. A configuração encaminha `/api` e probes para a API, possui fallback de rotas React e healthcheck próprio.
- Compose completo fixa PostgreSQL 17.6, persiste em volume nomeado, condiciona migration à saúde do banco, API à conclusão da migration e frontend à saúde da API. A API não é publicada no host; Nginx é a única origem do navegador.
- Seed é tarefa explícita do profile `demo`, desabilitada por padrão e sem entrypoint que altere o comando. `.dockerignore` separados evitam `.env`, dependências e artefatos nas imagens.
- `.env.example`, Makefile, README e `docs/DEPLOY.md` documentam configuração/precedência, comandos com e sem Make/Docker, portas, origem/cookies HTTP e HTTPS, desligamento preservando volume e reset deliberadamente destrutivo.

**Validação isolada efetivamente executada**

- Cópia exata sem `.env` em `/tmp/bit-project-compose-validation-source`; projeto `bit-project-compose-validation`, frontend `18085`, PostgreSQL `55440` e volume exclusivo. Nenhum serviço/volume do projeto normal foi usado.
- Builds `portal-solicitacoes-backend:validation` e `portal-solicitacoes-frontend:validation` concluíram. Inspeção confirmou os três binários/migrations/`wget`, backend como usuário `app`, Nginx/index e ausência de Node/npm no runtime frontend.
- `docker compose ... up -d --no-build --wait`: PostgreSQL, API e frontend ficaram `healthy`; migration aplicou versões 1 e 2 e saiu com código 0. `nginx -t` passou dentro da rede Compose.
- Seed explícito criou 2 usuários/5 solicitações; repetição registrou 2/5 existentes e criou 0/0.
- Fluxo HTTP pela origem pública: probes `200`, login e `/me` `200`, dashboard total 5, criação `201` com `Location`, filtro combinado encontrou o registro, edição e três mudanças/reabertura retornaram `200`, dashboard total 6, exclusão `204`, dashboard total 5, logout `204` e `/me` posterior `401`.
- O cookie observado tinha `HttpOnly`, `SameSite=Lax`, `Path=/` e não tinha `Secure` no HTTP local. A renderização de configuração para produção confirmou `APP_ENV=production`, `SESSION_COOKIE_SECURE=true`, origem `https://portal.exemplo.com`, banco em `postgres:5432`, nenhuma porta pública da API e seed fora dos serviços padrão.
- Fallback SPA respondeu `200 text/html` com o root React em URLs internas. Após reiniciar PostgreSQL, API e frontend, a sessão e a solicitação permaneceram consultáveis; o segundo `up` concluiu migrations sem reaplicar estrutura.

**Qualidade**

- `go vet ./...` e `go test -count=1 ./...` — passaram.
- Frontend no estágio de build: lint e typecheck passaram; build Vite foi concluído na criação da imagem.
- A primeira suíte frontend, concorrendo com quatro verificações pesadas, aprovou 43 testes mas falhou ao iniciar um worker por timeout. Repetida isoladamente, passou com 8 arquivos/53 testes em 95,71 s; a falha inicial não foi ocultada nem tratada como aprovação.
- O primeiro script pós-logout esperava por engano `unauthorized`; o contrato real usa `authentication_required`. A asserção foi corrigida e todo o ciclo com nova solicitação/reinício foi repetido com sucesso; não era defeito da aplicação.

**Arquivos**

- `backend/Dockerfile`, `backend/.dockerignore` — build/runtime Go e contexto seguro.
- `frontend/Dockerfile`, `frontend/.dockerignore`, `frontend/nginx.conf` — build Vite, runtime Nginx, proxy e SPA.
- `compose.yaml`, `.env.example`, `Makefile` — orquestração e comandos.
- `README.md`, `docs/DEPLOY.md`, `DECISOES.md`, `REQUISITOS.md`, `TESTES.md`, `PROGRESSO.md` — operação, decisões e evidências.

**Pendências e limites**

- Não houve deploy externo, TLS real nem automação E2E em navegador. HTTPS, backups e observabilidade dependem da infraestrutura escolhida.
- CI, memorial técnico e capturas/evidências finais continuam nas próximas etapas.
- A composição isolada foi encerrada preservando seu volume durante esta validação; a remoção do volume de teste, se desejada, é uma ação destrutiva separada. O volume de desenvolvimento não foi alterado.

### E6 — Testes automatizados e CI — 03/10/2026

**Implementado**

- Playwright 1.63.0 foi fixado no `package.json`/lockfile. A configuração usa Chromium, um worker, base URL configurável e guarda trace, screenshot e vídeo em falha.
- O E2E usa frontend, API Go e PostgreSQL reais. O fluxo principal cobre autenticação, dashboard, CRUD, filtro combinado com período civil de um dia, autoria com dois usuários, todos os estados/reabertura, exclusão e logout; dois smokes cobrem desktop e celular.
- Vitest foi limitado aos testes em `src`, separando corretamente os 53 testes de componente simulados dos 3 E2E reais.
- `.github/workflows/ci.yml` adiciona jobs de backend, frontend, E2E e build Docker em push para `main` e pull requests. O banco é descartável, migrations/seed são explícitos, os serviços têm espera de prontidão e artefatos Playwright são enviados apenas em falha.
- `Makefile`, `.gitignore`, `.dockerignore`, README, decisões, matriz e estratégia de testes foram atualizados. A UI passou a conter overflow horizontal no documento, mantendo somente a tabela como região rolável.

**Verificações efetivamente executadas**

- Frontend em filesystem Linux: `npm ci --no-audit --no-fund` instalou 195 pacotes; `npm run test` passou com 8 arquivos/53 testes em 38,44 s; typecheck, lint e build passaram. O build produziu 57 módulos, JavaScript 352,56 kB e CSS 11,28 kB.
- Ambiente Compose isolado `bit-project-e2e-validation`, frontend `18086` e PostgreSQL `55441`: a execução final de `npm run test:e2e` passou com 3 testes no Chromium em 1,3 min.
- Ambiente PostgreSQL isolado `bit-project-ci-validation`, porta `55442`: gofmt sem diferenças, `go vet ./...`, testes unitários/HTTP e `go build ./...` passaram; integração real passou em 5,653 s; `go test -race ./internal/service ./internal/http` passou.
- `docker compose --env-file .env.example config --quiet` passou. O build completo exportou com sucesso as imagens atuais de backend e frontend.
- Ao final, `bit-project-e2e-validation` foi encerrado com remoção de seu volume exclusivo e `bit-project-ci-validation` foi encerrado. Nenhum container ou volume do ambiente normal de desenvolvimento foi removido.

**Defeitos e condições observadas**

- O smoke móvel detectou que o documento ainda podia rolar horizontalmente 554 px por causa da tabela larga. O overflow raiz foi contido e o teste confirmou `scrollX=0`, largura do body igual ao viewport e rolagem disponível dentro da tabela.
- Vitest inicialmente coletou o arquivo Playwright e falhou apesar dos 53 componentes aprovados. O padrão de inclusão foi corrigido e a suíte completa foi repetida.
- O Chromium local encontrou bibliotecas Linux ausentes. `install-deps` não pôde usar `sudo`; os pacotes foram apenas baixados/extraídos em `/tmp` e o E2E passou via `LD_LIBRARY_PATH`. O workflow usa a instalação de dependências no runner descartável.
- A primeira conexão ao PostgreSQL do Docker Desktop excedeu o timeout padrão de 3 s mesmo após o healthcheck. Repetida com `DATABASE_CONNECT_TIMEOUT=10s`, migration e toda a integração passaram; não houve erro de schema ou teste.
- Um ajuste final para obter o dia do filtro a partir do `created_at` leu inicialmente o campo fora do envelope `{data: ...}` e causou `Invalid time value` no fluxo principal; screenshots, vídeo e trace foram retidos, demonstrando o diagnóstico configurado. A leitura foi corrigida conforme o contrato e os 3 casos foram repetidos com sucesso.

**Limites e pendências**

- O workflow foi criado e seus comandos foram reproduzidos localmente, mas ainda não foi executado no GitHub; não se afirma aprovação remota.
- O pipeline é CI. Não há publicação nem CD, conforme escopo.
- A validação automática de OpenAPI não foi adicionada porque `docs/` continua ignorado por decisão vigente e, em um checkout do CI, `docs/openapi.yaml` não existe. Esse passo depende de versionar o contrato ou realocá-lo para uma área rastreada.
- Capturas finais de entrega permanecem para a etapa 20; os screenshots/traces atuais são artefatos de diagnóstico em falha.

**Arquivos**

- `.github/workflows/ci.yml`, `frontend/playwright.config.ts`, `frontend/e2e/portal.spec.ts` — pipeline e E2E.
- `frontend/package.json`, `package-lock.json`, `vite.config.ts` — dependência, comando e separação das suítes.
- `frontend/src/styles/base.css`, `.gitignore`, `frontend/.dockerignore`, `Makefile` — correção responsiva, artefatos e comandos.
- `README.md`, `docs/DECISOES.md`, `REQUISITOS.md`, `TESTES.md`, `PROGRESSO.md` — uso, decisões e evidências reais.

### E6 — Auditoria e entrega final — 03/10/2026

**Implementado e revisado**

- PDF de 12 páginas relido; requisitos obrigatórios e diferenciais confrontados com código, rotas, schema, testes e telas reais.
- README final ampliado com funcionalidades, stack, pré-requisitos, tabela de variáveis, instalação Docker/manual, migrations, seed/credenciais públicas, acesso, testes, deploy, troubleshooting e mapa documental.
- Criados `MEMORIAL_TECNICO_DE_DESENVOLVIMENTO.md` e `CHECKLIST_ENTREGA.md`. Memorial relaciona somente tecnologias reais, alternativas, arquitetura, decisões, limitações/produção e uso factual do Codex.
- Dicionário confrontado com migrations e catálogo PostgreSQL; API.md/OpenAPI 1.0.0, DEPLOY, TESTES, DECISOES e matriz revisados para remover estados futuros superados.
- `docs/` deixou de ser ignorado porque README, Memorial, Dicionário e evidências são obrigatórios no PDF. `AGENTS.md` e o PDF de referência continuam locais; pacote e repositório incluem a documentação de entrega.
- Roteiro Playwright separado gerou nove PNGs reais em `docs/evidencias`, cobrindo login vazio, dashboard, lista, filtros, criação, detalhes, dois status e celular.
- CI passou a validar o OpenAPI com Redocly CLI 2.57.0. Makefile recebeu `evidence` e `openapi-check`.

**Instalação limpa e fluxo real**

- Projeto `bit-project-final-validation`, PostgreSQL `55450`, frontend `18090`, imagens `:final` e volume exclusivo: builds concluídos; PostgreSQL/API/frontend saudáveis; Goose aplicou versões 1 e 2.
- Seed explícito: primeira execução criou 2 usuários/5 solicitações; repetição criou 0 usuários/0 solicitações e reconheceu 2/5 existentes.
- Roteiro de evidências: 1 teste Chromium aprovado em 30,8 s; nove imagens abertas e inspecionadas, sem senha/token.
- E2E com dois usuários: 3 testes aprovados em 50,8 s, incluindo CRUD, autoria, filtros de um dia, três status/reabertura, dashboard, logout e smokes responsivos.
- Persistência: antes/depois de reiniciar PostgreSQL/API/frontend, login/lista responderam `200`; a mesma sessão (`colaborador1`) e a solicitação concluída permaneceram. `/solicitacoes/6` respondeu `200 text/html` com root React após refresh.
- Catálogo real confirmou Goose 1/2, 18 colunas, 19 constraints e 11 índices.

**Qualidade final**

- Backend: gofmt sem diferenças; `go vet`; testes unitários/HTTP sem cache; integração PostgreSQL em 2,113 s; `go test -race ./internal/service ./internal/http`; `go build ./...` — aprovados.
- Frontend: 8 arquivos/53 testes em 56,67 s; typecheck, lint e build aprovados; Vite gerou 57 módulos, JS 352,56 kB e CSS 11,28 kB.
- Docker build limpo executou `npm ci` com 195 pacotes e exportou as duas imagens.
- OpenAPI válido no Redocly CLI, com três avisos de estilo honestamente mantidos: licença não definida e probes sem 4xx. Não foram inventadas licença ou respostas.
- YAML da CI foi parseado com quatro jobs; `git diff --check` e auditoria final de pacote executados antes do encerramento. Workflow ainda não rodou no GitHub.

**Falhas intermediárias registradas**

- Cache Go padrão do ambiente era somente leitura; repetição com `GOCACHE` em `/tmp` passou.
- Sandbox bloqueou socket do PostgreSQL na primeira integração; repetição autorizada contra o mesmo banco tmpfs passou.
- Cópia relativa incorreta deixou um diretório temporário frontend vazio; após correção, uma tentativa de `npm ci` em `/tmp` encontrou erro interno `Exit handler never called`. O `npm ci` limpo no Docker passou, e todas as verificações frontend foram repetidas na árvore Linux válida.
- Consulta inicial de constraints ordenou por alias com cast inválido; consulta corrigida retornou 19 linhas.

**Higiene e entrega**

- `.env` e `frontend/.env` locais preexistentes permaneceram intocados e ignorados. A auditoria encontrou ainda dependências/builds/caches locais ignorados; todos foram excluídos do pacote.
- Busca por padrões de chaves privadas/tokens não encontrou segredo; somente credenciais públicas demo permanecem em exemplos/documentação.
- O pacote local `artifacts/bit-project-entrega-2026-10-05.tar.gz` exclui `.git`, `.git-local`, `.env` real, `node_modules`, builds, caches, relatórios, binários, volumes, `AGENTS.md` e o PDF de referência; preserva lockfiles, migrations, exemplos, CI, documentação e evidências.

**Pendências reais**

- Responsável ainda deve revisar e enviar por repositório ou pacote até 05/10/2026.
- Workflow foi criado, mas a execução no GitHub depende do futuro push; não há CD/deploy externo.
- Produção ainda requer HTTPS, segredos, backup/restauração, observabilidade e rate limit distribuído, conforme Memorial.

### E6 — Consolidação exclusivamente documental — 03/10/2026

**Documentos alterados**

- `README.md` passou a concentrar objetivo, funcionalidades, regras e permissões, arquitetura, diretórios, decisões técnicas, configuração, execução, uso, testes, deploy, evidências, limitações e uso de inteligência artificial.
- `docs/MEMORIAL_TECNICO_DE_DESENVOLVIMENTO.md` passou a explicitar contexto, decisão, motivo/benefício, alternativas/compromissos e evidências das tecnologias, além de separar exigências do PDF das decisões que preencheram lacunas.
- `docs/DICIONARIO_DE_DADOS.md` foi confrontado por leitura com as migrations e teve a tabela completada com os índices implícitos `users_pkey` e `requests_pkey`.
- `docs/TESTES.md` passou a identificar os resultados como históricos e removeu a afirmação superada de que as capturas ainda seriam produzidas.
- `docs/CHECKLIST_ENTREGA.md` registra que o pacote local anterior ficou desatualizado após esta revisão e deve ser regenerado somente se a entrega for feita por arquivo compactado.

**Conferência documental**

- PDF, AGENTS, requisitos, decisões, progresso, API, OpenAPI, testes, deploy, migrations, manifests, Compose, Makefile, rotas e componentes relevantes foram lidos para confrontar os textos com a implementação.
- O contrato OpenAPI e `docs/API.md` permaneceram coerentes com as rotas e tipos lidos; nenhuma alteração de contrato foi necessária.
- Corrigido o exemplo manual de seed no README para definir `DEMO_SEED_ENABLED=true` na própria invocação.
- Mantidos como resultados anteriores — e não como validações desta revisão — os números de testes, builds, banco, E2E e evidências registrados nos documentos.

**Escopo e pendências**

- Esta etapa alterou somente documentação. Não houve instalação, teste, build, lint, formatação, migration, seed, inicialização de aplicação/container, deploy, commit ou push.
- O comando `pdftotext` não estava instalado; o PDF foi lido com o leitor Python já existente em `/tmp`, sem instalação e sem escrita no repositório.
- Caso a entrega use `artifacts/bit-project-entrega-2026-10-05.tar.gz`, o pacote deve ser recriado para incorporar esta revisão documental. Se a entrega usar o repositório atualizado, essa pendência do pacote não se aplica.

## Modelo para próximos registros

### Publicação inicial no GitHub — 01/10/2026

**Implementado e validado**

- Repositório remoto vazio verificado em `github.com/vinijbarros/bit-project`.
- Identidade configurada apenas no repositório: `Vinicius Barros <barros_works@hotmail.com>`.
- Commit `7be2e12` (`chore: initialize project foundation`) criado com 30 arquivos.
- Branch `main` enviada por SSH e configurada para acompanhar `origin/main`.
- Estado final limpo e sincronizado: `main...origin/main`.
- `docs/`, `AGENTS.md`, `.env`, dependências e artefatos de build permaneceram fora do commit conforme `.gitignore`.

**Observação**

- O `.git` fornecido pelo ambiente é um ponto de montagem vazio e somente leitura. Os metadados locais foram mantidos em `.git-local/`, também ignorado; o repositório remoto resultante é normal.

### Ex — Nome — AAAA-MM-DD

**Implementado**

- Alterações efetivamente realizadas.

**Validado**

- Comando ou cenário: resultado observado.

**Arquivos**

- Caminho — finalidade.

**Bloqueios e pendências**

- Item, impacto e próximo passo.
