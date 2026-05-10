-- name: CreateDivision :one
INSERT INTO divisions (
    id,
    parent_id,
    short_name,
    full_name,
    description,
    regulation_url,
    media_links,
    is_archived
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING
    id,
    parent_id,
    short_name,
    full_name,
    description,
    regulation_url,
    media_links,
    is_archived,
    created_at,
    updated_at;

-- name: GetDivisionByID :one
SELECT
    id,
    parent_id,
    short_name,
    full_name,
    description,
    regulation_url,
    media_links,
    is_archived,
    created_at,
    updated_at
FROM divisions
WHERE id = $1
LIMIT 1;

-- name: GetRootDivision :one
SELECT
    id,
    parent_id,
    short_name,
    full_name,
    description,
    regulation_url,
    media_links,
    is_archived,
    created_at,
    updated_at
FROM divisions
WHERE parent_id IS NULL
LIMIT 1;

-- name: GetDivisionArchivedFlag :one
SELECT is_archived
FROM divisions
WHERE id = $1
LIMIT 1;

-- name: UpdateDivision :one
UPDATE divisions
SET
    parent_id = $2,
    short_name = $3,
    full_name = $4,
    description = $5,
    regulation_url = $6,
    media_links = $7,
    is_archived = $8,
    updated_at = now()
WHERE id = $1
RETURNING
    id,
    parent_id,
    short_name,
    full_name,
    description,
    regulation_url,
    media_links,
    is_archived,
    created_at,
    updated_at;

-- name: ListDivisionTree :many
WITH RECURSIVE division_tree AS (
    SELECT
        id,
        parent_id,
        short_name,
        full_name,
        description,
        regulation_url,
        media_links,
        is_archived,
        created_at,
        updated_at,
        1::int AS depth
    FROM divisions
    WHERE parent_id IS NULL
      AND (@include_archived::boolean = true OR is_archived = false)
  UNION ALL
    SELECT
        d.id,
        d.parent_id,
        d.short_name,
        d.full_name,
        d.description,
        d.regulation_url,
        d.media_links,
        d.is_archived,
        d.created_at,
        d.updated_at,
        dt.depth + 1
    FROM divisions d
    JOIN division_tree dt ON dt.id = d.parent_id
    WHERE (@include_archived::boolean = true OR d.is_archived = false)
)
SELECT
    id,
    parent_id,
    short_name,
    full_name,
    description,
    regulation_url,
    media_links,
    is_archived,
    created_at,
    updated_at,
    depth,
    (SELECT COUNT(*)::int FROM positions WHERE positions.division_id = division_tree.id AND positions.is_archived = false) AS positions_count,
    (
        SELECT COUNT(DISTINCT m.user_id)
        FROM memberships m
        JOIN positions p ON p.id = m.position_id
        WHERE p.division_id = division_tree.id
          AND p.is_archived = false
    ) AS members_count
FROM division_tree
ORDER BY depth, parent_id, short_name, id;

-- name: GetDivisionAncestors :many
WITH RECURSIVE ancestors AS (
    SELECT
        parent.id,
        parent.parent_id,
        1::int AS depth
    FROM divisions d
    JOIN divisions parent ON parent.id = d.parent_id
    WHERE d.id = $1
  UNION ALL
    SELECT
        parent.id,
        parent.parent_id,
        a.depth + 1
    FROM ancestors a
    JOIN divisions parent ON parent.id = a.parent_id
)
SELECT id, depth
FROM ancestors
ORDER BY depth;

-- name: ListDivisionSubtree :many
WITH RECURSIVE subtree AS (
    SELECT
        d.id,
        d.parent_id,
        0::int AS depth
    FROM divisions d
    WHERE d.id = $1
  UNION ALL
    SELECT
        d.id,
        d.parent_id,
        s.depth + 1
    FROM divisions d
    JOIN subtree s ON d.parent_id = s.id
)
SELECT
    d.id,
    d.parent_id,
    d.short_name,
    d.full_name,
    d.description,
    d.regulation_url,
    d.media_links,
    d.is_archived,
    d.created_at,
    d.updated_at,
    s.depth
FROM divisions d
JOIN subtree s ON s.id = d.id
ORDER BY s.depth DESC, d.id;

-- name: ArchiveChildDivisionsByRoot :execrows
WITH RECURSIVE subtree AS (
    SELECT d.id
    FROM divisions d
    WHERE d.id = $1
  UNION ALL
    SELECT d.id
    FROM divisions d
    JOIN subtree s ON d.parent_id = s.id
)
UPDATE divisions
SET is_archived = true,
    updated_at = now()
WHERE divisions.id <> $1
  AND EXISTS (
      SELECT 1
      FROM subtree
      WHERE subtree.id = divisions.id
  )
  AND divisions.is_archived = false;

-- name: ArchiveDivisionByID :execrows
UPDATE divisions
SET is_archived = true,
    updated_at = now()
WHERE id = $1
  AND is_archived = false;

-- name: ListDivisionChildrenByParent :many
SELECT
    d.id,
    d.parent_id,
    d.short_name,
    d.full_name,
    d.description,
    d.regulation_url,
    d.media_links,
    d.is_archived,
    d.created_at,
    d.updated_at,
    EXISTS (
        SELECT 1 FROM divisions c WHERE c.parent_id = d.id AND c.is_archived = false
    )::bool AS has_children,
    (
        SELECT COUNT(*)::int FROM divisions c WHERE c.parent_id = d.id AND c.is_archived = false
    ) AS children_count
FROM divisions d
WHERE d.parent_id = $1
  AND d.is_archived = false
ORDER BY d.short_name, d.id;

-- name: ListRootDivisions :many
SELECT
    d.id,
    d.parent_id,
    d.short_name,
    d.full_name,
    d.description,
    d.regulation_url,
    d.media_links,
    d.is_archived,
    d.created_at,
    d.updated_at,
    EXISTS (
        SELECT 1 FROM divisions c WHERE c.parent_id = d.id AND c.is_archived = false
    )::bool AS has_children,
    (
        SELECT COUNT(*)::int FROM divisions c WHERE c.parent_id = d.id AND c.is_archived = false
    ) AS children_count
FROM divisions d
WHERE d.parent_id IS NULL
  AND d.is_archived = false
ORDER BY d.short_name, d.id;
