-- name: SearchUsers :many
SELECT DISTINCT
    u.id,
    u.first_name,
    u.last_name,
    u.middle_name,
    u.login,
    u.group_number,
    u.institute,
    u.about,
    u.phone,
    u.social_links,
    u.birth_date,
    u.gradebook_number
FROM users u
LEFT JOIN memberships m ON m.user_id = u.id
LEFT JOIN positions p ON p.id = m.position_id
LEFT JOIN roles r ON r.id = p.role_id
LEFT JOIN divisions d ON d.id = p.division_id
WHERE (sqlc.narg(first_name)::text IS NULL OR u.first_name ILIKE '%' || sqlc.narg(first_name)::text || '%')
  AND (sqlc.narg(last_name)::text IS NULL OR u.last_name ILIKE '%' || sqlc.narg(last_name)::text || '%')
  AND (sqlc.narg(middle_name)::text IS NULL OR u.middle_name ILIKE '%' || sqlc.narg(middle_name)::text || '%')
  AND (sqlc.narg(login)::text IS NULL OR u.login ILIKE '%' || sqlc.narg(login)::text || '%')
  AND (sqlc.narg(group_number)::text IS NULL OR u.group_number ILIKE '%' || sqlc.narg(group_number)::text || '%')
  AND (sqlc.narg(institute)::text IS NULL OR u.institute ILIKE '%' || sqlc.narg(institute)::text || '%')
  AND (sqlc.narg(about)::text IS NULL OR u.about ILIKE '%' || sqlc.narg(about)::text || '%')
  AND (sqlc.narg(position_title)::text IS NULL OR p.title ILIKE '%' || sqlc.narg(position_title)::text || '%')
  AND (sqlc.narg(role_name)::text IS NULL OR r.name ILIKE '%' || sqlc.narg(role_name)::text || '%')
  AND (
    sqlc.arg(include_archived)::bool
    OR p.id IS NULL
    OR (p.is_archived = false AND d.is_archived = false)
  )
ORDER BY u.last_name, u.first_name, u.id
LIMIT sqlc.arg(page_limit)::int
OFFSET sqlc.arg(page_offset)::int;

-- name: CountSearchUsers :one
SELECT COUNT(DISTINCT u.id)
FROM users u
LEFT JOIN memberships m ON m.user_id = u.id
LEFT JOIN positions p ON p.id = m.position_id
LEFT JOIN roles r ON r.id = p.role_id
LEFT JOIN divisions d ON d.id = p.division_id
WHERE (sqlc.narg(first_name)::text IS NULL OR u.first_name ILIKE '%' || sqlc.narg(first_name)::text || '%')
  AND (sqlc.narg(last_name)::text IS NULL OR u.last_name ILIKE '%' || sqlc.narg(last_name)::text || '%')
  AND (sqlc.narg(middle_name)::text IS NULL OR u.middle_name ILIKE '%' || sqlc.narg(middle_name)::text || '%')
  AND (sqlc.narg(login)::text IS NULL OR u.login ILIKE '%' || sqlc.narg(login)::text || '%')
  AND (sqlc.narg(group_number)::text IS NULL OR u.group_number ILIKE '%' || sqlc.narg(group_number)::text || '%')
  AND (sqlc.narg(institute)::text IS NULL OR u.institute ILIKE '%' || sqlc.narg(institute)::text || '%')
  AND (sqlc.narg(about)::text IS NULL OR u.about ILIKE '%' || sqlc.narg(about)::text || '%')
  AND (sqlc.narg(position_title)::text IS NULL OR p.title ILIKE '%' || sqlc.narg(position_title)::text || '%')
  AND (sqlc.narg(role_name)::text IS NULL OR r.name ILIKE '%' || sqlc.narg(role_name)::text || '%')
  AND (
    sqlc.arg(include_archived)::bool
    OR p.id IS NULL
    OR (p.is_archived = false AND d.is_archived = false)
  );
