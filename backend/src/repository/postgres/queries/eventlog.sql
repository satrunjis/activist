-- name: InsertEventLog :one
INSERT INTO event_log (
    id,
    event_type,
    actor_id,
    subject_type,
    subject_id,
    payload,
    timestamp
) VALUES (
    sqlc.arg(id)::text,
    sqlc.arg(event_type)::text,
    sqlc.arg(actor_id)::text,
    sqlc.arg(subject_type)::text,
    sqlc.arg(subject_id)::text,
    sqlc.arg(payload)::jsonb,
    sqlc.arg(timestamp)::timestamptz
)
RETURNING *;

-- name: ListEventLog :many
SELECT *
FROM event_log
WHERE (sqlc.narg(event_type)::text IS NULL OR event_type = sqlc.narg(event_type)::text)
  AND (sqlc.narg(subject_type)::text IS NULL OR subject_type = sqlc.narg(subject_type)::text)
  AND (sqlc.narg(subject_id)::text IS NULL OR subject_id = sqlc.narg(subject_id)::text)
ORDER BY timestamp DESC, id DESC
LIMIT sqlc.arg(page_limit)::int
OFFSET sqlc.arg(page_offset)::int;

-- name: CountEventLog :one
SELECT count(*)
FROM event_log
WHERE (sqlc.narg(event_type)::text IS NULL OR event_type = sqlc.narg(event_type)::text)
  AND (sqlc.narg(subject_type)::text IS NULL OR subject_type = sqlc.narg(subject_type)::text)
  AND (sqlc.narg(subject_id)::text IS NULL OR subject_id = sqlc.narg(subject_id)::text);
