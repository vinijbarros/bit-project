# Dicionário de dados

## Visão geral

O schema é criado pelas migrations versionadas `backend/db/migrations/00001_create_core_tables.sql` e `00002_add_request_seed_key.sql`. A validação final anteriormente registrada consultou `information_schema`, `pg_constraint`, `pg_indexes` e `goose_db_version` em PostgreSQL 17.6 criado do zero: versões 1 e 2, 18 colunas, 19 constraints e 11 índices corresponderam às definições abaixo. Nesta revisão documental, o conteúdo foi novamente confrontado apenas por leitura das duas migrations; nenhum SQL foi executado. Todos os instantes usam `TIMESTAMPTZ`; a aplicação trabalha com UTC, independentemente da forma como uma sessão PostgreSQL exiba o valor.

```text
users (1) ─────< sessions (N)
   │
   └──────────< requests (N)
```

As duas chaves estrangeiras usam `ON DELETE RESTRICT`. A aplicação não oferece exclusão de usuários, preservando a identidade referenciada por sessões e solicitações.

## `users`

Armazena usuários previamente provisionados. Não existe cadastro público nesta versão.

| Campo | Tipo | Nulo | Chave | Default | Restrições | Significado |
| --- | --- | --- | --- | --- | --- | --- |
| `id` | `BIGINT` identity | Não | PK | Gerado pela identity | Sem `CHECK` adicional; a aplicação usa IDs gerados | Identificador interno do usuário. |
| `username` | `VARCHAR(50)` | Não | UK | — | 3–50 caracteres; único; minúsculo; sem espaços externos; começa e termina com letra ASCII minúscula ou número; miolo aceita `a-z`, `0-9`, `.`, `_` e `-` | Nome usado no login. |
| `display_name` | `VARCHAR(120)` | Não | — | — | 1–120 caracteres após trim; banco rejeita espaços externos | Nome apresentado na interface. |
| `password_hash` | `VARCHAR(255)` | Não | — | — | 20–255 caracteres | Hash de senha produzido pelo backend; nunca contém a senha original. |
| `created_at` | `TIMESTAMPTZ` | Não | — | `CURRENT_TIMESTAMP` | — | Instante de criação do usuário. O backend também o define explicitamente. |

### Normalização de `username`

Antes de validar e consultar/persistir, o backend aplica `trim` e conversão para minúsculas. O banco não transforma silenciosamente: ele rejeita valores não normalizados. A expressão aceita é equivalente a `^[a-z0-9][a-z0-9._-]*[a-z0-9]$`, combinada ao limite de 3–50 caracteres. Essa regra evita duplicidade por capitalização sem depender de extensão ou collation específica.

## `sessions`

Armazena sessões opacas revogáveis. O token entregue ao navegador nunca é persistido; somente seu hash SHA-256 em hexadecimal minúsculo é armazenado.

| Campo | Tipo | Nulo | Chave | Default | Restrições | Significado |
| --- | --- | --- | --- | --- | --- | --- |
| `token_hash` | `VARCHAR(64)` | Não | PK | — | Exatamente 64 caracteres hexadecimais minúsculos | SHA-256 do token de sessão; identifica a sessão sem guardar o segredo original. |
| `user_id` | `BIGINT` | Não | FK, índice | — | Referencia `users.id` com `ON DELETE RESTRICT` | Usuário dono da sessão. |
| `created_at` | `TIMESTAMPTZ` | Não | — | `CURRENT_TIMESTAMP` | — | Instante de criação. O backend também o define explicitamente. |
| `expires_at` | `TIMESTAMPTZ` | Não | Índice | — | Deve ser posterior a `created_at` | Limite absoluto de validade da sessão. |

Não há `id` adicional: `token_hash` já é uma chave estável e única para localizar e revogar a sessão. O índice de `expires_at` apoia consultas de validade e limpeza de sessões expiradas.

## `requests`

Armazena as solicitações internas. A exclusão é física e o backend a executa somente depois das regras de autoria e status.

| Campo | Tipo | Nulo | Chave | Default | Restrições | Significado |
| --- | --- | --- | --- | --- | --- | --- |
| `id` | `BIGINT` identity | Não | PK | Gerado pela identity | Sem `CHECK` adicional; a aplicação usa IDs gerados | Código interno numérico da solicitação. |
| `title` | `VARCHAR(150)` | Não | — | — | 3–150 caracteres após trim; banco rejeita espaços externos | Título informado pelo solicitante. |
| `description` | `TEXT` | Não | — | — | 10–5000 caracteres após trim; banco rejeita espaços externos | Descrição detalhada. |
| `category` | `VARCHAR(20)` | Não | Índice | — | `ti`, `rh`, `compras`, `financeiro` ou `infraestrutura` | Categoria interna estável. |
| `status` | `VARCHAR(20)` | Não | Índice | `'aberto'` | `aberto`, `em_atendimento` ou `concluido` | Estado interno estável. O backend define explicitamente `aberto` na criação. |
| `requester_id` | `BIGINT` | Não | FK, índice | — | Referencia `users.id` com `ON DELETE RESTRICT` | Autor da solicitação, obtido da sessão autenticada. |
| `created_at` | `TIMESTAMPTZ` | Não | Índice | `CURRENT_TIMESTAMP` | — | Instante de abertura. O backend também o define explicitamente. |
| `updated_at` | `TIMESTAMPTZ` | Não | — | `CURRENT_TIMESTAMP` | Igual ou posterior a `created_at` | Instante da última alteração; o backend deve atualizá-lo explicitamente. |
| `seed_key` | `VARCHAR(64)` | Sim | Índice único parcial | `NULL` | Quando presente, segue `^demo_[a-z0-9_]+$` e é único entre valores não nulos | Identificador técnico exclusivo dos exemplos gerenciados pelo seed; solicitações normais permanecem com `NULL`. |

`seed_key` não é um identificador funcional nem aparece para o usuário. O índice parcial permite várias solicitações normais com `NULL` e títulos repetidos, enquanto torna a carga demonstrativa repetível sem confundir um título real com um exemplo. O seed preserva registros sem essa marca e não apaga dados.

### Rótulos de categoria e status

O banco e a API usam valores internos sem acentos para manter contratos estáveis. A API fornece e a interface exibe estes rótulos:

| Valor interno | Rótulo |
| --- | --- |
| `ti` | TI |
| `rh` | RH |
| `compras` | Compras |
| `financeiro` | Financeiro |
| `infraestrutura` | Infraestrutura |
| `aberto` | Aberto |
| `em_atendimento` | Em Atendimento |
| `concluido` | Concluído |

### Código visual

O código apresentado é derivado de `requests.id` no formato `SOL-000001`. Não existe coluna duplicada para esse valor. No backend, a formatação é equivalente a `fmt.Sprintf("SOL-%06d", id)`; números com mais de seis dígitos não são truncados.

## Índices

| Índice | Coluna(s) | Uso planejado |
| --- | --- | --- |
| `users_pkey` | `users.id` | Índice único criado implicitamente pela chave primária do usuário. |
| `users_username_unique` | `users.username` | Garantir unicidade e localizar usuário no login. |
| `sessions_pkey` | `sessions.token_hash` | Localizar/revogar sessão pelo hash. |
| `idx_sessions_user_id` | `sessions.user_id` | Consultar ou revogar sessões de um usuário. |
| `idx_sessions_expires_at` | `sessions.expires_at` | Verificar/limpar sessões por validade. |
| `requests_pkey` | `requests.id` | Índice único criado implicitamente pela chave primária da solicitação. |
| `idx_requests_requester_id` | `requests.requester_id` | Consultar solicitações por autor e apoiar integridade referencial. |
| `idx_requests_status` | `requests.status` | Filtrar e agregar por status. |
| `idx_requests_category` | `requests.category` | Filtrar por categoria. |
| `idx_requests_created_at` | `requests.created_at DESC` | Ordenar e filtrar pelo período de abertura. |
| `idx_requests_seed_key_unique` | `requests.seed_key`, somente não nulos | Reconhecer cada exemplo do seed e impedir duplicação sem impor unicidade ao título. |

As chaves primárias e a restrição `UNIQUE` de username produzem índices automaticamente no PostgreSQL; os demais são declarados explicitamente nas migrations. Não há índice B-tree em `title`: ele não resolve de forma geral uma busca por substring como `ILIKE '%texto%'`. Para o volume do desafio, a busca usa `ILIKE` com varredura simples. Se medições futuras indicarem necessidade, poderá ser avaliado índice trigram com `pg_trgm`, sem adicioná-lo preventivamente.

## Defaults e responsabilidade do backend

Os defaults de status e timestamps protegem operações SQL diretas e mantêm o schema autoconsistente. Eles não transferem a regra de negócio ao banco: o backend envia explicitamente `status = 'aberto'`, `created_at` e `updated_at` na criação, além de atualizar `updated_at` nas alterações.
