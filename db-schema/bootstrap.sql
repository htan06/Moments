-- Khởi tạo role và quyền nền cho database. QUẢN TRỊ CHẠY TAY bằng tài khoản admin,
-- một lần khi dựng database (chạy lại nhiều lần vẫn an toàn). Không chạy trong moments-migrate.
--
--   psql -U <admin> -d moments_database -v migrator=<role migrator> -v app_user=<role ứng dụng> -f bootstrap.sql
--
-- Script không đặt mật khẩu; đặt riêng sau khi chạy (không để mật khẩu trong file / lịch sử shell):
--   psql -U <admin> -d moments_database -c '\password <role>'
--
-- Phân quyền:
--   admin     : sở hữu database và schema public. Chỉ dùng cho bước này.
--   migrator  : chạy goose + grants.sql. Sở hữu schema users / content / interaction và mọi bảng trong đó.
--   app_user  : role của app, chỉ có quyền thao tác dữ liệu cấp trong grants.sql.

\set ON_ERROR_STOP on

-- ---------------------------------------------------------------------------
-- Role
-- ---------------------------------------------------------------------------
SELECT format('CREATE ROLE %I LOGIN', :'migrator')
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'migrator') \gexec
SELECT format('CREATE ROLE %I LOGIN', :'app_user')
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'app_user') \gexec

ALTER ROLE :"migrator" NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS;
ALTER ROLE :"app_user" NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS;

-- ---------------------------------------------------------------------------
-- Database
-- ---------------------------------------------------------------------------
SELECT current_database() AS dbname \gset

-- CREATE: migration đầu của mỗi module chạy CREATE SCHEMA IF NOT EXISTS.
GRANT CONNECT, CREATE ON DATABASE :"dbname" TO :"migrator";
GRANT CONNECT         ON DATABASE :"dbname" TO :"app_user";

-- goose tạo bảng goose_version_* trong schema public.
GRANT USAGE, CREATE ON SCHEMA public TO :"migrator";

-- ---------------------------------------------------------------------------
-- Database đã migrate bằng admin trước đây: chuyển quyền sở hữu sang migrator.
-- Database mới thì các lệnh dưới không làm gì.
-- ---------------------------------------------------------------------------
SELECT format('ALTER SCHEMA %I OWNER TO %I', nspname, :'migrator')
FROM pg_namespace
WHERE nspname IN ('users', 'content', 'interaction') \gexec

SELECT format('ALTER %s %I.%I OWNER TO %I',
              CASE c.relkind WHEN 'S' THEN 'SEQUENCE' WHEN 'v' THEN 'VIEW'
                             WHEN 'm' THEN 'MATERIALIZED VIEW' ELSE 'TABLE' END,
              n.nspname, c.relname, :'migrator')
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE c.relkind IN ('r', 'p', 'S', 'v', 'm')
  AND NOT EXISTS (SELECT 1 FROM pg_depend d   -- sequence của cột identity/serial đi theo bảng
                  WHERE d.classid = 'pg_class'::regclass AND d.objid = c.oid AND d.deptype IN ('a', 'i'))
  AND (n.nspname IN ('users', 'content', 'interaction')
       OR (n.nspname = 'public' AND c.relname LIKE 'goose\_version\_%')) \gexec
