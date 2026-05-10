-- +goose Up
-- +goose StatementBegin
CREATE TABLE divisions (
    id text PRIMARY KEY,
    parent_id text REFERENCES divisions(id) ON DELETE RESTRICT,
    short_name text NOT NULL,
    full_name text NOT NULL DEFAULT '',
    description text NOT NULL DEFAULT '',
    regulation_url text NOT NULL DEFAULT '',
    media_links jsonb NOT NULL DEFAULT '[]'::jsonb,
    is_archived boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (id <> parent_id)
);

CREATE UNIQUE INDEX divisions_single_root_uq ON divisions ((parent_id IS NULL))
WHERE parent_id IS NULL;

CREATE INDEX divisions_parent_id_idx ON divisions (parent_id);
CREATE INDEX divisions_is_archived_idx ON divisions (is_archived);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS divisions;
-- +goose StatementEnd
