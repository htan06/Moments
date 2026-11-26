-- name: CreateProfile :exec
-- Idempotent: a profile that already exists is left unchanged.
INSERT INTO users.profiles (user_id, username, name)
VALUES ($1, $2, $3)
ON CONFLICT (user_id) DO NOTHING;

-- name: UpdateProfileUsername :execrows
-- Keeps the username copy in sync when the account module renames (ACCOUNT-R5).
UPDATE users.profiles
SET username = $2, updated_at = now()
WHERE user_id = $1;
