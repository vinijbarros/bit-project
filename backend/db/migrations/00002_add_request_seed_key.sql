-- +goose Up
ALTER TABLE requests
    ADD COLUMN seed_key VARCHAR(64),
    ADD CONSTRAINT requests_seed_key_format_check
        CHECK (seed_key IS NULL OR seed_key ~ '^demo_[a-z0-9_]+$');

CREATE UNIQUE INDEX idx_requests_seed_key_unique
    ON requests (seed_key)
    WHERE seed_key IS NOT NULL;

-- +goose Down
DROP INDEX idx_requests_seed_key_unique;

ALTER TABLE requests
    DROP CONSTRAINT requests_seed_key_format_check,
    DROP COLUMN seed_key;

