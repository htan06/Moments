-- name: CreateProfile :exec
INSERT INTO profile.users (user_id, name, username, followers_count, following_count, posts_count)
			VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetAvatarKeyByUserID :one
SELECT avatar_key FROM profile.users WHERE user_id = $1;

-- name: GetSelfProfileByUsername :one
SELECT
    user_id,
    username,
    name,
    avatar_key,
    avatar_thumbnail_key,
    bio,

    followers_count,
    following_count,
    posts_count,
    NULL as relationship

FROM profile.users
WHERE username = $1;

-- name: GetOtherProfileByUsername :one
SELECT
    u.user_id,
    u.username,
    u.name,
    u.avatar_key,
    u.avatar_thumbnail_key,
    u.bio,
    u.followers_count,
    u.following_count,
    u.posts_count,

    CASE
        WHEN f.following_id IS NOT NULL THEN 'FOLLOWING'
        ELSE 'NONE'
    END AS relationship_type

FROM profile.users AS u

LEFT JOIN social.follows AS f
    ON f.follower_id = sqlc.arg(current_user_id)
    AND f.following_id = u.user_id

WHERE u.username = sqlc.arg(target_username);

-- name: UpdateAvatarKeyAndAvatarThumbnailKey :exec
UPDATE profile.users SET avatar_key = $1, avatar_thumbnail_key = $2 WHERE user_id = $3;

-- name: FindProfilesByUsername :many
SELECT user_id, username, name, avatar_thumbnail_key FROM profile.users WHERE username LIKE $1 ||'%' LIMIT 10;