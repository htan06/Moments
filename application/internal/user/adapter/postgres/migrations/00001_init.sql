-- +goose Up
CREATE SCHEMA IF NOT EXISTS users;

-- ---------------------------------------------------------------------------
-- Account
-- ---------------------------------------------------------------------------
CREATE TABLE users.accounts (
    id            bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email         text        NOT NULL,                    -- lưu chữ thường, bất biến
    username      text        NOT NULL,                    -- lưu chữ thường (USER-R5)
    password_hash text        NOT NULL,
    status        text        NOT NULL DEFAULT 'ACTIVE',
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT accounts_email_lower_chk    CHECK (email = lower(email)),
    CONSTRAINT accounts_username_lower_chk CHECK (username = lower(username)),
    CONSTRAINT accounts_password_hash_chk  CHECK (password_hash <> ''),
    CONSTRAINT accounts_status_chk         CHECK (status IN ('ACTIVE'))
);

CREATE UNIQUE INDEX accounts_email_uq    ON users.accounts (email);
CREATE UNIQUE INDEX accounts_username_uq ON users.accounts (username);

-- ---------------------------------------------------------------------------
-- Profile (1–1 với Account, tạo cùng transaction – USER-R7)
-- ---------------------------------------------------------------------------
CREATE TABLE users.profiles (
    user_id          bigint      PRIMARY KEY REFERENCES users.accounts (id),
    name             text        NOT NULL,
    bio              text,                                 -- rỗng thì lưu NULL
    avatar_key       text,                                 -- key bucket media, không kèm bucket
    avatar_thumb_key text,
    follower_count   bigint      NOT NULL DEFAULT 0,
    following_count  bigint      NOT NULL DEFAULT 0,
    post_count       bigint      NOT NULL DEFAULT 0,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT profiles_name_len_chk        CHECK (char_length(name) BETWEEN 1 AND 50),
    CONSTRAINT profiles_bio_len_chk         CHECK (bio IS NULL OR char_length(bio) BETWEEN 1 AND 150),
    CONSTRAINT profiles_avatar_pair_chk     CHECK ((avatar_key IS NULL) = (avatar_thumb_key IS NULL)),
    CONSTRAINT profiles_follower_count_chk  CHECK (follower_count  >= 0),
    CONSTRAINT profiles_following_count_chk CHECK (following_count >= 0),
    CONSTRAINT profiles_post_count_chk      CHECK (post_count      >= 0)
);

-- ---------------------------------------------------------------------------
-- Follow
-- ---------------------------------------------------------------------------
CREATE TABLE users.follows (
    follower_id bigint      NOT NULL REFERENCES users.accounts (id),
    followee_id bigint      NOT NULL REFERENCES users.accounts (id),
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
    blocker_id bigint      NOT NULL REFERENCES users.accounts (id),
    blocked_id bigint      NOT NULL REFERENCES users.accounts (id),
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
DROP TABLE users.accounts;
DROP SCHEMA users;
