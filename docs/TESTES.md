# Estratégia de testes

Este documento separa verificações unitárias, HTTP e PostgreSQL. Os testes priorizam respostas observáveis e regras que protegem dados; não há meta artificial de cobertura nem duplicação de constantes apenas para elevar percentuais.

Os resultados abaixo são registros de etapas anteriores. A revisão documental posterior não repetiu instalação, testes, build, lint, banco, containers ou navegador; portanto, ela não cria uma nova validação e apenas preserva as evidências já observadas.

## Camadas

### Services e validações

Arquivos `backend/internal/service/*_test.go` verificam regras sem HTTP: normalização e bcrypt, token/hash e rotação de sessão, limites Unicode, campos parciais, autoria/estado, matriz de status, calendário de São Paulo, paginação e invariável do dashboard. Repositories são substituídos apenas para isolar decisões do service; regras SQL não são consideradas validadas por esses stubs.

### HTTP com `httptest`

Arquivos `backend/internal/http/*_test.go` verificam contrato, status e envelopes: cookies, identidade segura, autenticação, origem, rate limit, JSON estrito, corpo de 1 MiB, `Content-Type`, IDs/query, mass assignment, DTOs, erros internos sem detalhes e logs sem corpo/segredo. Esses testes não afirmam persistência.

### Integração PostgreSQL

`backend/internal/integration/backend_integration_test.go` conecta services/repositories reais ao router e usa `httptest.Server`. A suíte aplica as migrations Goose reais e cobre sessão persistida, CRUD/autorização com dois usuários, transições/reabertura, filtros e SQL literal, dashboard, falhas de infraestrutura e concorrência entre fechamento e edição.

## Proteções e preparação do banco

A integração é destrutiva somente dentro de um banco descartável. Três condições evitam limpeza acidental:

1. a suíte lê exclusivamente `TEST_DATABASE_URL`, nunca `DATABASE_URL`;
2. o nome extraído da conexão deve terminar em `_test`;
3. `TEST_DATABASE_ALLOW_RESET` deve ser exatamente `yes`.

Sem URL, os testes que dependem do PostgreSQL usam `Skip` com motivo explícito; o teste unitário da trava continua executando. URL insegura ou confirmação ausente falha, sem conectar nem limpar. Entre cenários, a limpeza é `TRUNCATE requests, sessions, users RESTART IDENTITY CASCADE`. O Compose dedicado usa `portal_test`, porta 55435 por padrão e `tmpfs`, portanto não persiste dados após remoção.

```sh
make db-test-up

TEST_DATABASE_URL='postgres://portal_test:portal_test@127.0.0.1:55435/portal_test?sslmode=disable' \
TEST_DATABASE_ALLOW_RESET=yes \
make test-integration

make db-test-down
```

Comandos gerais:

```sh
cd backend
gofmt -w .
go vet ./...
go test ./...
go test -race ./internal/service ./internal/http
```

`go test -race` é uma verificação adicional, não substitui os testes normais. Se toolchain/CGO/ambiente impedirem sua conclusão, isso deve ser registrado como pendência, nunca convertido em `Skip` silencioso.

## Cenários cobertos

| Área | Evidência automatizada principal |
| --- | --- |
| Login válido/inválido, bcrypt e erro genérico | `TestAuthenticationPersistenceAndFailures`, `TestLoginCreatesHashedPersistentSessionAndRotatesPrevious`, testes HTTP de login. |
| Expiração, revogação e persistência após reinício HTTP | `TestAuthenticationPersistenceAndFailures`. |
| `/me` sem senha/hash/token | `TestAuthenticationPersistenceAndFailures` e `TestMeRequiresAndReturnsAuthenticatedIdentity`. |
| Autorização de rotas | Loop anônimo de `TestRequestsAuthorizationStateFiltersAndDashboard` e testes HTTP específicos. |
| Criação automática e mass assignment | `TestRequestsAuthorizationStateFiltersAndDashboard` e `TestCreateRequestRejectsAutomaticOrUnknownFields`. |
| Autor/estado em editar e excluir | `TestRequestsAuthorizationStateFiltersAndDashboard` com dois usuários. |
| Status, reabertura e concorrência | `TestRequestsAuthorizationStateFiltersAndDashboard`, matriz unitária 3×3 e `TestConditionalMutationLosesRaceToStatusChange`. |
| Listagem, período São Paulo, LIKE literal e paginação | Integração combinada e testes unitários de calendário/escape/paginação. |
| Entrada SQL maliciosa como texto | Integração persiste título com aspas e aparência de SQL, filtra `%`/`_` literalmente e confirma tabela/usuários intactos. |
| Dashboard global e atual | Integração após criação/reabertura, testes HTTP vazio/populado e teste de service da soma. |
| JSON, tamanho, mídia, IDs e origem | `request_test.go`, `requests_test.go`, `auth_test.go` e `middleware_test.go`. |
| Falhas de banco preservam semântica | `TestDatabaseFailuresKeepTheirMeaning` e testes HTTP de erro interno seguro. |
| Logs, timeouts e fechamento | Testes de middleware/probes/config; revisão de `http.Server`, contexts, `rows.Close`, rollback/commit e `db.Close`. |
| Segurança da própria suíte | `TestDatabaseSafetyGuard`. |

## Limitações atuais

- O detector de corrida havia excedido 180 segundos em uma revisão anterior. Nesta etapa, `go test -race ./internal/service ./internal/http` foi repetido com tempo suficiente e passou; o custo de compilação no WSL continua significativamente maior que o teste normal.
- A concorrência é validada com lock PostgreSQL controlado e edição bloqueada até uma mudança de status vencer. Não há teste de carga/soak.
- O reinício recria router e services sobre um novo `sql.DB`, validando reconexão e persistência fora da memória do processo; não inicia, porém, um segundo binário do sistema operacional.
- Os fluxos E2E usam Chromium e um worker contra PostgreSQL/API/frontend reais. Não são testes de carga nem substituem validação em todos os navegadores. As capturas finais foram geradas separadamente pelo roteiro de evidências.

## Autenticação do frontend

`frontend/src/features/auth/AuthFlow.test.tsx` usa jsdom e API simulada por funções controladas. Verifica comportamento React observável: validação, login válido/inválido, botão durante envio, destino anterior seguro, ausência de conteúdo antes de `/me`, sessão restaurada, acesso direto protegido, logout confirmado/falho, retorno a URL protegida, expiração por `401` e indisponibilidade/retry de `/me`.

Esses testes não são apresentados como integração real. A validação real complementar sobe PostgreSQL descartável, API e frontend e usa HTTP através do proxy: anônimo `/me=401`, login válido `200`, `/me` com cookie `200`, logout `204`, novo `/me=401` e credencial inválida `401`. O E2E Playwright descrito adiante complementa essa evidência com renderização e interação em navegador real.

A execução reproduzível mais recente usou uma cópia exata do frontend em `/tmp`, porque o npm encontrou erros `ENOTEMPTY` ao recriar `node_modules` no volume Windows montado em `/mnt/d`. Nesse filesystem Linux, `npm ci` instalou o lockfile sem erro e a suíte passou com 2 arquivos/20 testes, seguida por typecheck, lint e build aprovados. O cenário de digitação/credencial inválida possui timeout local de 10 segundos para tolerar variação de I/O sem relaxar o restante da suíte.

## Listagem do frontend

`frontend/src/features/requests/listQuery.test.ts` verifica a transformação entre formulário, URL e comando da API sem criar objetos `Date` para dias civis, além da validação de intervalo invertido. `RequestsPage.test.tsx` usa jsdom e services simulados para testar renderização de todos os campos, metadata, filtros apenas no submit, limpeza/página 1, histórico do navegador, paginação, estados vazio/sem resultado/erro e cancelamento de uma resposta antiga.

A apresentação de fuso usa um timestamp próximo à meia-noite (`02:30Z`) e confirma o dia/hora anteriores em `America/Sao_Paulo`. `RequestDetailPage.test.tsx` confirma leitura completa por ID, retorno à URL filtrada, texto com quebras/trecho semelhante a HTML, mudança e reabertura, permissões de dois usuários, diálogo acessível, cancelamento/confirmação, conflitos `403/409`, remoção `404` e falha de infraestrutura. A integração real complementar já registrada usou os dois usuários do seed: ambos receberam a listagem global de 5 registros/2 autores; metadata retornou 5 categorias/3 status; um filtro combinado retornou apenas o registro esperado; página 99 preservou totais com `items=[]`. Ao concluir detalhe/status, a suíte passou com 7 arquivos/47 testes; o total mais recente está na seção de dashboard. Isso não é descrito como E2E de navegador: UI/URL foram exercitadas em jsdom.

## Formulários do frontend

`requestForm.test.ts` compara a contagem por pontos de código Unicode, trim e limites exatos de título/descrição, incluindo emoji, além de categoria desconhecida. `RequestForms.test.tsx` usa services simulados para observar criação com payload mínimo, bloqueio de envio duplicado, erros junto aos campos, preservação após rede, cancelamento confirmado, PATCH somente do campo alterado, acesso direto sem permissão e respostas `403/409` sem perda do conteúdo.

Esses testes de componente não validam sozinhos autoria. A evidência backend permanece `TestRequestsAuthorizationStateFiltersAndDashboard`, executado anteriormente contra PostgreSQL real com dois usuários: outro autor recebeu `403`, e o autor recebeu `409` após mudança de status. Na etapa dos formulários, a repetição real ficou impedida porque o comando Docker deixou de estar disponível na integração WSL; nenhuma nova execução PostgreSQL foi afirmada. Naquela etapa, a suíte frontend passou com 7 arquivos/40 testes; o total mais recente está registrado na seção de listagem/detalhes.

## Dashboard e revisão funcional do frontend

`frontend/src/features/dashboard/DashboardPage.test.tsx` usa jsdom e service controlado para confirmar os quatro valores globais, links de total/status para a listagem, base vazia com quatro zeros, erro/retry e uma nova consulta ao desmontar e revisitar a rota. `AuthFlow.test.tsx` também confirma `aria-current` em detalhes e o atalho para o conteúdo principal. Nenhum desses testes atribui persistência ao mock.

O fluxo HTTP manual usou PostgreSQL 17.6 descartável, migrations reais, seed, API compilada e cookies separados. Resultado observado:

- colaborador1 entrou (`200`), leu dashboard `5/2/2/1`, criou `SOL-000006` e o dashboard mudou para `6/3/2/1`;
- filtro combinado por título/categoria/status localizou exatamente o novo registro; detalhe e edição persistida retornaram os dados esperados;
- ao mudar para Em Atendimento, permissões ficaram falsas, o dashboard passou a `6/2/3/1` e editar/excluir retornaram `409 request_not_open`;
- conclusão e reabertura funcionaram; colaborador2 entrou, recebeu permissões falsas e `403 request_forbidden` ao editar/excluir, mas pôde fechar e reabrir o recurso;
- o autor excluiu a solicitação reaberta (`204`), consulta posterior retornou `404` e dashboard voltou a `5/2/2/1`;
- ambos os logouts retornaram `204`, seguidos por `/me` `401`.

A suíte PostgreSQL completa também passou em 2,174 s. O ambiente descartável e os processos temporários foram encerrados após a validação.

### Defeitos encontrados na revisão de interface

- A navegação não marcava “Solicitações” como página atual em detalhe/edição. Corrigido com determinação explícita da seção e `aria-current="page"`.
- O menu estreito dependia de rolagem horizontal e podia competir com a tabela. Corrigido para quebra de linha; em até 420 px os links ocupam duas colunas.
- Faltava atalho para teclado saltar a navegação. Incluído link “Pular para o conteúdo principal”, visível ao foco.
- Conteúdo da grade precisava de `min-width: 0`, e a tabela de `max-width`/overflow local explícitos, para evitar rolagem horizontal na página. Corrigido; somente a região identificada da tabela rola.
- Ações do cabeçalho ficavam próximas do texto ao empilhar. Incluído espaçamento no breakpoint móvel.

Os estilos foram revisados nos comportamentos definidos para celular (até 600 px, cards em uma coluna), tablet (601–1050 px, duas colunas) e desktop (quatro colunas). Um cálculo WCAG das combinações principais confirmou razões entre 5,63:1 e 8,32:1 para texto secundário, botão primário, badges e alerta de erro. Naquele momento isso era apenas revisão estrutural de CSS e semântica; as capturas finais foram produzidas posteriormente e estão descritas em `docs/evidencias/README.md`.

## Docker Compose

A validação de distribuição usa um projeto Compose separado, portas `55440`/`18085` e o volume nomeado `bit-project-compose-validation_postgres_data`. A fonte foi copiada sem `.env`, dependências ou artefatos locais para `/tmp/bit-project-compose-validation-source`; portanto, o teste não reutilizou o volume nem o banco de desenvolvimento.

Foram observados:

- build multi-stage das imagens de backend e frontend com as versões fixadas; os runtimes contêm os binários/arquivos necessários, e o frontend final não contém Node/npm;
- PostgreSQL saudável, migrations v1/v2 aplicadas do zero, tarefa `migrate` em `Exited (0)`, API e Nginx saudáveis;
- seed executado somente pelo profile `demo`: primeira execução criou 2 usuários/5 solicitações e a repetição criou 0/0;
- pela origem pública do Nginx, login, `/me`, dashboard 5→6, criação, filtro combinado, detalhe, edição, transições para atendimento/conclusão/reabertura, permissões, exclusão, dashboard 6→5 e logout;
- cookie local com `HttpOnly`, `SameSite=Lax`, `Path=/` e sem `Secure`; a configuração renderizada para produção apresentou `Secure=true` e origem HTTPS exata;
- fallback da SPA retornou `index.html` em `/solicitacoes/123` e `/requests/123`; `nginx -t` passou dentro da rede Compose;
- após reiniciar PostgreSQL, API e frontend, a sessão persistida e a solicitação criada continuaram consultáveis. Um novo `up` executou Goose sem reaplicar versões existentes.

A primeira execução concorrente de lint/tipos/testes frontend e verificações Go causou timeout ao iniciar um worker do Vitest após 43 testes aprovados. Executada novamente sozinha, a suíte terminou com 8 arquivos e 53 testes aprovados; não houve falha de asserção. `go vet ./...`, `go test -count=1 ./...`, lint e typecheck também terminaram com código zero. O build Vite ocorreu no estágio Docker e as duas imagens finais foram exportadas com sucesso.

Na etapa Docker ainda não havia automação de navegador; essa lacuna foi fechada posteriormente pelos testes Playwright abaixo. O Compose continua sendo validação local, não evidência de deploy externo ou alta disponibilidade.

## E2E Playwright e integração contínua

`frontend/e2e/portal.spec.ts` usa Playwright 1.63.0 no Chromium contra frontend, API Go e PostgreSQL reais. A suíte roda com um worker e títulos sintéticos exclusivos; a preparação e a limpeza consultam/excluem somente esses registros, sem truncar dados alheios. As credenciais são públicas e exclusivas de teste. Testes de componente continuam usando API simulada e não são apresentados como prova dessa integração.

O cenário funcional cobre login, baseline do dashboard, criação com categoria, recarga de URL interna, edição, filtro combinado com período de um dia em São Paulo, seis colunas da listagem, permissões reais de um segundo usuário (`403`), bloqueio do autor em status não aberto (`409`), atendimento, conclusão, reabertura, exclusão confirmada, atualização do dashboard, logout e nova proteção da rota. Dois smokes adicionais usam 1440×900 e 375×812 e verificam controles essenciais, ausência de rolagem horizontal no documento e rolagem contida na tabela.

Resultados locais de 03/10/2026:

- `npm ci --no-audit --no-fund`: 195 pacotes instalados a partir do lockfile;
- `npm run test`: 8 arquivos e 53 testes Vitest aprovados em 38,44 s;
- `npm run typecheck`, `npm run lint` e `npm run build`: aprovados; build Vite com 57 módulos, 352,56 kB de JavaScript e 11,28 kB de CSS;
- `npm run test:e2e`: execução final com 3 testes Playwright aprovados em 1,3 min;
- integração PostgreSQL: `go test -count=1 ./internal/integration` aprovado em 5,653 s;
- detector de corrida: `go test -race ./internal/service ./internal/http` aprovado;
- gofmt, `go vet ./...`, testes unitários/HTTP e `go build ./...`: aprovados;
- `docker compose --env-file .env.example config --quiet` e build das imagens backend/frontend: aprovados.

O navegador local foi baixado pelo Playwright, mas a instalação global das bibliotecas Linux pediu senha de `sudo`. Para não alterar a máquina, os pacotes runtime necessários foram baixados e extraídos em `/tmp`, e a execução usou `LD_LIBRARY_PATH` apontando para esse diretório. Em CI, `npx playwright install --with-deps chromium` prepara essas dependências no runner descartável.

A primeira execução E2E encontrou rolagem horizontal real no documento móvel; `html`/`body` passaram a conter o overflow, mantendo a tabela como região rolável, e o smoke foi repetido com sucesso. A primeira execução do Vitest após adicionar o E2E também tentou coletar o arquivo Playwright; o `include` foi restringido a `src/**/*.test.{ts,tsx}` e todos os 53 casos voltaram a passar.

Ao endurecer o cenário de período para derivar o dia do `created_at` real, uma execução intermediária leu por engano o campo fora do envelope `{data: ...}` e falhou com `Invalid time value`; os dois smokes passaram e os artefatos de falha foram gerados. A leitura foi alinhada ao contrato, e a suíte completa foi repetida com 3/3 casos aprovados.

O workflow `.github/workflows/ci.yml` possui jobs separados para backend, frontend/OpenAPI, E2E e build Docker. Ele prepara banco descartável, aplica migrations/seed explícito, espera health/readiness, executa o navegador e envia traces/screenshots/vídeos somente em falha. Os comandos equivalentes foram reproduzidos localmente, mas o workflow ainda não foi executado no GitHub. Portanto, CI está configurada; CD/deploy automatizado não está implementado.

## Auditoria final de entrega — 03/10/2026

A instalação final usou os projetos isolados `bit-project-final-validation` (PostgreSQL 55450, frontend 18090, volume exclusivo) e `bit-project-final-tests` (PostgreSQL tmpfs 55451). Nenhum banco ou volume de desenvolvimento foi reutilizado ou limpo.

Resultados observados:

- build Docker sem cache de fonte alterada: backend e frontend concluídos; o estágio frontend executou `npm ci` com 195 pacotes e Vite compilou 57 módulos, 352,56 kB de JavaScript e 11,28 kB de CSS;
- migration em banco vazio: versões Goose 1 e 2 aplicadas, serviço `migrate` em `Exited (0)` e API/frontend saudáveis;
- seed: primeira execução criou 2 usuários/5 solicitações; a segunda criou 0/0 e reconheceu 2/5 existentes;
- fluxo de evidências no Chromium: 1 teste aprovado em 30,8 s e nove PNGs reais produzidos;
- E2E final com dois usuários: 3 testes aprovados em 50,8 s;
- Go sem cache de resultado: cmd/config/http/repository/seed/service aprovados; integração PostgreSQL aprovada em 2,113 s; `go test -race ./internal/service ./internal/http` aprovado; `go build ./...` aprovado;
- frontend: 8 arquivos/53 testes aprovados em 56,67 s; typecheck, lint e build aprovados; Playwright listou 3 E2E e 1 roteiro de evidência;
- persistência: antes e depois do reinício de PostgreSQL/API/frontend, `/auth/me` e a consulta da solicitação criada retornaram `200`; usuário `colaborador1`, um item e status `concluido` foram preservados;
- refresh de `/solicitacoes/6`: `200 text/html` e root React presente, confirmando fallback SPA;
- catálogo real: Goose v1/v2, 18 colunas, 19 constraints e 11 índices conferidos contra as migrations/dicionário;
- OpenAPI: após acrescentar versão/descrições finais, Redocly CLI 2.57.0 considerou o contrato válido; restaram três avisos recomendados de estilo (licença não definida e ausência de resposta 4xx nos dois probes), sem erro estrutural/semântico. Não foi inventada licença nem resposta que a API não implementa apenas para silenciar o lint.

Falhas intermediárias não ocultadas:

- a primeira execução Go tentou o cache padrão somente leitura e falhou; repetida com `GOCACHE=/tmp/bit-project-final-go-cache`, passou;
- a integração sem permissão ampliada foi bloqueada pelo sandbox ao abrir o socket local; repetida contra o mesmo PostgreSQL isolado com acesso local permitido, passou;
- uma tentativa de `npm ci` em diretório temporário vazio usou origem relativa incorreta; após corrigir a cópia, o npm apresentou `Exit handler never called`. O `npm ci` limpo do estágio Docker já havia passado com o lockfile atual, e testes/tipos/lint/build foram repetidos sobre a árvore instalada válida em filesystem Linux;
- a primeira consulta de catálogo ordenou por alias com cast inválido; a consulta corrigida retornou as 19 constraints.

Não foram executados teste de carga, pentest, deploy externo, restauração de backup ou matriz de múltiplos navegadores. Esses itens não são apresentados como validados.
