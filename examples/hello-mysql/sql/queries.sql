-- name: GetUser :one
SELECT id, name, email, created_at, updated_at
FROM users
WHERE id = ?
LIMIT 1;

-- name: CreateUser :execresult
INSERT INTO users (name, email)
VALUES (?, ?);

-- name: ListUsers :many
SELECT id, name, email, created_at, updated_at
FROM users
ORDER BY id;

-- name: UpdateUser :execresult
UPDATE users
SET name = ?, email = ?
WHERE id = ?;

-- name: DeleteUser :execresult
DELETE FROM users
WHERE id = ?;
