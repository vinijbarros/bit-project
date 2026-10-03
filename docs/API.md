# Contrato da API

O contrato normativo está em `docs/openapi.yaml` (OpenAPI 3.0.3). Este documento resume convenções, segurança e o estado real de implementação. Qualquer mudança em um endpoint deve atualizar os dois arquivos e a rastreabilidade.

## Estado executável

| Operação | Estado atual |
| --- | --- |
| `GET /healthz` | Implementada e pública. |
| `GET /readyz` | Implementada e pública; consulta o banco. |
| `POST /api/v1/auth/login` | Implementada; exige origem confiável. |
| `POST /api/v1/auth/logout` | Implementada e idempotente; exige origem confiável. |
| `GET /api/v1/auth/me` | Implementada e protegida por sessão. |
| `GET /api/v1/requests` | Implementada e protegida; lista e combina filtros. |
| `POST /api/v1/requests` | Implementada e protegida; exige origem confiável. |
| `GET /api/v1/requests/{id}` | Implementada e protegida. |
| `PATCH /api/v1/requests/{id}` | Implementada para autor e estado aberto; exige origem confiável. |
| `DELETE /api/v1/requests/{id}` | Implementada para autor e estado aberto; exige origem confiável. |
| `PATCH /api/v1/requests/{id}/status` | Implementada para qualquer usuário autenticado; exige origem confiável. |
| `GET /api/v1/dashboard` | Implementada e protegida; retorna indicadores globais. |
| `GET /api/v1/metadata` | Implementada e protegida; retorna categorias/status e rótulos estáveis. |

## Convenções

- JSON UTF-8 e prefixo `/api/v1` para domínio; probes permanecem na raiz.
- Rotas com corpo aceitam apenas `Content-Type: application/json` (parâmetro `charset` é permitido).
- Corpo máximo: 1 MiB. Campos desconhecidos, JSON vazio, malformado ou com um segundo valor são rejeitados.
- IDs são inteiros positivos. Instantes de resposta usam RFC 3339 em UTC.
- Códigos internos, categorias e status são estáveis e sem acento; mensagens e rótulos exibíveis ficam em português.
- O servidor devolve e registra `X-Request-ID`; um valor recebido só é reutilizado se tiver 1–64 caracteres seguros (`A-Z`, `a-z`, `0-9`, `.`, `_`, `-`).
- Logs HTTP incluem request ID, método, padrão de rota, status e duração. Corpo, senha, cookie, token e parâmetros da URL não são registrados.

## Erros

Formato único:

```json
{
  "error": {
    "code": "validation_failed",
    "message": "Corrija os campos informados.",
    "fields": {
      "title": ["Informe entre 3 e 150 caracteres."]
    }
  },
  "request_id": "01J..."
}
```

`fields` aparece somente quando o cliente pode corrigir campos específicos. As chaves usam os nomes JSON ou dos parâmetros; cada valor é uma lista de mensagens. Códigos HTTP:

- `400`: JSON, ID, filtros, paginação ou campos inválidos.
- `401`: cookie ausente, inválido ou expirado; não representa falha do banco.
- `403`: autoria/permissão insuficiente ou origem não permitida.
- `404`: recurso ou rota inexistente.
- `409`: conflito com o estado atual, como editar/excluir solicitação não aberta.
- `415`: mídia diferente de `application/json` em rota com corpo.
- `429`: limite de tentativas de login atingido; inclui `Retry-After`.
- `500`: erro interno genérico; SQL, stack trace e detalhes de infraestrutura não são enviados.
- `503`: usado por `/readyz` quando o banco não responde.

Falha de autenticação usa mensagem genérica e não informa se o username existe. Indisponibilidade do banco é erro de infraestrutura (`500` em operações de domínio ou `503` no probe), nunca `401`.

## Sessão e proteção de origem

O login cria pelo menos 32 bytes aleatórios criptográficos e envia sua codificação base64url somente no cookie `portal_session`; apenas o SHA-256 hexadecimal é persistido. O cookie usa `HttpOnly`, `SameSite=Lax`, `Path=/`, duração absoluta configurável (`SESSION_DURATION`, padrão 8 horas) e `Secure=true` obrigatório em produção. O HTTP local usa `SESSION_COOKIE_SECURE=false` explicitamente. Não há renovação deslizante nem refresh token.

Um login bem-sucedido sempre gera sessão nova. Se houver cookie anterior com formato válido, seu hash é revogado na mesma transação que cria a nova sessão, reduzindo fixação. Login retorna apenas `id`, `username` e `display_name`; hash e token nunca aparecem no JSON.

Logout com cookie válido remove a sessão antes de limpar o cookie. Repetições sem cookie retornam `204`, tornando a operação idempotente sem deixar uma sessão conhecida reutilizável. Uma falha de banco ao revogar retorna `500` e não é convertida em falso sucesso.

Todas as operações de domínio, inclusive `metadata` e `dashboard`, exigem o cookie. Para reduzir CSRF, `POST`, `PUT`, `PATCH` e `DELETE` exigem cabeçalho `Origin` exatamente igual a uma origem confiável configurada. Desenvolvimento adota `http://127.0.0.1:5173`; não há correspondência por sufixo, curingas nem fallback para `Referer`. Origem ausente ou diferente retorna `403 origin_not_allowed`. O frontend chama caminhos relativos e não depende de CORS.

O guard de origem está conectado a login/logout e a todas as operações mutáveis existentes de solicitações.

## Criação, detalhe, edição e exclusão

`POST /api/v1/requests` aceita exclusivamente `title`, `description` e `category`. O servidor remove espaços externos, conta limites por pontos de código Unicode e exige título de 3–150 caracteres e descrição de 10–5000. A categoria deve ser `ti`, `rh`, `compras`, `financeiro` ou `infraestrutura`. Autor vem da sessão, status é explicitamente `aberto` e datas são definidas no servidor/banco; qualquer campo adicional é rejeitado. O sucesso retorna `201`, `Location: /api/v1/requests/{id}` e o recurso criado.

`GET /api/v1/requests/{id}` é permitido a qualquer usuário autenticado. A resposta não contém hash de senha nem dados de sessão e apresenta categoria/status como `{value, label}`, solicitante seguro e as permissões calculadas `can_edit`/`can_delete`. O código visual é derivado do ID como `SOL-000001`, sem coluna duplicada.

`PATCH /api/v1/requests/{id}` aceita somente os campos opcionais `title`, `description` e `category`. Campo ausente é preservado; `null`, valor vazio após trim e campo desconhecido são inválidos. Payload sem nenhum campo ou que resulte nos mesmos valores retorna `400 no_changes`. Apenas o autor pode editar (`403` para outro usuário), e apenas enquanto o status for `aberto` (`409 request_not_open`). Status não é alterado por esta rota.

`DELETE /api/v1/requests/{id}` segue a mesma regra de autoria/estado e retorna `204` sem corpo. A exclusão é física. Tanto atualização quanto exclusão usam SQL condicional por ID, autor e estado, de modo que uma mudança concorrente de status não abre uma janela para mutação indevida. O diagnóstico posterior diferencia `404`, `403` e `409` sem desfazer essa garantia.

Não há controle de versão otimista nesta versão: duas edições concorrentes válidas do mesmo autor sobre uma solicitação ainda aberta seguem a política de última gravação vence. Essa limitação não afeta a atomicidade da regra de autoria/status.

## Alteração de status

`PATCH /api/v1/requests/{id}/status` aceita exclusivamente `status`, com um dos valores `aberto`, `em_atendimento` ou `concluido`. Campo ausente, `null`, outro tipo, valor desconhecido e campo adicional retornam `400`. Os rótulos de resposta são, respectivamente, `Aberto`, `Em Atendimento` e `Concluído`.

Como o PDF não define perfis nem uma sequência obrigatória, qualquer usuário autenticado pode alterar qualquer solicitação entre quaisquer dos três estados, inclusive reabrir uma concluída. Não há papel administrativo nem máquina de estados adicional. Repetir o status atual é idempotente: retorna `200` com a representação atual e preserva `updated_at`. Uma mudança efetiva altera status e `updated_at` atomicamente.

O repository abre uma transação, bloqueia a linha com `SELECT ... FOR UPDATE` e decide sob o mesmo lock entre a resposta idempotente e o `UPDATE`. Edições/exclusões concorrentes usam condições de status em suas próprias escritas: após `em_atendimento` ou `concluido`, elas falham com `409`; após reabertura, somente o autor volta a poder executá-las. Se uma exclusão vencer a disputa, a mudança de status retorna `404`.

Exemplo completo usando cookie jar, sem expor o token:

```sh
# Login e armazenamento automático do cookie HttpOnly no arquivo local.
curl -i -c /tmp/portal-cookies.txt \
  -H 'Content-Type: application/json' \
  -H 'Origin: http://127.0.0.1:5173' \
  --data '{"username":"colaborador1","password":"DemoLocal-Colaborador1"}' \
  http://127.0.0.1:8080/api/v1/auth/login

# Em Atendimento.
curl -i -b /tmp/portal-cookies.txt \
  -X PATCH \
  -H 'Content-Type: application/json' \
  -H 'Origin: http://127.0.0.1:5173' \
  --data '{"status":"em_atendimento"}' \
  http://127.0.0.1:8080/api/v1/requests/1/status

# Conclusão.
curl -i -b /tmp/portal-cookies.txt \
  -X PATCH \
  -H 'Content-Type: application/json' \
  -H 'Origin: http://127.0.0.1:5173' \
  --data '{"status":"concluido"}' \
  http://127.0.0.1:8080/api/v1/requests/1/status

# Reabertura; repetir este comando é idempotente.
curl -i -b /tmp/portal-cookies.txt \
  -X PATCH \
  -H 'Content-Type: application/json' \
  -H 'Origin: http://127.0.0.1:5173' \
  --data '{"status":"aberto"}' \
  http://127.0.0.1:8080/api/v1/requests/1/status
```

## Dashboard

`GET /api/v1/dashboard` exige sessão válida e não recebe filtros. Os indicadores são globais, incluindo solicitações de todos os usuários. Cada chamada consulta o estado atual do PostgreSQL, sem cache, em uma única instrução de agregação; assim, os quatro números pertencem ao mesmo snapshot da instrução. A constraint de status e uma verificação no service asseguram `total = abertas + em_atendimento + concluidas`. Falha de banco retorna `500`, nunca contadores zerados artificiais.

Base vazia:

```json
{"data":{"total":0,"abertas":0,"em_atendimento":0,"concluidas":0}}
```

Exemplo com dados:

```json
{"data":{"total":8,"abertas":3,"em_atendimento":2,"concluidas":3}}
```

Os nomes JSON são estáveis. A interface implementada exibe `Quantidade total de solicitações`, `Abertas`, `Em Atendimento` e `Concluídas` em cards globais.

## Limite de login

Tentativas inválidas são limitadas por `username` normalizado combinado ao IP de `RemoteAddr`, sem confiar em `X-Forwarded-For`. O padrão permite 5 falhas em 15 minutos. Um username diferente não é bloqueado pela falha de outro; sucesso remove seu contador.

O limiter é em memória, limitado por `LOGIN_RATE_LIMIT_MAX_ENTRIES` (padrão 10.000) e remove a entrada mais antiga ao atingir o teto, evitando crescimento indefinido. Contadores não são compartilhados entre múltiplas instâncias e são perdidos no reinício; uma solução distribuída fica como melhoria de implantação horizontal.

## Metadados

`GET /api/v1/metadata` fornece as opções aceitas pela API na ordem de exibição, sob `data.categories` e `data.statuses`. Cada opção contém somente `value` interno e `label` em português. A rota é protegida e centraliza os rótulos usados pelos filtros/formulários, sem consultar ou expor dados sensíveis.

## Filtros e paginação

`GET /api/v1/requests` aceita:

| Parâmetro | Regra |
| --- | --- |
| `date_from` | Data local real `YYYY-MM-DD`, início do dia em `America/Sao_Paulo`, inclusivo. |
| `date_to` | Data local real `YYYY-MM-DD`, fim representado pelo início do dia seguinte, exclusivo. Deve ser igual ou posterior a `date_from`. |
| `category` | Um dos cinco valores internos. |
| `status` | Um dos três valores internos. |
| `q` | Substring somente do título, após trim, usando `ILIKE`; máximo de 150 caracteres. `%`, `_` e `\` são tratados literalmente. |
| `page` | Inteiro positivo; padrão `1`. |
| `page_size` | Inteiro entre 1 e 100; padrão `20`. |

Todos os filtros são opcionais e combinados por `AND`. Valor vazio ou composto apenas de espaços equivale à ausência para filtros textuais/datas; `page` e `page_size` vazios usam os padrões. Apenas uma das datas pode ser enviada. A conversão calcula o início de cada dia civil no fuso, inclusive em transições históricas, sem somar 24 horas cegamente.

Parâmetro singular repetido, desconhecido ou inválido retorna `400 invalid_query_parameter`. A ordenação é fixa em `created_at DESC, id DESC`; o cliente não envia SQL de ordenação. Count e itens usam a mesma cláusula parametrizada dentro de uma transação de leitura `REPEATABLE READ`. A pesquisa não consulta descrição e o escape `ILIKE ... ESCAPE` impede que `%` e `_` do usuário virem curingas. O B-tree existente não é apresentado como otimização para substring.

Resposta exata:

```json
{
  "items": [
    {
      "id": 9,
      "code": "SOL-000009",
      "title": "Relatório de acesso",
      "category": {"value": "ti", "label": "TI"},
      "requester": {"id": 2, "username": "colaborador2", "display_name": "Colaborador 2"},
      "created_at": "2026-10-01T15:00:00Z",
      "status": {"value": "aberto", "label": "Aberto"}
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_items": 1,
    "total_pages": 1
  }
}
```

Página além do fim retorna `items: []`, preservando os totais. Exemplo completo, reutilizando o cookie jar criado no login:

```sh
curl -i -b /tmp/portal-cookies.txt --get \
  --data-urlencode 'date_from=2026-10-01' \
  --data-urlencode 'date_to=2026-10-31' \
  --data-urlencode 'category=ti' \
  --data-urlencode 'status=aberto' \
  --data-urlencode 'q=acesso 100%_interno' \
  --data-urlencode 'page=1' \
  --data-urlencode 'page_size=20' \
  http://127.0.0.1:8080/api/v1/requests

# Somente limite inicial e defaults de paginação.
curl -i -b /tmp/portal-cookies.txt --get \
  --data-urlencode 'date_from=2026-10-01' \
  http://127.0.0.1:8080/api/v1/requests
```

## Responsabilidades das camadas

- Handlers: HTTP, JSON, cookies, parâmetros e tradução de erros.
- Services: autenticação, autorização, transições e validações de negócio.
- Repositories: SQL parametrizado via `database/sql`/pgx e tradução de resultados de persistência.

Handlers não contêm SQL. Repositories recebem `context.Context`; erros internos são registrados no servidor com request ID e convertidos em respostas seguras.
