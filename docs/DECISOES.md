# Decisões iniciais de arquitetura

## Status e contexto

Este documento começou com a arquitetura planejada na etapa E1. Decisão arquitetural, implementação e validação continuam sendo conceitos separados; os itens já executados trazem evidências em `PROGRESSO.md` e `REQUISITOS.md`. Mudanças posteriores devem preservar a justificativa histórica.

Na etapa E2 foi criada a base executável descrita adiante. A modelagem SQL e a autenticação foram acrescentadas em seguida; metadata, CRUD individual, mudança de status, listagem/filtros e dashboard estão implementados no backend e no frontend. Os fluxos funcionais obrigatórios da interface foram concluídos na E5; as etapas seguintes concentram qualidade e entrega.

## Versões fixadas na base executável

| Componente | Versão | Compatibilidade verificada |
| --- | --- | --- |
| Go | 1.22.2 | Versão instalada usada nos testes e builds. |
| pgx | 5.7.1 | O `go.mod` do pacote exige Go 1.21. Versões mais novas consultadas já exigem Go 1.23 ou 1.25. |
| Goose | 3.24.0 | O `go.mod` do pacote exige Go 1.21. Releases posteriores consultadas exigem Go 1.23–1.26. |
| Node.js | 24.14.0 | Versão instalada e aceita pelas dependências. |
| npm | 11.9.0 | Versão instalada usada para gerar e validar o lockfile. |
| React / React DOM | 19.3.0 | Instalados e compilados. |
| React Router DOM | 7.18.4 | Instalado e usado na rota inicial e no fallback. |
| Vite / plugin React | 8.3.2 / 6.1.1 | Exigem Node `^20.19.0` ou `>=22.12.0`; build validado em Node 24. |
| TypeScript | 6.0.3 | Série mais recente aceita pelo `typescript-eslint` selecionado (`<6.1.0`). TypeScript 7.0.2 foi rejeitado pela resolução de peers e não foi forçado. |
| ESLint / typescript-eslint | 10.11.0 / 8.71.0 | Lint executado sem avisos. |
| Vitest | 5.0.3 | Testes de componentes, páginas, roteamento e cliente HTTP; compatível com Node `^22.12`, `^24` ou `>=26`. |
| Testing Library React / user-event | 16.3.3 / 14.6.7 | Testes observáveis de formulário, roteamento e sessão sem acoplar à implementação interna. |
| jsdom | 29.1.1 | Ambiente DOM compatível com o Node 24.14 fixado; jsdom 30.1.1 foi rejeitado por exigir Node 24.15 ou superior. |
| Playwright Test | 1.63.0 | E2E no Chromium real; a dependência exige Node 20 ou superior e está fixada no lockfile. |
| Redocly CLI | 2.57.0 | Validação semântica do OpenAPI por comando fixado no Makefile/CI; exige Node compatível com o frontend. |
| PostgreSQL (Compose) | 17.6 Alpine | Imagem executada com healthcheck saudável e migrations validadas em banco descartável. |

As versões estão fixadas em `go.mod`/`go.sum` e `package.json`/`package-lock.json`. Atualizações futuras serão deliberadas, não automáticas.

## Visão arquitetural

Será usado um monólito modular: uma API Go, um frontend React separado e um banco PostgreSQL. Essa forma atende ao prazo e ao porte do desafio, mantém responsabilidades claras e evita o custo operacional de microserviços.

```text
Navegador
  └─ React + TypeScript + React Router
       └─ JSON/HTTP + cookie de sessão
            └─ API Go (net/http)
                 ├─ handlers: HTTP, DTOs e códigos de resposta
                 ├─ services: regras de negócio e autorização
                 └─ repositories: SQL via database/sql + pgx
                      └─ PostgreSQL

Goose ── aplica migrations versionadas no PostgreSQL
```

Estrutura final criada conforme cada etapa exigiu código real:

```text
backend/
  cmd/api/                 # composição e inicialização do servidor
  internal/
    config/                # leitura e validação de ambiente
    http/                  # roteador e handlers HTTP existentes
    repository/database/   # abertura/configuração de database/sql
    domain/                # criado quando houver tipos de negócio reais
    service/               # criado quando houver regras de negócio reais
  db/migrations/           # migrations Goose
frontend/
  src/
    api/                   # cliente HTTP e contratos
    components/            # componentes reutilizáveis
    features/              # auth, requests e dashboard
    pages/                 # composição das telas
    routes/                # rotas públicas e protegidas
    styles/                # estilos globais/tokens simples
docs/                      # requisitos, decisões, progresso e entrega
```

O módulo Go usa o nome local `portal-solicitacoes`, sem inventar endereço GitHub. O seed é um comando real; `internal/domain` e `internal/service` contêm os tipos e regras usados pela aplicação, sem arquivos de placeholder. Um `Makefile` substitui scripts duplicados.

## Decisões e alternativas

### Backend e camadas

- **Decisão:** Go com `net/http`, APIs JSON sob `/api/v1`, handlers finos, services com regras e repositories com SQL explícito.
- **Alternativas consideradas:** framework HTTP completo; regra de negócio diretamente nos handlers; ORM.
- **Justificativa:** `net/http` reduz dependências e demonstra fundamentos; as três camadas tornam autorização e regras testáveis; SQL explícito mantém a modelagem visível e atende à stack definida.

### Banco, migrations e transações

- **Decisão:** PostgreSQL, `database/sql` com pgx e migrations Goose. IDs numéricos gerados pelo banco; datas em `timestamptz`; constraints para categorias e status além da validação da aplicação.
- **Alternativas consideradas:** SQLite; UUID; enums nativos do PostgreSQL; ORM.
- **Justificativa:** PostgreSQL aproxima o exercício de um ambiente corporativo. IDs numéricos são suficientes e legíveis como código. Constraints `CHECK` são mais simples de evoluir que enums nativos neste escopo. Transações são usadas onde uma operação envolve escritas dependentes ou snapshot consistente.
- **Decisão de modelagem:** categorias (`ti`, `rh`, `compras`, `financeiro`, `infraestrutura`) e status (`aberto`, `em_atendimento`, `concluido`) são textos estáveis protegidos por `CHECK`, e não enums SQL. O código visual é calculado como `SOL-%06d` a partir do `id`.
- **Decisão de busca:** a pesquisa por título usa `ILIKE '%texto%'` no volume esperado para o desafio. Não foi criado B-tree para título, pois ele não atende de forma geral a esse padrão; `pg_trgm` só será considerado após evidência de necessidade.
- **Decisão de integridade:** usuários não são excluídos pela aplicação e as referências usam `ON DELETE RESTRICT`. Solicitações têm exclusão física, condicionada no service às regras de autoria e status.

### Autenticação e sessão

- **Decisão:** usuários previamente provisionados, senhas com hash bcrypt e sessões opacas persistidas no banco. O navegador recebe 32 bytes aleatórios criptográficos codificados em base64url somente no cookie `portal_session`; o banco guarda SHA-256 hexadecimal. O cookie usa `HttpOnly`, `SameSite=Lax`, `Path=/` e `Secure=true` obrigatório em produção. Prazo absoluto padrão de 8 horas, sem renovação/refresh.
- **Alternativas consideradas:** JWT no armazenamento local; autenticação Basic; cadastro público.
- **Justificativa:** sessão revogável torna o logout efetivo e reduz exposição do token a JavaScript. JWT adicionaria complexidade de revogação sem necessidade. Cadastro público e papéis estão fora das decisões de negócio.
- **Rotação:** cada login gera um token novo. Cookie anterior com formato válido é revogado na mesma transação que persiste a nova sessão; uma coincidência com o token anterior é rejeitada e o gerador tenta novamente.
- **Logout:** revoga pelo hash antes de limpar o cookie. Repetição sem cookie retorna `204`, mantendo idempotência. Falha do banco com token válido retorna `500`, não falso sucesso.
- **Rate limit:** 5 falhas em 15 minutos por username normalizado + IP de `RemoteAddr`, configuráveis. Não se confia em cabeçalhos encaminhados. O mapa em memória tem teto de 10.000 entradas e evicta a mais antiga; não é compartilhado entre instâncias e reinicia com o processo.
- **Cuidados:** resposta de login não revela se usuário ou senha falhou; usuário inexistente executa comparação bcrypt dummy; cookie seguro em produção; tokens/hashes/senhas não entram em respostas ou logs.
- **Normalização de usuário:** o backend aplicará trim e minúsculas antes da validação. O banco aceita 3–50 caracteres ASCII em `[a-z0-9._-]`, exige início/fim alfanuméricos e rejeita valores não normalizados.
- **Persistência de sessão:** somente SHA-256 hexadecimal minúsculo do token opaco será armazenado em `sessions.token_hash`; o token original existe apenas no cliente e no processamento imediato da requisição.

### Contrato HTTP e erros

- **Decisão:** OpenAPI 3.0.3 é o contrato normativo. JSON usa nomes `snake_case`, timestamps RFC 3339 em UTC e erros no formato `{"error":{"code":"...","message":"...","fields":{...}},"request_id":"..."}`. `fields` é um mapa de nomes JSON/parâmetros para listas de mensagens e aparece apenas em erros corrigíveis por campo.
- **Decisão:** `400` cobre JSON, parâmetros e validações de campo; `401` sessão ausente/expirada; `403` autorização/origem; `404` ausência; `409` conflito de estado; `415` mídia; `500` falha inesperada; readiness usa `503` para banco indisponível. A API não usa `422`, evitando dois códigos para a mesma classe de entrada inválida.
- **Decisão:** corpos JSON têm limite de 1 MiB, exigem `application/json`, rejeitam campos desconhecidos e um segundo valor. IDs são inteiros positivos; paginação padrão é 1/20 e limita `page_size` a 100.
- **Decisão:** toda resposta recebe `X-Request-ID`; logs incluem somente request ID, método, padrão de rota, status e duração. Corpo, query string, cookies, token e senha não são registrados. Panics e causas internas viram resposta genérica.
- **Decisão:** métodos mutáveis exigem `Origin` exatamente presente em `TRUSTED_ORIGINS`; sem curingas, comparação por sufixo ou fallback para `Referer`. O guard será conectado junto de cada handler real.
- **Alternativas consideradas:** GraphQL; respostas livres por endpoint.
- **Justificativa:** REST/JSON é suficiente e interoperável; envelope de erro previsível simplifica a interface e os testes.

### Datas e filtros

- **Decisão:** persistir instantes em UTC. `date_from`/`date_to` representam dias civis em `America/Sao_Paulo`; por exemplo, `date_from=2026-10-01&date_to=2026-10-01` é convertido para `[início do dia, início do dia seguinte)` em UTC.
- **Decisão:** cada limite é construído como início de uma data civil no fuso, inclusive em transições históricas, e depois convertido para UTC. O limite final usa a próxima data do calendário, não `Add(24*time.Hour)`.
- **Decisão:** filtros vazios equivalem a ausentes, filtros presentes são combinados por `AND`, e a ordenação não configurável é `created_at DESC, id DESC`.
- **Decisão:** `q` pesquisa somente título com `ILIKE`, após trim e limite de 150 pontos de código. `%`, `_` e barra invertida são escapados como literais; valores entram somente por parâmetros SQL.
- **Decisão:** count e página reutilizam a mesma cláusula estática de filtros dentro de transação somente leitura `REPEATABLE READ`, evitando metadados e itens de snapshots diferentes. O join muitos-para-um com `users` não duplica solicitações.
- **Decisão:** resposta usa `items` e `pagination` com `page`, `page_size`, `total_items` e `total_pages`; página além do fim é sucesso com lista vazia.
- **Alternativas consideradas:** comparar datas diretamente na timezone do banco; aceitar apenas timestamps completos.
- **Justificativa:** limites exclusivos evitam erros de precisão e a conversão explícita preserva o significado esperado pelo usuário.

### CRUD de solicitações

- **Decisão:** criação aceita somente título, descrição e categoria. O service aplica trim externo e conta os limites por pontos de código Unicode (`utf8.RuneCountInString`); o repository envia explicitamente autor, status `aberto` e instante corrente ao banco. Defaults e constraints SQL permanecem como defesa adicional.
- **Decisão:** o DTO expõe código visual derivado (`SOL-%06d`), categoria/status com valor estável e rótulo português, solicitante sem credenciais e flags `can_edit`/`can_delete` calculadas no servidor. Não há coluna para o código visual.
- **Decisão:** PATCH diferencia campo ausente, `null` e string vazia. Ausente preserva o valor; `null`, vazio após trim, payload sem campos e atualização sem diferença útil retornam `400`. Status é exclusivo da rota própria implementada.
- **Decisão:** edição e exclusão usam uma única instrução SQL condicional por `id`, `requester_id` e status `aberto`. Se nenhuma linha mudar, uma leitura diagnóstica classifica inexistência (`404`), autoria (`403`), estado (`409`) ou ausência de mudança (`400`). Assim, uma mudança concorrente de status nunca permite mutação indevida.
- **Limitação aceita:** não há versão otimista/ETag. Duas edições concorrentes do mesmo autor enquanto a solicitação permanece aberta usam última gravação vence. A proteção de autoria e estado continua atômica.
- **Alternativas consideradas:** transação com `SELECT ... FOR UPDATE`; coluna duplicada de código; contar bytes ou grafemas; aceitar `null` como limpeza.
- **Justificativa:** SQL condicional reduz round-trips e mantém a regra crítica junto à escrita; pontos de código produzem limites Unicode previsíveis sem dependência adicional; os campos obrigatórios não admitem limpeza para nulo/vazio.

### Transições de status

- **Origem:** o PDF exige alteração de status, mas não determina perfis, sequência obrigatória ou proibição de reabertura.
- **Decisão:** qualquer usuário autenticado pode alterar qualquer solicitação entre `aberto`, `em_atendimento` e `concluido`, em qualquer direção, inclusive reabrir. Não existem papéis administrativos nem máquina de estados mais restritiva.
- **Decisão:** repetir o status atual é idempotente: retorna a representação corrente sem executar `UPDATE` e sem alterar artificialmente `updated_at`.
- **Decisão:** o repository usa transação e `SELECT ... FOR UPDATE`. A existência, comparação e eventual mudança ocorrem sob o mesmo bloqueio; somente `status` e, em mudança efetiva, `updated_at` são escritos.
- **Coordenação:** updates/deletes do CRUD usam condições atômicas de autoria e status. O lock de linha serializa a disputa: fechar impede essas mutações posteriores; reabrir restabelece as permissões do autor; exclusão concluída antes do lock produz `404`.
- **Alternativas rejeitadas:** limitar a transição a uma sequência fixa; permitir apenas ao autor/administrador; fazer `SELECT` e `UPDATE` fora de transação; atualizar `updated_at` em repetição.

### Frontend

- **Decisão:** React + TypeScript + Vite, React Router, `fetch` encapsulado e CSS simples ou CSS Modules. Estado remoto permanecerá local às páginas/hooks enquanto a complexidade não justificar biblioteca adicional.
- **Alternativas consideradas:** Next.js; biblioteca global de estado; kit visual completo.
- **Justificativa:** é uma SPA autenticada sem necessidade de SSR. Poucas dependências mantêm o projeto compreensível e o CSS próprio permite demonstrar responsividade.
- **Cliente HTTP:** todas as chamadas usam `/api/v1` relativo, `credentials: same-origin`, `Accept` e JSON quando há corpo. O navegador fornece `Origin` nas mutações same-origin; o código não tenta forjar esse cabeçalho protegido. `204`, erro estruturado, resposta não JSON, rede, cancelamento e `401` central são tratados sem retry automático.
- **Sessão:** o frontend consulta `/auth/me`, mantém apenas o usuário seguro em memória e protege rotas com React Router. Token não é lido nem persistido por JavaScript; permanece no cookie HttpOnly. Login e logout usam a API real; a senha existe somente no estado transitório do formulário e é apagada após sucesso/desmontagem.
- **Fases:** autenticação usa `checking`, `authenticated` e `unauthenticated`. Falha de infraestrutura em `/me` mantém a fase de verificação com retry; `401` de `/me` significa ausência inicial de sessão, enquanto `401` operacional dispara expiração central. Login desabilita o disparo global de expiração para preservar o erro específico de credenciais.
- **Redirecionamento:** o destino anterior é aceito somente se for caminho local conhecido (`/dashboard` ou `/solicitacoes...`); URL absoluta, `//host` e rota desconhecida caem em `/dashboard`. A decisão fica no guard público para não depender da ordem entre atualização do contexto e navegação do formulário.
- **Logout:** o estado autenticado só é limpo depois do `204`. Falha de rede mantém usuário/tela e permite repetir. Após sucesso, o guard impede que histórico ou URL direta remontem conteúdo protegido.
- **Metadados:** `/metadata` protegido devolve, em ordem estável, os valores/rótulos definidos no domínio. A listagem usa essa resposta nos selects; badges usam os `{value,label}` retornados pelos próprios itens, sem mapas paralelos no frontend.
- **Listagem e URL:** filtros aplicados e página são a fonte da URL, com submissão explícita. Voltar/recarregar restaura o formulário; aplicar/limpar retorna à página 1. Datas civis permanecem strings `YYYY-MM-DD` até a API, evitando conversão UTC no navegador.
- **Formulários de solicitação:** criação e edição compartilham componente, normalização/validação por pontos de código Unicode e carregamento de metadata. Criação envia somente os três campos editáveis; edição monta PATCH apenas com campos efetivamente alterados. `permissions.can_edit` controla a entrada na tela, mas `403/409` do backend continuam tratados como autoridade contra estado obsoleto.
- **Alterações não salvas:** Cancelar pede confirmação apenas quando há mudanças e `beforeunload` protege fechamento/reload do navegador. Não foi introduzido bloqueador complexo de toda navegação SPA nesta etapa.
- **Atualização após mutações:** criação/edição navegam para o detalhe devolvido e a tela consulta novamente a API, exibindo confirmação. Mudança de status substitui a representação somente depois da resposta do servidor e recalcula as flags de permissão; exclusão navega para a lista somente após o `204`. Conflitos de estado/autoria disparam nova consulta do detalhe. Lista e dashboard não mantêm cache global e portanto recarregam ao serem visitados. Não foi adicionada biblioteca de consultas porque o volume atual não justifica invalidação global.
- **Dashboard:** cada montagem da rota executa um novo `GET /dashboard`; não há cache entre navegações ou sessões. Os quatro cards exibem somente os contadores globais do contrato e ligam para a listagem, usando `status` na URL quando aplicável.
- **Testes de frontend:** Vitest, Testing Library, user-event e jsdom exercitam componentes e fluxos observáveis com a fronteira HTTP controlada. Esses testes rápidos não são chamados de integração. Playwright executa os fluxos E2E no Chromium contra Nginx/Vite, API Go e PostgreSQL reais, com dados sintéticos identificáveis, limpeza restrita a esses registros e um worker para evitar dependência concorrente entre cenários.
- **Diagnóstico E2E:** trace, screenshot e vídeo são retidos em falha; seletores usam papel, label e texto. Viewports desktop e celular verificam que navegação, filtros e tabela continuam acessíveis.

### Execução, testes e CI

- **Decisão:** Dockerfiles e Compose para frontend, API e PostgreSQL; testes unitários de services/handlers, integração dos repositories com banco real, testes de componentes e E2E no navegador. O workflow de CI possui jobs separados para backend, frontend, E2E e build Docker, prepara um PostgreSQL descartável, espera a prontidão dos serviços e publica artefatos do Playwright somente em falha.
- **Alternativas consideradas:** somente execução manual; banco mockado em todos os testes; pipeline de deploy externo.
- **Justificativa:** Compose torna o ambiente reproduzível; regras críticas merecem testes rápidos e SQL precisa de testes contra PostgreSQL. O pipeline implementa integração contínua, não publicação: deploy externo automatizado continua melhoria futura por decisão de escopo. A configuração local foi reproduzida, mas a execução no GitHub só poderá ser afirmada depois que o workflow rodar no serviço.

### Configuração, banco e ciclo do processo

- **Decisão:** configuração exclusivamente por variáveis de ambiente, com `DATABASE_URL` obrigatória e padrões validados para timeouts/pool. URL ausente ou malformada encerra a aplicação como erro de configuração. Falha de conectividade gera aviso sem conteúdo sensível, permite manter o processo vivo e faz a prontidão responder `503`.
- **Decisão:** pool limitado a 10 conexões abertas e 5 ociosas por padrão, com vida máxima de 30 minutos e ociosidade máxima de 5 minutos. Ping e prontidão têm timeout de 3 segundos por padrão.
- **Decisão:** servidor com timeouts de cabeçalho, leitura, escrita e ociosidade, além de encerramento gracioso em `SIGINT`/`SIGTERM` por até 10 segundos.
- **Justificativa:** distingue erro corrigível de infraestrutura de configuração inválida, evita esperas e recursos ilimitados e não transforma indisponibilidade transitória do banco em falso resultado de processo morto.

### Probes e prefixo HTTP

- **Decisão:** `GET /healthz` verifica somente o processo e sempre retorna `200` enquanto o servidor responde. `GET /readyz` executa `PingContext` e retorna `200` ou `503`. Respostas não exibem URL, credencial ou erro interno.
- **Decisão:** endpoints funcionais ficam sob `/api/v1`. No desenvolvimento, o Vite encaminha `/api` para a API em `127.0.0.1:8080`; o navegador usa caminhos relativos.
- **Justificativa:** liveness e readiness têm responsabilidades diferentes. Origem relativa evita CORS desnecessário e simplifica cookies de sessão.

### Execução completa com containers

- **Decisão:** o Compose possui PostgreSQL 17.6 com volume/healthcheck, migration como tarefa concluída antes da API, API interna saudável e frontend Nginx como única origem pública. O seed fica fora do fluxo padrão, no profile `demo`, e exige `DEMO_SEED_ENABLED=true`.
- **Decisão:** o backend é construído em `golang:1.22.2-alpine3.19` e entregue em `alpine:3.19.1`, com usuário sem privilégios, `tzdata`, certificados, três binários e migrations. O frontend usa `node:24.14.0-alpine3.23` somente no build e `nginx:1.27.4-alpine3.21` no runtime.
- **Decisão:** Nginx encaminha `/api`, `/healthz` e `/readyz` para `api:8080` e aplica fallback para `index.html`. A API não publica porta no host; o navegador usa somente a origem do frontend, coerente com cookies e proteção por `Origin`.
- **Decisão:** `COMPOSE_DATABASE_URL` separa a URL interna com host `postgres` da `DATABASE_URL` usada fora de containers. O volume não é removido por nenhum alvo normal; reset com volumes é sempre deliberado e documentado como destrutivo.
- **Alternativas consideradas:** imagem única contendo Node/Nginx/Go; frontend servido diretamente pela API; seed automático no startup; API exposta em uma segunda origem pública.
- **Justificativa:** estágios separados reduzem o runtime, a tarefa de migration torna a ordem verificável, mesma origem evita CORS e o seed explícito impede criação silenciosa de credenciais. A composição continua simples para avaliação e servidor único, sem afirmar alta disponibilidade ou deploy externo.

## Modelo de dados inicial

- `users`: `id`, nome de usuário único/normalizado, nome de exibição, hash da senha e criação.
- `sessions`: hash único do token como chave primária, usuário, criação e expiração. Sessões expiradas não autenticam; logout excluirá a linha correspondente.
- `requests`: `id` numérico, título, descrição, categoria, status, autor, criação e atualização.

Não há tabelas de categoria/status: os conjuntos são pequenos e fixos, validados na aplicação e em constraints. Não haverá histórico de status ou auditoria nesta versão, pois não são exigidos; essa limitação deve constar no memorial. O detalhamento de campos, constraints, relações e índices está em `docs/DICIONARIO_DE_DADOS.md`.

## Seed de demonstração

- **Decisão:** o seed é executável separado da API e exige `DEMO_SEED_ENABLED=true`; também é bloqueado quando `APP_ENV=production`.
- **Decisão:** os dois usuários vêm de variáveis de ambiente. Senhas têm 8–72 bytes e são persistidas com bcrypt no custo padrão da biblioteca. A validação é em bytes por causa do limite de 72 bytes do bcrypt.
- **Decisão:** se um usuário já existir, a senha configurada é comparada com o hash. Divergência aborta a transação; somente `--reset-passwords` permite redefinição deliberada.
- **Decisão:** solicitações do seed possuem `seed_key` técnico, anulável, com índice único parcial. Registros comuns usam `NULL`; título não é único e não identifica carga demonstrativa.
- **Decisão:** o seed insere datas retroativas calculadas em UTC apenas para permitir demonstração do filtro por período. A API define a data corrente e não aceita `created_at` do cliente.
- **Decisão:** conflitos preservam o exemplo existente, e dados sem `seed_key` nunca são apagados ou atualizados pelo seed. A operação inteira usa transação serializável.
- **Justificativa:** o avaliador recebe dados realistas e hashes prontos para autenticação sem acoplar a inicialização da API, contaminar produção ou tornar a aplicação dependente de dados simulados.

Alternativas rejeitadas: usar título como chave natural, pois títulos reais podem repetir; apagar/recriar dados, pois destruiria trabalho do avaliador; executar seed automaticamente na API, pois mistura provisionamento com ciclo normal do serviço; guardar senha reversível ou em texto puro, por ser inseguro.

## Telas implementadas

| Rota | Tela | Conteúdo e comportamento principal |
| --- | --- | --- |
| `/login` | Login | Usuário, senha, erro genérico, envio e redirecionamento após autenticação. |
| `/` | Entrada autenticada | Redireciona para `/dashboard`. |
| `/dashboard` | Dashboard | Cartões com total e contagens por status, navegação e logout. |
| `/solicitacoes` | Lista e filtros | Tabela/cartões responsivos, período, categoria, status, título, paginação e acesso aos detalhes. |
| `/solicitacoes/nova` | Nova solicitação | Título, descrição, categoria, validação e confirmação de criação. |
| `/solicitacoes/:id` | Detalhes | Todos os dados, alteração de status e ações de editar/excluir apenas quando autorizadas. |
| `/solicitacoes/:id/editar` | Editar solicitação | Edição de título, descrição e categoria, disponível apenas ao autor enquanto aberta. |
| `*` | Não encontrada | Mensagem clara e retorno seguro à área apropriada. |

Todas as telas internas são protegidas. Estados de carregamento, vazio, erro, sucesso e indisponibilidade da sessão fazem parte da interface real.

As rotas acima possuem shell responsivo e fluxos reais. Login, sessão, logout, dashboard, listagem/filtros, detalhes, criação, edição, mudança de status e exclusão consomem a API; não restam placeholders nas telas obrigatórias.

## Endpoints implementados

| Método e caminho | Autenticação | Finalidade | Respostas principais |
| --- | --- | --- | --- |
| `GET /healthz` | Não | Saúde do processo, sem consultar dependências. | `200`. |
| `GET /readyz` | Não | Prontidão baseada na conectividade PostgreSQL. | `200`, `503`. |
| `POST /api/v1/auth/login` | Não | Validar credenciais, criar sessão/cookie e retornar o usuário. | `200`, `400`, `401`, `403`, `415`, `429`, `500`. |
| `POST /api/v1/auth/logout` | Cookie opcional para idempotência | Revogar sessão existente e expirar cookie; repetição retorna sucesso. | `204`, `403`, `500`. |
| `GET /api/v1/auth/me` | Sim | Obter usuário da sessão. | `200`, `401`, `500`. |
| `GET /api/v1/metadata` | Sim | Retornar categorias/status e rótulos em português. | `200`, `401`, `500`. |
| `GET /api/v1/requests` | Sim | Listar e filtrar por `date_from`, `date_to`, `category`, `status`, `q`, `page`, `page_size`. | `200`, `400`, `401`, `500`. |
| `POST /api/v1/requests` | Sim | Criar solicitação com campos automáticos no servidor. | `201`, `400`, `401`, `403`, `415`, `500`. |
| `GET /api/v1/requests/{id}` | Sim | Consultar detalhes. | `200`, `400`, `401`, `404`, `500`. |
| `PATCH /api/v1/requests/{id}` | Sim/autor | Editar título, descrição e categoria de solicitação aberta. | `200`, `400`, `401`, `403`, `404`, `409`, `415`, `500`. |
| `DELETE /api/v1/requests/{id}` | Sim/autor | Excluir solicitação aberta. | `204`, `400`, `401`, `403`, `404`, `409`, `500`. |
| `PATCH /api/v1/requests/{id}/status` | Sim | Alterar status, inclusive reabrir. | `200`, `400`, `401`, `403`, `404`, `415`, `500`. |
| `GET /api/v1/dashboard` | Sim | Retornar contagens globais total/por status. | `200`, `401`, `500`. |

O contrato detalhado está em `docs/openapi.yaml` e `docs/API.md`; qualquer mudança futura deve atualizá-los. Probes, autenticação, metadata, CRUD individual, mudança de status, listagem e dashboard estão registrados como implementados. A API não aceita autor, data de criação ou status inicial no comando de criação.

O dashboard é uma leitura global sem filtros e sem restrição por solicitante. Uma única instrução PostgreSQL usa `COUNT(*)` e agregações `FILTER` para obter total e os três status no mesmo snapshot; não há cache. O service verifica a invariável de soma e trata inconsistência ou falha de banco como erro, em vez de devolver zeros falsos.

Testes ficam separados por responsabilidade: regras puras em services, contrato HTTP com `httptest` e persistência/concorrência em `internal/integration` contra PostgreSQL real. A integração nunca lê `DATABASE_URL`: exige `TEST_DATABASE_URL`, banco com nome terminado em `_test` e `TEST_DATABASE_ALLOW_RESET=yes` antes de qualquer limpeza. O Compose de teste usa serviço próprio, porta dedicada e `tmpfs` para reduzir o risco de confundir dados de desenvolvimento com dados descartáveis.

## Etapas executadas

1. **E1 — Escopo e organização:** ler enunciado, registrar requisitos, decisões, dúvidas, arquitetura, telas, endpoints e processo de progresso.
2. **E2 — Fundação e autenticação:** módulos Go/Vite, configuração, schema inicial, migrations, conexão PostgreSQL, usuários demo, sessões e proteção básica; primeiros testes.
3. **E3 — API de solicitações:** CRUD autorizado, mudança de status, filtros, dashboard, erros, validações e testes backend.
4. **E4 — Frontend base:** shell responsivo, rotas, cliente HTTP, sessão, login/logout e componentes compartilhados.
5. **E5 — Fluxos completos:** dashboard, lista/filtros, criação, detalhes, edição, exclusão, status e testes frontend.
6. **E6 — Qualidade e entrega:** Compose, CI, revisão responsiva/segurança, dicionário, README final, memorial, evidências e validação integral.

## Fora do escopo atual

- Cadastro público, recuperação de senha, papéis/permissões administrativas e login social.
- Anexos, comentários, notificações, SLA, auditoria e histórico de status.
- Aplicativos móveis nativos e internacionalização.
- Deploy automatizado em provedor externo. A documentação de deploy continua obrigatória.
