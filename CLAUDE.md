# CLAUDE.md – Moments backend

Moments là mạng xã hội chia sẻ hình ảnh, video. Repo này là backend, tổ chức theo **modular monolith**: một binary, nhiều module nghiệp vụ tách biệt.

> **Đang chuyển đổi kiến trúc.** Repo trước đây là microservice (`auth`, `user`, `post`…, mỗi thư mục một Go module; nay gom trong `legacy/`, chỉ để tham khảo). Đang chuyển về modular monolith theo file này. **Code mới nằm trong `services/application/`** (Go module `github.com/htan06/Moments/application`, không dùng `go.work`); mọi đường dẫn trong mục "Cấu trúc dự án" tính từ `services/application/`. Code microservice cũ **không phải mẫu**; code mới theo cấu trúc và quy ước bên dưới. Chưa rõ phần nào của code cũ được giữ, xóa hay chuyển vào module nào thì hỏi dev.

## Cách làm việc

- **Mô tả domain** (entity, invariant, trạng thái, quy tắc nghiệp vụ của các module): [`docs/domain.md`](./docs/domain.md). Đây là nguồn sự thật về nghiệp vụ; thiếu hoặc mâu thuẫn thì dừng và hỏi dev, không tự đặt quy tắc. Cập nhật file này khi quy tắc nghiệp vụ đổi.
- **Không có đặc tả từng use case.** Luồng chi tiết, mã lỗi, giá trị policy: code, test và `config` là spec. Làm theo lát cắt dọc nhỏ, chạy được end-to-end, rồi mới mở rộng.
- Yêu cầu chưa rõ hoặc có nhiều cách hiểu: hỏi dev một câu ngắn, không tự bịa quy tắc nghiệp vụ.
- Quyết định đáng nhớ (khó đổi về sau) ghi thành vài dòng trong mục "Quyết định đã chốt" cuối file này, không viết tài liệu riêng.
- Mọi giá trị chính sách (TTL, giới hạn số lần, độ dài, dung lượng…) lấy từ `config`, không hard-code.
- **Giải thích việc đã làm vừa đủ để hiểu, không nén quá ngắn** (tóm tắt khi xong việc, mô tả PR, commit message, ghi chú quyết định). Với mỗi quyết định thiết kế, nêu: vấn đề cần giải quyết (kèm rule liên quan, ví dụ ACCOUNT-R1), cách làm cụ thể, và hạn chế / điểm cần dev xem lại. Dùng ví dụ cụ thể khi cơ chế khó hình dung (ví dụ "sai 3 lần qua username + 2 lần qua email → bị chặn"). Không viết kiểu liệt kê từ khóa như "WATCH + KEEPTTL, hash giả, cửa sổ cố định" mà không giải thích vì sao.

## Phạm vi hiện tại

| Module | Phạm vi |
|---|---|
| `account` | Account: đăng ký bằng Email OTP, đăng nhập, làm mới token, quên / đổi mật khẩu, đổi username |
| `user` | Profile (kèm avatar), Follow, Block |
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
1. Mỗi module sở hữu dữ liệu riêng: một **Postgres schema** (`account`, `users`, `content`, `interaction`), một **prefix key Redis** (`account:`, `user:`, `content:`, `interaction:`). Module khác không đọc / ghi trực tiếp.
2. Không join, không foreign key giữa schema của các module. Tham chiếu chéo chỉ bằng ID.
3. **Module gọi nhau qua facade, không dùng chéo data.**
   - Bên cung cấp (ví dụ `user`) mở method public ở package `module`; input / output chỉ là kiểu built-in hoặc DTO khai báo trong package `module`. Không trả entity domain hay model sqlc.
   - Bên dùng (ví dụ `content`) khai báo interface nhỏ ở nơi dùng, với kiểu dữ liệu của chính nó.
   - Cầu nối là adapter của bên dùng, ví dụ `internal/content/adapter/user/`: nơi **duy nhất** được import module khác. Wiring ở `cmd/api`.
4. Không import `domain`, `command`, `query`, `adapter/...` của module khác. Việc bất đồng bộ (gửi OTP, xử lý avatar) đi qua RabbitMQ.
5. Code dùng chung chỉ là hạ tầng kỹ thuật trong `internal/platform/`, không có logic nghiệp vụ.

## Cấu trúc dự án

Giữ cấu trúc phẳng nhất có thể. Mọi logic nằm trong `services/application/internal/`. Lệnh `go`, `sqlc`, `goose` chạy từ thư mục `services/application/`.

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
      httpapi/          # gin handler, map lỗi domain → mã lỗi + HTTP status; openapi.yaml của module
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
- Tên biến môi trường tìm trong `.env.example` của từng thành phần (`platform/`, `services/application/`, `services/mail/`, `services/migrate/`); không có thì báo dev. Thêm biến mới thì thêm vào file mẫu của đúng thành phần.

## Database và migration

| Module | Schema | Bảng theo dõi version goose |
|---|---|---|
| `account` | `account` | `goose_version_account` |
| `user` | `users` (không dùng `user` vì là từ khóa Postgres) | `goose_version_user` |
| `content` | `content` | `goose_version_content` |
| `interaction` | `interaction` | `goose_version_interaction` |

- Dùng goose, SQL viết tay, thư mục riêng mỗi module: `internal/<module>/adapter/postgres/migrations/`.
  ```
  goose -dir internal/user/adapter/postgres/migrations create <ten_thay_doi> sql
  ```
  Version là timestamp (`YYYYMMDDHHMMSS`, mặc định của `goose create`, không dùng `-s`), tên chữ thường nối bằng `_`.
- Mỗi file có đủ `-- +goose Up` và `-- +goose Down`; Down đưa schema về đúng trạng thái trước Up.
- Migration đầu của module tạo schema (`CREATE SCHEMA IF NOT EXISTS users;`). Mọi bảng, index, constraint ghi kèm tên schema.
- Lệnh không chạy được trong transaction (`CREATE INDEX CONCURRENTLY`) để riêng một file với `-- +goose NO TRANSACTION`.
- **Không sửa file migration đã merge**; thay đổi thì tạo file mới.
- Ràng buộc unique (email, username chuẩn hóa chữ thường) phải có ở DB; adapter dịch lỗi vi phạm unique thành lỗi domain.
- **Số liệu đếm** (follower, like…): không dùng `COUNT(...)`; đọc từ field `*_count` lưu sẵn, cập nhật trong cùng transaction với thao tác insert / delete, chỉ khi dữ liệu thực sự đổi. Chưa có field `*_count` cần dùng thì dừng và báo dev, không tự thêm.
- Thao tác ghi quan hệ giữa hai user (follow, block…) tuần tự hóa theo cặp user bằng `pg_advisory_xact_lock`.
- Chạy migration bằng lệnh `services/migrate` như bước riêng trước khi khởi động app; app không tự migrate. Truyền module nào thì migrate module đó (goose `up` rồi `grants.sql`, xem "Quyết định đã chốt"):
  ```
  cd services/migrate                                  # đọc services/migrate/.env
  go run ./cmd/migrate -module user=../application/internal/user/adapter/postgres/migrations
  ```
  Thử `down` một module thì dùng goose CLI trực tiếp:
  ```
  goose -dir internal/user/adapter/postgres/migrations -table goose_version_user postgres "<URI của migrator>" down
  ```
  Migration mới phải chạy được `up` → `down` → `up`.
- **Thêm bảng mới thì thêm GRANT vào `migrations/grants.sql`** của module đó trong cùng thay đổi; chỉ cấp thao tác nghiệp vụ cần. Thiếu GRANT thì app bị `permission denied`.
- sqlc đọc thẳng thư mục migration làm schema; mỗi module một entry trong `sqlc.yaml`. Query viết trong `adapter/postgres/queries/`, chạy `sqlc generate`.

## Redis và messaging

- Mỗi module gom **toàn bộ** key Redis vào **một file** (ví dụ `internal/account/adapter/session/keys.go`): prefix là hằng số, mỗi loại key là một hàm dựng key. Ngoài file đó không viết chuỗi key trực tiếp.
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
- **Task mail OTP chỉ mang dữ liệu, mail service render tiêu đề và nội dung**: body JSON `{"type":"otp","to":"<địa chỉ, không có tên hiển thị>","otp":"<chỉ chữ số>","expires_at":"<RFC 3339 UTC>"}`, publish lên `MAIL_EXCHANGE` / `MAIL_OTP_ROUTING_KEY`, có `message_id`. Vì sao: template (HTML đen trắng + text) thuộc mail service, backend không giữ nội dung email.
- **Chưa lên production nên sửa thẳng migration khởi tạo**: mỗi module một migration goose `internal/<module>/adapter/postgres/migrations/20261125000000_init.sql`, là **nguồn schema duy nhất**; thay đổi schema thì sửa trực tiếp file này. Role và quyền DB ở `platform/postgres/init.sql`. Quy tắc "không sửa migration đã merge" áp dụng từ khi lên production.
- **Quyền DB của role ứng dụng (`POSTGRES_APPLICATION_USER`) khai báo ở `grants.sql` của từng module** (`internal/<module>/adapter/postgres/migrations/grants.sql`, **trong** thư mục migration), chỉ cấp thao tác nghiệp vụ cần (ví dụ `content.songs` chỉ SELECT). Mỗi file thu hồi hết quyền của app trên schema của module (`REVOKE ALL ON ALL TABLES` + `REVOKE ALL ON SCHEMA`) rồi cấp lại theo file, nên bỏ một dòng GRANT là quyền đó mất. Role viết là `"__app_role__"`: lệnh migrate thay bằng tên role đã quote (escape, không chèn được SQL). Đặt trong thư mục migration không ảnh hưởng gì: goose bỏ qua file không có số version ở đầu tên; sqlc đọc file này như schema, nên placeholder phải là identifier hợp lệ (cú pháp psql `:"app_user"` làm sqlc lỗi parse). `init.sql` không cấp quyền nào trên schema / bảng nghiệp vụ (không dùng default privileges). Vì sao: module sở hữu schema thì sở hữu luôn quyền trên schema đó; DB không cho app thao tác ngoài nghiệp vụ (thay quyết định cũ "default privileges, chấp nhận thừa DELETE").
  - *Hạn chế:* thêm bảng mà quên GRANT thì app lỗi `permission denied` lúc chạy, không phải lúc migrate. Cột `id` dùng `IDENTITY` nên không cần quyền sequence; nếu dùng `serial` thì phải GRANT thêm `USAGE` trên sequence.
- **Migrate là lệnh Go chạy một lần rồi thoát** (**tạm thời, chỉ cho giai đoạn local**; dev sẽ chọn kiến trúc migrate khác khi lên production, không tự mở rộng hướng này). Những điểm production cần mà bản này chưa có: migration chép vào image theo từng bản phát hành (hiện mount từ mã nguồn), khóa chống chạy song song (`goose.WithSessionLocker` / advisory lock), quy trình deploy đảm bảo migrate xong mới cập nhật api. (`services/migrate/`, Go module `github.com/htan06/Moments/migrate`, dùng goose dạng thư viện). Migrate **cố ý** không phụ thuộc cấu trúc application và không biết có những module nào: chỉ migrate đúng các module được truyền, không cảnh báo về module không được truyền (để chạy riêng một module là việc bình thường). Đầu vào là danh sách `-module tên=thư_mục_migration` (lặp lại cờ, chạy theo thứ tự truyền; module nào cần migrate thì truyền), tài khoản migrator (`POSTGRES_MIGRATOR_*`), tên role application (`POSTGRES_APPLICATION_USER`), địa chỉ Postgres qua biến chuẩn libpq `PGHOST` / `PGPORT` / `PGDATABASE`. Với mỗi module: goose `up` thư mục đó (bảng version `goose_version_<tên>`), rồi áp `grants.sql` trong cùng thư mục **trong một transaction**, nên app không bao giờ thấy trạng thái đã REVOKE mà chưa GRANT. Module lỗi thì dừng, không chạy module sau. Danh sách module cho container nằm ở `command` của `moments-migrate` trong `services/compose.yml` (mount `services/application/internal` vào `/migrations`); thêm module mới thì thêm một dòng ở đó. Tên module chỉ gồm chữ thường, số, `_` vì nằm trong tên bảng version. Migration và `grants.sql` không nằm trong image, nên sửa chúng chỉ cần `docker compose -f services/compose.yml up moments-migrate`, không cần `--build`. Không dùng `docker-entrypoint-initdb.d`. **Tạo role, mật khẩu là việc của dev / quản trị, chạy tay**: volume postgres mới thì chạy `platform/postgres/init.sql` bằng admin rồi `\password` từng role (lệnh ở đầu file), trước migrate; quên thì migrate lỗi đăng nhập, không hỏng dữ liệu.
  - *Hạn chế:* goose và grant không chung transaction. Goose xong mà grant lỗi (ví dụ GRANT trên bảng chưa có) thì bảng mới đã tạo nhưng app chưa có quyền; sửa `grants.sql` rồi chạy lại migrate (goose bỏ qua migration đã chạy).
- **Ba role DB, migration không dùng admin**: admin sở hữu database và schema `public`, chỉ chạy tay `platform/postgres/init.sql` (tạo role, cấp `CONNECT`/`CREATE` database + `CREATE` trên `public` cho migrator, `CONNECT` cho app, thu hồi quyền của `PUBLIC` trên schema `public` để app không đụng bảng `goose_version_*`); **migrator** (không superuser) sở hữu schema và bảng của các module và cấp quyền cho app qua `grants.sql`; **app** chỉ có quyền thao tác dữ liệu (DML) theo `grants.sql`, không tạo / sửa / xóa bảng. Vì sao: lộ thông tin của migrate không lộ superuser; mật khẩu migrator chỉ đưa cho job migrate, không đưa cho app; container postgres không giữ mật khẩu migrator / app.
- **Tổ chức repo: platform tách hẳn khỏi services.**
  ```
  platform/          hạ tầng: compose.yml (postgres, redis, rabbitmq, kafka, rustfs), .env, postgres/init.sql
  services/          compose.yml (migrate, api, mail, nginx)
    application/     API (Go module), .env, .env.docker, secrets/ (key RS256)
    mail/            mail service (Go module), .env, .env.docker
    migrate/         lệnh migrate (Go module), .env, .env.docker (tài khoản migrator, tên role app)
    nginx/nginx.conf
  docs/  legacy/     legacy: code microservice cũ, không phải mẫu
  ```
  Vì sao: platform dựng một lần và có vòng đời riêng; services build lại thường xuyên. Mỗi thành phần giữ env / secret của chính nó, không có `.env` chung ở gốc repo.
- **Docker compose tách hai project dùng chung network `moments-net`** (tên cố định, do platform tạo). Chạy từ gốc repo, platform trước: `docker compose -f platform/compose.yml up -d`, rồi `docker compose -f services/compose.yml up -d --build`. Platform đọc `platform/.env` (compose tự tìm `.env` cạnh file compose). Nginx chuyển `/api/v1/*` sang api, bỏ prefix (`/api/v1/auth/login` → `/auth/login`).
- **Env và secret tách theo thành phần; mỗi service hai file env**: `.env` cho `go run` trên máy (địa chỉ `localhost`), `.env.docker` cho container (tên container, `PRIVATE_KEY_PATH=/secrets/...`). Compose nạp `.env.docker` qua `env_file` và không ghi đè gì. `platform/.env` chỉ chứa tài khoản quản trị hạ tầng (admin Postgres, user RabbitMQ, mật khẩu Redis, key RustFS); `services/migrate/.env*` chỉ chứa tài khoản migrator và tên role app. Nhờ vậy container api / mail không nhận mật khẩu admin hay migrator. Key JWT nằm ở `services/application/secrets/`, mount vào `/secrets`. Mẫu tên biến: `.env.example` của từng thành phần.
  - *Hạn chế:* mật khẩu service dùng phải ghi hai lần (`.env` và `.env.docker`) và giữ khớp tay. RabbitMQ, Redis, RustFS hiện chỉ có một tài khoản, nên mật khẩu đó nằm cả ở `platform/.env` lẫn env của api / mail. Muốn tách hẳn thì platform tạo tài khoản riêng cho từng service (user RabbitMQ, Redis ACL, access key RustFS).
- **Tập giá trị cố định (status, type…) lưu `text` + `CHECK`, không dùng enum Postgres**: đổi / bỏ giá trị chỉ cần thay constraint, migration Down viết được.
- **Chưa viết integration test** (Postgres / Redis / RabbitMQ / RustFS); hiện chỉ unit test theo mục "Test". Dev làm khi đến bước CI.
- **Mail service là service riêng ở `services/mail/`** (Go module `github.com/htan06/Moments/mail`, ngang cấp `services/application/`, không dùng chung code); consume queue `MAIL_QUEUE`, gửi qua SMTP (Gmail + app password, `SMTP_*`). Backend chỉ publish task. Task hết hạn thì bỏ, không gửi (ACCOUNT-R2).
- **Media service là service riêng, xử lý bất đồng bộ**: giữ RabbitMQ (task) + Kafka (kết quả) ngay từ giai đoạn đầu, không xử lý ảnh trong monolith.
- **Tên exchange / queue / topic / consumer group và struct message đặt khi code**, gom tập trung theo module trong `internal/<module>/adapter/mq/` (giống `keys.go` của Redis); tên thật lấy từ `config`.
- **Topology RabbitMQ / Kafka do dev quản lý, app không tự tạo**: dev đọc code, tạo queue / topic cần thiết và ghi lại thành file schema trong repo. Thêm exchange / queue / topic mới thì báo dev trong phần tóm tắt khi xong việc.
- **Đăng ký: gửi đủ thông tin trước, tạo user khi xác thực OTP** (thay luồng 4 bước cũ).
  1. Client kiểm tra username qua `GET /auth/username-availability` (chỉ đọc DB).
  2. `POST /auth/register {username, name, email, password}`: kiểm tra định dạng, hash mật khẩu, kiểm tra email (đã tồn tại thì tạo phiên dummy), tạo phiên, gửi OTP; luôn trả `sessionId`.
  3. `POST /auth/register/verify {sessionId, otp}`: OTP đúng thì insert Account + Profile và trả token.

  Gửi lại OTP qua `/auth/register/resend-otp`; dummy, cooldown và giới hạn sai OTP giữ nguyên (ACCOUNT-R1, ACCOUNT-R2).
  - *Vì sao:* luồng cũ cho phép ai có `sessionId` sau bước verify-otp cũng tạo được account bằng email đó. Giờ OTP nằm trong chính request tạo account. Phiên không còn `status`.
  - *Không giữ chỗ username:* nếu giữ chỗ thì API kiểm tra username phải đọc thêm Redis. Phiên dummy không giữ chỗ thì kẻ xấu dò được email đã đăng ký (gửi `email=X, username=u_test` rồi kiểm tra `u_test`). Còn nếu dummy cũng giữ chỗ thì kẻ xấu dùng email rác để chiếm username hàng loạt. Unique index trên `username` là trọng tài cuối: bị lấy mất giữa chừng (hiếm, tối đa 10 phút) thì trả `USERNAME_TAKEN`, hủy phiên.
  - *Bcrypt chạy ở bước register cho cả hai nhánh*, để thời gian phản hồi như nhau (ACCOUNT-R1).
- **VerificationSession lưu Redis** dạng JSON, TTL bằng thời gian sống còn lại nên phiên tự biến mất khi hết hạn. Cập nhật (nhập OTP, gửi lại) dùng `WATCH` để hai request đồng thời không làm mất một lần đếm sai, và `KEEPTTL` để không gia hạn phiên (ACCOUNT-R2). Chỉ lưu SHA-256 của OTP ghép với `session_id`, nên hash của phiên này không dùng được cho phiên khác.
- **Gửi mail OTP chạy nền (goroutine), request không chờ.** Vì sao: nhánh email mới phải publish lên RabbitMQ và chờ broker xác nhận, nhánh email đã tồn tại (phiên dummy) thì không; nếu chờ thì đo thời gian phản hồi là đoán được email đã đăng ký (ACCOUNT-R1). Publish lỗi lúc đó không còn đổi được response nên chỉ ghi log (không ghi email / OTP); user dùng "gửi lại". Hạn chế: app tắt đúng lúc đang publish thì mất task đó.
- **Mật khẩu: bcrypt (cost 12) trên SHA-256 (base64) của mật khẩu**, vì bcrypt chỉ đọc 72 byte đầu mà mật khẩu 64 ký tự UTF-8 (có dấu) có thể dài hơn; không pre-hash thì phần đuôi bị bỏ qua.
- **Đếm đăng nhập sai (ACCOUNT-R3).** *(Tạm thời: giới hạn thử sai là vấn đề chưa giải quyết, xem "Để sau".)* Identifier khớp tài khoản thì đếm theo account id, nên email và username của cùng tài khoản dùng chung bộ đếm (sai 3 lần qua username + 2 lần qua email → bị chặn). Không khớp tài khoản thì vẫn đếm theo identifier (băm SHA-256 trong key); nếu không, chỉ tài khoản có thật mới bị chặn và kẻ xấu dựa vào đó dò tài khoản. Cửa sổ cố định 15 phút tính từ lần sai đầu; đủ 5 lần thì chặn tới khi cửa sổ hết, trả 429 `TOO_MANY_ATTEMPTS` kèm `retryAfterSeconds` và header `Retry-After`. Đăng nhập thành công không xóa bộ đếm. Không tìm thấy user vẫn chạy bcrypt với một hash giả tạo lúc khởi động, để hai nhánh tốn thời gian như nhau.
  - *Cần dev xem lại:* (1) sai lần 1 lúc 0:00, lần 5 lúc 14:00 thì chỉ bị chặn 1 phút; muốn chặn đủ 15 phút kể từ lần thứ 5 thì đặt lại TTL khi chạm ngưỡng. (2) Kiểm tra chặn và tăng đếm là hai lệnh Redis riêng, request song song có thể vượt 5 lần một chút; muốn chính xác thì gộp vào một Lua script.
- **Giá trị policy của module `account`** (TTL OTP, cooldown gửi lại, số lần sai, bcrypt cost, username dành riêng) hiện là default lấy từ `docs/domain.md`, đặt trong `config.defaultAccount()`; chưa đọc từ biến môi trường để không tự đặt thêm tên biến. Tên exchange / routing key mail bắt buộc lấy từ `MAIL_EXCHANGE`, `MAIL_OTP_ROUTING_KEY`.
- **Port trong `command` tách theo khả năng, đặt tên có hậu tố** `Reader` / `Creator` / `Updater` / `Deleter` / `Hasher` / `Verifier` / `Issuer`… (ví dụ `AccountReader`, `SessionUpdater`). Mỗi handler chỉ nhận đúng port nó dùng; cần nhiều port do **cùng một object** thỏa thì khai báo interface ghép embed các port con (`SessionReadDeleter`, `TokenParseIssuer`), không nhận cùng object qua nhiều tham số. Vì sao: đọc chữ ký constructor là biết handler đọc / ghi gì.
- **`httpapi` gọi thẳng struct command handler (`*command.XHandler`), không khai báo interface**; field của `httpapi.Auth` export để `module` truyền command handler vào. Vì sao: chỉ có một implementation; test handler đã làm ở tầng `command`.
- **Kiểm tra định dạng chuỗi (email, username, chữ / số trong mật khẩu) bằng `regexp` biên dịch sẵn trong `domain`**; độ dài vẫn kiểm tra riêng vì RE2 không gộp được giới hạn độ dài với cấu trúc "nhóm nối bằng một dấu chấm".
- **Hash giả để chống dò tài khoản thuộc password adapter** (`password.Bcrypt` tạo lúc khởi động với đúng cost, `CompareDummy`), không nằm ở `module`.
- **Tách module `account` khỏi `user`.** `account` sở hữu Account (schema `account`, prefix Redis `account:`, migration riêng `internal/account/adapter/postgres/migrations/`): đăng ký, đăng nhập, token, mật khẩu, username, status. `user` chỉ còn Profile, Follow, Block (schema `users`); `users.profiles.user_id` = `account.accounts.id`, không có FK chéo schema. Mã rule của Account đổi thành ACCOUNT-R1…R7; USER-R8 trở đi giữ nguyên.
  - *Tạo Profile khi đăng ký (ACCOUNT-R7):* hai module không chung transaction được, nên `account` tạo Account trước rồi gọi facade `user` (`CreateProfile`, idempotent nhờ `ON CONFLICT DO NOTHING`) qua adapter `internal/account/adapter/user/`. Tạo Profile lỗi thì xóa Account vừa tạo (cần quyền `DELETE` trên `account.accounts`) và giữ phiên đăng ký để user thử lại với cùng OTP. Tên profile được kiểm tra ngay ở bước register qua `ValidateProfileName` của `user`, vì quy tắc tên thuộc Profile.
  - *Cần dev xem lại:* nếu cả bước tạo Profile và bước xóa Account đều lỗi (hoặc app tắt giữa hai bước) thì còn Account thiếu Profile; lần thử lại sẽ gặp email đã tồn tại → `REGISTRATION_FAILED`. Hiện chỉ trả lỗi kèm id account để tra log, chưa có job dọn hay tạo bù.
  - *Bản sao username trong `users.profiles`:* `user` lưu `username` cùng Profile để đọc profile, danh sách follow… không phải gọi sang `account` mỗi lần. `account` là nguồn gốc; bản sao được ghi khi tạo Profile. Bản sao **cũng có unique index** (`profiles_username_uq`). Tạo Profile gặp trùng bản sao thì `user` trả `ErrUsernameTaken`; `account` xóa Account vừa tạo, hủy phiên và trả `USERNAME_TAKEN`, giống trùng ở bảng `account.accounts`.
  - *Đổi username dùng saga (chưa làm):* đổi ở `account` và cập nhật bản sao ở `user` là hai bước ở hai module, mỗi bước một transaction riêng; bước nào lỗi thì chạy bước bù trừ để đưa cả hai về username cũ. Ví dụ: A đổi `alice` → `bob`; `account` đổi xong nhưng `user` báo `bob` còn bị giữ trong bản sao (B vừa đổi `bob` → `alice` mà chưa đồng bộ) → `account` đổi A về `alice` và báo lỗi. Thứ tự bước, cách lưu trạng thái saga và retry chốt khi làm chức năng đổi username.
  - *Phụ thuộc hai chiều:* `account` gọi `user` (tạo Profile); sau này `user` sẽ cần hỏi `account` (tồn tại, `ACTIVE`). Không bị vòng import Go vì mỗi bên chỉ import package `module` của bên kia trong adapter, wiring ở `cmd/api`.
- **Tài liệu API (OpenAPI 3.1) nằm cùng module**: `internal/<module>/adapter/httpapi/openapi.yaml`, chỉ chứa route của module đó; route hệ thống (`/healthz`) ở `cmd/api/openapi.yaml`. Sửa route, body hay mã lỗi thì sửa spec trong cùng thay đổi và chạy `npx @redocly/cli lint` trên file đó. Vì sao: spec nằm cạnh handler nên dễ giữ khớp với code.
  - **Tag chỉ dùng để gom nhóm chức năng đi chung** (ví dụ `register` cho các bước đăng ký); endpoint đứng riêng không gắn tag. Import vào Postman chọn "Folder organization: Tags" để collection phẳng: chỉ nhóm có tag mới thành thư mục, không lồng thư mục theo path (`auth › register › resend-otp`).
- **Quên mật khẩu theo cùng mẫu với đăng ký** (ACCOUNT-R1, ACCOUNT-R2): `POST /auth/password/forgot {email}` luôn trả `sessionId` (email chưa đăng ký thì phiên dummy, không gửi mail); `POST /auth/password/forgot/resend-otp`; `POST /auth/password/reset {sessionId, otp, newPassword}` kiểm tra OTP ngay trong request đổi mật khẩu. Một `command.ResendOTPHandler` phục vụ cả hai luồng, mỗi luồng một instance theo `VerificationType`; phiên của luồng kia trả `SESSION_NOT_FOUND`.
  - *Reset trả 204, không trả token* (dev chọn): client đăng nhập lại bằng mật khẩu mới. Token đã cấp vẫn sống tới khi hết hạn (ACCOUNT-R6).
  - *Không bắt mật khẩu mới khác mật khẩu cũ* (dev chọn): quy tắc "khác mật khẩu cũ" của ACCOUNT-R4 chỉ cho đổi mật khẩu khi đã đăng nhập; người quên mật khẩu không biết mật khẩu cũ, báo trùng sẽ làm lộ nó.
  - `newPassword` sai định dạng bị từ chối trước khi kiểm tra OTP nên không tốn lượt nhập OTP. Lỗi DB lúc lưu mật khẩu thì giữ phiên để thử lại với cùng OTP.
  - *Chưa làm:* xóa bộ đếm đăng nhập sai khi reset thành công (ACCOUNT-R3), chờ thiết kế giới hạn thử sai chung (xem "Để sau").
- **Session id OTP là UUID v4**, sinh bằng `github.com/google/uuid` (`uuid.NewRandom`, đọc từ `crypto/rand`) trong `command.CryptoRandom`. Dependency này dev đã đồng ý thêm.

## Để sau – không tự làm

- File schema topology RabbitMQ / Kafka (dev làm).
- Integration test và CI (dev làm khi đến bước CI).
- **Xóa bộ đếm đăng nhập sai khi đặt lại mật khẩu thành công (ACCOUNT-R3)**: chưa làm, làm cùng thiết kế giới hạn thử sai bên dưới.
- **Giới hạn số lần thử sai: vấn đề chưa giải quyết, dev sẽ thiết kế một cơ chế chung.** Nhiều nơi cần nó (đăng nhập, mật khẩu cũ khi đổi mật khẩu, OTP, gửi lại OTP…). Bản hiện tại (ACCOUNT-R3, xem "Đếm đăng nhập sai") giữ nguyên tới khi có thiết kế mới; không tự sửa hay mở rộng. Các vấn đề đã thấy, làm đầu vào cho thiết kế:
  - *Lộ email thuộc username nào:* vì email và username của một tài khoản dùng chung bộ đếm, kẻ xấu sai 5 lần với `alice@x.com` rồi thử `alice`: nhận 429 nghĩa là email đó là của `alice` (va chạm ACCOUNT-R3 với ACCOUNT-R1).
  - *Khóa tài khoản người khác:* ai biết username cũng chặn được chủ tài khoản 15 phút bằng 5 lần nhập sai, lặp lại mãi được (không có mở khóa thủ công).
  - *Thêm IP vào khóa đếm:* (identifier, IP) hết khóa người khác và hết lộ email ↔ username, nhưng mất giới hạn tổng khi kẻ xấu dùng nhiều IP; cần thêm lớp đếm theo identifier và / hoặc theo IP. Phải lấy IP thật sau nginx (`SetTrustedProxies`, không tin `X-Forwarded-For` từ client), gộp IPv6 theo `/64`, cân nhắc nhiều người chung IP (NAT).
  - *Định danh thiết bị:* id do client gửi lên không tin được (đổi tùy ý); cách dùng được là device cookie do server ký cấp sau khi đăng nhập thành công, thiết bị có cookie hợp lệ không bị chặn theo lớp identifier.
  - *Cơ chế hiện tại:* cửa sổ cố định tính từ lần sai đầu (sai lần 5 ở phút 14 chỉ bị chặn ~1 phút); kiểm tra chặn và tăng đếm là hai lệnh Redis riêng nên request song song có thể vượt ngưỡng một chút.
