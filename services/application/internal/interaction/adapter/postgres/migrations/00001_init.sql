-- +goose Up
CREATE SCHEMA IF NOT EXISTS interaction;

-- ---------------------------------------------------------------------------
-- Like
-- ---------------------------------------------------------------------------
CREATE TABLE interaction.likes (
    user_id    bigint      NOT NULL,
    post_id    bigint      NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),

    PRIMARY KEY (user_id, post_id)                         -- isLiked, idempotent like / unlike
);

-- ---------------------------------------------------------------------------
-- Comment (phẳng, xóa mềm)
-- ---------------------------------------------------------------------------
CREATE TABLE interaction.comments (
    id         bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    post_id    bigint      NOT NULL,
    author_id  bigint      NOT NULL,
    content    text        NOT NULL,
    status     text        NOT NULL DEFAULT 'ACTIVE',
    created_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,

    CONSTRAINT comments_content_len_chk CHECK (char_length(content) BETWEEN 1 AND 1000),
    CONSTRAINT comments_status_chk      CHECK (status IN ('ACTIVE', 'DELETED')),
    CONSTRAINT comments_deleted_at_chk  CHECK ((status = 'DELETED') = (deleted_at IS NOT NULL))
);

-- Danh sách comment ACTIVE của một post: keyset (created_at, id).
CREATE INDEX comments_post_idx
    ON interaction.comments (post_id, created_at DESC, id DESC)
    WHERE status = 'ACTIVE';

-- ---------------------------------------------------------------------------
-- Repost
-- ---------------------------------------------------------------------------
CREATE TABLE interaction.reposts (
    user_id    bigint      NOT NULL,
    post_id    bigint      NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),

    PRIMARY KEY (user_id, post_id)                         -- isReposted, idempotent repost / unrepost
);

-- Repost trên trang cá nhân của một user: keyset (created_at, post_id).
CREATE INDEX reposts_user_idx
    ON interaction.reposts (user_id, created_at DESC, post_id DESC);


-- +goose Down
DROP TABLE interaction.reposts;
DROP TABLE interaction.comments;
DROP TABLE interaction.likes;
DROP SCHEMA interaction;
