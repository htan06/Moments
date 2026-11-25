-- name: CreateProfile :exec
-- Idempotent: a profile that already exists is left unchanged.
INSERT INTO users.profiles (user_id, username, name)
VALUES ($1, $2, $3)
ON CONFLICT (user_id) DO NOTHING;
