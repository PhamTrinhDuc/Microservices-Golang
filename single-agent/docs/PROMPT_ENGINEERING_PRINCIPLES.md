# Nguyên Tắc Cốt Lõi: Viết Prompt & Tool Description cho LLM Agent

> Tài liệu này đúc kết các nguyên tắc thiết kế dành cho hệ thống `single-agent`.  
> Áp dụng cho mọi agent, mọi nhóm tools mới được thêm vào trong tương lai.

---

## Mục lục

1. [Kiến trúc Prompt 3 tầng](#1-kiến-trúc-prompt-3-tầng)
2. [Nguyên tắc viết System Prompt](#2-nguyên-tắc-viết-system-prompt)
3. [Nguyên tắc viết Agent Instruction](#3-nguyên-tắc-viết-agent-instruction)
4. [Nguyên tắc viết Tool Description](#4-nguyên-tắc-viết-tool-description)
5. [Approval-Required Tools](#5-approval-required-tools)
6. [Checklist trước khi deploy](#6-checklist-trước-khi-deploy)

---

## 1. Kiến trúc Prompt 3 tầng

Toàn bộ prompt của agent được chia thành **3 tầng độc lập**, mỗi tầng có trách nhiệm riêng biệt:

```
┌──────────────────────────────────────────────────────┐
│  TẦNG 1: system_prompt  (config.yaml)                │
│  Persona + Meta-principles bất biến                   │
│  → Không mention tên tool cụ thể nào                 │
│  → Không thay đổi khi thêm tools mới                 │
└──────────────────────────────────────────────────────┘
                         ↓
┌──────────────────────────────────────────────────────┐
│  TẦNG 2: instruction  (config.yaml / per-agent)      │
│  Workflow tư duy + Approval awareness                  │
│  → Định nghĩa CÁC BƯỚC suy nghĩ, không hard-code     │
│     tên tool                                          │
│  → Cập nhật khi thêm nhóm tool có workflow khác      │
└──────────────────────────────────────────────────────┘
                         ↓
┌──────────────────────────────────────────────────────┐
│  TẦNG 3: Tool Description  (MCP server Go code)      │
│  Self-documenting tools                               │
│  → Mỗi tool tự mô tả khi nào nên và không nên dùng  │
│  → Là nơi DUY NHẤT chứa logic "chọn tool nào"       │
└──────────────────────────────────────────────────────┘
```

**Lý do phân tầng:**  
- Khi thêm tool mới → chỉ cần viết description tốt ở Tầng 3, Tầng 1 & 2 không cần đổi.  
- Khi thêm nhóm workflow mới → chỉ cập nhật Tầng 2.  
- Tầng 1 luôn ổn định.

---

## 2. Nguyên tắc viết System Prompt

### 2.1 Chỉ chứa "điều bất biến"

`system_prompt` là nơi định nghĩa:
- **Persona** (nhân vật, tên, phong cách, tông giọng)
- **Meta-principles** (nguyên tắc tư duy cốt lõi không phụ thuộc tools)
- **Hard rules** (những điều tuyệt đối không bao giờ làm)

**✅ Đúng:**
```yaml
system_prompt: |
  # META-PRINCIPLES
  1. Dữ liệu thực tế trước tiên: Luôn lấy thông tin từ hệ thống trước khi trả lời.
  2. Đọc mô tả công cụ: Mỗi công cụ tự định nghĩa khi nào nên dùng.
  3. Tối thiểu hóa số lần gọi công cụ.
  4. Lọc tại nguồn: Truyền đầy đủ filter vào tool, không lấy bulk rồi tự lọc.
  5. Giải quyết phụ thuộc dữ liệu: Thiếu ID thì lookup trước.
  6. Tự sửa lỗi có giới hạn: Thử lại tối đa 2 lần.
```

**❌ Sai — mention tên tool cụ thể:**
```yaml
system_prompt: |
  - Luôn gọi list_categories trước khi gọi list_products
  - Chỉ gọi get_product_by_id khi user hỏi 1 sản phẩm cụ thể
```

> **Tại sao sai?** Khi thêm tool mới, bạn phải sửa `system_prompt` → vi phạm nguyên tắc Open/Closed.

### 2.2 Thứ tự section trong buildInstruction

LLM có **primacy & recency bias** — nhớ phần đầu và cuối nhất.

```go
// Thứ tự tối ưu trong buildInstruction:
instruction (Tool Rules)   // Đầu tiên → xử lý sớm nhất, được "nạp" vào reasoning
system_prompt (Persona)    // Giữa → identity luôn active
response_format            // Cuối → recency effect, áp dụng khi sinh text
```

### 2.3 Không chồng chéo giữa các tầng

- `system_prompt` nói về **tư duy**
- `instruction` nói về **workflow**
- Tool description nói về **khi nào dùng tool đó**

Nếu hai tầng đề cập cùng một điều → LLM bị confuse hoặc bỏ qua một trong hai.

---

## 3. Nguyên tắc viết Agent Instruction

### 3.1 Dạy "cách suy nghĩ", không dạy "tên tool"

**✅ Đúng — dạy nguyên tắc:**
```yaml
instruction: |
  Bước 2 – Lập kế hoạch công cụ:
  - Xác định số công cụ tối thiểu cần gọi.
  - Nếu thiếu tham số định danh (ID): tra cứu trước, rồi mới gọi công cụ chính.
  - Đọc mô tả từng công cụ để quyết định dùng đúng công cụ, đúng lúc.
```

**❌ Sai — hard-code tên tool:**
```yaml
instruction: |
  - Gọi list_categories trước để lấy category_id
  - Sau đó gọi list_products với đầy đủ filter
  - Chỉ gọi get_product_by_id khi user hỏi chi tiết
```

> **Tại sao sai?** Khi thêm tool mới (`list_brands`, `search_orders`...), bạn phải sửa instruction.

### 3.2 Luôn có "When NOT to call tools"

Đây là phần quan trọng thường bị bỏ sót. LLM mặc định muốn dùng tool, nên phải dạy nó biết **khi nào không cần**:

```yaml
Bước 3 – Khi KHÔNG cần gọi công cụ:
- Câu chào hỏi, câu xã giao thông thường.
- Thông tin đã có đầy đủ trong lịch sử hội thoại.
- Câu hỏi về chính sách/quy trình bạn đã biết sẵn.
```

### 3.3 Workflow phải là bước tuần tự, không phải bullet rules

**✅ Đúng — dạng workflow:**
```
Bước 1 → Phân tích câu hỏi
Bước 2 → Lập kế hoạch công cụ
Bước 3 → Kiểm tra có cần gọi không
Bước 4 → Xử lý approval nếu cần
Bước 5 → Tổng hợp & phản hồi
```

**❌ Sai — chỉ là bullet rules không có thứ tự:**
```
- Tối ưu số lần gọi tool
- Filter at source
- Không gọi get_product_by_id hàng loạt
```

> Bullet rules dễ bị LLM "bỏ qua" khi context dài. Workflow buộc LLM phải đi theo thứ tự.

---

## 4. Nguyên tắc viết Tool Description

Tool description là **nguồn thông tin duy nhất** để LLM quyết định dùng tool nào, khi nào.  
Một description tốt loại bỏ nhu cầu hard-code tên tool trong prompt.

### 4.1 Template chuẩn cho mọi tool

```
[Một câu mô tả tool làm gì — output là gì]
USE when: [điều kiện cụ thể nên dùng tool này]
DO NOT USE when: [điều kiện không nên dùng — quan trọng như USE]
NOTE: [dependency, side effect, hoặc lưu ý quan trọng về tham số]
```

### 4.2 Ví dụ thực tế

**❌ Description yếu (cũ):**
```go
Description: "List all product categories available in the catalog."
```
→ LLM không biết khi nào dùng, khi nào không → dùng bừa.

**✅ Description mạnh (mới):**
```go
Description: `List all product categories available in the catalog. Returns category name, ID, and slug.
USE when: you need a category_id but don't have it yet, or the user asks what product types are available.
DO NOT USE when: you already know the category_id, or the query is specific enough to search directly by keyword.`
```
→ LLM biết chính xác khi nào dùng tool này.

### 4.3 Viết description cho parameter

Parameter description phải trả lời được: **"Tôi nên truyền gì vào đây?"**

**❌ Yếu:**
```go
"search_term": "Keyword to search for categories"
```

**✅ Mạnh:**
```go
"search_term": "Keyword to search for a specific category by name (e.g. 'laptop', 'smartwatch'). Use this to resolve category_id without fetching all categories."
```

### 4.4 Quy tắc "DO NOT USE" — quan trọng như "USE"

Mọi tool đều cần có **ít nhất 1 DO NOT USE** để:
1. Ngăn LLM lạm dụng tool (quá nhiều calls)
2. Hướng LLM sang tool phù hợp hơn trong trường hợp đó
3. Giảm token cost và latency

| Loại tool | DO NOT USE phổ biến |
|---|---|
| Lookup/List tool | Khi đã có ID rồi |
| Detail tool | Khi chỉ cần tổng quan / bulk |
| Spec/metadata tool | Khi đã biết tên tham số |
| Write/Mutation tool | Khi chưa có xác nhận từ user |

### 4.5 Tính nhất quán ngôn ngữ

Chọn **một ngôn ngữ duy nhất** cho tool description — khuyến nghị **tiếng Anh** vì:
- LLM được pre-train nhiều hơn với English
- Dễ maintainability khi team mở rộng
- Tránh inconsistency khi mix Việt-Anh

---

## 5. Approval-Required Tools

### 5.1 Khi nào cần approval?

Áp dụng cho bất kỳ tool nào có **tác động không thể hoàn tác**:
- Tạo đơn hàng, đặt hàng
- Thanh toán, xử lý giao dịch
- Cập nhật thông tin tài khoản người dùng
- Xóa dữ liệu

### 5.2 Cách đánh dấu tool cần approval

Thêm annotation `[REQUIRES_APPROVAL]` vào đầu description:

```go
Description: `[REQUIRES_APPROVAL] Place an order for the specified product and quantity.
This action creates a real order in the system and CANNOT be undone without contacting support.
USE only after the user has explicitly confirmed the product, quantity, and delivery address.
Before calling: summarize the order details clearly and wait for explicit user confirmation.`
```

### 5.3 Logic xử lý approval trong instruction

```yaml
Bước 4 – Công cụ cần phê duyệt (Approval-Required):
  Nhận biết qua annotation [REQUIRES_APPROVAL] trong mô tả công cụ.
  Quy trình bắt buộc:
    1. Tóm tắt rõ hành động sắp thực hiện cho khách.
    2. Chờ khách xác nhận đồng ý một cách tường minh.
    3. Chỉ gọi công cụ sau khi có xác nhận.
```

### 5.4 Mapping với `approvedTools` trong config

```yaml
agents:
  ecommerce_agent:
    allowedTools:    # Gọi tự do
      - list_products
      - list_categories
    approvedTools:   # Cần xác nhận trước khi execute
      - place_order
      - process_payment
```

> Backend handler cần intercept tools trong `approvedTools`, gửi confirmation request về frontend trước khi thực thi.

---

## 6. Checklist trước khi deploy

### Checklist cho Prompt mới

- [ ] `system_prompt` không mention tên tool cụ thể nào
- [ ] `system_prompt` có đủ 6 meta-principles cốt lõi
- [ ] `instruction` có dạng workflow tuần tự (Bước 1 → Bước N)
- [ ] `instruction` có section "Khi KHÔNG cần gọi công cụ"
- [ ] `instruction` có section "Approval-Required" workflow
- [ ] Không có nội dung chồng chéo giữa `system_prompt` và `instruction`

### Checklist cho Tool Description mới

- [ ] Câu đầu mô tả tool làm gì + output là gì
- [ ] Có `USE when:` với ít nhất 1 điều kiện cụ thể
- [ ] Có `DO NOT USE when:` với ít nhất 1 điều kiện
- [ ] Có `NOTE:` nếu tool có dependency với tool khác
- [ ] Có `[REQUIRES_APPROVAL]` nếu tool có side effect không thể hoàn tác
- [ ] Mỗi parameter có description đủ để biết truyền gì vào
- [ ] Viết bằng tiếng Anh nhất quán

### Checklist khi thêm Tool mới vào agent

- [ ] Tool đã được thêm vào `allowedTools` hoặc `approvedTools` trong config.yaml
- [ ] `system_prompt` không cần sửa
- [ ] `instruction` không cần sửa (chỉ sửa nếu tool thuộc một nhóm workflow hoàn toàn mới)
- [ ] Tool description đủ mạnh để LLM tự suy ra khi nào dùng

---

## Tham khảo

- [config.yaml](../config.yaml) — System prompt, instruction, response format
- [api_category.go](../../mcp-server/internal/tools/api_category.go) — Tool definitions mẫu
- [api_products.go](../../mcp-server/internal/tools/api_products.go) — Tool definitions mẫu
- [agents/agent.go](../agents/agent.go) — buildInstruction function
