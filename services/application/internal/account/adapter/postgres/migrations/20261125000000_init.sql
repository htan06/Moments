-- +goose Up
CREATE SCHEMA IF NOT EXISTS account;

-- ---------------------------------------------------------------------------
-- Account (thông tin đăng nhập). Profile thuộc module user (schema users),
-- tham chiếu Account chỉ bằng id, không có foreign key giữa hai schema.
-- ---------------------------------------------------------------------------
CREATE TABLE account.accounts (
    id            bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email         text        NOT NULL,                    -- lưu chữ thường, bất biến
    username      text        NOT NULL,                    -- lưu chữ thường (ACCOUNT-R5)
    password_hash text        NOT NULL,
    status        text        NOT NULL DEFAULT 'ACTIVE',
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT accounts_email_lower_chk    CHECK (email = lower(email)),
    CONSTRAINT accounts_username_lower_chk CHECK (username = lower(username)),
    CONSTRAINT accounts_password_hash_chk  CHECK (password_hash <> ''),
    CONSTRAINT accounts_status_chk         CHECK (status IN ('ACTIVE'))
);

CREATE UNIQUE INDEX accounts_email_uq    ON account.accounts (email);
CREATE UNIQUE INDEX accounts_username_uq ON account.accounts (username);


-- +goose Down
DROP TABLE account.accounts;
DROP SCHEMA account;
