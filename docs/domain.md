# Moments – Mô tả domain

Chỉ mô tả nghiệp vụ: module, entity, rule. Usecase không mô tả ở đây, sẽ thiết kế khi làm. Không mô tả hạ tầng, lưu trữ hay giao thức. Giá trị cụ thể (hạn dùng, số lần tối đa…) ghi ở đây là quy tắc nghiệp vụ mặc định và có thể cấu hình.

**Quy ước chung**
- Người dùng luôn được định danh bằng `userId`, không bằng `username` (username đổi được). `userId` của người thao tác luôn lấy từ phiên đăng nhập, không bao giờ từ dữ liệu client gửi lên.
- Thao tác idempotent: đưa dữ liệu về trạng thái đã có sẵn thì báo thành công và không thay đổi gì (follow người đã follow, unfollow người chưa follow, like post đã like…).
- Số đếm hiển thị (follower, following, like, comment, repost) luôn khớp dữ liệu nguồn, không âm, và chỉ thay đổi khi dữ liệu nguồn thực sự thay đổi. 
- Danh sách phân trang theo thời gian, mới nhất trước; không lặp và không sót bản ghi khi dữ liệu thay đổi giữa hai lần xem.
- Đối tượng "không tồn tại hoặc không được phép xem" luôn báo **không tìm thấy**, để người ngoài không phân biệt được hai trường hợp.

**Các module:** `account`, `user`, `content`, `interaction`. Module không đọc dữ liệu của nhau; cần gì thì hỏi module sở hữu. Phụ thuộc: `account` yêu cầu `user` tạo Profile khi đăng ký; `content` hỏi `user`; `interaction` hỏi `content` và `user`.

---

# Module: account

Tài khoản đăng nhập: đăng ký, đăng nhập, token, mật khẩu, username, trạng thái.

## Entity

### Account
- `id`: số; duy nhất; bất biến
- `email`: đúng định dạng email; duy nhất (không phân biệt hoa thường); bất biến
- `username`: duy nhất (không phân biệt hoa thường); theo Rule username
- `passwordHash`: không rỗng; chỉ lưu hash, không lưu mật khẩu gốc
- `status`: `ACTIVE`; hiện chỉ có giá trị này, giữ để mở rộng
- `createdAt`: bất biến
- `updatedAt`

### VerificationSession (phiên xác thực OTP)
- `id`: duy nhất; trả cho client làm `session_id`
- `type`: `REGISTER` | `FORGOT_PASSWORD`; bất biến
- `email`: bất biến
- `otp`: 6 chữ số, sinh ngẫu nhiên an toàn; chỉ lưu hash
- `username`, `name`, `passwordHash`: chỉ có ở phiên `REGISTER`; dữ liệu user gửi lúc đăng ký, đã kiểm tra định dạng; mật khẩu chỉ lưu hash; bất biến
- không có trạng thái: OTP được kiểm tra ngay trong thao tác cuối của luồng (tạo user khi đăng ký), không có bước "đã xác thực OTP" riêng
- phiên bị xóa khi luồng hoàn tất, khi sai OTP đủ số lần, hoặc khi hết hạn
- `failedAttempts`: ≥ 0; số lần nhập sai OTP; đạt 5 thì xóa phiên
- `lastSentAt`: thời điểm gửi OTP gần nhất, dùng để tạo khoảng cách giữa các lần yêu cầu gửi lại
- `isDummy`: bất biến; `true` khi email không hợp lệ với luồng (xem Rule không lộ tài khoản)
- hạn dùng: 10 phút kể từ lúc tạo; không gia hạn khi gửi lại OTP

## Rule

**ACCOUNT-R1 – Không lộ sự tồn tại của tài khoản.** Email đã / chưa tồn tại, sai username, sai mật khẩu đều cho cùng một phản hồi chung chung. Bước gửi email (đăng ký, quên mật khẩu) luôn trả `session_id`; khi email không hợp lệ với luồng (đã tồn tại khi đăng ký, chưa tồn tại khi quên mật khẩu) thì tạo phiên dummy: không gửi OTP, mọi OTP nhập vào đều bị coi là sai. Thời gian phản hồi hai nhánh không được chênh lệch đáng kể (đăng nhập sai user vẫn so sánh với hash giả).

**ACCOUNT-R2 – OTP.**
- OTP có hạn theo phiên (10 phút); mỗi phiên chỉ dùng một lần.
- OTP được kiểm tra trong chính request thực hiện thao tác cuối của luồng (đăng ký: OTP đúng thì tạo Account + Profile ngay), để chỉ người sở hữu email mới hoàn tất được luồng; `session_id` một mình không đủ.
- Gửi lại OTP cách nhau tối thiểu 60 giây; OTP mới vô hiệu OTP cũ; không gia hạn phiên.
- Nhập sai OTP tối đa 5 lần mỗi phiên, vượt thì xóa phiên và bắt đầu lại.
- Gửi email OTP là bất đồng bộ: hệ thống không chờ mail đi. Mail chưa gửi mà OTP đã hết hạn thì không gửi nữa. Gửi thất bại không làm đổi phản hồi (theo ACCOUNT-R1); người dùng dùng chức năng gửi lại.

**ACCOUNT-R3 – Giới hạn thử sai.** Đăng nhập sai tối đa 5 lần trong 15 phút cho mỗi identifier (tính chung cho 1 cặp email username). Nhập sai mật khẩu cũ khi đổi mật khẩu cùng giới hạn đó cho mỗi user. Vượt giới hạn thì bị chặn tới khi hết thời gian, không có mở khóa thủ công, và phản hồi cho biết còn bao lâu. Đăng nhập thành công không xóa bộ đếm; đặt lại mật khẩu thành công thì xóa bộ đếm đăng nhập của tài khoản. *(Chưa chốt: quy tắc này sẽ được thiết kế lại thành cơ chế giới hạn thử sai dùng chung; vấn đề đã biết ghi ở mục "Để sau" của CLAUDE.md.)*

**ACCOUNT-R4 – Mật khẩu.** 6–64 ký tự, có ít nhất một chữ và một số. Đổi mật khẩu thì mật khẩu mới phải khác mật khẩu cũ. Đặt lại mật khẩu qua quên mật khẩu (xác thực bằng OTP) thì không áp dụng điều kiện này: người dùng không biết mật khẩu cũ, và báo trùng sẽ làm lộ nó.

**ACCOUNT-R5 – Username.** Chỉ gồm `a-z`, `0-9`, `.`, `_`; dài 3–30; lưu chữ thường; không bắt đầu hoặc kết thúc bằng `.`; không có hai dấu `.` liền nhau; không chứa `@`; không thuộc danh sách tên dành riêng (`admin`, `moments`, `support`, `api`…). Đổi username giải phóng username cũ ngay; sau khi đổi, `account` yêu cầu `user` cập nhật bản sao username trong Profile. Hai bước này ràng buộc theo saga: bước sau thất bại thì bù trừ, đưa username ở cả hai module về như trước khi đổi. Username trùng thì báo cụ thể (username là thông tin công khai).

**ACCOUNT-R6 – Phiên đăng nhập và token.**
- Token truy cập và token làm mới chứa `userId`, `username`, `status`. Hệ thống chỉ dùng `userId` để định danh và phân quyền; `username` chỉ để hiển thị.
- Token truy cập sống 15 phút, token làm mới sống 30 ngày. Làm mới token lấy `username`, `status` hiện tại, không sao chép từ token cũ.
- Đổi username hoặc status thì trả cặp token mới.
- Chỉ user `ACTIVE` được dùng chức năng cần đăng nhập. Token cũ có thể mang thông tin cũ tới khi hết hạn; chấp nhận độ trễ này.
- Đăng xuất chỉ xảy ra ở client; token đã cấp vẫn hợp lệ tới khi hết hạn.

**ACCOUNT-R7 – Tạo tài khoản.** Mỗi Account có đúng một Profile (thuộc module `user`). Khi đăng ký, Account được tạo trước, ngay sau đó `account` yêu cầu `user` tạo Profile; yêu cầu này idempotent (gọi lại không tạo thêm). Tạo Profile thất bại thì xóa Account vừa tạo và giữ phiên, user thử lại với cùng OTP. Trùng email khi hoàn tất đăng ký (hai luồng đua nhau) thì báo lỗi chung và hủy phiên. Username không được giữ chỗ trong lúc chờ OTP; nếu bị người khác lấy trước khi hoàn tất thì báo username trùng và hủy phiên, user đăng ký lại.

## Việc `account` cung cấp cho module khác
- **Thông tin tài khoản của một user:** có tồn tại không, có `ACTIVE` không, `username`.

---

# Module: user

Hồ sơ (Profile) và quan hệ follow / block giữa các user. Tài khoản đăng nhập thuộc module `account`.

## Entity

### Profile
- `userId`: = id của Account (module `account`); bất biến; mỗi Account đúng một Profile, do `account` yêu cầu tạo ngay sau Account (ACCOUNT-R7)
- `username`: bản sao `username` của Account, duy nhất, để đọc profile không phải hỏi `account`; `account` là nguồn gốc, đổi username thì `account` yêu cầu `user` cập nhật bản sao theo saga (ACCOUNT-R5)
- `name`: 1–50 ký tự sau khi trim; do user nhập khi đăng ký (bắt buộc)
- `bio`: tối đa 150 ký tự; chuỗi rỗng thì lưu `null`
- `avatar`: ảnh đại diện đã xử lý; `null` khi chưa có; luôn vuông 1:1
- `avatarThumb`: bản thu nhỏ của avatar; luôn có / không có cùng lúc với `avatar`; chỉ thay đổi bởi kết quả xử lý avatar, không sửa qua cập nhật profile
- `followerCount`: ≥ 0; mặc định 0; số người đang follow user (số Follow có `followeeId = userId`)
- `followingCount`: ≥ 0; mặc định 0; số người user đang follow (số Follow có `followerId = userId`)
- `postCount`: ≥ 0; mặc định 0; số post `PUBLISHED` của user (không tính `PENDING`, `FAILED`, `DELETED`)
- `createdAt`: bất biến
- `updatedAt`

### Follow
- `followerId`: người đi follow; bất biến
- `followeeId`: người được follow; bất biến; ≠ `followerId`
- `createdAt`: bất biến
- mỗi cặp (`followerId`, `followeeId`) tối đa một bản ghi
- không tồn tại khi giữa hai user có Block theo bất kỳ chiều nào

### Block
- `blockerId`: người chặn; bất biến
- `blockedId`: người bị chặn; bất biến; ≠ `blockerId`
- `createdAt`: bất biến
- mỗi cặp (`blockerId`, `blockedId`) tối đa một bản ghi
- hai chiều độc lập: A block B và B block A có thể cùng tồn tại; mỗi người chỉ gỡ block của chính mình

## Rule

**USER-R8 – Avatar.** Ảnh jpeg / png / webp, tối đa 5 MB, phải đúng tỉ lệ 1:1 (cho phép sai số 1%). Hệ thống tự kiểm tra ảnh thật, không tin thông tin client khai báo. Upload avatar mới ghi đè avatar đang chờ xử lý. Xử lý lỗi thì giữ avatar cũ; không có retry, user upload lại.

**USER-R9 – Xem profile.** Profile không tồn tại, không `ACTIVE`, hoặc có Block giữa người xem và chủ profile theo bất kỳ chiều nào thì báo không tìm thấy. Không bao giờ trả `email`.

**USER-R10 – Block chặn mọi thao tác follow.** Giữa người thao tác và đối tượng có Block theo bất kỳ chiều nào thì follow, unfollow, remove follower bị từ chối bằng "không tìm thấy".

**USER-R11 – Block xóa mọi quan hệ follow.** A block B xóa cả Follow(A→B) và Follow(B→A) trong cùng một lần xử lý. Unblock không khôi phục Follow đã xóa.

**USER-R12 – Nhất quán theo cặp user.** Các thao tác ghi về quan hệ giữa cùng một cặp user (follow, unfollow, remove follower, block, unblock) được xử lý tuần tự, để không bao giờ tồn tại Follow trong khi có Block.

**USER-R13 – Thao tác quan hệ.** Thao tác lên chính mình thì báo dữ liệu không hợp lệ. Đối tượng không tồn tại hoặc không `ACTIVE` thì báo không tìm thấy (riêng unblock vẫn cho gỡ khi đối tượng không còn `ACTIVE`, để dọn dữ liệu). Mọi thao tác idempotent. Đối phương không nhận thông báo khi bị block/unlock hoặc bị remove follower/unfollow.

**USER-R14 – Số đếm follow.** Xử lý theo batch sau

**USER-R15 – Danh sách quan hệ.** Danh sách followers / following: đối tượng khác người xem thì phải tồn tại, `ACTIVE` và không có Block với người xem. Bỏ khỏi danh sách các user không `ACTIVE` và các user có Block với người xem, lọc ngay khi lấy từng trang để trang không bị thiếu. Danh sách đã block chỉ của chính mình, không lọc user không `ACTIVE` để vẫn gỡ block được.

## Việc `user` cung cấp cho module khác
- **Tạo Profile cho Account mới** (ACCOUNT-R7), kèm bản sao `username`: idempotent; kiểm tra tên profile hợp lệ trước khi tạo.
- **Cập nhật bản sao `username`** khi `account` đổi username (chưa làm, cùng lúc với chức năng đổi username).
- **Cập nhật `postCount`:** `content` yêu cầu tăng / giảm `postCount` của tác giả khi post của họ được publish hoặc bị xóa.
- **Thông tin tóm tắt của một user:** có tồn tại không, có `ACTIVE` không, `username`, `name`, `avatarThumb`. `username` lấy từ bản sao trong Profile; tồn tại và `ACTIVE` do `user` hỏi lại `account` (chưa làm).
- **Quan hệ giữa hai user:** người xem có follow đối tượng không; có Block giữa hai bên theo bất kỳ chiều nào không.

---

# Module: content

Post, media của post, bài hát.

## Entity

### Post
- `id`: số; duy nhất; bất biến
- `authorId`: `userId` của tác giả; bất biến
- `caption`: nullable; tối đa 2200 ký tự sau khi trim; rỗng thì `null`
- `media`: danh sách Media có thứ tự; 1–20 phần tử; trộn được ảnh và video; bất biến sau khi tạo (không thêm, bớt, đổi thứ tự)
- `song`: PostSong; nullable; bất biến sau khi tạo
- `aspectRatio`: `SQUARE` (1:1) | `PORTRAIT` (4:5) | `LANDSCAPE` (16:9); bắt buộc khi tạo, không có mặc định; bất biến
- `cover`: ảnh bìa hiển thị dạng lưới, tỉ lệ 3:4 dọc; nullable; lấy từ media đầu tiên (với video là một khung hình đại diện); chỉ có khi media đầu tiên đã `READY`; gán một lần, sau đó không đổi
- `visibility`: `PUBLIC` | `FOLLOWERS` | `PRIVATE`; mặc định `PUBLIC`
- `commentEnabled`: true / false; mặc định `true`
- `status`: `PENDING` | `PUBLISHED` | `FAILED` | `DELETED` (xem Rule trạng thái)
- `retryable`: true / false; mặc định `true`; chỉ chuyển một chiều `true → false`
- `likeCount`: ≥ 0; mặc định 0; số Like của post
- `commentCount`: ≥ 0; mặc định 0; số Comment `ACTIVE` của post
- `repostCount`: ≥ 0; mặc định 0; số Repost của post
- `createdAt`: bất biến
- `publishedAt`: nullable; gán khi chuyển `PUBLISHED`
- `deletedAt`: nullable; gán khi chuyển `DELETED`
- `updatedAt`
- Ba số đếm `likeCount`, `commentCount`, `repostCount` do module `interaction` yêu cầu cập nhật (INTERACTION-R7); `content` chỉ đọc.

### Media (thuộc Post)
- `id`: số; duy nhất; bất biến
- `postId`: bất biến
- `position`: ≥ 0; thứ tự hiển thị, bắt đầu từ 0; bất biến
- `type`: `IMAGE` | `VIDEO`; bất biến. Ảnh: jpeg / png / webp, tối đa 10 MB. Video: mp4 / mov, tối đa 200 MB, dài tối đa 90 giây
- `rawKey`: định danh file upload (file gốc); bất biến; thường là key của file trên object storage như "post/anh.jpeg"
- `fileKey`: định danh file đã xử lý; nullable; chỉ có khi `READY`
- `status`: `PENDING` | `READY` | `FAILED`
- `attempt`: ≥ 1; lần xử lý hiện tại; tăng 1 mỗi lần retry; tối đa 4 (tức tối đa 3 lần retry)
- `dispatchedAt`: nullable; thời điểm gửi yêu cầu xử lý gần nhất.
- `metadata`: nullable; chỉ có khi `READY`; gồm kích thước (rộng, cao > 0), thời lượng (bắt buộc với video), dung lượng > 0
- `failureCode`: nullable; chỉ có khi `FAILED`; `DISPATCH_FAILED` (không gửi đi xử lý được, retry được) | `PROCESSING_ERROR` (lỗi tạm thời, retry được) | `INVALID_MEDIA` (file hỏng hoặc vi phạm giới hạn, **không** retry)

### PostSong
- `songId`: tham chiếu Song
- `startMs`: ≥ 0
- `endMs`: > `startMs`; ≤ độ dài bài hát
- độ dài đoạn (`endMs − startMs`): 5–90 giây
- Song bị xóa sau đó thì PostSong của các post dùng nó bị xóa theo; post giữ nguyên và coi như không có nhạc

### Song (dữ liệu dùng chung)
- `id`: số; duy nhất; bất biến
- `name`: không rỗng
- `artist`: không rỗng
- `url`: địa chỉ file nhạc; không rỗng
- `durationMs`: > 0
- `createdAt`: bất biến
- Danh mục nạp sẵn, user chỉ đọc và chọn; thêm / sửa / xóa Song: chưa thiết kế. Xóa Song không bị chặn bởi post đang dùng.

### UploadSession (phiên soạn post)
- `uploadId`: duy nhất; trả cho client
- `items`: danh sách file đã được cấp quyền upload (`rawKey`, loại, content type); chỉ thêm, không xóa; tối đa 40
- hạn dùng: 30 phút từ lúc tạo; không gia hạn khi cấp thêm quyền
- một user có thể có nhiều phiên cùng lúc

## Rule

**CONTENT-R1 – Trạng thái Post.**

| Từ → Tới | Khi nào |
|---|---|
| `PENDING → PUBLISHED` | Media cuối cùng chuyển `READY` |
| `PENDING → FAILED` | Một media chuyển `FAILED` |
| `FAILED → PENDING` | Tác giả retry thành công |
| `FAILED → PUBLISHED` | Không còn media `FAILED` và mọi media `READY` (kết quả thành công đến muộn) |
| `PENDING / PUBLISHED / FAILED → DELETED` | Tác giả xóa (xóa mềm) |

Trạng thái Media: `PENDING → READY`, `PENDING → FAILED`, `FAILED → READY` (kết quả thành công của lần xử lý trước đến muộn), `FAILED → PENDING` (retry); `READY` là trạng thái cuối. `PUBLISHED` thì mọi media phải `READY`. Post và toàn bộ Media được tạo cùng một lúc.

**CONTENT-R2 – Publish khi mọi media sẵn sàng.** Post `PENDING` / `FAILED` chuyển `PUBLISHED` ngay khi mọi media đều `READY`, cùng lúc với việc cập nhật media cuối cùng; gán `publishedAt`. Khi post chuyển sang `PUBLISHED`, `postCount` của tác giả tăng 1 (do module `user` cập nhật).

**CONTENT-R3 – Post lỗi khi có media lỗi.** Một media chuyển `FAILED` thì post `PENDING` chuyển `FAILED` ngay, dù các media khác còn `PENDING` (chúng vẫn tiếp tục xử lý). Post `FAILED` không tự xóa; tác giả retry hoặc xóa.

**CONTENT-R4 – Ảnh bìa.** Chỉ media đầu tiên (`position = 0`) có ảnh bìa. Ảnh bìa gán cùng lúc với việc media đó `READY`, kể cả khi kết quả đến muộn; media đầu chưa `READY` thì `cover = null` (client hiển thị ảnh chờ).

**CONTENT-R5 – Quyền xem post của một tác giả** (hai lớp).
- Lớp 1, quan hệ (khi người xem không phải tác giả): tác giả phải tồn tại và `ACTIVE`, và không có Block giữa hai bên theo bất kỳ chiều nào; sai thì không tìm thấy cho toàn bộ yêu cầu.
- Lớp 2, từng post:

| Người xem | `status` thấy được | `visibility` thấy được |
|---|---|---|
| Chính tác giả | `PENDING`, `PUBLISHED`, `FAILED` | Mọi giá trị |
| Người đang follow tác giả | `PUBLISHED` | `PUBLIC`, `FOLLOWERS` |
| Người khác | `PUBLISHED` | `PUBLIC` |

- `DELETED` không ai thấy (mọi thao tác coi như không tồn tại). `PRIVATE` chỉ tác giả thấy.
- Tác giả không có post nào người xem được thấy thì trả danh sách rỗng, không phải "không tìm thấy".

**CONTENT-R6 – Nhất quán theo post.** Áp kết quả media, cập nhật, xóa, retry trên cùng một post được xử lý tuần tự, để hai media cuối cùng cùng `READY` đồng thời không làm post không bao giờ được publish.

**CONTENT-R7 – Chỉ tác giả được sửa, xóa, retry.** Người khác nhận "không tìm thấy". Sau khi tạo, tác giả chỉ sửa được `caption`, `visibility`, `commentEnabled`. `commentEnabled = false` chỉ chặn bình luận mới; bình luận đã có vẫn hiển thị.

**CONTENT-R8 – Điều kiện retry.**
- Media **cần retry** khi: `FAILED` với `DISPATCH_FAILED` hoặc `PROCESSING_ERROR`; hoặc `PENDING` mà chưa từng gửi đi xử lý, hoặc đã gửi quá 5 phút mà chưa có kết quả (coi như bị mất).
- Post **không retry được** (`retryable = false`) khi: có media `INVALID_MEDIA`; hoặc file gốc của media cần retry không còn (file gốc chỉ được giữ 1 ngày, nên post `FAILED` quá 1 ngày thường không retry được); hoặc media cần retry đã hết lượt (`attempt` tối đa). Đã `false` thì không bao giờ trở lại `true`; khi đó tác giả chỉ còn cách xóa post.
- Retry: media cần retry quay về `PENDING`, `attempt + 1`, xóa `dispatchedAt` và `failureCode`; post về `PENDING`; gửi xử lý lại.

**CONTENT-R9 – Media đúng tỉ lệ của post.** Mọi media phải khớp `aspectRatio` của post (cho phép sai số 1%, so theo kích thước hiển thị tức đã tính hướng xoay). Client cắt đúng tỉ lệ trước khi upload, nhưng hệ thống **không tin client**: tự đo file thật trước khi tạo post. Sai tỉ lệ hoặc không đọc được kích thước thì báo dữ liệu không hợp lệ kèm file lỗi, không tạo post. Hệ thống xử lý media kiểm tra lại; sai thì trả `INVALID_MEDIA` và post `FAILED` không retry được.

**CONTENT-R10 – Kết quả xử lý lặp lại hoặc đến muộn.** Một media có thể nhận nhiều kết quả (nhiều lần xử lý, hoặc gửi lặp):
- Media đã `READY`: bỏ qua mọi kết quả sau.
- Kết quả thành công của **bất kỳ** lần xử lý nào: chấp nhận nếu media chưa `READY`.
- Kết quả lỗi của lần xử lý khác lần hiện tại, hoặc media đã `FAILED`: bỏ qua.
- Media, file hoặc post không khớp, hoặc post đã `DELETED`: bỏ qua.
- Nhận lại cùng một kết quả cho cùng một kết quả cuối.
- Kết quả lỗi `INVALID_MEDIA` thì đặt post `retryable = false`.

**CONTENT-R11 – Xóa post.** Xóa mềm: chuyển `DELETED`, gán `deletedAt`; cho phép từ `PENDING`, `PUBLISHED`, `FAILED`; kết quả media đến sau bị bỏ qua. Xóa post đang `PUBLISHED` thì `postCount` của tác giả giảm 1. Dọn dẹp dữ liệu và file của post đã xóa: chưa thiết kế.

**CONTENT-R12 – Phiên soạn post.** Chỉ file nằm trong `items` của phiên mới được dùng để tạo post; danh sách file khi tạo post không trùng lặp và đúng thứ tự hiển thị mong muốn(tức là các file phải được xác nhận đúng session và thứ tự theo thứ tự mà client gửi lên). File xin quyền nhưng không dùng thì bỏ.

**CONTENT-R13 – Hiển thị.** Thông tin tác giả (`username`, `name`, `avatarThumb`) lấy từ module `user`, không lưu ở post. Số like / comment / repost là field của Post; việc người xem đã like / repost chưa (`isLiked`, `isReposted`) lấy từ module `interaction`. Chỉ tác giả mới thấy `retryable` và `failureCode`.

## Việc `content` cung cấp cho module khác
- **Cập nhật số đếm của post:** `interaction` yêu cầu tăng / giảm `likeCount`, `commentCount`, `repostCount` của một post khi dữ liệu tương tác thực sự thay đổi.
- **Khả năng tương tác với một post của một user:** post có tồn tại, đã `PUBLISHED`, người xem có xem được không (theo R5 đầy đủ cả quan hệ và `visibility`), `authorId`, `visibility`, `commentEnabled`.

---

# Module: interaction

Like, comment, repost trên post. Module sở hữu toàn bộ dữ liệu tương tác của post (like / comment / repost).

## Entity

### Like
- `userId`: người like; bất biến
- `postId`: bất biến
- `createdAt`: bất biến
- mỗi cặp (`userId`, `postId`) tối đa một bản ghi

### Comment
- `id`: số; duy nhất; bất biến
- `postId`: bất biến
- `authorId`: người viết; bất biến
- `content`: 1–1000 ký tự sau khi trim; không được rỗng; bất biến (không có sửa comment)
- `status`: `ACTIVE` | `DELETED`
- `createdAt`: bất biến
- `deletedAt`: nullable; gán khi `DELETED`
- Comment phẳng, mỗi comment thuộc trực tiếp một post; trả lời lồng nhau: chưa làm

### Repost
- `userId`: người repost; bất biến
- `postId`: bất biến
- `createdAt`: bất biến
- mỗi cặp (`userId`, `postId`) tối đa một bản ghi
- Repost là chia sẻ lại nguyên post gốc lên trang cá nhân của mình, không kèm nội dung riêng

## Rule

**INTERACTION-R1 – Chỉ tương tác với post xem được.** Like, comment, repost, và xem comment của một post chỉ khi post tồn tại, đã `PUBLISHED` và người thao tác xem được post đó (theo CONTENT-R5: tác giả `ACTIVE`, không có Block, đúng `visibility`). Ngược lại báo không tìm thấy. Riêng **unlike**, **xóa comment của mình** và **unrepost** vẫn cho phép khi post không còn xem được (trừ khi post `DELETED`), để người dùng dọn dữ liệu của mình.

**INTERACTION-R2 – Like.** Mỗi user like một post tối đa một lần. Like lại post đã like và unlike post chưa like đều idempotent. Tác giả được like post của chính mình.

**INTERACTION-R3 – Comment.**
- Chỉ comment khi post có `commentEnabled = true`; ngược lại báo không được bình luận.
- Nội dung 1–1000 ký tự sau trim.
- Comment mới luôn `ACTIVE`. Không sửa comment.
- Người được xóa comment: chính người viết, hoặc tác giả của post. Người khác nhận không tìm thấy. Xóa là xóa mềm; xóa comment đã `DELETED` thì idempotent.
- Tắt `commentEnabled` chỉ chặn comment mới; comment đã có vẫn hiển thị và vẫn xóa được.

**INTERACTION-R4 – Repost.**
- Chỉ repost post `PUBLISHED` với `visibility = PUBLIC`; post `FOLLOWERS` hoặc `PRIVATE` không được repost (báo không được repost).
- Không repost post của chính mình.
- Mỗi user repost một post tối đa một lần. Repost lại và unrepost post chưa repost đều idempotent.
- Repost hiển thị cho người xem khi người xem xem được post gốc; post gốc không còn xem được với người xem (đổi `visibility`, bị xóa, hoặc có Block) thì repost đó ẩn khỏi người xem, nhưng vẫn tính trong số đếm.

**INTERACTION-R5 – Block.** Danh sách comment ẩn comment của user có Block với người xem theo bất kỳ chiều nào. Block không xóa Like, Comment, Repost đã có và không làm đổi số đếm. Số đếm là số tổng, không tính riêng theo người xem.

**INTERACTION-R6 – Post bị xóa.** Post `DELETED` thì mọi tương tác trên post đó coi như không tồn tại (không tìm thấy), kể cả unlike / unrepost / xóa comment. Dọn dữ liệu tương tác của post đã xóa: chưa thiết kế.

**INTERACTION-R7 – Số đếm.** Số like / comment / repost nằm trong Post (module `content`), `interaction` yêu cầu `content` cập nhật. Thêm Like / Repost / Comment thực sự mới thì `+1`; xóa Like / Repost thực sự có, hoặc xóa mềm Comment thực sự `ACTIVE` thì `−1`; thao tác idempotent không đụng số đếm. Hai thao tác đồng thời của cùng một user trên cùng một post không làm lệch số đếm.

**INTERACTION-R8 – User không còn `ACTIVE`.** Chỉ user `ACTIVE` được tương tác. Comment của user không còn `ACTIVE` ẩn khỏi danh sách comment.

## Việc `interaction` cung cấp cho module khác
- **Viewer đã like / repost post chưa** (`isLiked`, `isReposted`), dùng khi hiển thị post cho người xem.

## Chưa thiết kế
Like cho comment, trả lời comment lồng nhau, sửa comment, repost kèm nội dung riêng, danh sách người đã like, thông báo, dọn dữ liệu tương tác của post đã xóa, feed.
