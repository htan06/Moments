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

-- name: UpdatePasswordHash :execrows
-- Reset password (forgot-password flow).
UPDATE account.accounts
SET password_hash = $2, updated_at = now()
WHERE id = $1;

-- name: UpdateUsername :execrows
-- Runs inside the rename transaction; the profile copy is updated before commit.
UPDATE account.accounts
SET username = $2, updated_at = now()
WHERE id = $1;

-- name: GetAccountStatus :one
-- Facade for other modules: exists / ACTIVE / username.
SELECT username, status
FROM account.accounts
WHERE id = $1;
