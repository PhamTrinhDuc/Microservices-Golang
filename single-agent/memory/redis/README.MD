Dưới đây là bản tóm tắt chi tiết file `session.go`, giải thích cách nó hiện thực hóa (implement) bộ tiêu chuẩn quản lý phiên làm việc của **Google ADK-go** bằng **Redis**.

---

### 1. Tổng quan (High-Level)
File này đóng vai trò là một **Redis Storage Provider** cho ADK. Mục tiêu là chuyển đổi các yêu cầu quản lý trạng thái của Agent (như: "Nhớ tên khách hàng", "Lưu lại câu trả lời vừa rồi") thành các thao tác đọc/ghi trên database Redis.

Để làm được điều này, file implement **4 Interface cốt lõi** của ADK:
1.  **`session.Service`**: Bộ não điều khiển (Tạo, Xóa, Lấy Session).
2.  **`session.Session`**: Đại diện cho một phiên làm việc đang tồn tại.
3.  **`session.State`**: Kho chứa dữ liệu (biến số, thông tin ngữ cảnh).
4.  **`session.Events`**: Nhật ký các sự kiện/hội thoại đã xảy ra.

> Các dòng code `var _ session.Session = (*RedisSession)(nil)`, v.v... để ép sang interface của ADK `(session.Session)`. Ép RedisSession có kiểu `session.Session` và phải đủ các method đã định nghĩa trước trong `session.Session`
---

### 2. Kiến trúc lưu trữ (Data Organization)
File sử dụng một chiến lược đặt tên Key (Namespace) rất chặt chẽ để quản lý dữ liệu trong Redis:
####   `appstate:{app}`: Lưu dữ liệu dùng chung cho toàn bộ ứng dụng.
```bash
"appstate:app_001": 
    "theme": "dark", 
    "lang": "vi,
```
####   `userstate:{app}:{user}`: Lưu dữ liệu riêng của từng khách hàng.
```bash
"userstate:app_001:user_001": 
    "name": "Jiyuu", 
    "email": "Jiyuu@gmail.com", 
    "favorite_stylist" : "Alex", 
    "note": "thích uống trà đào ít đường",
```
Tại sao dùng `HSet` để thêm dữ liệu của App state và User state? 

- **Nhanh chóng**: Dữ liệu có app thường thay đổi liên tục `theme -> dark; lang -> eng` do nhu cầu của user. `HSet` cho phép cập nhật 1 vài trường dữ liệu trong json của value nhanh chóng, nó không lấy cả cụm json ra để cập nhật và ghi lại
 vào redis giống `Set` 
- **Tránh xung đột dữ liệu (Concurrency)**: Nếu dùng Set (JSON), khi 2 Agent cùng lúc muốn cập nhật 2 trường khác nhau (ví dụ: một thằng đổi theme, một thằng đổi lang), chúng sẽ bị ghi đè dữ liệu của nhau (Race Condition). Với HSet, mỗi thằng cập nhật một field riêng biệt nên không bao giờ bị xung đột.
-*** Tiết kiệm RAM (Memory Optimization)***: Redis có cơ chế tối ưu (ziplist) cho các Hash nhỏ. Lưu 1000 field trong 1 Hash tốn ít RAM hơn rất nhiều so với lưu 1000 Key-Value lẻ tẻ hoặc 1 cục JSON to đùng.

####   `session:{app}:{user}:{sess}`: Lưu thông tin cơ bản và metadata của 1 phiên.
```bash
"session:app_001:user_001:session_001": {
    "id": "session_001",
    "app_name": "app_001",
    "user_id": "user_001",
    "state": { "last_viewed": "item_1" },
    "last_update_time": "2026-03-05T10:10:00Z"
}
```
Session sử dụng `Set` thay vì `HSet` như app và user state để lưu dữ liệu vào Redis? 

- **Tính toàn vẹn (Atomic Object)**: Session thường được coi là 1 snapshot của cuộc hội thoại. Khi agent cần load phiên nó cần tất cả các thông tin cùng 1 lúc, lúc này dùng `Get` lấy nguyên cục json nhanh chóng hơn `Hget` gọi lẻ tẻ từng field
- **Hỗ trợ dữ liệu lồng nhau (Nested Data)**: Trường state trong Session có thể chứa các object phức tạp (ví dụ: một cái map nằm trong một cái map). Redis Hash (HSet) không hỗ trợ tốt việc lưu object lồng nhau, trong khi JSON (Set) thì cân được hết.
- **Vòng đời ngắn**: Session thường có thời gian sống (TTL) ngắn (ví dụ: 24h). Việc ghi đè một cục JSON nhỏ không tốn nhiều tài nguyên bằng việc duy trì một cấu trúc Hash phức tạp cho hàng triệu session ngắn hạn.

####   `sessions:{app}:{user}`: Một Redis `SET` chứa danh sách ID của tất cả session thuộc về một user.
```bash
"sessions:app_001:user_001": ["session_001", "session_002", "session_003"]
```
- **Tính duy nhất (Uniqueness)**: Vì lưu danh sách của các session cho 1 user nên trong TH này ta sử dụng `SAdd` để thêm vào 1 tập Set(), viêc này tránh được tập keys trùng nhau
- **Tốc độ kiểm tra (Existence check)**: Lệnh `SISMEMBER` kiểm tra cực nhanh xem một session có thuộc về user này không (O(1)).
- **Nhược điểm (The Flip side)**: `Set` không có thứ tự. Nếu muốn lấy danh sách "10 session mới nhất", sẽ không làm được với `Set`. Lúc đó phải đổi sang Sorted `Set` (ZSet) và dùng Timestamp làm Score để sắp xếp.

####   `events:{app}:{user}:{sess}`: Một Redis `LIST` lưu lịch sử hội thoại (Events).
```bash
"events:app_001:user_001:session_001": [
    // Phần tử [0] trong List (User nói)
    {
    "timestamp": "2026-05-12T16:00:00Z",
    "type": "message",
    "payload": {
    "role": "user",
    "text": "Hello, I want to book a haircut"
    }
    // Phần tử [1] trong List (AI đáp)
    {
    "timestamp": "2026-05-12T16:00:05Z",
    "type": "message",
    "payload": {
    "role": "assistant",
    "text": "Sure! Which stylist do you prefer?"
    }
]
- Lịch sử hội thoại lưu danh sách các Events có key có naming convention của session key, sử dụng RPush (đẩy vào bên phải/cuối), không dùng SAdd thì không có thứ tự. 
- Cho phép lấy N Events gần nhất thông qua LRange
```

---

### 3. Đi sâu vào chi tiết hiện thực (Detailed Implementation)

### Hàm `Create` (Khởi tạo phiên làm việc)

Khi bắt đầu một cuộc hội thoại mới:

1. **Kiểm tra trùng lặp:** Check `Exists` để xem Session ID này đã có trong Redis chưa. Nếu có rồi thì báo lỗi để tránh ghi đè dữ liệu cũ.
2. **Phân tách State:**
- Gọi `extractStateDeltas` để tách dữ liệu người dùng gửi lên thành 3 giỏ: App, User, và Session từ trường State sau đó update lại vào trong Redis nếu có thay đổi thông qua hàm: `updateAppState` và `updateUserState`

3. **Lưu trữ song song:**
* Lưu thông tin Session vào một Key String (`session:appName:userID:sessionID`).
* Thêm ID này vào "Mục lục" bằng `SAdd` để sau này biết User này có những Session nào.


4. **Expire (Hết hạn):** Đây là khái niệm cực quan trọng. Tác giả gọi `Expire(key, ttl)` để dặn Redis: "Sau X giờ, nếu không ai động vào cái này thì tự xóa đi cho đỡ chật nhà".

### Hàm `AppendEvent` (Ghi nhật ký và cập nhật trạng thái)

Đây là hàm chạy liên tục mỗi khi có tin nhắn mới:

1. **Dọn dẹp (`trimTempStateDelta`):** Nó xóa các key có tiền tố `temp:`. Đây là những dữ liệu "nháp" chỉ dùng cho xử lý tức thời, không đáng để lưu vào ổ cứng.
2. **Ghi Log:** Dùng `RPush` để thêm Event mới vào cuối danh sách Events trong Redis. Việc này giúp bảo toàn thứ tự thời gian của cuộc hội thoại.
3. **Cập nhật thông minh:** Thay vì sửa Session, nếu Event có thay đổi App State hay User State, nó dùng `HSet` để ghi đè đúng cái Field đó trong Redis Hash. Điều này rất tối ưu về hiệu năng.

### Khái niệm `Pipeline` (Đường ống)

Trong hàm `Delete`, bác sẽ thấy `s.client.Pipeline()`.

* **Giải thích:** Bình thường, mỗi lệnh gửi đến Redis là một lần đi-về qua mạng (Round-trip). Nếu xóa 3 thứ, mất 3 lần đi-về.
* **Cách làm:** `Pipeline` gom 3 lệnh (Xóa Session, Xóa Events, Xóa Index) gửi đi một lượt duy nhất. Nó giúp giảm độ trễ mạng và đảm bảo các lệnh được thực hiện liên tục.
