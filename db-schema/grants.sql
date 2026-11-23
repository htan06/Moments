-- Chạy bằng psql với role migrator (chủ sở hữu schema / bảng), biến psql app_user = tên role ứng dụng.
-- Xem docker/migrate/migrate.sh. CONNECT trên database do quản trị cấp ở bootstrap.sql.

-- ---------------------------------------------------------------------------
-- Thu hồi toàn bộ quyền trên bảng rồi cấp lại theo file này,
-- để quyền trong DB luôn khớp đúng file (bỏ một dòng GRANT là quyền đó mất).
-- ---------------------------------------------------------------------------
REVOKE ALL ON ALL TABLES IN SCHEMA users, content, interaction FROM :"app_user";

-- ---------------------------------------------------------------------------
-- Schema
-- ---------------------------------------------------------------------------
GRANT USAGE ON SCHEMA users, content, interaction TO :"app_user";

-- ---------------------------------------------------------------------------
-- users
-- ---------------------------------------------------------------------------
GRANT SELECT, INSERT, UPDATE ON users.accounts TO :"app_user";
GRANT SELECT, INSERT, UPDATE ON users.profiles TO :"app_user";
GRANT SELECT, INSERT, DELETE ON users.follows  TO :"app_user";
GRANT SELECT, INSERT, DELETE ON users.blocks   TO :"app_user";

-- ---------------------------------------------------------------------------
-- content
-- ---------------------------------------------------------------------------
GRANT SELECT                 ON content.songs      TO :"app_user";
GRANT SELECT, INSERT, UPDATE ON content.posts      TO :"app_user";
GRANT SELECT, INSERT         ON content.post_songs TO :"app_user";
GRANT SELECT, INSERT, UPDATE ON content.media      TO :"app_user";

-- ---------------------------------------------------------------------------
-- interaction
-- ---------------------------------------------------------------------------
GRANT SELECT, INSERT, DELETE ON interaction.likes    TO :"app_user";
GRANT SELECT, INSERT, UPDATE ON interaction.comments TO :"app_user";
GRANT SELECT, INSERT, DELETE ON interaction.reposts  TO :"app_user";
