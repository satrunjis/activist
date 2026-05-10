-- name: CreateMembership :exec
INSERT INTO memberships (user_id, position_id) VALUES ($1, $2);

-- name: DeleteMembership :exec
DELETE FROM memberships WHERE user_id = $1 AND position_id = $2;

-- name: DeleteMembershipsByPosition :execrows
DELETE FROM memberships WHERE position_id = $1;

-- name: DeleteMembershipsByDivisionSubtree :execrows
WITH RECURSIVE subtree AS (
    SELECT d.id
    FROM divisions d
    WHERE d.id = $1
  UNION ALL
    SELECT d.id
    FROM divisions d
    JOIN subtree s ON d.parent_id = s.id
)
DELETE FROM memberships m
USING positions p
WHERE m.position_id = p.id
  AND p.division_id IN (SELECT id FROM subtree);

-- name: MembershipExists :one
SELECT exists(
    SELECT 1 FROM memberships WHERE user_id = $1 AND position_id = $2
);

-- name: ListMembershipsByUser :many
SELECT
    m.user_id,
    m.position_id,
    p.title       AS position_title,
    p.role_id,
    r.name        AS role_name,
    p.division_id,
    p.is_archived AS position_archived,
    d.short_name  AS division_short_name,
    d.is_archived AS division_archived
FROM memberships m
JOIN positions p ON p.id = m.position_id
JOIN roles r ON r.id = p.role_id
JOIN divisions d ON d.id = p.division_id
WHERE m.user_id = $1
  AND p.is_archived = false
  AND d.is_archived = false;

-- name: ListMembersByPosition :many
SELECT
    m.user_id,
    m.position_id,
    p.title       AS position_title,
    p.role_id,
    r.name        AS role_name,
    p.division_id,
    p.is_archived AS position_archived,
    d.short_name  AS division_short_name,
    d.is_archived AS division_archived
FROM memberships m
JOIN users u ON u.id = m.user_id
JOIN positions p ON p.id = m.position_id
JOIN roles r ON r.id = p.role_id
JOIN divisions d ON d.id = p.division_id
WHERE m.position_id = $1
  AND p.is_archived = false
  AND d.is_archived = false
ORDER BY u.first_name ASC, u.id ASC, m.user_id ASC, m.position_id ASC;
