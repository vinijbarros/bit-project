# Migrations

As migrations SQL deste diretório são aplicadas em ordem pelo Goose através de `cmd/migrate`.

- `00001_create_core_tables.sql`: cria `users`, `sessions`, `requests`, constraints, relacionamentos e índices iniciais.
- `00002_add_request_seed_key.sql`: adiciona identificador interno, anulável e único apenas para registros sintéticos criados pelo seed.

Use `up` e `status` no fluxo normal. `down` reverte apenas a última migration e deve ser usado somente em banco descartável ou quando houver um procedimento de rollback aprovado. O comando bloqueia `down` quando `APP_ENV=production`.

`testdata/constraints.sql` é uma verificação de integração: insere casos válidos e inválidos dentro de uma transação e sempre executa `ROLLBACK`. Não é seed e não deixa dados de demonstração no banco.
