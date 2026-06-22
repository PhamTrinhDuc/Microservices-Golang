### Bảng so sánh tổng quát

| Tiêu chí | **Fixed Window** (Cửa sổ cố định) | **Sliding Window Log** (Cửa sổ trượt) | **Token Bucket** (Thùng Token) |
| :--- | :--- | :--- | :--- |
| **Độ khó code** | Rất dễ (vài dòng Go) | Trung bình (Dùng Redis ZSet) | Khó hơn (Cần Lua Script) |
| **Độ chính xác** | Thấp (Bị lỗi bùng nổ ở ranh giới) | **Cao nhất** (Chính xác từng mili giây) | Cao (Rất mượt mà) |
| **Tốn bộ nhớ** | Rất ít (Chỉ lưu 1 con số) | **Rất nhiều** (Lưu timestamp mọi request) | Ít (Lưu số lượng và thời gian) |
| **Khả năng Burst** | Không hỗ trợ | Không hỗ trợ | **Có (Rất tốt)** |
| **Dùng khi nào?** | Project nhỏ, không quan trọng UX | Cần độ chính xác tuyệt đối | Hệ thống lớn, API trả phí, SaaS |

---

### 1. Fixed Window (Cách bạn đang dùng)
* **Cơ chế:** Chia thời gian thành các khối cứng nhắc (ví dụ: phút 10, phút 11). Hết phút là reset bộ đếm về 0.
* **Điểm yếu chết người:** Giả sử giới hạn 100 req/phút. Nếu user gửi 100 req vào giây thứ 59 và 100 req vào giây thứ 01 của phút tiếp theo -> Hệ thống phải gánh **200 req trong 2 giây**.
* **Phù hợp:** Chặn spam cơ bản, các task không quá nhạy cảm.

### 2. Sliding Window Log (Dùng Sorted Set)
* **Cơ chế:** Lưu mọi thời điểm (timestamp) của request vào một danh sách trong Redis. Mỗi khi có req mới, nó quét lại danh sách và xóa những req cũ hơn 1 phút trước.
* **Ưu điểm:** Chính xác tuyệt đối. Không bao giờ có chuyện "vượt rào" ở ranh giới phút.
* **Nhược điểm:** Nếu một user gửi 10.000 request, Redis phải lưu 10.000 cái timestamp đó. Rất tốn RAM nếu hệ thống có hàng triệu user.
* **Phù hợp:** Các API cực kỳ quan trọng cần độ chính xác cao nhưng lượng traffic mỗi user không quá lớn.

### 3. Token Bucket (Cách dùng Lua Script)
* **Cơ chế:** Như một cái vòi nước chảy vào xô. Nước (token) chảy vào xô với tốc độ đều đặn. Khi cần dùng API, bạn múc 1 gáo nước ra. Nếu xô đầy, nước tràn ra ngoài (bỏ qua). Nếu xô cạn, bạn phải đợi nước chảy vào thêm.
* **Ưu điểm:**
    * **Cho phép Burst:** Nếu user không dùng API trong 10 phút, cái xô sẽ đầy. Khi đó họ có thể "xả" một lúc 100 request (bằng dung tích xô) mà không bị chặn. Điều này rất thân thiện với thực tế sử dụng.
    * **Tiết kiệm:** Chỉ cần lưu 2 con số: `số token hiện tại` và `thời gian cập nhật cuối`.
* **Phù hợp:** **Lựa chọn số 1 cho các hệ thống hiện đại**, các nền tảng SaaS (như cách Stripe, AWS đang làm).

```bash
local key = KEYS[1]           -- Key của tenant
local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = 1

-- Lấy dữ liệu cũ từ Redis
local bucket = redis.call('hmget', key, 'tokens', 'last_updated')
local tokens = tonumber(bucket[1]) or capacity
local last_updated = tonumber(bucket[2]) or now

-- 1. Tính toán lượng token nạp thêm theo thời gian
local duration = math.max(0, now - last_updated)
local generated = duration * refill_rate
tokens = math.min(capacity, tokens + generated)

-- 2. Kiểm tra xem có đủ token không
local allowed = false
if tokens >= requested then
    tokens = tokens - requested
    allowed = true
end

-- 3. Lưu lại trạng thái mới
redis.call('hmset', key, 'tokens', tokens, 'last_updated', now)
redis.call('expire', key, 600) -- Tự xóa sau 10 phút không hoạt động để dọn rác

return allowed and 1 or 0
```

---

### Chốt lại: Bạn nên chọn cái nào?

1.  Nếu bạn muốn **nhanh, gọn, lẹ** và không sợ hệ thống bị "Peak" nhẹ: Cứ dùng **Fixed Window** (nhớ sửa lỗi Race Condition như mình nhắc ở trên).
2.  Nếu bạn muốn **chuyên nghiệp, mượt mà** và xử lý được cho nhiều khách hàng (Tenant): Hãy đầu tư viết **Token Bucket**. Nó là "tiêu chuẩn vàng" trong ngành rồi.
