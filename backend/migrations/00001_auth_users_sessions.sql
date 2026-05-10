-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id text PRIMARY KEY,
    login text NOT NULL,
    password_hash text NOT NULL,
    first_name text NOT NULL,
    last_name text,
    middle_name text,
    gradebook_number text,
    group_number text,
    institute text,
    birth_date date,
    phone text,
    social_links jsonb NOT NULL DEFAULT '[]'::jsonb,
    about text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX users_login_uq ON users (login);
CREATE UNIQUE INDEX users_gradebook_number_uq ON users (gradebook_number);

CREATE TABLE sessions (
    id text PRIMARY KEY,
    user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash text NOT NULL,
    csrf_secret text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    absolute_expires_at timestamptz NOT NULL,
    idle_expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    revoked_reason text,
    created_ip text NOT NULL,
    created_user_agent text NOT NULL
);

CREATE UNIQUE INDEX sessions_token_hash_uq ON sessions (token_hash);
CREATE INDEX sessions_user_revoked_idx ON sessions (user_id, revoked_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
