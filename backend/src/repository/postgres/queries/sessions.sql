-- name: CreateSession :one
INSERT INTO sessions (
    id,
    user_id,
    token_hash,
    csrf_secret,
    absolute_expires_at,
    idle_expires_at,
    created_ip,
    created_user_agent
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING
    id,
    user_id,
    token_hash,
    csrf_secret,
    created_at,
    last_seen_at,
    absolute_expires_at,
    idle_expires_at,
    revoked_at,
    revoked_reason,
    created_ip,
    created_user_agent;

-- name: GetSessionByTokenHash :one
SELECT
    id,
    user_id,
    token_hash,
    csrf_secret,
    created_at,
    last_seen_at,
    absolute_expires_at,
    idle_expires_at,
    revoked_at,
    revoked_reason,
    created_ip,
    created_user_agent
FROM sessions
WHERE token_hash = $1
LIMIT 1;

-- name: TouchSession :one
UPDATE sessions
SET
    last_seen_at = $2,
    idle_expires_at = $3
WHERE id = $1
  AND revoked_at IS NULL
RETURNING
    id,
    user_id,
    token_hash,
    csrf_secret,
    created_at,
    last_seen_at,
    absolute_expires_at,
    idle_expires_at,
    revoked_at,
    revoked_reason,
    created_ip,
    created_user_agent;

-- name: RevokeSession :execrows
UPDATE sessions
SET
    revoked_at = $2,
    revoked_reason = $3
WHERE id = $1
  AND revoked_at IS NULL;

-- name: RevokeAllUserSessions :execrows
UPDATE sessions
SET
    revoked_at = $2,
    revoked_reason = $3
WHERE user_id = $1
  AND revoked_at IS NULL;

-- name: DeleteExpiredSessions :execrows
DELETE FROM sessions
WHERE absolute_expires_at <= $1
   OR idle_expires_at <= $1;
