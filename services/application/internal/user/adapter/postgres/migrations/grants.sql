-- Quyền của role ứng dụng trên schema users. services/migrate áp file này ngay sau goose up của module,
-- bằng role migrator, trong một transaction. goose bỏ qua file này vì tên không có số version.
-- "__app_role__" được thay bằng tên role đã quote (POSTGRES_APPLICATION_USER). Placeholder là
-- identifier hợp lệ để sqlc (đọc thư mục này làm schema) parse được. Chạy tay thì thay chuỗi đó.
-- Thu hồi hết rồi cấp lại, để quyền trong DB luôn khớp đúng file (bỏ một dòng GRANT là quyền đó mất).
-- Thêm bảng mới thì thêm GRANT vào đây; chỉ cấp thao tác nghiệp vụ cần.

REVOKE ALL ON ALL TABLES IN SCHEMA users FROM "__app_role__";
REVOKE ALL ON SCHEMA users FROM "__app_role__";
GRANT USAGE ON SCHEMA users TO "__app_role__";

GRANT SELECT, INSERT, UPDATE ON users.profiles TO "__app_role__";
GRANT SELECT, INSERT, DELETE ON users.follows  TO "__app_role__";
GRANT SELECT, INSERT, DELETE ON users.blocks   TO "__app_role__";
