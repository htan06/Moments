-- name: GetPostsByAuthorID :many
SELECT  
    sqlc.embed(p),
    sqlc.embed(m)
FROM content.posts p
JOIN content.authors a
    ON p.author_id = a.user_id
JOIN content.medias m
    ON p.id = m.post_id
WHERE p.created_at <= sqlc.arg(cursor_timestamp);

-- name: CreatePost :one
INSERT INTO content.posts (
    author_id,
    content,
    visibility,
    thumbnail_key,
    media_count,
    aspect_ratio,
    status
) VALUES (
    sqlc.arg(author_id),
    sqlc.arg(content)::jsonb,
    sqlc.arg(visibility)::content.visibility,
    sqlc.arg(thumbnail_key),
    sqlc.arg(media_count),
    sqlc.arg(aspect_ratio)::content.aspect_ratio,
    sqlc.arg(status)::content.post_status
)
RETURNING id;

-- name: CreateMedias :batchexec
INSERT INTO content.medias (
    post_id,
    type,
    media_key,
    display_order,
    width,
    height
) VALUES (
    sqlc.arg(post_id),
    sqlc.arg(type),
    sqlc.arg(media_key),
    sqlc.arg(display_order),
    sqlc.arg(width),
    sqlc.arg(height)
);

-- name: CreateMentions :batchexec
INSERT INTO content.post_mentions (
    post_id, 
    user_id
) 
VALUES (
    sqlc.arg(post_id),
    sqlc.arg(user_id)
);

--name: CreateLikePost