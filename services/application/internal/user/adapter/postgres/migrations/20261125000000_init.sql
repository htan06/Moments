-- +goose Up
CREATE SCHEMA IF NOT EXISTS users;

-- ---------------------------------------------------------------------------
-- Profile (1–1 với Account của module account, tạo ngay sau Account – ACCOUNT-R7).
-- user_id = account.accounts.id; tham chiếu chéo module chỉ bằng id, không có FK.
-- ---------------------------------------------------------------------------
CREATE TABLE users.profiles (
    user_id          bigint      PRIMARY KEY,
    username         text        NOT NULL,                 -- bản sao account.accounts.username, đồng bộ khi đổi (saga)
    name             text        NOT NULL,
    bio              text,                                 -- rỗng thì lưu NULL
    avatar_key       text,                                 -- key bucket media, không kèm bucket
    avatar_thumb_key text,
    follower_count   bigint      NOT NULL DEFAULT 0,
    following_count  bigint      NOT NULL DEFAULT 0,
    post_count       bigint      NOT NULL DEFAULT 0,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT profiles_username_lower_chk  CHECK (username = lower(username) AND username <> ''),
    CONSTRAINT profiles_name_len_chk        CHECK (char_length(name) BETWEEN 1 AND 50),
    CONSTRAINT profiles_bio_len_chk         CHECK (bio IS NULL OR char_length(bio) BETWEEN 1 AND 150),
    CONSTRAINT profiles_avatar_pair_chk     CHECK ((avatar_key IS NULL) = (avatar_thumb_key IS NULL)),
    CONSTRAINT profiles_follower_count_chk  CHECK (follower_count  >= 0),
    CONSTRAINT profiles_following_count_chk CHECK (following_count >= 0),
    CONSTRAINT profiles_post_count_chk      CHECK (post_count      >= 0)
);

-- Bản sao username cũng duy nhất; đổi username giữa hai module ràng buộc bằng saga.
CREATE UNIQUE INDEX profiles_username_uq ON users.profiles (username);

-- ---------------------------------------------------------------------------
-- Follow
-- ---------------------------------------------------------------------------
CREATE TABLE users.follows (
    follower_id bigint      NOT NULL,
    followee_id bigint      NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),

    PRIMARY KEY (follower_id, followee_id),                -- danh sách following (prefix follower_id)
    CONSTRAINT follows_not_self_chk CHECK (follower_id <> followee_id)
);

-- Danh sách followers của một user: keyset (created_at, follower_id).
CREATE INDEX follows_followers_idx
    ON users.follows (followee_id, created_at DESC, follower_id DESC);

-- Danh sách following của một user: keyset (created_at, followee_id).
CREATE INDEX follows_following_idx
    ON users.follows (follower_id, created_at DESC, followee_id DESC);

-- ---------------------------------------------------------------------------
-- Block
-- ---------------------------------------------------------------------------
CREATE TABLE users.blocks (
    blocker_id bigint      NOT NULL,
    blocked_id bigint      NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),

    PRIMARY KEY (blocker_id, blocked_id),                  -- kiểm tra A block B
    CONSTRAINT blocks_not_self_chk CHECK (blocker_id <> blocked_id)
);

-- Kiểm tra chiều ngược lại (B block A) và lọc danh sách theo người bị chặn.
CREATE INDEX blocks_blocked_idx ON users.blocks (blocked_id, blocker_id);

-- Danh sách đã block của chính mình: keyset (created_at, blocked_id).
CREATE INDEX blocks_list_idx
    ON users.blocks (blocker_id, created_at DESC, blocked_id DESC);


-- +goose Down
DROP TABLE users.blocks;
DROP TABLE users.follows;
DROP TABLE users.profiles;
DROP SCHEMA users;
