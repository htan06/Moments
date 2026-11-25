-- name: CreateAccount :one
INSERT INTO account.accounts (email, username, password_hash, status)
VALUES ($1, $2, $3, $4)
RETURNING id, created_at, updated_at;

-- name: DeleteAccount :exec
-- Undoes CreateAccount when the profile could not be created (ACCOUNT-R7).
DELETE FROM account.accounts WHERE id = $1;

-- name: GetAccountByID :one
SELECT id, email, username, password_hash, status, created_at, updated_at
FROM account.accounts
WHERE id = $1;

-- name: GetAccountByEmail :one
SELECT id, email, username, password_hash, status, created_at, updated_at
FROM account.accounts
WHERE email = $1;

-- name: GetAccountByUsername :one
SELECT id, email, username, password_hash, status, created_at, updated_at
FROM account.accounts
WHERE username = $1;

-- name: EmailExists :one
SELECT EXISTS (SELECT 1 FROM account.accounts WHERE email = $1);

-- name: UsernameExists :one
SELECT EXISTS (SELECT 1 FROM account.accounts WHERE username = $1);
