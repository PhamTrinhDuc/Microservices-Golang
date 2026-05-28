---
name: golang-code-style
description: Quy ước viết code Go sao cho đẹp, dễ đọc — bao gồm: độ dài dòng code, cách khai báo biến, viết control flow cho rõ ràng, khi nào nên/không nên viết comment. Dùng khi viết code, review code, hoặc thiết lập chuẩn code cho project
user-invocable: true
license: MIT
compatibility: Dùng cho Claude Code hoặc AI coding agents tương tự, và project viết bằng Golang
metadata:
  author: samber
  version: "1.2.0"
  openclaw:
    emoji: "🎨"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent
---
> **Community default.** A company skill that explicitly supersedes `samber/cc-skills-golang@golang-code-style` skill takes precedence.

# Go Code Style

## Triết lý chung

*"Code rõ ràng quan trọng hơn code khôn khéo."* — Go Proverbs 

Đừng cố viết code "ngầu" với trick phức tạp. Code mà đồng đội đọc 5 giây đã hiểu > code 10 dòng nhưng phải ngồi phân tích 5 phút.

> *Quy tắc vàng*: Khi cố tình phá vỡ một quy tắc nào đó, hãy viết comment giải thích lý do ngay trong code.

## 1. Độ dài dòng & Cách ngắt dòng

- Không có giới hạn cứng, nhưng dòng nào dài quá ~120 ký tự thì nên ngắt.
- Ngắt ở chỗ có nghĩa, không ngắt bừa theo số cột.
- Hàm có từ 4 tham số trở lên: mỗi tham số một dòng, kể cả khi AI gợi ý viết gộp.
```go
// ✅ Nên làm: mỗi argument một dòng, dấu đóng ngoặc riêng
mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
    handleUsers(
        w,
        r,
        serviceName,
        cfg,
        logger,
        authMiddleware,
    )
})
```
> Hint: Nếu signature hàm dài quá, thường là do hàm nhận quá nhiều tham số. Giải pháp thật sự là gom tham số vào một struct (options pattern), chứ không phải cố ngắt dòng cho đẹp.

## 2. Cách khai báo biến

### Quy tắc chọn := hay var



### Slice & Map: Luôn khởi tạo rõ ràng

```go
// ✅ Luôn khởi tạo, không để nil
users := []User{}                       // slice rỗng, an toàn
m := map[string]int{}                   // map rỗng, không panic khi ghi

// ✅ Preallocate khi biết trước kích thước
users := make([]User, 0, len(ids))      // tránh realloc nhiều lần
m := make(map[string]int, len(items))
```
> *Important*: 
> - Map nil sẽ panic khi ghi dữ liệu vào.
> - Slice nil khi marshal JSON ra null, còn slice rỗng [] ra [] — dễ gây nhầm lẫn cho API consumer.
> - Đừng preallocate "cho chắc": make([]T, 0, 1000) tốn memory nếu thực tế chỉ có 10 phần tử.

### Composite literals: Luôn ghi tên field
```go 
// ✅ Rõ ràng, dễ bảo trì
srv := &http.Server{
    Addr:         ":8080",
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
}

// ❌ Dễ break khi struct thay đổi thứ tự field
srv := &http.Server{":8080", 5 * time.Second, 10 * time.Second}
```

### Control Flow: Viết sao cho dễ đọc
#### 1.  Giảm thiểu nesting (lồng nhau)
Xử lý lỗi và edge case trước, return sớm. Giữ "happy path" (luồng chính) ở mức thụt ít nhất.
```go
// ✅ Dễ đọc: lỗi xử lý trước, logic chính nằm ở cuối
func process(data []byte) (*Result, error) {
    if len(data) == 0 {
        return nil, errors.New("empty data")
    }
    parsed, err := parse(data)
    if err != nil {
        return nil, fmt.Errorf("parsing: %w", err)
    }
    return transform(parsed), nil  // happy path
}
```

#### 2. Bỏ else khi không cần thiết
Nếu khối if đã kết thúc bằng return/break/continue, thì không cần else.

```go
// ✅ Clean: default value rồi override bằng switch
level := slog.LevelInfo
switch {
case debug:
    level = slog.LevelDebug
case verbose:
    level = slog.LevelWarn
}

// ❌ Else-if chain dài, khó thấy default value ở đâu
if debug {
    level = slog.LevelDebug
} else if verbose {
    level = slog.LevelWarn
} else {
    level = slog.LevelInfo
}
```

#### 3. Điều kiện phức tạp: Tách thành biến boolean có tên

Khi if có từ 3 điều kiện trở lên, nên tách ra thành biến boolean đặt tên rõ nghĩa.

```go
// ✅ Đọc là hiểu logic business
isAdmin := user.Role == RoleAdmin
isOwner := resource.OwnerID == user.ID
isPublicVerified := resource.IsPublic && user.IsVerified

if isAdmin || isOwner || isPublicVerified || permissions.Contains(PermOverride) {
    allow()
}
```
> *Hint*: Các check tốn kém (expensive) vẫn nên giữ inline để tận dụng short-circuit evaluation (Go dừng kiểm tra ngay khi gặp điều kiện đúng).

#### 4. Scope biến: Chỉ khai báo khi cần
```go
// ✅ err chỉ dùng trong khối if này
if err := validate(input); err != nil {
    return err
}
// Ngoài này không thấy err nữa → giảm noise
```

#### 5. Ưu tiên switch thay vì chuỗi if-else
Khi so sánh cùng một biến nhiều lần, switch rõ ràng hơn:
```go
switch status {
case StatusActive:
    activate()
case StatusInactive:
    deactivate()
default:
    panic(fmt.Sprintf("unexpected status: %d", status))
}
```

### Thiết kế hàm

| Chủ đề                | Quy tắc                                                                    |
| --------------------- | -------------------------------------------------------------------------- |
| Độ dài & phạm vi      | Ngắn & tập trung                                                           |
| Single Responsibility | Một hàm chỉ làm một việc                                                   |
| Số lượng tham số      | ≤ 4 tham số                                                                |
| Options struct        | Quá 4 tham số → gom vào options struct                                     |
| Thứ tự tham số        | `context.Context` trước, rồi input, rồi output destination                 |
| Naked return          | Chỉ dùng trong hàm rất ngắn (1–3 dòng); hàm dài nên ghi rõ tên biến return |


```go
// ✅ Signature rõ ràng, context đầu tiên
func FetchUser(ctx context.Context, id string) (*User, error)

// ✅ Gom nhiều field liên quan vào struct
func SendEmail(ctx context.Context, msg EmailMessage) error
```


#### Ưu tiên range khi lặp
```go 
// ✅ Tự nhiên, ít lỗi hơn index-based loop
for _, user := range users {
    process(user)
}

// ✅ Go 1.22+: range số nguyên cho counting đơn giản
for i := range 10 {
    fmt.Println(i)
} 
```

### Truyền tham số: Value hay Pointer? [Detail](references\details.md)

| Loại dữ liệu                                    | Cách truyền | Lý do                                              |
| ----------------------------------------------- | ----------- | -------------------------------------------------- |
| Type nhỏ (`string`, `int`, `bool`, `time.Time`) | By value    | Copy rẻ, tránh side effect                         |
| Struct lớn (~128+ bytes)                        | By pointer  | Tránh copy tốn kém                                 |
| Cần sửa giá trị gốc                             | By pointer  | Hàm thay đổi được giá trị bên ngoài                |
| `nil` có ý nghĩa đặc biệt                       | By pointer  | Phân biệt “không có giá trị” vs “giá trị mặc định” |

### Tổ chức code trong file
#### 1. Nhóm các khai báo liên quan lại với nhau: type → constructor → methods
#### 2. Thứ tự ưu tiên:
```go
package doc
imports
constants
types
constructors
methods
helper functions
```
#### 3. Một file, một type chính (nếu type đó có nhiều methods)
#### 4. Blank imports (_ "pkg"): Chỉ dùng trong main hoặc test package, để side effect (init) dễ theo dõi
#### 5. Dot imports (. "pkg"): Không dùng trong library code — làm rối namespace, khó trace nguồn gốc tên
#### 6. Unexport mặc định: Chỉ export khi thật sự cần. Export rồi mà muốn unexport sau là breaking change.

### Xử lý string

| Tình huống                          | Cách làm                           | Lý do                             |
| ----------------------------------- | ---------------------------------- | --------------------------------- |
| Convert đơn giản (`int` ↔ `string`) | `strconv.Itoa`, `strconv.ParseInt` | Nhanh hơn `fmt`                   |
| Format phức tạp                     | `fmt.Sprintf`                      | Linh hoạt, dễ đọc                 |
| In string trong error message       | Dùng `%q`                          | Hiện rõ dấu ngoặc, tránh nhầm lẫn |
| Nối string trong loop               | `strings.Builder`                  | Tránh tạo nhiều object trung gian |
| Nối 2–3 string đơn giản             | `+`                                | Ngắn gọn, đủ dùng                 |

### Type conversions & Generics
- Ưu tiên convert rõ ràng, hẹp scope: int64(val) thay vì any(val)
- Dùng generics thay vì any khi có thể xác định type cụ thể:
```go 
// ✅ Type-safe, compiler check được
func Contains[T comparable](slice []T, target T) bool

// ❌ Mất type safety, dễ lỗi runtime
func Contains(slice []any, target any) bool
```

### Triết lý thiết kế (Philosophy)

1. **"Copy một chút còn hơn phụ thuộc vào thư viện ngoài"**
→ Đừng import thư viện chỉ để làm việc đơn giản mà stdlib đã làm được.
2. **Dùng slices và maps từ stdlib**
→ Cần filter/group-by/chunk phức tạp hơn? Dùng github.com/samber/lo.
3. **"Reflection không bao giờ rõ ràng"**
→ Tránh reflect trừ khi thật sự cần (ví dụ: framework, ORM).
4. **Đừng abstract quá sớm**
→ Chờ pattern lặp lại 3 lần rồi hãy extract thành hàm/utils.
5. **Giữ public surface nhỏ nhất có thể**
→ Mỗi tên export là một "lời hứa" với user của package — khó thay đổi sau này.

### Parallelizing Code Style Reviews
Khi review code style trên codebase lớn, có thể chạy song song tối đa 5 sub-agents, mỗi agent phụ trách một mảng:

| Sub-agent | Nhiệm vụ                                                       |
| --------- | -------------------------------------------------------------- |
| 1         | Control flow: early return, nesting, switch vs if-else         |
| 2         | Function design: số tham số, naked return, `context` placement |
| 3         | Variable declarations: `:=` vs `var`, slice/map init           |
| 4         | String handling: `strconv` vs `fmt`, `strings.Builder`         |
| 5         | Code organization: grouping, import style, export strategy     |

### Enforce với Linters
Nhiều quy tắc trên đã được linter hỗ trợ tự động:
- **Format**: gofmt, gofumpt, goimports
- **Style**: gocritic, revive, wsl_v5
→ Xem skill `samber/cc-skills-golang@golang-lint` để config chi tiết.

### Liên kết liên quan
- @golang-naming: Quy ước đặt tên biến, hàm, type (ErrNotFound, NewUser...)
- @golang-structs-interfaces: Khi nào dùng pointer receiver, thiết kế interface
- @golang-design-patterns: Functional options, builder pattern, constructor
- @golang-lint: Config linter để enforce style tự động
- @golang-continuous-integration: Tích hợp AI review vào CI/CD