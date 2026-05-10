-- name: CreateUser :one
INSERT INTO users (
    id,
    login,
    password_hash,
    first_name,
    last_name,
    middle_name,
    gradebook_number,
    group_number,
    institute,
    birth_date,
    phone,
    social_links,
    about
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
RETURNING
    id,
    login,
    password_hash,
    first_name,
    last_name,
    middle_name,
    gradebook_number,
    group_number,
    institute,
    birth_date,
    phone,
    social_links,
    about,
    created_at,
    updated_at;

-- name: GetUserByLogin :one
SELECT
    id,
    login,
    password_hash,
    first_name,
    last_name,
    middle_name,
    gradebook_number,
    group_number,
    institute,
    birth_date,
    phone,
    social_links,
    about,
    created_at,
    updated_at
FROM users
WHERE login = $1
LIMIT 1;

-- name: GetUserByID :one
SELECT
    id,
    login,
    password_hash,
    first_name,
    last_name,
    middle_name,
    gradebook_number,
    group_number,
    institute,
    birth_date,
    phone,
    social_links,
    about,
    created_at,
    updated_at
FROM users
WHERE id = $1
LIMIT 1;

-- name: UpdateUserProfile :one
UPDATE users
SET
    first_name = $2,
    last_name = $3,
    middle_name = $4,
    gradebook_number = $5,
    group_number = $6,
    institute = $7,
    birth_date = $8,
    phone = $9,
    social_links = $10,
    about = $11,
    updated_at = now()
WHERE id = $1
RETURNING
    id,
    login,
    password_hash,
    first_name,
    last_name,
    middle_name,
    gradebook_number,
    group_number,
    institute,
    birth_date,
    phone,
    social_links,
    about,
    created_at,
    updated_at;
