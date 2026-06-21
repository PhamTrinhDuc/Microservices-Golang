## 1. Telemetry.go 

## 2. Metrics.go 
Đăng ký các loại metrics cho MCP Server với OTEL:

### 1. Nhóm Request (Giao tiếp với App)
Nhóm này đo xem người dùng đang tương tác với app như thế nào:
*   **RequestCount:** Đếm tổng số yêu cầu gửi đến app. (Bác biết được app có đang bị "overload" do nhiều người dùng quá không).
*   **RequestDuration:** Đo thời gian phản hồi của mỗi request. (Giúp bác biết app đang chạy nhanh hay chậm).
*   **ActiveRequests:** Đếm số lượng request đang xử lý tại đúng thời điểm đó. (Nếu con số này tăng liên tục mà không giảm tức là app đang bị "treo").

### 2. Nhóm Tool Execution (Xử lý tác vụ)
Vì đây là app MCP, nó sẽ gọi các "Tool" (công cụ) để làm việc:
*   **ToolExecutionCount:** Đếm xem các tool đã được gọi bao nhiêu lần.
*   **ToolExecutionDuration:** Đo xem mỗi tool chạy mất bao lâu (Ví dụ: tool đọc file mất 10ms, tool gọi AI mất 2s).

### 3. Nhóm Database (Tương tác cơ sở dữ liệu)
Đo xem việc đọc/ghi dữ liệu có ổn định không:
*   **DBQueryDuration / Count:** Đo tốc độ và số lượng truy vấn database.
*   **DBConnectionPool (Active/Idle):** Đây là cái rất hay. Nó đo xem bác đang dùng bao nhiêu kết nối vào DB. Nếu "Active" quá cao và "Idle" bằng 0, tức là bác sắp hết kết nối và app sẽ bị lỗi kết nối DB.

### 4. Nhóm AI & Search (Dành riêng cho RAG/Search)
Nếu bác làm app tìm kiếm tài liệu bằng AI:
*   **SearchResultCount:** Một lần search bác trả về bao nhiêu kết quả.
*   **HybridSearchScore:** Đo chất lượng (điểm số) của các kết quả tìm kiếm. Nếu điểm quá thấp, bác biết là hệ thống search của mình đang hoạt động không hiệu quả.

### 5. Nhóm Error (Quản lý lỗi)
*   **ErrorCount:** Gom tất cả các lỗi lại để đếm. Bác có thể biết lỗi phát sinh từ đâu (từ Tool, từ DB, hay từ Request) thông qua cái gọi là `attribute` (nhãn) đi kèm.

### Giải thích nhanh về "Cách đo" cho bác dễ hình dung:
Trong code có 3 kiểu đo (Instrument) khác nhau:
1.  **Counter (Cái đếm):** Chỉ có tăng lên, không bao giờ giảm (giống công tơ mét xe máy). Dùng để đếm tổng số lượng.
2.  **Histogram (Biểu đồ phân phối):** Không chỉ đo trung bình mà còn đo xem bao nhiêu cái nhanh, bao nhiêu cái chậm. Dùng cho thời gian (Duration).
3.  **UpDownCounter (Bộ đếm lên xuống):** Có thể tăng và giảm (giống số người trong thang máy). Dùng cho những thứ có trạng thái thay đổi liên tục như "số kết nối đang mở".

Câu hỏi rất thực tế bác ạ. Cái `WithUnit` này tưởng nhỏ nhưng nó sẽ giúp các biểu đồ sau này (như trên Grafana) hiển thị đúng đơn vị theo trục Y (y-axis) mà bác không cần cấu hình bằng tay nhiều.


## Giải thích về các Unit Metrics:  
### 1. `WithUnit` là cái gì?
Nó đơn giản là một cái **"nhãn" (label)** đánh dấu cho hệ thống biết con số bác đang gửi lên đại diện cho cái gì.
*   Nếu bác gửi con số `10` và unit là `ms` -> Grafana sẽ hiểu là `10 milliseconds`.
*   Nếu bác gửi con số `10` và unit là `By` -> Grafana sẽ hiểu là `10 Bytes`.

### 2. Các đơn vị này do OTEL định nghĩa hay mình tự chế?
Thực tế, OpenTelemetry khuyến khích sử dụng tiêu chuẩn **UCUM (Unified Code for Units of Measure)**, nhưng về mặt kỹ thuật, nó chỉ là một chuỗi `string`. Bác có thể viết gì vào đó cũng được, nhưng nên dùng theo chuẩn để các công cụ hiển thị (Prometheus/Grafana) tự động nhận diện được.

**Các đơn vị phổ biến theo chuẩn UCUM:**
*   **Thời gian:** `ms` (milliseconds), `s` (seconds), `us` (microseconds).
*   **Dung lượng:** `By` (Bytes), `KBy` (Kilobytes), `MBy` (Megabytes).
*   **Tỉ lệ:** `1` (dimensionless - không đơn vị), `%` (phần trăm).

### 3. Tại sao trong code của bác lại có dấu ngoặc nhọn `{...}`?
Bác sẽ thấy trong code gốc có những đoạn như: `metric.WithUnit("{request}")` hay `metric.WithUnit("{error}")`.

Đây là một quy ước trong UCUM:
*   Nếu đơn vị **không phải là đơn vị vật lý** (không phải là gram, mét, giây...), bác nên cho nó vào trong dấu `{}`.
*   **Ví dụ:** `{request}`, `{order}`, `{connection}`. 
*   **Tác dụng:** Giúp người xem dashboard hiểu ngay con số đó đang đếm cái gì mà không bị nhầm lẫn với các đơn vị tính toán khác.

### 4. Cách chọn đúng Unit cho từng Metric:
Bác có thể tham khảo bảng "bí kíp" này của tôi để chọn cho đúng:

| Loại Metric | Trường hợp sử dụng | Unit khuyên dùng |
| :--- | :--- | :--- |
| **Đếm số lượng** | Counter (Request, Error, Items) | `{tên_vật_thể}` (Ví dụ: `{request}`) |
| **Đo thời gian** | Histogram (Latency, Duration) | `ms` hoặc `s` (thường dùng `ms` cho chính xác) |
| **Đo dung lượng** | Memory, File size, Disk | `By`, `KBy`, `MBy` |
| **Trạng thái** | Active connections, CPU usage | `1` (nếu là số lượng) hoặc `%` (nếu là tỉ lệ) |

**Tóm lại:**
Bác cứ chọn Unit sao cho **người xem biểu đồ hiểu được ngay con số đó là gì**. 
*   Đo tốc độ -> dùng `ms`.
*   Đếm cái gì đó -> dùng `{tên_cái_đó}`.
*   Không biết dùng gì -> dùng `1` hoặc bỏ trống.

## 3. Tracing.go