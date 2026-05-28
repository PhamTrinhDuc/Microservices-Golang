# CHEATSHEET: Ký hiệu UML và Mermaid tương ứng

## 1. Quan hệ trong Class Diagram

| Loại quan hệ | Mô tả | Mermaid syntax |
|---|---|---|
| Association | Liên kết đơn giản giữa 2 lớp | `A -- B` |
| Directed Association | Liên kết có hướng | `A --> B` |
| Inheritance (Generalization) | Kế thừa (con → cha) | `A <|-- B` (B kế thừa A) |
| Realization | Thực thi interface | `A <|.. B` |
| Dependency | Phụ thuộc | `A <.. B` |
| Aggregation | Cộng hợp (A là phần của B, A tồn tại độc lập) | `B o-- A` |
| Composition | Gắn chặt (A là phần của B, A không tồn tại độc lập) | `B *-- A` |

**Ghi chú số lượng:**
```
A "1" --> "n" B : label
A "0..1" --> "1..*" B
```

---

## 2. Các kiểu lớp đặc biệt

```mermaid
classDiagram
    class LopThucThe {
        <<entity>>
    }
    class LopBien {
        <<boundary>>
    }
    class LopDieuKhien {
        <<control>>
    }
    class Interface {
        <<interface>>
    }
    class AbstractClass {
        <<abstract>>
    }
```

---

## 3. Use Case — Quan hệ giữa các UC

```
%% Include: A phải thực hiện B
UC_A ..> UC_B : <<include>>

%% Extend: B tùy chọn mở rộng A
UC_B ..> UC_A : <<extend>>

%% Generalization: UC_Con kế thừa UC_Cha
UC_Cha <|-- UC_Con
```

---

## 4. Sequence Diagram — Các loại message

```mermaid
sequenceDiagram
    A->>B: Gọi đồng bộ (call)
    A-->>B: Trả về (return, nét đứt)
    A-)B: Gọi bất đồng bộ
    A->>+B: Kích hoạt (activate)
    B-->>-A: Trả về & deactivate
    Note over A,B: Ghi chú
    loop Vòng lặp
        A->>B: Lặp lại
    end
    alt Điều kiện đúng
        A->>B: Nhánh 1
    else Điều kiện sai
        A->>B: Nhánh 2
    end
    opt Tùy chọn
        A->>B: Thực hiện nếu có
    end
```

---

## 5. State Machine — Cú pháp

```mermaid
stateDiagram-v2
    [*] --> State1
    
    State1 --> State2 : event [guard] / action
    State2 --> State3 : event
    
    %% Trạng thái tổ hợp
    state State2 {
        [*] --> SubState1
        SubState1 --> SubState2
    }
    
    State3 --> [*]
```

---

## 6. Activity Diagram — Cú pháp đầy đủ

```mermaid
flowchart TD
    Start([●]) --> A

    A[Hoạt động 1] --> B{Điều kiện?}
    
    B -- "true [guard]" --> C[Hoạt động 2]
    B -- "false" --> D[Hoạt động 3]
    
    %% Song song
    C --> Fork[ ]
    Fork --> E[Nhánh 1]
    Fork --> F[Nhánh 2]
    E --> Join[ ]
    F --> Join[ ]
    Join --> G[Tiếp tục]
    
    D --> G
    G --> End([◉])
```

---

## 7. Ký hiệu phạm vi (Visibility)

| Ký hiệu | Tên | Mô tả |
|---|---|---|
| `+` | public | Truy cập từ mọi nơi |
| `-` | private | Chỉ trong lớp |
| `#` | protected | Trong lớp và lớp con |
| `~` | package | Trong cùng package |

---

## 8. Cú pháp thuộc tính và phương thức

**Thuộc tính:**
```
phạm_vi tên : kiểu [số_lượng] = mặc_định
```
Ví dụ: `- ngaySinh : Date[1] = "01-01-2000"`

**Phương thức:**
```
phạm_vi tên(tham_so : kiểu) : kiểu_tra_ve
```
Ví dụ: `+ tinhDiem(monHocId : int) : float`

---

## 9. CSDL — Quy tắc đặt tên

| Đối tượng | Quy tắc | Ví dụ |
|---|---|---|
| Tên bảng | `tbl` + TenLop | `tblSinhVien` |
| Khóa chính | `id` | `id` (int, auto increment) |
| Khóa ngoại | `tbl{A}id` | `tblKhoaid` |
| Cột thông thường | camelCase | `hoTen`, `ngaySinh` |

---

## 10. Bảng kiểu dữ liệu Java → SQL

| Java | MySQL |
|---|---|
| `String` | `VARCHAR(n)` |
| `int` / `Integer` | `INT` |
| `float` / `Float` | `FLOAT` |
| `double` / `Double` | `DOUBLE` |
| `Date` | `DATE` |
| `boolean` / `Boolean` | `TINYINT(1)` |
| `long` / `Long` | `BIGINT` |