-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_users_first_name_trgm ON users USING gin (first_name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_users_last_name_trgm ON users USING gin (last_name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_users_middle_name_trgm ON users USING gin (middle_name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_users_login_trgm ON users USING gin (login gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_users_group_number_trgm ON users USING gin (group_number gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_users_institute_trgm ON users USING gin (institute gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_users_about_trgm ON users USING gin (about gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_positions_title_trgm ON positions USING gin (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_roles_name_trgm ON roles USING gin (name gin_trgm_ops);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_roles_name_trgm;
DROP INDEX IF EXISTS idx_positions_title_trgm;
DROP INDEX IF EXISTS idx_users_about_trgm;
DROP INDEX IF EXISTS idx_users_institute_trgm;
DROP INDEX IF EXISTS idx_users_group_number_trgm;
DROP INDEX IF EXISTS idx_users_login_trgm;
DROP INDEX IF EXISTS idx_users_middle_name_trgm;
DROP INDEX IF EXISTS idx_users_last_name_trgm;
DROP INDEX IF EXISTS idx_users_first_name_trgm;
-- +goose StatementEnd
