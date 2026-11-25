-- Khởi tạo role và quyền nền cho database. DEV CHẠY TAY bằng tài khoản admin,
-- sau khi dựng volume postgres mới và trước moments-migrate. Chạy lại nhiều lần vẫn an toàn.
-- Postgres không tự chạy file này (không dùng docker-entrypoint-initdb.d).
--
--   docker exec -i moments-postgres psql -U <admin> -d moments_database \
--     -v migrator=<role migrator> -v app_user=<role ứng dụng> < platform/postgres/init.sql
--
-- Script không đặt mật khẩu; đặt riêng cho từng role (nhập tương tác, không để mật khẩu
-- trong file / lịch sử shell). Phải khớp mật khẩu trong services/migrate/.env* và DATABASE_POSTGRES_URI
-- của services/application/.env*:
--   docker exec -it moments-postgres psql -U <admin> -d moments_database -c '\password <role>'
--
-- Phân quyền:
--   admin     : sở hữu database và schema public. Chỉ dùng cho bước này.
--   migrator  : chạy services/migrate (goose + grants.sql). Sở hữu schema account / users /
--               content / interaction và mọi bảng trong đó.
--   app_user  : role của app. Ở đây chỉ được CONNECT; quyền trên schema / bảng do
--               services/migrate cấp theo grants.sql của từng module.

\set ON_ERROR_STOP on

-- ---------------------------------------------------------------------------
-- Role
-- ---------------------------------------------------------------------------
SELECT format('CREATE ROLE %I LOGIN', :'migrator')
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'migrator') \gexec
SELECT format('CREATE ROLE %I LOGIN', :'app_user')
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'app_user') \gexec

ALTER ROLE :"migrator" LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS;
ALTER ROLE :"app_user" LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS;

-- ---------------------------------------------------------------------------
-- Database
-- ---------------------------------------------------------------------------
SELECT current_database() AS dbname \gset

-- CREATE: migration đầu của mỗi module chạy CREATE SCHEMA IF NOT EXISTS.
GRANT CONNECT, CREATE ON DATABASE :"dbname" TO :"migrator";
GRANT CONNECT         ON DATABASE :"dbname" TO :"app_user";

-- goose tạo bảng goose_version_* trong schema public.
GRANT USAGE, CREATE ON SCHEMA public TO :"migrator";

-- app_user không cần schema public; chặn để app không đọc / ghi bảng goose_version_*.
REVOKE ALL ON SCHEMA public FROM PUBLIC;

-- ---------------------------------------------------------------------------
-- Quyền của app_user trên schema / bảng nghiệp vụ KHÔNG cấp ở đây: services/migrate
-- áp grants.sql của từng module (internal/<module>/adapter/postgres/migrations/grants.sql) sau goose up.
-- Gỡ default privileges của phiên bản init.sql cũ (nếu từng chạy); không có thì không làm gì.
-- ---------------------------------------------------------------------------
ALTER DEFAULT PRIVILEGES FOR ROLE :"migrator" REVOKE USAGE ON SCHEMAS FROM :"app_user";
ALTER DEFAULT PRIVILEGES FOR ROLE :"migrator"
    REVOKE SELECT, INSERT, UPDATE, DELETE ON TABLES FROM :"app_user";
