-- +goose Up
-- +goose StatementBegin
CREATE TABLE event_log (
    id text PRIMARY KEY,
    event_type text NOT NULL,
    actor_id text NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    subject_type text NOT NULL,
    subject_id text NOT NULL,
    payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    timestamp timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (event_type <> ''),
    CHECK (subject_type <> ''),
    CHECK (subject_id <> ''),
    CHECK (jsonb_typeof(payload) = 'object')
);

CREATE INDEX event_log_event_type_idx ON event_log (event_type);
CREATE INDEX event_log_subject_type_idx ON event_log (subject_type);
CREATE INDEX event_log_subject_id_idx ON event_log (subject_id);
CREATE INDEX event_log_timestamp_id_idx ON event_log (timestamp DESC, id DESC);

CREATE OR REPLACE FUNCTION event_log_prevent_mutation() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'event_log is append-only';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER event_log_no_update
BEFORE UPDATE ON event_log
FOR EACH ROW EXECUTE FUNCTION event_log_prevent_mutation();

CREATE TRIGGER event_log_no_delete
BEFORE DELETE ON event_log
FOR EACH ROW EXECUTE FUNCTION event_log_prevent_mutation();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS event_log_no_delete ON event_log;
DROP TRIGGER IF EXISTS event_log_no_update ON event_log;
DROP FUNCTION IF EXISTS event_log_prevent_mutation();
DROP TABLE IF EXISTS event_log;
-- +goose StatementEnd
