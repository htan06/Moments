# Post Domain

## 0. Unique Languages

* **`Post`**: Một bài viết của người dùng
* **`Media`**: Ảnh hoặc video của bài viết
* **`Mention`**: Người được nhắc đến trong comment, bài viết. Ví dụ: `@username`
* **`Hashtag`**: Tag được gắn vào content của bài viết. Ví dụ: `#tag`
* **`Comment`**: Bình luận của 1 người dùng trong bài viết.
* **`Repost`**: Bài đăng lại của một người ( người đăng lại không sở hữu bài viết này).

## 1. Entities

### `Post` (Aggregate Root)

* **Thuộc tính**
  * `id`: Number, ≥ 0, sequence, immutable.
  * `authorId`: Number, ≥ 0, immutable. Tham chiếu User context theo ID.
  * `content`: Content. Được phép null.
  * `visibility`: Enum:Visibility.
  * `aspectRatio`: Enum:AspectRatio, immutable.
  * `thumbnailKey`: String, not empty.
  * `likeCount`: Number, ≥ 0.
  * `commentCount`: Number, ≥ 0.
  * `repostCount`: Number, ≥ 0.
  * `status`: Enum:PostStatus.
  * `createdAt`: UtcTime, immutable.
  * `updatedAt`: UtcTime.
* **Invariants**
  * `INV-POST-01`: `displayOrder` của các Media trong cùng Post là duy nhất.
  * `INV-POST-02`: `status = PUBLISHED` chỉ khi mọi Media của post có `status = PUBLISHED`.
* **Hành vi**: 
  * `Create`
  * `UpdateContent`
  * `UpdateVisibility` 
  * `MarkMediaPublished` 
  * `Delete`

### `Media` (Child Entity của Post)

* **Thuộc tính**
  * `mediaKey`: String, not empty, immutable. Ví dụ `uuid.jpeg`, `uuid/index.m3u8`.
  * `mediaType`: Enum:MediaType, immutable.
  * `displayOrder`: Number, > 0.
  * `status`: Enum:MediaStatus.
  * `width`: Number, > 0, optional khi `status = PROCESSING`.
  * `height`: Number, > 0, optional khi `status = PROCESSING`.
  * `createdAt`: UtcTime, immutable.
  * `updatedAt`: UtcTime.
* **Invariants**
  * `INV-MEDIA-01`: `width` và `height` bắt buộc có giá trị khi `status = PUBLISHED`.

### `Session` (Aggregate Root)

Phiên tạo post, quản lý các presigned upload URL.

* **Thuộc tính**
  * `id`: UUID, immutable.
  * `userId`: Number, immutable. Sẽ là `authorId` của Post được tạo.
  * `imageKeys`: List<String>. Ví dụ `uuid.jpeg`.
  * `videoKeys`: List<String>. Ví dụ `uuid/index.m3u8`.
  * `expiresAt`: UtcTime.
* **Invariants**
  * `INV-SESSION-01`: thời hạn presigned URL của mỗi key phải sớm hơn `expiresAt`.
  * `INV-SESSION-02`: `imageKeys.size + videoKeys.size ≤ 20`.

### `Like` (Aggregate Root)

* **Thuộc tính**
  * `userId`: Number, ≥ 0, immutable.
  * `postId`: Number, ≥ 0, immutable.
  * `createdAt`: UtcTime, immutable.
* **Invariants**
  * `INV-LIKE-01`: mỗi cặp `(userId, postId)` có tối đa một bản ghi.

### `Comment` (Aggregate Root)

* **Thuộc tính**
  * `userId`: Number, ≥ 0, immutable.
  * `postId`: Number, ≥ 0, immutable.
  * `content`: Content, không rỗng, không chứa node `hashtag`.
  * `createdAt`: UtcTime, immutable.
* **Invariants**
  * `INV-COMMENT-01`: Cặp (`userId và postId`) có tối đa 1 bản ghi.
  * `INV-COMMENT-02`: cây `content` không chứa node có `type = hashtag`.
  * Xem OQ-04 về khóa chính.

### `Repost` (Aggregate Root)
* **Thuộc tính**
  * `userId`: Number, ≥ 0, immutable.
  * `postId`: Number, ≥ 0, immutable.
  * `createdAt`: UtcTime, immutable.
* **Invariants**
  * `INV-REPOST-01`: Cặp (`userId và postId`) có tối đa 1 bản ghi.
---

## 2. Value Objects

### `Content` (cây AST, bất biến)

* **Thuộc tính**
  * `type`: Enum:ContentNodeType.
  * `text`: String.
  * `attributes`: Map<String,String>, optional.
    * `mention` bắt buộc có `userId`.
    * `hashtag` bắt buộc có `hashtagId`.
  * `children`: List<Content>, optional. Chỉ `paragraph` mới có.
* **Ví dụ**

```json
{
  "type": "paragraph",
  "children": [
    { "type": "text", "text": "Hello " },
    { "type": "mention", "text": "@username", "attributes": { "userId": "12" } },
    { "type": "text", "text": " " },
    { "type": "hashtag", "text": "#hashtag", "attributes": { "hashtagId": "12" } }
  ]
}
```

---

## 3. Enums

### Domain enum

* **`Visibility`** (phân loại, kèm tỉ lệ width:height)
  * `PUBLIC`: Ai cũng có thể xem
  * `PRIVATE`: Chỉ có bản thân mới có thể xem
* **`AspectRatio`** (phân loại, kèm tỉ lệ width:height)
  * `SQUARE`: 1:1
  * `CLASSIC`: 3:2
  * `STANDARD`: 4:3
  * `PORTRAIT`: 3:4
  * `WIDESCREEN`: 16:9
* **`PostStatus`** (trạng thái)
  * `PENDING`: đã tạo, còn media đang xử lý, chưa hiển thị. Chuyển tới `PUBLISHED`.
  * `PUBLISHED`: hiển thị theo BR-01. Trạng thái cuối.
* **`MediaType`** (phân loại)
  * `IMAGE`: ảnh tĩnh, key dạng `uuid.jpeg`.
  * `VIDEO`: video HLS, key dạng `uuid/index.m3u8`.
* **`MediaStatus`** (trạng thái)
  * `PROCESSING`: đã upload, đang xử lý. Chuyển tới `PUBLISHED`.
  * `PUBLISHED`: sẵn sàng phát. Trạng thái cuối.
* **`ContentNodeType`** (giá trị chữ thường trong JSON, in hoa khi là enum)
  * `paragraph`, `text`, `mention`, `hashtag`.

### Reference Enum from User domain

* **`Relationship`** (Đọc theo chiều từ trái sang phải. Ví dụ quan hệ giữa A và B là FOLLOWING, có nghĩa là A đang FOLLOWING B)
  * `FOLLOWING`: Người đang theo dõi.
  * `FOLLOWER`: Người theo dõi của bản thân.
  * `NONE`: Không có quan hệ.
  * `BLOCKING`: Người đã bị chặn.
  * `BLOCKED_BY`: Đã bị người này chặn.

## 4. Business rules (liên entity)

### BR-01: Quyền xem và tương tác với các tài nguyên của người khác

* **Áp dụng cho:** Post, Like, Comment, Repost.
* **Rule:** Các thao tác liên quan đến các entity trên phải kiểm tra:
  * **1. Visibility:** `Kiểm tra visibility của post là public hay private`.
  * **2. Quan hệ giữa 2 user:** `Kiểm tra quan hệ giữa 2 user`.
  * **3. Status của Post:** `Status phải là PUBLISH`.
* **Ví dụ:** *Post để chế độ public nhưng 2 user có mối quan hệ chặn -> `Không hiển thị`*

## 5. Domain service

* **LikePost:**
  * **Input:** `userId`, `postId`, `quan hệ giữa 2 user`
  * **Logic:** Kiểm tra BR-01, thực hiện tạo entity nếu thỏa mãn
  * **Output:** `like entity` hoặc `error`

* **Comment:**
  * **Input:** `userId`, `postId`, `quan hệ giữa 2 user`, `comment content`
  * **Logic:** Kiểm tra BR-01, thực hiện tạo entity nếu thỏa mãn
  * **Output:** `comment entity` hoặc `error`

* **Repost:**
  * **Input:** `userId`, `postId`, `quan hệ giữa 2 user`
  * **Logic:** Kiểm tra BR-01, thực hiện tạo entity nếu thỏa mãn
  * **Output:** `repost entity` hoặc `error`