# TGDD Product Crawler Suite

Bộ công cụ thu thập dữ liệu (crawler) từ Thế Giới Di Động (TGDD), được thiết kế mô-đun hóa để cào các loại dữ liệu độc lập, khớp hoàn hảo với cấu trúc cơ sở dữ liệu PostgreSQL của bạn và cung cấp dữ liệu văn bản phục vụ hệ thống RAG (Retrieval-Augmented Generation).

---

## Cấu trúc thư mục hiện tại

```text
craw_data/
├── data/
│   ├── stores.json                       # Dữ liệu hệ thống siêu thị
│   ├── products.json                     # Dữ liệu sản phẩm hàng loạt (thông số, màu, dung lượng, ảnh gallery, khuyến mãi)
│   ├── reviews.json                      # Dữ liệu đánh giá của khách hàng cho từng sản phẩm
│   ├── policies/                         # Thư mục chứa cơ sở dữ liệu tri thức RAG
│   │   ├── chinh-sach-bao-hanh.md        # Văn bản chính sách bảo hành của TGDD
│   │   └── ...                           # Các văn bản chính sách/nội quy khác
│   └── test_product_schema_aligned.json   # Kết quả chạy thử nghiệm cào 1 sản phẩm
├── crawler_stores.py                     # Crawler bóc tách hệ thống cửa hàng & tọa độ GPS
├── crawler_products.py                   # Crawler bóc tách sản phẩm, cấu hình, gallery, khuyến mãi
├── crawler_reviews.py                    # Crawler bóc tách bình luận & xếp hạng sao
├── crawler_policies.py                   # Crawler bóc tách chính sách/quy chế phục vụ RAG
├── pyproject.toml                        # Cấu hình dự án & dependencies
├── uv.lock                               # Quản lý lock dependencies (nếu dùng uv)
└── README.md                             # Hướng dẫn này
```

---

## Yêu cầu hệ thống & Cài đặt

1. **Python**: Đã cài đặt phiên bản `>=3.12`.
2. **Thư viện cần thiết**: `selenium`, `webdriver-manager`, `pandas`.

### Cài đặt nhanh bằng `pip` hoặc `uv`

Nếu bạn dùng `uv` (khuyên dùng):
```bash
# Đồng bộ môi trường ảo và cài đặt dependencies tự động
uv sync
```

Hoặc cài đặt thủ công qua `pip` trong môi trường ảo của bạn:
```bash
pip install selenium webdriver-manager pandas
```

---

## Hướng dẫn sử dụng chi tiết từng Crawler

### 1. Cào hệ thống cửa hàng (`crawler_stores.py`)
Mục đích: Cào địa chỉ siêu thị, hotline, chia nhỏ tỉnh/phường/đường và lấy tọa độ GPS (`lat`, `lng`) ánh xạ vào bảng `Store`.

*   **Chạy cào thử nghiệm (1 tỉnh, tối đa 3 cửa hàng):**
    ```bash
    .venv\Scripts\python crawler_stores.py --limit-provinces 1 --limit-stores 3
    ```
*   **Chạy cào toàn bộ cửa hàng trên cả nước:**
    ```bash
    .venv\Scripts\python crawler_stores.py
    ```
    *Dữ liệu sẽ được lưu tại `data/stores.json`.*

---

### 2. Cào danh mục sản phẩm & Gallery & Khuyến mãi (`crawler_products.py`)
Mục đích: Cào sản phẩm, thuộc tính màu/dung lượng, thông số phân tách số và đơn vị, ảnh bộ sưu tập (`ProductImage`) và chương trình khuyến mãi (`Promotions`).

*   **Cào 1 link sản phẩm cụ thể:**
    ```bash
    .venv\Scripts\python crawler_products.py --url "https://www.thegioididong.com/dtdd/iphone-17"
    ```
    *Lưu kết quả tại `data/test_product_schema_aligned.json`.*
*   **Cào hàng loạt sản phẩm theo trang danh mục (giới hạn 5 sản phẩm):**
    ```bash
    .venv\Scripts\python crawler_products.py --category "https://www.thegioididong.com/dtdd" --name "Điện thoại" --limit 5
    ```
    *Lưu kết quả tại `data/products.json`.*
*   **Chạy demo mặc định (không truyền tham số):**
    ```bash
    .venv\Scripts\python crawler_products.py
    ```

---

### 3. Cào đánh giá & bình luận của người mua (`crawler_reviews.py`)
Mục đích: Cào các đánh giá thật của khách hàng, số sao bình luận, nội dung phản hồi, thời gian dùng và ảnh feedback thực tế, ánh xạ vào bảng `Reviews`.

*   **Cào đánh giá của 1 sản phẩm cụ thể (Giới hạn 1 trang đầu, max 5 bình luận):**
    ```bash
    .venv\Scripts\python crawler_reviews.py --url "https://www.thegioididong.com/dtdd/iphone-17" --pages 1 --limit 5
    ```
*   **Cào đánh giá sâu hơn (ví dụ cào 3 trang, giới hạn 30 bình luận):**
    ```bash
    .venv\Scripts\python crawler_reviews.py --url "https://www.thegioididong.com/dtdd/iphone-17" --pages 3 --limit 30
    ```
    *Dữ liệu được lưu tại `data/reviews.json`.*

---

### 4. Cào chính sách/quy chế phục vụ RAG (`crawler_policies.py`)
Mục đích: Cào các bài viết chính sách bảo hành, đổi trả, giao hàng, hướng dẫn mua online, quy chế hoạt động... để làm dữ liệu phi cấu trúc phục vụ Agent RAG.

*   **Chạy cào thử nghiệm (chỉ cào 1 chính sách đầu tiên):**
    ```bash
    .venv\Scripts\python crawler_policies.py --limit 1
    ```
*   **Chạy cào toàn bộ 10 chính sách cốt lõi của TGDD:**
    ```bash
    .venv\Scripts\python crawler_policies.py
    ```
    *Các file Markdown sẽ được lưu tại thư mục `data/policies/`.*

---

## Định dạng dữ liệu đầu ra khớp DB (JSON fields mapping)

### Store:
- `name`: Tên cửa hàng ("Thế giới di động Nguyễn Duy Trinh")
- `hotline`: "18001060"
- `province`, `ward`, `road`: Địa chỉ được chia nhỏ hoàn chỉnh
- `lat`, `lng`: Tọa độ GPS trích xuất từ bản đồ

### Product Image:
- Mảng `gallery_images` trong sản phẩm chứa:
  - `url`: Link ảnh đầy đủ
  - `alt_text`: Mô tả ảnh
  - `sort_order`: Thứ tự hiển thị (từ 1 tăng dần)
  - `is_thumbnail`: `true` cho ảnh đại diện chính, `false` cho các ảnh slide chi tiết

### Promotions:
- Mảng `promotions` trong sản phẩm chứa:
  - `name`: "Khuyến mãi Thế Giới Di Động"
  - `description`: Nội dung chi tiết quà tặng/giảm giá
  - `discount_percent`: % giảm giá trích xuất được từ mô tả (nếu có, VD: 10)

### Reviews:
- `reviewer_name`: Tên khách hàng
- `rating`: Số sao (1 đến 5)
- `comment`: Nội dung bình luận
- `usage_time`: Thời gian đã dùng máy (VD: "Đã dùng khoảng 4 tháng")
- `images`: Mảng URL ảnh feedback thực tế do khách hàng tải lên

### RAG Policies:
- Định dạng: File `.md` (Markdown) chuẩn
- Cấu trúc: Tiêu đề trang (`# Title`), link gốc (`**Source URL**`) và toàn bộ nội dung văn bản chính sách đã được lọc bỏ quảng cáo, menu và footer.
