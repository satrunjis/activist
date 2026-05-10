-- name: ListPermissionContextsByUser :many
SELECT
    p.division_id,
    r.permissions
FROM memberships m
JOIN positions p ON p.id = m.position_id
JOIN roles r ON r.id = p.role_id
JOIN divisions d ON d.id = p.division_id
WHERE m.user_id = $1
  AND p.is_archived = false
  AND d.is_archived = false;

-- name: ListActiveDivisionIDsByUser :many
SELECT DISTINCT
    p.division_id
FROM memberships m
JOIN positions p ON p.id = m.position_id
JOIN divisions d ON d.id = p.division_id
WHERE m.user_id = $1
  AND p.is_archived = false
  AND d.is_archived = false
ORDER BY p.division_id;
