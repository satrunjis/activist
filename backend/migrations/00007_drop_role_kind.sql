-- +goose Up
-- +goose StatementBegin
ALTER TABLE roles DROP COLUMN kind;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE roles ADD COLUMN kind text NOT NULL DEFAULT 'standard' CHECK (kind IN ('standard', 'leader', 'deputy'));
-- +goose StatementEnd
