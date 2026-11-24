#!/bin/sh
# Chạy migration goose cho từng module, rồi cấp quyền cho role ứng dụng.
# Role migrator và role ứng dụng do quản trị tạo sẵn bằng db-schema/bootstrap.sql;
# script này không tạo role, không đụng mật khẩu, không dùng tài khoản admin.
# Kết nối bằng role migrator qua biến PGHOST / PGPORT / PGUSER / PGPASSWORD / PGDATABASE
# (goose và psql đều đọc các biến này).
set -eu

: "${POSTGRES_APPLICATION_USER:?missing}"

# module:bảng-version (bảng goose_version_<schema>)
for entry in user:users content:content interaction:interaction; do
    module=${entry%%:*}
    table=goose_version_${entry#*:}
    echo "migrate: $module"
    goose -dir "/migrations/$module/adapter/postgres/migrations" -table "$table" \
        postgres "dbname=$PGDATABASE" up
done

echo "migrate: grants"
# --single-transaction: REVOKE + GRANT áp dụng cùng lúc, app không bị mất quyền giữa chừng.
psql -X -q --single-transaction -v ON_ERROR_STOP=1 -v app_user="$POSTGRES_APPLICATION_USER" -f /sql/grants.sql

echo "migrate: done"
