# MEMORIAL TÉCNICO DE DESENVOLVIMENTO

## 1. Identificação e objetivo

Este memorial descreve a solução efetivamente construída para o Portal de Solicitações Internas da seleção de Desenvolvedor(a) de Sistemas Júnior da bit Soluções, com entrega prevista para 05/10/2026. O sistema permite autenticar colaboradores, registrar e acompanhar solicitações, consultar e filtrar a visão global, alterar status e visualizar indicadores.

O documento distingue requisitos do enunciado de decisões complementares do projeto. Ele não afirma deploy externo, alta disponibilidade, teste de carga ou certificação de segurança: essas atividades não foram realizadas.

## 2. Tecnologias efetivamente utilizadas

As versões abaixo estão fixadas nos manifestos, imagens ou workflow. Dependências transitivas não são apresentadas como escolhas arquiteturais próprias. Em cada linha, “uso real e motivo” reúne o contexto, a decisão e o benefício buscado; “alternativa e impacto” registra a análise técnica atual e o compromisso aceito, sem afirmar que todas as alternativas foram experimentadas. A última coluna aponta a evidência no repositório.

| Tecnologia | Uso real e motivo: contexto, decisão e benefício | Alternativa e impacto: comparação, limitação ou compromisso | Evidência |
| --- | --- | --- | --- |
| Go 1.22.2 | Linguagem da API, migrations e seed. A tipagem estática, compilação em binário e biblioteca padrão atendem bem uma API pequena. | Em comparação com um runtime dinâmico, exige mais código explícito, mas facilita rastrear erros e contratos. Binário único simplifica distribuição; isso não implica desempenho comprovado por benchmark. | `backend/go.mod`, `backend/cmd/` e `backend/Dockerfile`. |
| `net/http` | Roteamento por método/padrão, servidor, cookies e middleware sem framework web adicional. | Gin/Echo ofereceriam conveniências, porém adicionariam uma abstração desnecessária ao porte do desafio. O custo aceito foi implementar helpers de JSON, erros e middleware. | `backend/internal/http/router.go` e demais arquivos de `backend/internal/http/`. |
| `log/slog` | Produz logs estruturados da API e dos comandos com método, rota, status, duração e request ID, sem registrar corpos ou credenciais. | Uma biblioteca externa de logging poderia oferecer mais formatadores e integrações, mas a biblioteca padrão atende ao escopo. Logs locais não substituem centralização e alertas de produção. | `backend/cmd/api/main.go`, `backend/cmd/migrate/main.go`, `backend/cmd/seed/main.go` e middleware HTTP. |
| `database/sql` | Pool, transações, contexts e execução de SQL parametrizado. | Um ORM reduziria SQL repetitivo, mas esconderia parte da modelagem e das condições atômicas importantes. SQL explícito exige scans e fechamento de recursos manuais. | `backend/internal/repository/` e `backend/internal/repository/database/`. |
| pgx 5.7.1 (`stdlib`) | Driver PostgreSQL usado por `database/sql`, combinando a interface padrão com suporte específico ao banco escolhido. | A API nativa do pgx ofereceria recursos adicionais, mas afastaria a implementação da interface `database/sql` definida para o projeto. | `backend/go.mod` e abertura do driver em `backend/internal/repository/database/`. |
| Goose 3.24.0 | Aplica migrations SQL versionadas pelos comandos `up`, `status`, `down` e `version`, evitando criação manual não rastreada. | Scripts soltos exigiriam controle externo de versão. Goose torna a criação reproduzível, mas `down` continua destrutivo e requer decisão operacional. | `backend/cmd/migrate/main.go` e `backend/db/migrations/`. |
| `golang.org/x/crypto` 0.31.0 / bcrypt | Gera e compara hashes de senha no seed/login, atendendo à necessidade de não persistir senha reversível. | SHA-256 simples seria inadequado para senhas por ser rápido. bcrypt tem limite de 72 bytes, validado explicitamente; custo e migração de hashes precisariam de revisão periódica em produção. | `backend/internal/service/auth.go`, `backend/internal/seed/seed.go` e `backend/go.mod`. |
| `crypto/rand` e SHA-256 da biblioteca padrão | Gera token opaco imprevisível e persiste apenas seu digest para permitir localização/revogação sem guardar o segredo original. | JWT reduziria consultas, mas tornaria revogação imediata mais complexa. A sessão no banco adiciona uma consulta a cada requisição protegida. | `backend/internal/service/auth.go` e `backend/internal/repository/auth.go`. |
| PostgreSQL 17.6 | Banco relacional para `TIMESTAMPTZ`, constraints, chaves estrangeiras, índices, isolamento e locks. | SQLite simplificaria a demonstração, mas reproduziria menos da concorrência usada pelas regras. PostgreSQL aumenta o pré-requisito local, compensado pelo Compose. | `compose.yaml`, `compose.test.yaml` e `backend/db/migrations/`. |
| SQL (dialeto PostgreSQL) | Define schema e queries parametrizadas, deixando relações e operações atômicas visíveis. | Query builders poderiam reduzir montagem estrutural, mas adicionariam abstração. SQL direto exige revisão cuidadosa de parâmetros e scans. | `backend/db/migrations/` e `backend/internal/repository/`. |
| SQL `CHECK`, FKs e índices B-tree | Reforça formato, limites, categorias, status e integridade mesmo fora do fluxo HTTP. | Validar só na aplicação permitiria gravações diretas inconsistentes. As regras críticas ficam duplicadas de propósito; B-tree não é apresentado como solução para `ILIKE '%texto%'`. | Migrations `00001_create_core_tables.sql` e `00002_add_request_seed_key.sql`. |
| React 19.3.0 / React DOM 19.3.0 | Implementa a SPA e seus estados de autenticação, carga, vazio, erro e mutação em componentes. | Templates server-side reduziriam JavaScript, mas não atenderiam tão diretamente ao frontend separado adotado. React exige cuidado com efeitos e estado remoto. | `frontend/package.json`, `frontend/src/pages/` e `frontend/src/components/`. |
| TypeScript 6.0.3 | Tipifica DTOs, formulários, filtros e estado em modo strict para reduzir divergências durante manutenção. | JavaScript puro teria configuração menor, mas menos verificação estática. Os tipos não substituem validação de dados recebidos nem o backend. | `frontend/tsconfig*.json` e `frontend/src/types/api.ts`. |
| Vite 8.3.2 / plugin React 6.1.1 | Fornece servidor de desenvolvimento, proxy `/api`, transformação e build da SPA com configuração curta. | Um bundler configurado manualmente permitiria mais controle, mas aumentaria trabalho sem benefício proporcional ao prazo. | `frontend/vite.config.ts` e `frontend/package.json`. |
| React Router DOM 7.18.4 | Organiza rotas públicas/protegidas, parâmetros, navegação, URL direta e página 404. | Um roteador artesanal eliminaria uma dependência, mas recriaria histórico, parâmetros e guards. | `frontend/src/app/routes/router.tsx` e guards no mesmo diretório. |
| Fetch do navegador | Centraliza URL relativa, cookie same-origin, JSON, cancelamento e erros tipados sem armazenar token no frontend. | Axios daria interceptors, mas seria dependência adicional para capacidades já atendidas. O cliente ainda precisa tratar manualmente respostas não JSON e `204`. | `frontend/src/api/client.ts` e `frontend/src/services/`. |
| CSS próprio | Implementa layout, badges, formulários, tabela rolável, cards, foco e breakpoints responsivos sem pacote visual. | Um framework de UI aceleraria componentes prontos, mas aumentaria dependências. CSS próprio demanda revisão visual e de acessibilidade manual. | `frontend/src/styles/`. |
| HTML semântico / HTTP / JSON | Usa labels, roles, formulários, códigos HTTP e DTOs como contrato claro entre navegador e API. | GraphQL ou protocolos adicionais não trariam benefício proporcional. REST/JSON permanece simples, mas exige manter OpenAPI e tipos coerentes. | Componentes React, `backend/internal/http/` e `docs/openapi.yaml`. |
| Node.js 24.14.0 / npm 11.9.0 | Executa as ferramentas do frontend e instala versões fixadas por `npm ci`. | Outros gerenciadores seriam válidos, mas npm já acompanha o runtime. O lockfile melhora reprodução, embora não elimine diferenças de sistema operacional. | `frontend/package.json`, `frontend/package-lock.json` e `frontend/Dockerfile`. |
| Docker / Compose v2 | Constrói imagens multi-stage e coordena PostgreSQL, migration, API, frontend e seed opcional. | A instalação manual é suportada, mas mais sensível a versões e portas. Compose melhora a execução local; não oferece alta disponibilidade. | `backend/Dockerfile`, `frontend/Dockerfile`, `compose.yaml` e `compose.test.yaml`. |
| Alpine Linux | Base versionada das imagens de build/runtime para conter apenas o necessário. | Debian costuma facilitar diagnóstico e compatibilidade, mas tende a produzir imagens maiores. Alpine pode exigir atenção a bibliotecas nativas. | `backend/Dockerfile` e `frontend/Dockerfile`. |
| Nginx 1.27.4 | Serve os arquivos estáticos, aplica fallback da SPA e encaminha `/api` e probes para Go. | Servir a SPA pela API acoplaria responsabilidades. Nginx não fornece sozinho TLS, gestão de certificados ou observabilidade de produção. | `frontend/Dockerfile` e `frontend/nginx.conf`. |
| OpenAPI 3.0.3 | Registra rotas, schemas, cookie, filtros, respostas e erros em contrato legível por ferramentas. | Texto narrativo isolado seria mais fácil de divergir. O YAML ainda depende de revisão quando handlers mudam. | `docs/openapi.yaml` e resumo em `docs/API.md`. |
| Redocly CLI 2.57.0 | Faz lint semântico do OpenAPI com versão fixada na instrução e na CI. | Validar só YAML encontraria apenas problemas sintáticos. Avisos de estilo não equivalem a erro nem a teste da API. | `Makefile`, `README.md` e `.github/workflows/ci.yml`. |
| Teste padrão do Go / `httptest` | Verifica services, configuração, handlers e middleware com foco em comportamento observável. | Mocks extensos poderiam isolar mais, mas esconderiam integração; por isso interfaces pequenas são complementadas pelo banco real. | Arquivos `*_test.go` em `backend/internal/`. |
| PostgreSQL de integração | Exercita migrations, repositories, autenticação, autorização, filtros, dashboard e concorrência com semântica real do banco. | Banco em memória seria mais rápido, porém não reproduziria locks, `ILIKE` e PostgreSQL. Exige banco descartável e trava explícita de reset. | `backend/internal/integration/backend_integration_test.go` e `compose.test.yaml`. |
| Detector de corrida do Go | Instrumenta os pacotes HTTP e service para detectar acessos concorrentes durante os testes executados. | Tem compilação mais lenta e não prova ausência universal de corridas. | Comandos registrados em `docs/TESTES.md` e no job backend da CI. |
| Vitest 5.0.3 | Executa os testes do frontend compartilhando a base ESM/Vite. | Jest seria alternativa madura, mas implicaria configuração paralela. Vitest não substitui navegador real. | `frontend/package.json`, `frontend/vite.config.ts` e `frontend/src/**/*.test.*`. |
| React Testing Library 16.3.3 / user-event 14.6.7 | Exercita a interface por papel, label e texto, reduzindo acoplamento a CSS e estrutura interna. | Testes de componente usam API simulada e não comprovam persistência nem rede. | Testes em `frontend/src/features/` e `frontend/src/api/client.test.ts`. |
| jsdom 29.1.1 | Fornece DOM leve aos testes de componente. | É mais rápido que E2E, porém não reproduz layout completo nem navegador real. | Dependência em `frontend/package.json` e configuração em `frontend/vite.config.ts`. |
| Playwright 1.63.0 / Chromium | Executa E2E contra a aplicação real, smokes responsivos e roteiro das capturas. | Testes manuais seriam menos repetíveis. A suíte cobre Chromium, não todos os navegadores ou dispositivos. | `frontend/playwright.config.ts`, `frontend/e2e/`, `frontend/evidence/` e `docs/evidencias/`. |
| ESLint 10.11.0 / typescript-eslint 8.71.0 | Aplica regras estáticas ao frontend e falha com warnings. | Não substitui typecheck nem teste, mas identifica padrões problemáticos antes da execução. | `frontend/eslint.config.js`, `frontend/package.json` e workflow. |
| `tsc`, `gofmt`, `go vet` | Fazem verificação de tipos/build, formatação e análise estática com ferramentas das próprias stacks. | Reduzem divergências mecânicas, mas não provam correção funcional ou segurança. | Scripts do `frontend/package.json`, `Makefile` e `.github/workflows/ci.yml`. |
| GitHub Actions | Define CI em quatro jobs: backend, frontend/OpenAPI, E2E e imagens Docker. | Execução apenas local depende de disciplina; a CI automatiza em push/PR. O workflow ainda não tem execução remota registrada e não implementa CD. | `.github/workflows/ci.yml`. |
| Git | Versiona código, lockfiles, migrations, workflow e documentação. | Um pacote isolado perderia histórico e revisão incremental; ainda existe como alternativa local. | Estrutura do repositório e arquivos de controle de versão. |
| Make 4.3 (opcional) | Centraliza atalhos, mantendo comandos equivalentes documentados para quem não usa Make. | Scripts separados seriam possíveis. O Makefile é conveniência, não requisito da aplicação. | `Makefile` e seção de execução do `README.md`. |

Não foram utilizados cloud, ORM, Next.js, microserviços, biblioteca global de estado, framework visual ou serviço externo de autenticação.

## 3. Estrutura e arquitetura

A solução é um monólito modular com três processos de execução: PostgreSQL, uma API Go e uma SPA React servida pelo Nginx. “Monólito” aqui não significa arquivo único: o backend separa tradução HTTP, negócio e persistência, enquanto o frontend possui páginas, componentes, features, services e tipos.

```text
Navegador
  └─ React + React Router
       └─ JSON/HTTP relativo + cookie HttpOnly
            └─ Nginx (/api -> API Go)
                 └─ handler -> service -> repository -> PostgreSQL

Goose -> migrations versionadas
seed  -> usuários/dados sintéticos, somente quando habilitado
```

### Backend

- `cmd/api`: composição das dependências, timeouts, servidor e encerramento gracioso.
- `cmd/migrate`: interface explícita do Goose.
- `cmd/seed`: carga demonstrativa separada da inicialização.
- `internal/http`: handlers, DTOs, cookies, parsing estrito, erros e middlewares.
- `internal/service`: validações, autorização, sessão e regras de negócio.
- `internal/repository`: SQL parametrizado, transações e interpretação de resultados.
- `internal/domain`: tipos de domínio compartilhados internamente.
- `db/migrations`: fonte versionada do schema.

As dependências são montadas por construtores. Services dependem de interfaces pequenas dos repositories para testes; isso é injeção de dependência explícita, não um container de DI. O padrão repository isola SQL, e a camada service impede que regras de autoria/status sejam apenas decisões da tela.

### Frontend

- `app/routes`: roteamento e guards público/protegido.
- `pages`: composição de cada tela.
- `components`: layout, campos, feedback, badges e confirmação reutilizáveis.
- `features`: lógica compartilhada de autenticação, filtros e formulários.
- `services` e `api`: operações HTTP e tratamento comum.
- `types`: contrato TypeScript.
- `styles`: CSS base, componentes e layout.

Não há cache remoto global. Cada página consulta novamente a API ao montar; isso simplifica invalidação após mutações e impede que dados de uma sessão sejam reaproveitados depois do logout.

## 4. Modelagem e persistência

O schema possui `users`, `sessions` e `requests`. Usuários são provisionados; não há exclusão de usuários nem cadastro público. `sessions` usa o hash do token como chave. `requests` referencia o solicitante e possui uma `seed_key` opcional, única apenas quando não nula, para reconhecer exemplos sem tornar títulos globalmente únicos.

Categorias e status são `VARCHAR` com `CHECK`, em vez de enum PostgreSQL. O conjunto é pequeno, mas uma migration futura pode alterá-lo sem depender de operações específicas de enum. O código `SOL-000001` é calculado do `id`; não existe coluna duplicada.

Os instantes usam `TIMESTAMPTZ` e são tratados como UTC. Índices atendem autenticação/sessão, autoria, categoria, status e criação. A busca no título usa `ILIKE` com `%`, `_` e barra escapados como literais; não se afirma que o B-tree resolva substring. Para escala maior, `pg_trgm` só deveria ser introduzido após medição.

Edição e exclusão utilizam escrita SQL condicional por ID, autor e status. Alteração de status usa transação e `SELECT ... FOR UPDATE`. Essas estratégias evitam a janela simples entre “consultar regra” e “gravar”. Não existe controle de versão otimista para duas edições simultâneas válidas do mesmo autor: prevalece a última gravação.

## 5. Autenticação, sessão, logout e CSRF

O login normaliza/valida o username, compara bcrypt e devolve mensagem genérica para credenciais inválidas. Em sucesso, gera no mínimo 32 bytes por `crypto/rand`, envia o token base64url apenas no cookie `portal_session` e persiste somente SHA-256 hexadecimal. Um cookie anterior válido é revogado na mesma transação da nova sessão.

O cookie usa `HttpOnly`, `SameSite=Lax`, `Path=/` e expiração absoluta padrão de oito horas. `Secure=false` existe somente para HTTP local; produção recusa essa configuração. Não há refresh ou renovação deslizante. O middleware consulta a sessão persistida, portanto reiniciar a API não desconecta uma sessão ainda válida.

Logout revoga a linha antes de limpar o cookie. Sem cookie, retorna `204`; com falha do banco, retorna erro e não apresenta falso sucesso. O frontend só remove o usuário em memória depois da confirmação.

Mutações exigem `Origin` exatamente presente em `TRUSTED_ORIGINS`, combinado a `SameSite=Lax`. Não há curinga ou confiança automática em `Referer`/`X-Forwarded-For`. Essa política reduz CSRF no modelo same-origin adotado, mas não substitui HTTPS, configuração correta do proxy e revisão de segurança para produção.

O limite de login usa username normalizado + `RemoteAddr`, cinco falhas em quinze minutos por padrão e mapa limitado. Por estar em memória, reinicia com o processo e não é compartilhado entre réplicas; uma implantação horizontal exigiria Redis ou outro armazenamento central, definição segura do IP do cliente e política operacional.

## 6. Comunicação e contrato

O navegador chama `/api/v1` por caminho relativo. Em desenvolvimento, Vite encaminha `/api` à API; na imagem final, Nginx faz o proxy. Assim, cookie e origem permanecem same-origin e CORS não é necessário.

Corpos JSON são limitados a 1 MiB, exigem `application/json`, rejeitam campos desconhecidos e conteúdo adicional. Erros seguem `error.code`, `error.message`, `error.fields` opcional e `request_id`. Logs usam método, padrão da rota, status, duração e request ID, sem corpo, query string, cookie ou credenciais.

`/healthz` informa processo vivo. `/readyz` testa o banco e retorna `503` quando indisponível, sem fingir que falha de infraestrutura é ausência de usuário/recurso.

## 7. Decisões de negócio complementares

O PDF define o núcleo funcional, mas deixa ambiguidades. A tabela separa expressamente o que foi exigido daquilo que foi decidido para tornar o comportamento determinístico. As alternativas são análises técnicas atuais; não há alegação de que todas tenham sido implementadas e comparadas experimentalmente.

| Tema | Exigência do PDF | Decisão implementada, motivo e benefício | Alternativa, compromisso e evidência |
| --- | --- | --- | --- |
| Visibilidade | Somente usuários autenticados acessam o sistema; o escopo da consulta não é definido. | Todo autenticado consulta todas as solicitações. Isso atende à ideia de portal interno sem introduzir perfis não solicitados. | Restringir ao autor ou criar equipes seria possível, mas exigiria papéis. Evidência: services de consulta/listagem e rotas protegidas. |
| Edição e exclusão | Solicitações abertas podem ser editadas e excluídas; autoria não é especificada. | Somente o autor altera conteúdo ou exclui, e apenas no status `aberto`. A regra protege o conteúdo criado por outro usuário e mantém o significado de “aberta”. | Permissão administrativa seria alternativa, mas não há papel administrativo. SQL condicional em `backend/internal/repository/request.go` aplica autoria/status atomicamente. |
| Alteração de status | Deve existir alteração entre Aberto, Em Atendimento e Concluído; perfis e sequência não são definidos. | Qualquer autenticado muda qualquer solicitação para qualquer dos três estados, inclusive reabrir. Repetir o estado é idempotente e não altera artificialmente `updated_at`. | Uma máquina linear ou aprovação administrativa seria mais restritiva, porém inventaria regras. A transação com `SELECT ... FOR UPDATE` está no repository. |
| Abrangência do dashboard | Exige quatro indicadores, sem dizer se são pessoais ou globais. | Os indicadores são globais e não herdam filtros. Uma única agregação mantém os quatro valores na mesma visão da consulta. | Dashboard por usuário seria válido em outro contexto, mas divergiria da visão global escolhida. Evidência: service/repository de dashboard e `GET /api/v1/dashboard`. |
| Categorias | O PDF sugere TI, RH, Compras, Financeiro e Infraestrutura. | As cinco sugestões foram adotadas como conjunto fechado e expostas por metadata para evitar rótulos divergentes no frontend. | Categorias cadastráveis dariam flexibilidade, com custo de CRUD e governança fora do prazo. Evidência: `domain/request.go`, constraint SQL e `/metadata`. |
| Limites de texto | Título e descrição são exigidos, sem limites. | Após trim, título aceita 3–150 e descrição 10–5000 pontos de código Unicode. A regra evita conteúdo vazio e payloads desnecessariamente grandes. | Limites configuráveis seriam possíveis, mas aumentariam configuração e inconsistência com constraints. Evidência: service, formulário e migration `00001`. |
| Campos automáticos | Data, solicitante e status Aberto devem ser automáticos. | O backend usa a identidade da sessão, define `aberto` e controla datas; campos adicionais no JSON são rejeitados. Defaults SQL permanecem como defesa adicional. | Confiar no cliente permitiria falsificação. Evidência: DTO/handler de criação, service e repository. |
| Datas e período | O PDF exige filtro por período, sem fuso ou semântica dos limites. | Banco usa `TIMESTAMPTZ`/UTC; dias de entrada são interpretados em `America/Sao_Paulo`, com início inclusivo e início do dia seguinte exclusivo. | Somar 24 horas ou tratar `YYYY-MM-DD` como UTC poderia deslocar dias civis. Evidência: `backend/internal/service/request.go` e filtros do repository. |
| Busca e filtros | Exige período, categoria, status e texto livre no título. | Filtros opcionais combinam por `AND`; `q` busca substring sem diferenciar caixa apenas no título, com `%`, `_` e barra escapados. | Busca na descrição ou full-text não foi solicitada. `ILIKE` é simples para o volume do desafio, mas pode escalar mal. Evidência: parser/service e SQL parametrizado de listagem. |
| Paginação e ordenação | Não são definidas pelo PDF. | Página 1 e 20 itens são padrões, máximo 100; ordem fixa por `created_at DESC, id DESC`. Isso limita respostas e torna o resultado determinístico. | Cursor pagination escalaria melhor em grande volume, mas aumentaria o contrato. Evidência: tipos/query da listagem e OpenAPI. |
| Código visual | O PDF pede código, sem formato ou armazenamento. | `SOL-` mais ID com no mínimo seis dígitos é calculado na resposta, evitando coluna duplicada. | Um identificador independente permitiria outro ciclo de numeração, mas exigiria constraint e geração próprias. Evidência: mapper HTTP; não existe coluna de código nas migrations. |
| Exclusão | O PDF pede exclusão, sem definir lógica ou física. | A exclusão é física e autorizada somente ao autor enquanto aberta. É simples para o desafio. | Soft delete facilitaria auditoria/recuperação, mas afetaria todas as consultas e índices. Evidência: `DELETE` parametrizado no repository. |
| Usuários | O PDF exige login, sem cadastro ou provisionamento. | Dois usuários demonstrativos são criados por seed separado, explícito, transacional e repetível; não há cadastro público. | Integração com diretório corporativo ou painel administrativo seria adequada em produção. Evidência: `backend/cmd/seed`, `backend/internal/seed/` e profile `demo`. |

Essas escolhas não são apresentadas como obrigações impostas pela empresa quando o enunciado não as determina.

## 8. Testes e evidências

Foram executados testes unitários/services, HTTP com `httptest`, integração PostgreSQL com migrations reais, detector de corrida, 53 testes Vitest/Testing Library e três Playwright E2E. O fluxo E2E inclui dois usuários e autorização real; os testes de componente usam API simulada e não são chamados de integração.

As capturas de `docs/evidencias/` foram produzidas em Chromium contra Compose, PostgreSQL, migrations, seed, API e Nginx reais. Não são mockups. O OpenAPI foi validado pelo Redocly CLI; o resultado foi válido com avisos de estilo documentados em `TESTES.md`.

Não houve teste de carga, pentest, análise SAST dedicada, matriz completa de browsers, teste de restauração de backup ou deploy externo.

## 9. Análise crítica e melhorias

### Limitações atuais

- Rate limit em memória não coordena múltiplas instâncias e perde contadores no reinício.
- Exclusão física e ausência de histórico de status impedem auditoria e recuperação funcional.
- Não há gestão de usuários, troca/recuperação de senha, papéis ou trilha administrativa.
- Política de transição é deliberadamente aberta; processos reais podem exigir responsáveis, SLA e estados adicionais.
- Edições concorrentes do mesmo autor usam última gravação vence; não há `ETag`, versão ou detecção de conflito de conteúdo.
- `ILIKE '%...%'` é adequado ao volume do desafio, mas pode exigir `pg_trgm` ou busca dedicada após medição.
- Sessão no banco acrescenta consulta a cada requisição. Em maior escala, pool, cache/revogação e limpeza de sessões expiradas precisariam de operação explícita.
- O Compose é local/servidor único. Não entrega TLS, backups, observabilidade central, balanceamento ou alta disponibilidade.
- O CI foi configurado e reproduzido localmente, mas ainda não executado no GitHub. Não existe CD.

### Mudanças para produção

- HTTPS obrigatório, `Secure=true`, HSTS no proxy e origem pública exata.
- Segredos fornecidos por secret manager/ambiente protegido, nunca `.env` versionado; credenciais demo removidas.
- Banco privado com backup, retenção, restauração testada, monitoração e migrations controladas.
- Auditoria de login, mudança de status e exclusão, com política de privacidade/retenção.
- Rate limit distribuído e política confiável para IP do cliente atrás de proxies autorizados.
- Gestão de identidade corporativa ou provisionamento administrativo, papéis e menor privilégio conforme o processo real.
- Métricas, traces, logs centralizados, alertas e objetivos operacionais definidos.
- Teste de carga, revisão de ameaça/pentest, varredura de dependências/imagens e matriz de browsers antes de compromissos de desempenho ou segurança.
- Estratégia de escala baseada em métricas reais, não em abstrações preventivas.

## 10. Uso de inteligência artificial

Durante o desenvolvimento do projeto, ChatGPT e Codex foram utilizados como apoio à tomada de decisões técnicas, à elaboração da documentação e ao design do frontend. Esse apoio contribuiu para analisar alternativas, organizar as justificativas e propor soluções para a interface.

Na tomada de decisões, a assistência ajudou a comparar opções como sessão opaca versus JWT, SQL explícito versus ORM, proxy same-origin versus configuração de CORS e ferramentas de teste por camada. As escolhas descritas neste memorial correspondem ao que aparece no código e nas configurações; a comparação com alternativas é análise técnica, não benchmark nem experimento histórico quando não há registro disso.

Na documentação, o apoio foi usado para estruturar requisitos, decisões, contrato, instruções de execução e esta análise crítica, além de revisar coerência entre os textos. No design do frontend, foram consideradas sugestões de organização das telas, componentes reutilizáveis, navegação, estados de feedback, responsividade e apresentação visual. Não se atribui à IA uma validação independente: os resultados mencionados continuam limitados às execuções e evidências já registradas no repositório.

O uso dessas ferramentas não elimina a necessidade de compreender e explicar a solução. Por isso, o projeto conserva SQL explícito, poucas dependências, alternativas documentadas e rastreabilidade entre requisito, implementação, teste e evidência.
