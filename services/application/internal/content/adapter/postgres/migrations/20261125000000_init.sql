-- +goose Up
CREATE SCHEMA IF NOT EXISTS content;

-- ---------------------------------------------------------------------------
-- Song (danh mục nạp sẵn)
-- ---------------------------------------------------------------------------
CREATE TABLE content.songs (
    id          bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        text        NOT NULL,
    artist      text        NOT NULL,
    file_key    text        NOT NULL,                      -- key bucket media; API ghép thành url
    duration_ms integer     NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT songs_name_chk     CHECK (name     <> ''),
    CONSTRAINT songs_artist_chk   CHECK (artist   <> ''),
    CONSTRAINT songs_file_key_chk CHECK (file_key <> ''),
    CONSTRAINT songs_duration_chk CHECK (duration_ms > 0)
);

-- ---------------------------------------------------------------------------
-- Post
-- ---------------------------------------------------------------------------
CREATE TABLE content.posts (
    id              bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    author_id       bigint      NOT NULL,
    caption         text,                                  -- rỗng thì NULL
    aspect_ratio    text        NOT NULL,
    cover_key       text,                                  -- gán một lần khi media position 0 READY
    visibility      text        NOT NULL DEFAULT 'PUBLIC',
    comment_enabled boolean     NOT NULL DEFAULT true,
    status          text        NOT NULL DEFAULT 'PENDING',
    retryable       boolean     NOT NULL DEFAULT true,

    like_count      bigint      NOT NULL DEFAULT 0,
    comment_count   bigint      NOT NULL DEFAULT 0,
    repost_count    bigint      NOT NULL DEFAULT 0,

    created_at      timestamptz NOT NULL DEFAULT now(),
    published_at    timestamptz,
    deleted_at      timestamptz,
    updated_at      timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT posts_caption_len_chk  CHECK (caption IS NULL OR char_length(caption) BETWEEN 1 AND 2200),
    CONSTRAINT posts_aspect_ratio_chk CHECK (aspect_ratio IN ('SQUARE', 'PORTRAIT', 'LANDSCAPE')),
    CONSTRAINT posts_visibility_chk   CHECK (visibility   IN ('PUBLIC', 'FOLLOWERS', 'PRIVATE')),
    CONSTRAINT posts_status_chk       CHECK (status       IN ('PENDING', 'PUBLISHED', 'FAILED', 'DELETED')),
    CONSTRAINT posts_published_at_chk CHECK (status <> 'PUBLISHED' OR published_at IS NOT NULL),
    CONSTRAINT posts_deleted_at_chk   CHECK ((status = 'DELETED') = (deleted_at IS NOT NULL)),
    CONSTRAINT posts_like_count_chk    CHECK (like_count    >= 0),
    CONSTRAINT posts_comment_count_chk CHECK (comment_count >= 0),
    CONSTRAINT posts_repost_count_chk  CHECK (repost_count  >= 0)
);

-- Danh sách post của một tác giả: keyset (created_at, id), bỏ DELETED.
-- Lọc status / visibility theo người xem (CONTENT-R5) chạy trên index này.
CREATE INDEX posts_author_idx
    ON content.posts (author_id, created_at DESC, id DESC)
    WHERE status <> 'DELETED';

-- ---------------------------------------------------------------------------
-- PostSong (liên kết Post – Song: 1 post tối đa 1 song, 1 song nhiều post)
-- Tạo cùng transaction với Post; bất biến. Xóa Song thì chỉ xóa liên kết (CASCADE),
-- post giữ nguyên và coi như không có nhạc.
-- ---------------------------------------------------------------------------
CREATE TABLE content.post_songs (
    post_id    bigint      PRIMARY KEY REFERENCES content.posts (id),
    song_id    bigint      NOT NULL    REFERENCES content.songs (id) ON DELETE CASCADE,
    start_ms   integer     NOT NULL,
    end_ms     integer     NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),

    -- end_ms ≤ durationMs của song kiểm tra ở domain (CHECK không đọc được bảng khác).
    CONSTRAINT post_songs_start_chk CHECK (start_ms >= 0),
    CONSTRAINT post_songs_range_chk CHECK (end_ms - start_ms BETWEEN 5000 AND 90000)
);

-- Phục vụ ON DELETE CASCADE khi xóa Song (tránh quét toàn bảng).
CREATE INDEX post_songs_song_idx ON content.post_songs (song_id);

-- ---------------------------------------------------------------------------
-- Media (thuộc Post, tạo cùng transaction với Post)
-- ---------------------------------------------------------------------------
CREATE TABLE content.media (
    id            bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    post_id       bigint      NOT NULL REFERENCES content.posts (id),
    position      smallint    NOT NULL,
    type          text        NOT NULL,
    raw_key       text        NOT NULL,                    -- key file gốc trên bucket raw
    file_key      text,                                    -- key file đã xử lý trên bucket media; có khi READY
    status        text        NOT NULL DEFAULT 'PENDING',
    attempt       smallint    NOT NULL DEFAULT 1,
    dispatched_at timestamptz,
    failure_code  text,

    -- metadata: chỉ có khi READY
    width         integer,
    height        integer,
    duration_ms   integer,                                 -- bắt buộc với VIDEO khi READY
    size_bytes    bigint,

    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT media_position_chk CHECK (position BETWEEN 0 AND 19),
    CONSTRAINT media_type_chk     CHECK (type   IN ('IMAGE', 'VIDEO')),
    CONSTRAINT media_status_chk   CHECK (status IN ('PENDING', 'READY', 'FAILED')),
    CONSTRAINT media_attempt_chk  CHECK (attempt BETWEEN 1 AND 4),
    CONSTRAINT media_failure_chk  CHECK (
        (status = 'FAILED') = (failure_code IS NOT NULL)
        AND (failure_code IS NULL
             OR failure_code IN ('DISPATCH_FAILED', 'PROCESSING_ERROR', 'INVALID_MEDIA'))
    ),
    CONSTRAINT media_ready_chk CHECK (
        status <> 'READY'
        OR (file_key IS NOT NULL
            AND width > 0 AND height > 0 AND size_bytes > 0
            AND (type = 'IMAGE' OR (type = 'VIDEO' AND duration_ms > 0)))
    )
);

-- Thứ tự hiển thị trong post; cũng dùng để lấy media của một post.
CREATE UNIQUE INDEX media_post_position_uq ON content.media (post_id, position);

-- raw_key sinh từ mediaName (UUID); dùng để khớp kết quả xử lý và chống dùng một file gốc cho hai media.
CREATE UNIQUE INDEX media_raw_key_uq ON content.media (raw_key);

-- Mỗi file đã xử lý chỉ thuộc một media. NULL (chưa READY) không bị ràng buộc.
CREATE UNIQUE INDEX media_file_key_uq ON content.media (file_key);

-- +goose Down
DROP TABLE content.media;
DROP TABLE content.post_songs;
DROP TABLE content.posts;
DROP TABLE content.songs;
DROP SCHEMA content;
