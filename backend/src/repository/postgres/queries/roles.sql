-- name: CreateRole :one
INSERT INTO roles (id, name, permissions, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, name, permissions, created_at, updated_at;

-- name: GetRoleByID :one
SELECT id, name, permissions, created_at, updated_at FROM roles WHERE id = $1;

-- name: ListRoles :many
SELECT id, name, permissions, created_at, updated_at
FROM roles
ORDER BY name
LIMIT sqlc.arg(page_limit)::int
OFFSET sqlc.arg(page_offset)::int;

-- name: CountRoles :one
SELECT COUNT(*)::bigint FROM roles;

-- name: UpdateRole :one
UPDATE roles SET name = $2, permissions = $3, updated_at = $4
WHERE id = $1
RETURNING id, name, permissions, created_at, updated_at;

-- name: DeleteRole :exec
DELETE FROM roles WHERE id = $1;
