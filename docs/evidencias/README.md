# Evidências da aplicação

Este diretório contém capturas reais geradas no Chromium por `npm run evidence`, contra a aplicação completa em execução. O cenário usa PostgreSQL, migrations, seed, API Go e frontend Nginx reais; não são mockups nem telas alimentadas por arrays estáticos.

O roteiro automatizado está em `frontend/evidence/capture.spec.ts`. Ele abre a tela de login vazia antes de preencher as credenciais em memória, portanto senha, cookie e token de sessão não aparecem nas imagens. As credenciais usadas são exclusivamente as públicas de demonstração.

## Cenário validado

Em 03/10/2026 foi criado o projeto Compose isolado `bit-project-final-validation`, com PostgreSQL 17.6 vazio, migrations Goose 1/2, seed explícito, API Go e frontend Nginx. O seed criou dois usuários e cinco solicitações. O roteiro entrou como `colaborador1`, consultou os dados do banco, aplicou filtros reais, criou uma sexta solicitação e alterou seu status no backend.

O Playwright concluiu o roteiro de captura com 1 teste aprovado em 30,8 s. Depois, a suíte E2E independente com dois usuários passou com 3 testes em 50,8 s. As imagens foram abertas e inspecionadas; a tela de login foi capturada antes do preenchimento e nenhuma mostra senha, cookie ou token.

## Arquivos

| Arquivo | Evidência observável |
| --- | --- |
| `01-login.png` | Login real com campos vazios, labels e aviso de ausência de cadastro público. |
| `02-dashboard.png` | Quatro indicadores globais carregados da API após login. |
| `03-listagem-completa.png` | Cinco registros do seed e as seis colunas exigidas. |
| `04-filtros-combinados.png` | Categoria Infraestrutura, status Em Atendimento e texto “iluminação”, retornando um registro. |
| `05-criacao-sucesso.png` | Confirmação após `POST` real e detalhe da solicitação criada. |
| `06-detalhes-aberto.png` | Recarregamento por URL direta, dados completos e ações permitidas ao autor. |
| `07-status-em-atendimento.png` | Resposta confirmada pelo servidor, status Em Atendimento e edição/exclusão indisponíveis. |
| `08-status-concluido.png` | Segunda transição real, agora para Concluído. |
| `09-layout-movel.png` | Viewport 375×812 com navegação, criação e filtros acessíveis. |

As datas e IDs pertencem exclusivamente ao banco descartável da execução. Não devem ser interpretados como dados de produção.

## Como reproduzir

Com a aplicação demonstrativa já iniciada e populada conforme o README:

```sh
cd frontend
npx playwright install --with-deps chromium
E2E_BASE_URL='http://127.0.0.1:8080' npm run evidence
```

O comando escreve neste diretório e pode substituir capturas de mesmo nome. Execute-o somente contra base demonstrativa controlada, pois cria uma solicitação sintética e altera seu status até Concluído.
