# CLAUDE.md – Moments backend

Moments là mạng xã hội chia sẻ hình ảnh, video. Repo này là backend, tổ chức theo **modular monolith**: một binary, nhiều module nghiệp vụ tách biệt.

> **Đang chuyển đổi kiến trúc.** Repo trước đây là microservice (`auth`, `user`, `post`, `media-processor`…, mỗi thư mục một Go module). Đang chuyển về modular monolith theo file này. **Code mới nằm trong `application/`** (Go module `github.com/htan06/Moments/application`, không dùng `go.work`); mọi đường dẫn trong mục "Cấu trúc dự án" tính từ `application/`. Code microservice cũ **không phải mẫu**; code mới theo cấu trúc và quy ước bên dưới. Chưa rõ phần nào của code cũ được giữ, xóa hay chuyển vào module nào thì hỏi dev.

## Cách làm việc

- **Mô tả domain** (entity, invariant, trạng thái, quy tắc nghiệp vụ của `user` và `content`): [`docs/domain.md`](./docs/domain.md). Đây là nguồn sự thật về nghiệp vụ; thiếu hoặc mâu thuẫn thì dừng và hỏi dev, không tự đặt quy tắc. Cập nhật file này khi quy tắc nghiệp vụ đổi.
- **Không có đặc tả từng use case.** Luồng chi tiết, mã lỗi, giá trị policy: code, test và `config` là spec. Làm theo lát cắt dọc nhỏ, chạy được end-to-end, rồi mới mở rộng.
- Yêu cầu chưa rõ hoặc có nhiều cách hiểu: hỏi dev một câu ngắn, không tự bịa quy tắc nghiệp vụ.
- Quyết định đáng nhớ (khó đổi về sau) ghi thành vài dòng trong mục "Quyết định đã chốt" cuối file này, không viết tài liệu riêng.
- Mọi giá trị chính sách (TTL, giới hạn số lần, độ dài, dung lượng…) lấy từ `config`, không hard-code.

## Phạm vi hiện tại

| Module | Phạm vi |
|---|---|
| `user` | Account (đăng ký bằng Email OTP, đăng nhập, quên / đổi mật khẩu, đổi username), Profile (kèm avatar), Follow, Block |
| `content` | Post (kèm Media ảnh / video), Song |
| `interaction` | Like, Comment, Repost (số đếm lưu trong Post của `content`, cập nhật qua facade của `content`) |

Feed và các mục ở phần "Chưa thiết kế" của `docs/domain.md`: **chưa làm**. Không tự thiết kế khi chưa được yêu cầu.

## Tech stack

| Thành phần | Công nghệ |
|---|---|
| Ngôn ngữ | Go 1.26.2 |
| HTTP | gin |
| Database | PostgreSQL, truy cập qua code sinh bởi sqlc |
| Migration | goose (versioned, SQL viết tay) |
| Session / bộ đếm có TTL | Redis |
| Work queue | RabbitMQ |
| Event stream | Kafka (kết quả xử lý media) |
| Object storage | RustFS (S3-compatible) |
| Email | Mail service riêng (Gmail SMTP). Backend chỉ đẩy task vào RabbitMQ |

Thư viện mặc định (dùng nếu `go.mod` chưa có lựa chọn khác): `jackc/pgx/v5`, `pressly/goose/v3`, `redis/go-redis/v9`, `rabbitmq/amqp091-go`, `segmentio/kafka-go`, `aws/aws-sdk-go-v2/service/s3`, `golang-jwt/jwt/v5`, `golang.org/x/crypto/bcrypt`, `log/slog`, `joho/godotenv` (chỉ trong `cmd/api`, nạp `.env` khi chạy local). Cần đọc kích thước ảnh / video từ header thì dùng `golang.org/x/image`, `rwcarlsen/goexif`, `abema/go-mp4`. Muốn thêm dependency ngoài danh sách thì hỏi trước.

**Hạ tầng**
- RustFS: S3 client với endpoint tùy chỉnh và path-style addressing. Endpoint, bucket, key lấy từ config.
- Media service: **service riêng, không nằm trong application này** (ffmpeg, libvips). Backend publish task vào RabbitMQ và consume kết quả từ Kafka; không gọi HTTP, không dùng chung code.
- Email: backend không gửi trực tiếp và không giữ thông tin SMTP; chỉ publish task (có OTP, `expires_at`) lên RabbitMQ.

## Modular monolith

**Quy tắc ranh giới**
1. Mỗi module sở hữu dữ liệu riêng: một **Postgres schema** (`users`, `content`, `interaction`), một **prefix key Redis** (`user:`, `content:`, `interaction:`). Module khác không đọc / ghi trực tiếp.
2. Không join, không foreign key giữa schema của các module. Tham chiếu chéo chỉ bằng ID.
3. **Module gọi nhau qua facade, không dùng chéo data.**
   - Bên cung cấp (ví dụ `user`) mở method public ở package `module`; input / output chỉ là kiểu built-in hoặc DTO khai báo trong package `module`. Không trả entity domain hay model sqlc.
   - Bên dùng (ví dụ `content`) khai báo interface nhỏ ở nơi dùng, với kiểu dữ liệu của chính nó.
   - Cầu nối là adapter của bên dùng, ví dụ `internal/content/adapter/user/`: nơi **duy nhất** được import module khác. Wiring ở `cmd/api`.
4. Không import `domain`, `command`, `query`, `adapter/...` của module khác. Việc bất đồng bộ (gửi OTP, xử lý avatar) đi qua RabbitMQ.
5. Code dùng chung chỉ là hạ tầng kỹ thuật trong `internal/platform/`, không có logic nghiệp vụ.

## Cấu trúc dự án

Giữ cấu trúc phẳng nhất có thể. Mọi logic nằm trong `application/internal/`. Lệnh `go`, `sqlc`, `goose` chạy từ thư mục `application/`.

```
cmd/
  api/                  # main: config, platform, từng module, route, chạy server
config/                 # struct cấu hình + load từ biến môi trường
sqlc.yaml               # một entry cho mỗi module
internal/
  platform/             # hạ tầng dùng chung, không có nghiệp vụ
    pg/ redisx/ mq/ objstore/ mediaprobe/ token/ httpx/
  user/
    domain/             # entity, invariant, lỗi domain. Chỉ import stdlib
    command/            # write side: mỗi use case một file, một handler
    query/              # read side: mỗi use case một file, một model riêng
    adapter/            # mọi code nói chuyện với bên ngoài, mỗi công nghệ một package
      postgres/         # migrations/, queries/, sqlc/ (code sinh, không sửa tay)
      session/          # Redis; keys.go gom mọi key của module
      httpapi/          # gin handler, map lỗi domain → mã lỗi + HTTP status
    module/             # facade: New(deps), RegisterRoutes, method public cho module khác
```

Module `content` làm theo cùng mẫu, chỉ tạo package con nào thực sự cần.

**Quy tắc `adapter/`:** adapter phụ thuộc `domain`, `command`, `query`, không ngược lại. Adapter dùng client từ `internal/platform/`, không tự tạo kết nối, và không được module khác dùng.

## DDD và CQRS (trong mỗi module)

**Write side (`command`)**
1. Handler HTTP parse request, lấy `user_id` từ JWT (không bao giờ từ body / query), gọi command handler.
2. Command handler tải aggregate qua repository, gọi method của domain, lưu lại.
3. Invariant kiểm tra trong `domain`, không rải rác ở handler hay adapter.
4. Thao tác nhiều aggregate trong một transaction (ví dụ Account + Profile khi đăng ký) gói trong một method repository; adapter `postgres` tự mở và commit transaction.

**Read side (`query`)** – không đi qua domain
- Mỗi use case đọc có một model riêng và một câu query sqlc riêng, chỉ chứa trường cần trả. Không trả entity domain.
- Dữ liệu của module khác lấy qua facade.
- Phân trang bằng keyset: `(created_at, id) < ($cursor_created_at, $cursor_id)`, `ORDER BY created_at DESC, id DESC`, `LIMIT`. Không dùng `OFFSET`. Cursor là chuỗi mờ với client. Mỗi truy vấn phân trang cần index khớp bộ lọc và thứ tự.

## Bảo mật

- Không bao giờ nhận định danh user từ request; luôn lấy từ JWT.
- Không tiết lộ sự tồn tại của tài khoản: email đã / chưa tồn tại, sai username, sai mật khẩu đều trả cùng response và mã lỗi chung; thời gian phản hồi không chênh lệch đáng kể (ví dụ vẫn so sánh hash giả khi không tìm thấy user). Publish task gửi OTP thất bại chỉ log, response giữ nguyên.
- Chỉ lưu hash mật khẩu và hash OTP.
- Token chỉ chứa `user_id`, `username`, `user_status`; chỉ `user_status = ACTIVE` được gọi API cần xác thực.
- Response không bao giờ trả key object storage hay tên bucket; chỉ trả URL công khai.

## Quy ước Go

- Go idioms: package nhỏ, tên ngắn, số ít; không có `util` / `common` / `helper`; không đặt tên trùng thư viện đang dùng.
- **Interface nhỏ, khai báo ở nơi dùng.** Không tạo interface khi chỉ có một implementation và không ai cần thay thế.
- Constructor nhận dependency qua tham số, trả struct cụ thể. Không biến global, không `init()` để khởi tạo dependency. Wiring ở `cmd/api` và `module.New`.
- `context.Context` là tham số đầu tiên của mọi hàm có I/O.
- Lỗi: sentinel hoặc kiểu lỗi trong `domain`, so sánh bằng `errors.Is` / `errors.As`; wrap bằng `fmt.Errorf("...: %w", err)`. Chỉ `httpapi` biết mã lỗi và HTTP status.
- Không log mật khẩu, OTP, token, presigned URL, thông tin đăng nhập RustFS, nội dung message gửi mail service.
- Tên biến môi trường tìm trong [`docs/env-name.txt`](./docs/env-name.txt); không có thì báo dev.

## Database và migration

| Module | Schema | Bảng theo dõi version goose |
|---|---|---|
| `user` | `users` (không dùng `user` vì là từ khóa Postgres) | `goose_version_users` |
| `content` | `content` | `goose_version_content` |
| `interaction` | `interaction` | `goose_version_interaction` |

- Dùng goose, SQL viết tay, thư mục riêng mỗi module: `internal/<module>/adapter/postgres/migrations/`.
  ```
  goose -dir internal/user/adapter/postgres/migrations -s create <ten_thay_doi> sql
  ```
  Số thứ tự tuần tự (không dùng timestamp), tên chữ thường nối bằng `_`.
- Mỗi file có đủ `-- +goose Up` và `-- +goose Down`; Down đưa schema về đúng trạng thái trước Up.
- Migration đầu của module tạo schema (`CREATE SCHEMA IF NOT EXISTS users;`). Mọi bảng, index, constraint ghi kèm tên schema.
- Lệnh không chạy được trong transaction (`CREATE INDEX CONCURRENTLY`) để riêng một file với `-- +goose NO TRANSACTION`.
- **Không sửa file migration đã merge**; thay đổi thì tạo file mới.
- Ràng buộc unique (email, username chuẩn hóa chữ thường) phải có ở DB; adapter dịch lỗi vi phạm unique thành lỗi domain.
- **Số liệu đếm** (follower, like…): không dùng `COUNT(...)`; đọc từ field `*_count` lưu sẵn, cập nhật trong cùng transaction với thao tác insert / delete, chỉ khi dữ liệu thực sự đổi. Chưa có field `*_count` cần dùng thì dừng và báo dev, không tự thêm.
- Thao tác ghi quan hệ giữa hai user (follow, block…) tuần tự hóa theo cặp user bằng `pg_advisory_xact_lock`.
- Chạy migration bằng goose CLI như bước riêng trước khi khởi động app; app không tự migrate.
  ```
  goose -dir internal/user/adapter/postgres/migrations -table goose_version_users postgres "$DATABASE_POSTGRES_URI" up
  ```
  Migration mới phải chạy được `up` → `down` → `up`.
- sqlc đọc thẳng thư mục migration làm schema; mỗi module một entry trong `sqlc.yaml`. Query viết trong `adapter/postgres/queries/`, chạy `sqlc generate`.

## Redis và messaging

- Mỗi module gom **toàn bộ** key Redis vào **một file** (ví dụ `internal/user/adapter/session/keys.go`): prefix là hằng số, mỗi loại key là một hàm dựng key. Ngoài file đó không viết chuỗi key trực tiếp.
- Messaging làm tương tự: tên exchange / queue / topic, routing key và struct message của module khai báo tập trung trong một package (`internal/<module>/adapter/mq/`); tên thật lấy từ `config`. `internal/platform/mq` chỉ cung cấp kết nối và publisher chung.
- Tập trung theo module, không gom key hay kênh của mọi module vào một chỗ.
- Key / kênh mới thì ghi vào đúng file tập trung của module; không có bảng đăng ký riêng.

## Media (dùng chung các module)

Chỉ giữ nguyên tắc; chi tiết quyết định khi code và ghi lại trong "Quyết định đã chốt".

- Client upload thẳng lên bucket **raw** (private) bằng presigned `PUT`; module chỉ `HEAD` / `GET` theo range để kiểm tra. Bucket raw tự xóa object sau 1 ngày. Không bao giờ trả URL đọc của bucket raw.
- File đã xử lý nằm ở bucket **media**, client đọc qua CDN.
- **DB chỉ lưu key**, không lưu tên bucket và không ghép `bucket/key`.
- API trả **URL, không trả key** (`avatarUrl`, `coverUrl`, `url`…). Ghép URL ở `httpapi` bằng `PublicURL(key)` trong `internal/platform/objstore`; `domain`, `command`, `query` chỉ làm việc với key.
- Tên object dùng `mediaName` (UUID) do module sinh khi cấp presigned URL, không dùng ID trong DB. Hàm dựng object key gom vào một file của module.
- Xử lý bất đồng bộ: module publish task lên RabbitMQ (exchange `media.task`, routing key theo `image` / `video`), media service trả kết quả qua Kafka; mỗi module một topic kết quả riêng, khai báo trong field `reply_to` của task. Kết quả có thể đến nhiều lần: nhận lại cùng message phải cho cùng kết quả; commit offset Kafka sau khi commit transaction.
- Server không tin client: kiểm tra dung lượng, content type và tỉ lệ khung hình của file thật trên bucket raw trước khi publish task.
- Phiên bản đầu: ưu tiên **ảnh**. Video / HLS để sau khi lát cắt ảnh chạy được.

## Test

- Test `domain` bằng table-driven test, không cần I/O.
- Test command / query handler bằng fake viết tay thỏa interface ở nơi dùng; không dùng thư viện mock.
- Mỗi luồng lỗi quan trọng cần ít nhất một test, tên test nêu rõ tình huống (ví dụ `TestRegister_UsernameTaken`).

## Trước khi báo xong

```
gofmt -l .        # phải rỗng
go vet ./...
sqlc generate     # nếu có sửa queries hoặc migration
# nếu có migration mới: chạy up → down → up trên DB local
go test ./...
```

## Quyết định đã chốt

*(Thêm vào đây khi chốt, mỗi quyết định một–hai dòng: gì, vì sao.)*

- **JWT ký bằng RS256** (private key ký, public key xác thực). Client gửi token qua header `Authorization: Bearer <token>`.
- **Key RS256 nằm trong file**, đường dẫn lấy từ `PRIVATE_KEY_PATH` và `PUBLIC_KEY_PATH`. Access token và refresh token dùng **hai cặp key riêng**, phân biệt bằng loại block PEM: file private chứa `ACCESS PRIVATE KEY` và `REFRESH PRIVATE KEY`, file public chứa `ACCESS PUBLIC KEY` và `REFRESH PUBLIC KEY`. Private key định dạng PKCS#8 (`x509.ParsePKCS8PrivateKey`), public key định dạng PKIX (`x509.ParsePKIXPublicKey`).
- **Response lỗi HTTP** có dạng `{"code": "...", "message": "..."}` (`httpx.Error`); `message` tùy chọn. Auth sai / thiếu token → 401 `UNAUTHORIZED`; user không `ACTIVE` → 403 `FORBIDDEN`.
- **Nội dung email OTP tạm thời:** `Your OTP is <otp>`. Template là task riêng, chưa làm.
- **Chưa lên production nên sửa thẳng migration khởi tạo**: mỗi module một migration goose `internal/<module>/adapter/postgres/migrations/00001_init.sql`, là **nguồn schema duy nhất**; thay đổi schema thì sửa trực tiếp file này. `db-schema/` ở gốc repo chỉ còn `bootstrap.sql`, `grants.sql` và `moments.sql` (schema microservice cũ, không sửa). Quy tắc "không sửa migration đã merge" áp dụng từ khi lên production.
- **Quyền DB của role ứng dụng (`POSTGRES_APPLICATION_USER`) khai báo ở `db-schema/grants.sql`**, chỉ cấp thao tác nghiệp vụ cần. Thêm bảng mới thì thêm GRANT vào file này.
- **Migration local chạy bằng service `moments-migrate`** (`docker/migrate/`, goose + psql, role `POSTGRES_MIGRATOR_*`): goose `up` từng module → `grants.sql`. Không dùng `docker-entrypoint-initdb.d`. **Tạo role, mật khẩu là việc của quản trị**; repo chỉ lo schema, migration và quyền. Thêm module mới thì thêm vào vòng lặp trong `migrate.sh`.
- **Ba role DB, migration không dùng admin**: admin sở hữu database và schema `public`, chỉ chạy tay `db-schema/bootstrap.sql` (tạo role, cấp `CONNECT`/`CREATE` database + `CREATE` trên `public` cho migrator, `CONNECT` cho app); **migrator** (không superuser) sở hữu schema của các module và tự cấp quyền trong `grants.sql`; **app** chỉ có quyền thao tác dữ liệu. Vì sao: lộ thông tin của migrate không lộ superuser; mật khẩu migrator chỉ đưa cho job migrate, không đưa cho app.
- **Tập giá trị cố định (status, type…) lưu `text` + `CHECK`, không dùng enum Postgres**: đổi / bỏ giá trị chỉ cần thay constraint, migration Down viết được.
- **Chưa viết integration test** (Postgres / Redis / RabbitMQ / RustFS); hiện chỉ unit test theo mục "Test". Dev làm khi đến bước CI.
- **Mail service là service riêng, nằm ngoài repo này**; nhận task qua work queue RabbitMQ. Backend chỉ publish task.
- **Media service là service riêng, xử lý bất đồng bộ**: giữ RabbitMQ (task) + Kafka (kết quả) ngay từ giai đoạn đầu, không xử lý ảnh trong monolith.
- **Tên exchange / queue / topic / consumer group và struct message đặt khi code**, gom tập trung theo module trong `internal/<module>/adapter/mq/` (giống `keys.go` của Redis); tên thật lấy từ `config`.
- **Topology RabbitMQ / Kafka do dev quản lý, app không tự tạo**: dev đọc code, tạo queue / topic cần thiết và ghi lại thành file schema trong repo. Thêm exchange / queue / topic mới thì báo dev trong phần tóm tắt khi xong việc.
- **Đăng ký 4 bước**: `POST /auth/register` (gửi OTP, luôn trả `sessionId`) → `/auth/register/verify-otp` → `/auth/register/complete` (chọn username + mật khẩu, trả token). Gửi lại OTP: `/auth/register/resend-otp`. Sai input / username trùng ở bước complete giữ phiên để user sửa.
- **VerificationSession lưu Redis** (JSON, TTL = thời gian còn lại, cập nhật bằng `WATCH` + `KEEPTTL`); OTP lưu dạng SHA-256 gắn với `session_id`. Gửi mail OTP chạy nền (goroutine), để nhánh email thật / dummy có thời gian phản hồi như nhau.
- **Mật khẩu: bcrypt trên SHA-256 (base64) của mật khẩu**, vì bcrypt chỉ đọc 72 byte đầu mà 64 ký tự UTF-8 có thể vượt quá.
- **Đếm đăng nhập sai (USER-R3)**: identifier khớp tài khoản thì đếm theo account id (email và username dùng chung bộ đếm), không khớp thì đếm theo identifier (hash). Cửa sổ cố định tính từ lần sai đầu tiên. Bị chặn trả 429 `TOO_MANY_ATTEMPTS` kèm `retryAfterSeconds` và header `Retry-After`.
- **Giá trị policy của module `user`** (TTL OTP, cooldown, số lần sai, bcrypt cost, username dành riêng) hiện là default trong `config.defaultUser()`, chưa đọc từ biến môi trường. Tên exchange / routing key mail lấy từ `MAIL_EXCHANGE`, `MAIL_OTP_ROUTING_KEY`.

## Để sau – không tự làm

- Template email OTP (task riêng).
- File schema topology RabbitMQ / Kafka (dev làm).
- Integration test và CI (dev làm khi đến bước CI).
