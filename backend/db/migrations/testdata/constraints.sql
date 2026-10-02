\set ON_ERROR_STOP on

BEGIN;

DO $migration_checks$
DECLARE
    test_user_id BIGINT;
    test_request_status VARCHAR(20);
    test_request_created_at TIMESTAMPTZ;
    test_request_updated_at TIMESTAMPTZ;
BEGIN
    INSERT INTO users (username, display_name, password_hash)
    VALUES ('teste.user', 'Usuário de Teste', repeat('x', 60))
    RETURNING id INTO test_user_id;

    INSERT INTO sessions (token_hash, user_id, expires_at)
    VALUES (repeat('a', 64), test_user_id, CURRENT_TIMESTAMP + INTERVAL '8 hours');

    INSERT INTO requests (title, description, category, requester_id)
    VALUES (
        'Falha no notebook',
        'O notebook não inicializa após a atualização.',
        'ti',
        test_user_id
    )
    RETURNING status, created_at, updated_at
    INTO test_request_status, test_request_created_at, test_request_updated_at;

    IF test_request_status <> 'aberto' THEN
        RAISE EXCEPTION 'expected default request status aberto, got %', test_request_status;
    END IF;
    IF test_request_created_at IS NULL OR test_request_updated_at IS NULL THEN
        RAISE EXCEPTION 'expected automatic request timestamps';
    END IF;

    BEGIN
        INSERT INTO users (username, display_name, password_hash)
        VALUES ('UsuarioInvalido', 'Inválido', repeat('x', 60));
        RAISE EXCEPTION 'expected normalized username constraint violation';
    EXCEPTION WHEN check_violation THEN
        NULL;
    END;

    BEGIN
        INSERT INTO users (username, display_name, password_hash)
        VALUES ('teste.user', 'Duplicado', repeat('x', 60));
        RAISE EXCEPTION 'expected username unique violation';
    EXCEPTION WHEN unique_violation THEN
        NULL;
    END;

    BEGIN
        INSERT INTO sessions (token_hash, user_id, expires_at)
        VALUES ('token-original-invalido', test_user_id, CURRENT_TIMESTAMP + INTERVAL '8 hours');
        RAISE EXCEPTION 'expected token hash format constraint violation';
    EXCEPTION WHEN check_violation THEN
        NULL;
    END;

    BEGIN
        INSERT INTO sessions (token_hash, user_id, expires_at)
        VALUES (repeat('b', 64), test_user_id, CURRENT_TIMESTAMP - INTERVAL '1 minute');
        RAISE EXCEPTION 'expected session expiration constraint violation';
    EXCEPTION WHEN check_violation THEN
        NULL;
    END;

    BEGIN
        INSERT INTO requests (title, description, category, requester_id)
        VALUES (' x ', 'Descrição suficientemente longa.', 'ti', test_user_id);
        RAISE EXCEPTION 'expected title constraint violation';
    EXCEPTION WHEN check_violation THEN
        NULL;
    END;

    BEGIN
        INSERT INTO requests (title, description, category, requester_id)
        VALUES ('Título válido', 'curta', 'ti', test_user_id);
        RAISE EXCEPTION 'expected description constraint violation';
    EXCEPTION WHEN check_violation THEN
        NULL;
    END;

    BEGIN
        INSERT INTO requests (title, description, category, requester_id)
        VALUES ('Título válido', 'Descrição suficientemente longa.', 'juridico', test_user_id);
        RAISE EXCEPTION 'expected category constraint violation';
    EXCEPTION WHEN check_violation THEN
        NULL;
    END;

    BEGIN
        INSERT INTO requests (title, description, category, status, requester_id)
        VALUES ('Título válido', 'Descrição suficientemente longa.', 'rh', 'cancelado', test_user_id);
        RAISE EXCEPTION 'expected status constraint violation';
    EXCEPTION WHEN check_violation THEN
        NULL;
    END;

    BEGIN
        INSERT INTO requests (title, description, category, requester_id, seed_key)
        VALUES (
            'Título válido',
            'Descrição suficientemente longa.',
            'financeiro',
            test_user_id,
            'identificador_invalido'
        );
        RAISE EXCEPTION 'expected seed key format constraint violation';
    EXCEPTION WHEN check_violation THEN
        NULL;
    END;

    INSERT INTO requests (title, description, category, requester_id, seed_key)
    VALUES (
        'Exemplo identificado',
        'Solicitação válida usada para testar a chave do seed.',
        'financeiro',
        test_user_id,
        'demo_constraint_test'
    );

    BEGIN
        INSERT INTO requests (title, description, category, requester_id, seed_key)
        VALUES (
            'Outro título permitido',
            'O título pode variar, mas a chave do seed deve ser única.',
            'rh',
            test_user_id,
            'demo_constraint_test'
        );
        RAISE EXCEPTION 'expected seed key unique violation';
    EXCEPTION WHEN unique_violation THEN
        NULL;
    END;

    BEGIN
        INSERT INTO requests (
            title,
            description,
            category,
            requester_id,
            created_at,
            updated_at
        )
        VALUES (
            'Título válido',
            'Descrição suficientemente longa.',
            'compras',
            test_user_id,
            CURRENT_TIMESTAMP,
            CURRENT_TIMESTAMP - INTERVAL '1 second'
        );
        RAISE EXCEPTION 'expected updated_at constraint violation';
    EXCEPTION WHEN check_violation THEN
        NULL;
    END;

    BEGIN
        DELETE FROM users WHERE id = test_user_id;
        RAISE EXCEPTION 'expected user foreign key restriction';
    EXCEPTION WHEN foreign_key_violation THEN
        NULL;
    END;

    IF to_regclass('public.idx_sessions_user_id') IS NULL
        OR to_regclass('public.idx_sessions_expires_at') IS NULL
        OR to_regclass('public.idx_requests_requester_id') IS NULL
        OR to_regclass('public.idx_requests_status') IS NULL
        OR to_regclass('public.idx_requests_category') IS NULL
        OR to_regclass('public.idx_requests_created_at') IS NULL
        OR to_regclass('public.idx_requests_seed_key_unique') IS NULL THEN
        RAISE EXCEPTION 'one or more expected indexes are missing';
    END IF;
END;
$migration_checks$;

ROLLBACK;
