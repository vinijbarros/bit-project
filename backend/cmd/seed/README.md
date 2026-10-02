# Seed de demonstração

O comando `go run ./cmd/seed` é separado da API e só executa quando
`DEMO_SEED_ENABLED=true`. Ele cria dois usuários e cinco solicitações sintéticas
em uma única transação. As senhas são armazenadas somente como hashes bcrypt.

A repetição reconhece solicitações por `requests.seed_key`, não pelo título, e
não duplica registros. Usuários existentes precisam aceitar a senha configurada;
uma senha diferente causa erro sem alterações. O reset é deliberado:

```sh
go run ./cmd/seed --reset-passwords
```

O comando é bloqueado em `APP_ENV=production`, não remove dados e não altera
solicitações que já pertencem ao seed. Consulte o README da raiz para as
variáveis, credenciais públicas locais e autoria dos exemplos.
