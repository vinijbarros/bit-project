# Execução e deploy

Este documento descreve execução reproduzível local e preparação para servidor. Nenhum deploy externo foi realizado.

## Docker Compose local

Pré-requisitos: Docker Engine/Desktop com Docker Compose v2 e portas locais disponíveis. Os exemplos publicados em `.env.example` são exclusivamente de demonstração.

```sh
cp .env.example .env
docker compose up -d --build --wait
DEMO_SEED_ENABLED=true docker compose --profile demo run --rm seed
```

Abra `http://127.0.0.1:8080`. O navegador usa somente essa origem: o Nginx serve a SPA e encaminha `/api` para a API na rede interna. `/healthz` e `/readyz` também são encaminhados para os probes da API.

O fluxo padrão é deliberadamente separado:

1. `postgres` inicia e precisa ficar saudável;
2. `migrate` executa `/app/migrate up` e termina com código zero;
3. `api` inicia somente após a migration concluída e precisa ficar saudável;
4. `frontend` inicia após a API saudável;
5. `seed` pertence ao profile `demo`, executa `/app/seed` e nunca participa do `up` padrão.

Os alvos equivalentes são `make compose-up`, `make compose-seed` e `make compose-down`.

### Estado e desligamento

O banco usa o volume nomeado `postgres_data`. Para encerrar containers e rede preservando os dados:

```sh
docker compose down
```

`docker compose down --volumes` é destrutivo e não faz parte do fluxo normal. Use-o somente quando quiser apagar explicitamente uma base local de demonstração, após conferir o nome do projeto Compose e eventual backup. Nunca use essa opção em dados que devam ser preservados.

Reiniciar sem recriar o banco:

```sh
docker compose restart postgres api frontend
docker compose ps
```

O serviço `migrate` pode permanecer com estado `Exited (0)`: ele é uma tarefa concluída, não um processo permanente.

### Configuração e precedência

O Compose lê os padrões de `compose.yaml`, depois o arquivo `.env` do projeto; variáveis exportadas no shell têm precedência sobre o `.env`. O bloco `environment` do Compose tem precedência sobre qualquer `ENV` da imagem.

Variáveis específicas da composição:

| Variável | Uso | Padrão local |
| --- | --- | --- |
| `FRONTEND_BIND_ADDRESS` | Endereço publicado pelo Nginx | `127.0.0.1` |
| `FRONTEND_PORT` | Porta pública única da aplicação | `8080` |
| `POSTGRES_PORT` | Porta PostgreSQL ligada somente ao loopback | `5432` |
| `COMPOSE_DATABASE_URL` | URL interna da API/migrate/seed; host deve ser `postgres` | URL demo do `.env.example` |
| `COMPOSE_TRUSTED_ORIGINS` | Origens públicas exatas aceitas nas mutações | `http://127.0.0.1:8080` |
| `BACKEND_IMAGE` / `FRONTEND_IMAGE` | Nome/tag local das imagens | tags `:local` |

`DATABASE_URL` continua sendo usada na execução sem Docker e aponta para `127.0.0.1`. Dentro do Compose, `COMPOSE_DATABASE_URL` evita o erro de usar `localhost`, que apontaria para o próprio container. Se usuário, senha ou banco do PostgreSQL forem alterados, mantenha essa URL interna coerente e aplique URL-encoding aos componentes quando necessário.

No HTTP local use `APP_ENV=development` e `SESSION_COOKIE_SECURE=false`. Para HTTPS use `APP_ENV=production`, `SESSION_COOKIE_SECURE=true` e `COMPOSE_TRUSTED_ORIGINS=https://portal.exemplo.com` sem curingas, caminhos ou barra final.

### Seed demonstrativo

O comando abaixo habilita o seed somente naquela execução:

```sh
DEMO_SEED_ENABLED=true docker compose --profile demo run --rm seed
```

O Dockerfile não define entrypoint que substitua argumentos. Assim, o serviço executa diretamente o binário real `/app/seed`. O comando falha em `APP_ENV=production`, mesmo que alguém tente habilitá-lo.

Reset de senhas demonstrativas continua sendo uma ação separada e consciente:

```sh
DEMO_SEED_ENABLED=true docker compose --profile demo run --rm seed /app/seed --reset-passwords
```

Não mantenha senhas demonstrativas em um ambiente publicado.

## Execução sem Docker

Instale Go 1.22.2, Node.js 24.14.0/npm 11.9.0 e PostgreSQL 17. Copie os exemplos e ajuste `DATABASE_URL` para o banco local:

```sh
cp .env.example .env
cp frontend/.env.example frontend/.env
set -a
. ./.env
set +a
```

Crie o banco/usuário PostgreSQL de acordo com a URL e execute, a partir da raiz:

```sh
cd backend
go run ./cmd/migrate up
DEMO_SEED_ENABLED=true go run ./cmd/seed
go run ./cmd/api
```

Em outro terminal:

```sh
cd frontend
npm ci
npm run dev
```

Acesse `http://127.0.0.1:5173`. O proxy do Vite encaminha `/api` para `http://127.0.0.1:8080`; por isso `TRUSTED_ORIGINS` deve conter exatamente `http://127.0.0.1:5173` nesse modo.

## Orientação para servidor com HTTPS

- Termine TLS em um proxy reverso ou balanceador e encaminhe a origem pública para o container `frontend`; não publique a API separadamente ao navegador.
- Defina `FRONTEND_BIND_ADDRESS` conforme a topologia. Se houver proxy no mesmo host, manter `127.0.0.1` reduz exposição.
- Use `APP_ENV=production`, `SESSION_COOKIE_SECURE=true` e a origem HTTPS exata em `COMPOSE_TRUSTED_ORIGINS`.
- Forneça credenciais fortes por mecanismo seguro do ambiente. Não versione `.env`, senhas, tokens ou chaves.
- Não publique a porta do PostgreSQL na internet. Use rede privada, backup periódico, teste de restauração e monitoração de espaço/saúde.
- Execute migrations como tarefa única antes de liberar a API. Não execute `down` automaticamente em produção.
- Preserve o volume/banco entre recriações das imagens e faça backup antes de mudanças estruturais.

Limitações atuais: não houve deploy externo; TLS, backup, observabilidade centralizada, escalabilidade horizontal do rate limit e rotação de segredos dependem da infraestrutura escolhida. O Compose fornecido é adequado para avaliação/local e uma base para servidor único, não uma afirmação de alta disponibilidade.
