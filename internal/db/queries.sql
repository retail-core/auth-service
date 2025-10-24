-- name: CreateUser :exec
INSERT INTO users (email, username, password, role, tenant_id)
VALUES ($1, $2, $3, $4, $5);

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;