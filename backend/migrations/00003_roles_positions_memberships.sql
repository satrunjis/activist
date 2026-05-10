-- +goose Up
-- +goose StatementBegin
CREATE TABLE roles (
    id text PRIMARY KEY,
    name text NOT NULL,
    permissions bytea NOT NULL DEFAULT ''::bytea,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX roles_name_uq ON roles (name);

CREATE TABLE positions (
    id text PRIMARY KEY,
    title text NOT NULL,
    role_id text NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
    division_id text NOT NULL REFERENCES divisions(id) ON DELETE RESTRICT,
    max_count integer,
    is_archived boolean NOT NULL DEFAULT false,
    CHECK (max_count IS NULL OR max_count > 0)
);

CREATE INDEX positions_division_id_idx ON positions (division_id);
CREATE INDEX positions_role_id_idx ON positions (role_id);
CREATE INDEX positions_is_archived_idx ON positions (is_archived);

CREATE TABLE memberships (
    user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    position_id text NOT NULL REFERENCES positions(id) ON DELETE RESTRICT,
    PRIMARY KEY (user_id, position_id)
);

CREATE INDEX memberships_user_id_idx ON memberships (user_id);
CREATE INDEX memberships_position_id_idx ON memberships (position_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS memberships;
DROP TABLE IF EXISTS positions;
DROP TABLE IF EXISTS roles;
-- +goose StatementEnd
