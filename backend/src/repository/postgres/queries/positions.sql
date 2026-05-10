-- name: CreatePosition :one
INSERT INTO positions (id, title, role_id, division_id, max_count, is_archived)
VALUES ($1, $2, $3, $4, $5, false)
RETURNING *;

-- name: GetPositionByID :one
SELECT * FROM positions WHERE id = $1;

-- name: GetPositionByIDForUpdate :one
SELECT * FROM positions WHERE id = $1 FOR UPDATE;

-- name: ListPositionsByDivision :many
SELECT * FROM positions WHERE division_id = $1 AND is_archived = false ORDER BY title;

-- name: UpdatePosition :one
UPDATE positions SET title = $2, role_id = $3, max_count = $4
WHERE id = $1
RETURNING *;

-- name: ArchivePosition :one
UPDATE positions SET is_archived = true WHERE id = $1
RETURNING *;

-- name: ArchivePositionsByDivisionSubtree :execrows
WITH RECURSIVE subtree AS (
    SELECT d.id
    FROM divisions d
    WHERE d.id = $1
  UNION ALL
    SELECT d.id
    FROM divisions d
    JOIN subtree s ON d.parent_id = s.id
)
UPDATE positions
SET is_archived = true
WHERE division_id IN (SELECT id FROM subtree)
  AND is_archived = false;

-- name: CountMembersByPosition :one
SELECT count(*) FROM memberships WHERE position_id = $1;
