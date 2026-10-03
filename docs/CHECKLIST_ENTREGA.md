# Checklist de entrega

Data da auditoria: 03/10/2026  
Prazo do enunciado: 05/10/2026

## Resultado executivo

Os itens obrigatórios do PDF estão implementados e foram verificados em instalação limpa isolada. A entrega está tecnicamente pronta para revisão/envio pelo responsável. Nenhum push, e-mail, publicação ou deploy externo foi realizado automaticamente.

As pendências reais que não bloqueiam os obrigatórios são: primeira execução do workflow no GitHub, deploy/TLS externos e melhorias futuras descritas no Memorial. O pacote local existente em `artifacts/bit-project-entrega-2026-10-05.tar.gz` foi gerado antes da revisão documental posterior e, por isso, deve ser recriado se for escolhido como meio de entrega. Um repositório atualizado não depende desse pacote.

## Auditoria funcional e técnica

| Item auditado | Implementação real | Teste ou evidência executada | Resultado |
| --- | --- | --- | --- |
| Login por usuário/senha | `POST /api/v1/auth/login`; `service/auth.go`; `http/auth.go`; `LoginPage.tsx` | `TestAuthenticationPersistenceAndFailures`, AuthFlow, E2E e `01-login.png` | Validado |
| Controle/persistência de sessão | Cookie HttpOnly + hash em `sessions`; middleware `/auth/me` | integração de expiração/reinício; após reiniciar containers, `/me=200` com o mesmo cookie | Validado |
| Logout/revogação | `POST /auth/logout` remove sessão antes de limpar cookie | integração, testes HTTP/frontend e E2E com acesso protegido posterior | Validado |
| Acesso somente autenticado | Middleware em metadata, requests e dashboard; `ProtectedRoute` | HTTP anônimo `401`, testes de rota e E2E redirecionando ao login | Validado |
| Criação com três campos | `POST /requests`; formulário compartilhado | testes de mass assignment/Unicode; E2E; `05-criacao-sucesso.png` | Validado |
| Autor, data e status automáticos | Service/repository usam usuário da sessão, UTC atual e `aberto` | integração PostgreSQL e DTO real da criação | Validado |
| Cinco categorias | domínio/metadata/constraints e selects | constraints, metadata, seed e formulários | Validado |
| Editar somente aberta e pelo autor | `PATCH /requests/{id}` com SQL condicional | dois usuários: `403`; fechada: `409`; E2E edita como autor | Validado |
| Excluir somente aberta e pelo autor | `DELETE /requests/{id}` com SQL condicional e confirmação | integração/E2E: cancelamento, `403`, `409`, `204` e `404` posterior | Validado |
| Seis colunas da listagem | `RequestsPage` exibe Código, Título, Categoria, Solicitante, Data e Status | RTL, E2E e `03-listagem-completa.png` | Validado |
| Consulta de detalhes | `GET /requests/{id}` e `RequestDetailPage` por URL | testes HTTP/RTL, reload E2E e `06-detalhes-aberto.png` | Validado |
| Três status e alteração | `PATCH /requests/{id}/status`; metadata e seletor | matriz 3×3, integração, E2E, `07`/`08` PNGs | Validado |
| Mudança global/reabertura | qualquer autenticado muda qualquer solicitação; `SELECT FOR UPDATE` | segundo usuário no E2E; conclusão e reabertura; conflito concorrente | Validado |
| Filtro por período | datas civis São Paulo convertidas a limites UTC | unitários de dias/horário de verão, integração e E2E de um dia | Validado |
| Filtros categoria/status/título | SQL parametrizado, `ILIKE` literal e URL sincronizada | integração com `%`, `_`, acentos/aspas; RTL/E2E; `04-filtros-combinados.png` | Validado |
| Paginação/ordem | `page`, `page_size`, total e ordem criação/ID decrescente | unitários/integração e página além do fim | Validado |
| Dashboard com quatro indicadores | `GET /dashboard`, agregação única e tela global | vazio/populado, mutações 5→6→5, E2E e `02-dashboard.png` | Validado |
| API/persistência/regras | handlers → services → repositories → PostgreSQL | Go unitário/HTTP, integração real, E2E e reinício | Validado |
| JSON, validações e erros | parser 1 MiB/estrito, erros estáveis, IDs e mídia | testes de JSON inválido, excesso, unknown fields, 400/401/403/404/409/415/500 | Validado |
| CSRF e logs seguros | origem exata; logs por rota/request ID sem corpo/query | testes de origem/logs e revisão dos middlewares | Validado |
| Organização de backend | `cmd`, `domain`, `http`, `service`, `repository`, `db/migrations` | build/vet/testes e auditoria de imports/SQL | Validado |
| Componentes/navegação/UX | layout, guards, feedback, badges, modal e formulários reutilizados | 53 testes RTL/Vitest, E2E e capturas | Validado |
| Responsividade/acessibilidade | CSS próprio, foco, labels, badges textuais, tabela rolável | smokes 1440×900 e 375×812; `09-layout-movel.png` | Validado no escopo |
| Schema/scripts | migrations `00001` e `00002`, Goose e comandos | banco vazio até v2; repetição sem reaplicar; catálogo consultado | Validado |
| Dicionário de dados | `DICIONARIO_DE_DADOS.md` | comparação com 18 colunas, 19 constraints e 11 índices reais | Validado |
| README | visão, stack, funcionalidades, variáveis, instalação, acesso, testes, deploy e troubleshooting | comandos confrontados com Makefile/Compose/manifests | Revisado |
| Memorial obrigatório | `MEMORIAL_TECNICO_DE_DESENVOLVIMENTO.md` | tecnologias e decisões confrontadas com código/manifestos | Revisado |
| Evidências reais | nove PNGs + roteiro/README | Playwright contra instalação limpa, 1 teste em 30,8 s; imagens inspecionadas | Validado |
| Instruções de deploy | README + `DEPLOY.md` | Compose isolado e configuração de produção renderizada em revisão anterior | Validado localmente |
| Docker/Compose | imagens multi-stage, healthchecks, migration task, seed profile | build limpo; serviços healthy; persistência após restart | Validado |
| Testes automatizados | Go, integração, race, Vitest/RTL e Playwright | resultados finais abaixo | Validado |
| CI | `.github/workflows/ci.yml`, quatro jobs, PostgreSQL/OpenAPI/E2E/Docker | comandos equivalentes reproduzidos localmente; YAML validado | Implementado; GitHub pendente |
| CD externo | inexistente por decisão de escopo | nenhum deploy/publicação executado | Melhoria futura, não alegado |

## Verificações finais observadas

- Docker build: backend e frontend concluídos; `npm ci` instalou 195 pacotes no estágio limpo.
- Instalação Compose: PostgreSQL/API/frontend `healthy`; migration v1/v2 em `Exited (0)`.
- Seed: primeira execução `users_created=2`, `requests_created=5`; repetição criou `0/0`.
- Backend: gofmt limpo; `go vet`; testes unitários/HTTP; integração PostgreSQL em 2,113 s; race; build — aprovados.
- Frontend: 8 arquivos/53 testes em 56,67 s; typecheck, lint e build — aprovados.
- E2E: 3 casos com dois usuários em 50,8 s — aprovados.
- Evidências: 1 roteiro em 30,8 s — aprovado; nove PNGs inspecionados.
- Persistência: mesma sessão e solicitação consultáveis após restart; refresh de rota interna retornou SPA.
- OpenAPI: válido no Redocly CLI 2.57.0, com três avisos de estilo e nenhum erro; não foram inventadas licença ou respostas inexistentes para silenciá-los.
- `git diff --check`, auditoria de arquivos sensíveis e conteúdo do pacote foram executados na auditoria final anterior. A revisão documental posterior não repetiu esses comandos.

## Higiene do material

- `.env` real, `.git`, `.git-local`, `node_modules`, builds, caches, relatórios Playwright, binários, volumes e dumps são excluídos do pacote.
- `.env.example`, `frontend/.env.example`, `go.sum`, `package-lock.json`, migrations, workflow e documentação são preservados.
- As únicas senhas documentadas são públicas e exclusivas de demonstração/teste.
- O PDF de referência e `AGENTS.md` são contexto interno e não são necessários para executar a entrega; não entram no pacote fonte.

## Pendências não bloqueantes

1. Executar o workflow após o responsável versionar/enviar as alterações ao GitHub.
2. Configurar infraestrutura real de HTTPS, segredos, backup, observabilidade e rate limit distribuído antes de produção.
3. Opcionalmente produzir vídeo; o PDF o trata como opcional e as capturas reais já foram geradas.
4. O responsável deve revisar autoria/compreensão e efetuar o envio até 05/10/2026.
5. Se a entrega ocorrer por arquivo compactado, regenerar pacote e checksum para incluir a consolidação documental mais recente; esta ação não foi executada na revisão exclusivamente documental.

## Pacote local

- Caminho: `artifacts/bit-project-entrega-2026-10-05.tar.gz`
- Estado: **desatualizado em relação à revisão documental mais recente**; não usar para envio sem regenerar.
- O tamanho e SHA-256 históricos estão no arquivo externo `artifacts/bit-project-entrega-2026-10-05.tar.gz.sha256`, mas mudarão quando o pacote for recriado.
