-- name: CreateAccount :one
INSERT INTO users.accounts (email, username, password_hash, status)
VALUES ($1, $2, $3, $4)
RETURNING id, created_at, updated_at;

-- name: CreateProfile :exec
INSERT INTO users.profiles (user_id, name)
VALUES ($1, $2);

-- name: GetAccountByID :one
SELECT id, email, username, password_hash, status, created_at, updated_at
FROM users.accounts
WHERE id = $1;

-- name: GetAccountByEmail :one
SELECT id, email, username, password_hash, status, created_at, updated_at
FROM users.accounts
WHERE email = $1;

-- name: GetAccountByUsername :one
SELECT id, email, username, password_hash, status, created_at, updated_at
FROM users.accounts
WHERE username = $1;

-- name: EmailExists :one
SELECT EXISTS (SELECT 1 FROM users.accounts WHERE email = $1);
