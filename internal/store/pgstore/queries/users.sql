-- name: CreateUser :one
INSERT INTO users ("user_name", "email", "password_hash", "bio")
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: UpdateUser :exec
UPDATE users
SET user_name = $2, password_hash = $3, bio = $4, updated_at = NOW()
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: GetUserById :one
SELECT
    id,
    user_name,
    password_hash,
    email,
    bio,
    created_at,
    updated_at
FROM users
WHERE id = $1;


-- name: GetUserByEmail :one
SELECT
    id,
    user_name,
    password_hash,
    email,
    bio,
    created_at,
    updated_at
FROM users
WHERE email = $1;
