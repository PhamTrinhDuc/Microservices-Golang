---
name: golang-testing 
description: Viết test Go chuẩn production — bao gồm: table-driven tests, testify (suites & mocks), chạy test song song, fuzzing, fixtures, phát hiện goroutine leak với goleak, snapshot testing, đo coverage, integration tests, và quy tắc đặt tên test. Dùng khi viết/review test, chọn cách tiếp cận testing, setup CI cho Go test, hoặc debug test flaky/chạy chậm.
user-invocable: true 
license: MIT 
combatibility: thiết kế cho Claude code hoặc các AI coding agents tương tự và cho các dự án sử dụng Golang 
metadata:
  author: samber
  version: "1.2.0"
  openclaw:
    emoji: "⚠"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent
---

## Persona
Bạn là Go engineer xem test như "specification chạy được". Bạn viết test để ràng buộc hành vi của code, không phải để "đạt chỉ tiêu coverage". Test tốt là test mô tả rõ "code này phải làm gì", chứ không phải test viết cho có. Khi code thay đổi, test sẽ "kêu" nếu hành vi bị phá vỡ.

## Chế độ hoạt động (Modes)

| Mode        | Khi nào dùng                                  | Cách làm                                                                                                                       |
| ----------- | --------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| Write mode  | Viết test mới cho code có sẵn hoặc code mới   | Dùng `gotests` để tạo khung table-driven test trước, rồi bổ sung edge cases và error paths                                     |
| Review mode | Review PR có thay đổi test                    | Kiểm tra: coverage cho behavior mới, chất lượng assertion, cấu trúc table-driven, dấu hiệu flaky test                          |
| Audit mode  | Audit toàn bộ test suite xem có lỗ hổng không | Chạy song song 3 sub-agents: (1) chất lượng unit test, (2) isolation của integration test, (3) goroutine leak & race condition |
| Debug mode  | Test bị fail hoặc flaky                       | Reproduce lỗi → isolate assertion bị fail → trace root cause trong code hoặc test setup                                        |


## 1. Quy tắc best practices 

| STT | Quy tắc                                                           | Giải thích ngắn                                                                     |
| --- | ----------------------------------------------------------------- | ----------------------------------------------------------------------------------- |
| 1   | Table-driven tests MUST dùng named subtests (`t.Run`)             | Mỗi test case phải có name để dễ debug và chạy riêng lẻ                             |
| 2   | Integration tests MUST dùng build tags (`//go:build integration`) | Tách biệt với unit test, tránh chạy nhầm trong CI nhanh                             |
| 3   | Tests MUST NOT phụ thuộc thứ tự chạy                              | Mỗi test phải chạy độc lập, không được giả định test trước đã chạy                  |
| 4   | Independent tests SHOULD dùng `t.Parallel()`                      | Chạy song song để test suite chạy nhanh hơn                                         |
| 5   | NEVER test implementation details                                 | Chỉ test behavior quan sát được và public API, không test “code bên trong viết sao” |
| 6   | Package có goroutine SHOULD dùng `goleak.VerifyTestMain`          | Phát hiện goroutine bị rò rỉ — lỗi khó debug trong production                       |
| 7   | Dùng `testify` làm helper, không thay thế stdlib                  | `assert`/`require` tiện nhưng đừng lạm dụng, vẫn ưu tiên `testing` package          |
| 8   | Mock interfaces, không mock concrete types                        | Mock chỗ interface được dùng, không mock struct cụ thể                              |
| 9   | Unit test phải nhanh (<1ms), integration test dùng build tags     | Test chậm = CI chậm = developer ngại chạy test                                      |
| 10  | Chạy test với `-race` trong CI                                    | Phát hiện race condition sớm, trước khi lên production                              |
| 11  | Include examples như executable documentation                     | `ExampleXxx()` vừa là docs vừa là test — 2 trong 1                                  |


## 2. Cấu trúc & tổ chức test

#### Quy ước tên file

```go 
// package_test.go — test trong cùng package (white-box, access được unexported)
package mypackage

// mypackage_test.go — test trong package riêng (black-box, chỉ dùng public API)
package mypackage_test
```

#### Khi nào dùng cái nào?
- `white-box`: Cần test logic nội bộ, unexported function → dùng cùng package.
- `black-box`: Muốn test như user thực tế dùng → dùng package _test, chỉ import public API.

#### Quy ước đặt tên function test

```go
func TestAdd(t *testing.T) { ... }               // test function
func TestMyStruct_MyMethod(t *testing.T) { ... } // test method
func BenchmarkAdd(b *testing.B) { ... }          // benchmark performance
func ExampleAdd() { ... }                        // example (chạy được, verify bởi go test)
func FuzzAdd(f *testing.F) { ... }               // fuzz test (tự sinh input để tìm bug)
```

#### Table-Driven Tests (Cách viết test "chuẩn Go")
Đây là pattern idiomatic nhất trong Go để test nhiều scenario cùng lúc. Luôn đặt tên cho mỗi test case.
```go 
func TestCalculatePrice(t *testing.T) {
    tests := []struct {
        name      string
        quantity  int
        unitPrice float64
        expected  float64
    }{
        {
            name:      "single item",
            quantity:  1,
            unitPrice: 10.0,
            expected:  10.0,
        },
        {
            name:      "bulk discount - 100 items",
            quantity:  100,
            unitPrice: 10.0,
            expected:  900.0, // giảm 10%
        },
        {
            name:      "zero quantity",
            quantity:  0,
            unitPrice: 10.0,
            expected:  0.0,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := CalculatePrice(tt.quantity, tt.unitPrice)
            if got != tt.expected {
                t.Errorf("CalculatePrice(%d, %.2f) = %.2f, want %.2f",
                    tt.quantity, tt.unitPrice, got, tt.expected)
            }
        })
    }
}
```

#### Unit Tests: Nhanh, cô lập, deterministic
- **Nhanh**: <1ms/test — nếu chậm, xem lại có gọi DB/network không.
- **Cô lập**: Không phụ thuộc external service (DB, API, file system).
- **Deterministic**: Chạy 100 lần, kết quả giống nhau 100 lần.


#### Testing HTTP Handlers
Dùng `httptest` để test handler, kết hợp với table-driven pattern.
```go 
func TestHandler(t *testing.T) {
    tests := []struct {
        name           string
        method         string
        url            string
        body           string
        expectedStatus int
    }{
        {"valid request", "POST", "/api/users", `{"name":"Alice"}`, http.StatusCreated},
        {"missing body", "POST", "/api/users", "", http.StatusBadRequest},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest(tt.method, tt.url, strings.NewReader(tt.body))
            w := httptest.NewRecorder()
            
            handler.ServeHTTP(w, req)
            
            if w.Code != tt.expectedStatus {
                t.Errorf("status = %d, want %d", w.Code, tt.expectedStatus)
            }
        })
    }
}
```

#### Phát hiện goroutine leak với `goleak`
Code dùng goroutine mà không cleanup kỹ sẽ bị "rò" — tốn memory, khó debug. Dùng `go.uber.org/goleak` để bắt lỗi này trong test.
```go 
// Trong file *_test.go
func TestMain(m *testing.M) {
    goleak.VerifyTestMain(m)
}
```
##### Bỏ qua goroutine "biết trước" (known leaks)
```go 
func TestMain(m *testing.M) {
    goleak.VerifyTestMain(m,
        goleak.IgnoreCurrent(), // bỏ qua goroutine đang chạy khi test bắt đầu
    )
}
```

##### Hoặc check per-test
```go 
func TestWorkerPool(t *testing.T) {
    defer goleak.VerifyNone(t)
    // ... code test ...
}
```

#### `testing/synctest` — Test goroutine deterministically (Go 1.25+)

`synctest` giúp test code concurrent không phụ thuộc vào timing thật — time sẽ "nhảy" chỉ khi tất cả goroutine block, giúp test ổn định, reproducible.
##### Khi nào dùng synctest?
- Test code có `time.Sleep`, `time.After`, `time.Ticker`
- Test bị flaky do race condition về timing
- Cần reproduce bug liên quan đến timeout/deadline

```go 
func TestContextTimeout(t *testing.T) {
    synctest.Test(t, func(t *testing.T) {
        const timeout = 5 * time.Second

        ctx, cancel := context.WithTimeout(t.Context(), timeout)
        defer cancel()

        // Trước khi timeout: context vẫn OK
        time.Sleep(timeout - time.Nanosecond)
        synctest.Wait()
        if err := ctx.Err(); err != nil {
            t.Fatalf("before timeout: %v", err)
        }

        // Sau khi timeout: context phải báo DeadlineExceeded
        time.Sleep(time.Nanosecond)
        synctest.Wait()
        if err := ctx.Err(); err != context.DeadlineExceeded {
            t.Fatalf("after timeout: got %v, want DeadlineExceeded", err)
        }
    })
}
```
> ##### Lưu ý version:
> - Go 1.25+: Dùng synctest.Test
> - Go 1.24 (experimental): Dùng synctest.Run (cũ, không khuyến khích)
> - Nếu module target Go 1.24 + GOEXPERIMENT=synctest, chỉ dùng API cũ làm fallback.


#### Test Timeouts — Tránh test bị treo
Test có thể hang do deadlock, network timeout, bug infinite loop. Dùng helper timeout để fail nhanh + báo vị trí caller.
```go 
// Helper trong test_helpers.go
func timeout(t *testing.T, d time.Duration) func() {
    done := make(chan struct{})
    go func() {
        time.Sleep(d)
        close(done)
    }()
    select {
    case <-done:
        _, file, line, _ := runtime.Caller(2)
        panic(fmt.Sprintf("test timeout at %s:%d", file, line))
    case <-t.Context().Done():
        return
    }
}
```
>  Mẹo: Đặt timeout ngắn (5-10s) cho unit test, dài hơn cho integration test.

#### Benchmarks — Đo performance, phát hiện regression
>💡 Benchmark không phải test "đúng/sai", mà test "nhanh/chậm". Dùng để:
> - So sánh 2 cách implement
> - Phát hiện code mới làm chậm hệ thống
> - Tune performance trước khi deploy
```go 
func BenchmarkStringConcatenation(b *testing.B) {
    b.Run("plus-operator", func(b *testing.B) {
        for b.Loop() {  // Go 1.24+: dùng b.Loop() thay vì for i := 0; i < b.N; i++
            result := "a" + "b" + "c"
            _ = result
        }
    })

    b.Run("strings.Builder", func(b *testing.B) {
        for b.Loop() {
            var builder strings.Builder
            builder.WriteString("a")
            builder.WriteString("b")
            builder.WriteString("c")
            _ = builder.String()
        }
    })
}
```
##### Benchmark với nhiều input sizes
```go 
func BenchmarkFibonacci(b *testing.B) {
    sizes := []int{10, 20, 30}
    for _, size := range sizes {
        b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
            b.ReportAllocs()  // báo cáo số allocation
            for b.Loop() {
                Fibonacci(size)
            }
        })
    }
}
```
> Go 1.24+: Dùng b.Loop() thay vì b.N — cú pháp mới, rõ nghĩa hơn. Chỉ dùng b.N khi module target Go <1.24 hoặc giữ code cũ.

#### Go 1.26+: Test Artifacts với `ArtifactDir()`
Khi test/benchmark/fuzz cần lưu file để inspect sau (ví dụ: golden file, rendered output), dùng `ArtifactDir()` thay vì path tự tạo.
```go 
func TestRenderGoldenArtifact(t *testing.T) {
    dir := t.ArtifactDir()  // Go 1.26+: trả về path chuẩn để lưu artifact

    out := filepath.Join(dir, "rendered.json")
    if err := os.WriteFile(out, renderedBytes, 0o644); err != nil {
        t.Fatal(err)
    }

    t.Logf("artifact written: %s", out)
}
```
> 💡 Lợi ích:
> - Path nhất quán, dễ tìm trong CI/CD
> - Không pollute repo với file tạm
> - Hỗ trợ upload artifact tự động trong GitHub Actions/GitLab CI

#### Parallel Tests — Chạy test song song cho nhanh
Dùng `t.Parallel()` để các test độc lập chạy concurrently.
```go 
func TestParallelOperations(t *testing.T) {
    tests := []struct {
        name string
        data []byte
    }{
        {"small data", make([]byte, 1024)},
        {"medium data", make([]byte, 1024*1024)},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()  // ✅ Quan trọng: gọi trong subtest, không gọi ở test cha
            is := assert.New(t)

            result := Process(tt.data)
            is.NotNil(result)
        })
    }
}
```
#### Fuzzing — Tự sinh input để tìm bug
Fuzz test tự động sinh input ngẫu nhiên, giúp phát hiện edge case mà developer không nghĩ tới.
```go
func FuzzReverse(f *testing.F) {
    // Seed với vài input mẫu
    f.Add("hello")
    f.Add("")
    f.Add("a")

    f.Fuzz(func(t *testing.T, input string) {
        reversed := Reverse(input)
        doubleReversed := Reverse(reversed)
        if input != doubleReversed {
            t.Errorf("Reverse(Reverse(%q)) = %q, want %q", input, doubleReversed, input)
        }
    })
}
```
> 💡 Khi nào nên fuzz?
> - Hàm xử lý string, parser, encoder/decoder
> - Hàm có logic phức tạp, nhiều branch
> - Code nhận input từ user/network (risk cao)

#### Examples — Documentation chạy được
`ExampleXxx()` vừa là docs (hiển thị trong `godoc`), vừa là test (verify bởi `go test`).
```go 
func ExampleCalculatePrice() {
    price := CalculatePrice(100, 10.0)
    fmt.Printf("Price: %.2f\n", price)
    // Output: Price: 900.00
}

func ExampleCalculatePrice_singleItem() {
    price := CalculatePrice(1, 25.50)
    fmt.Printf("Price: %.2f\n", price)
    // Output: Price: 25.50
}
```
> 💡 Lưu ý:
> - Comment // Output: phải khớp chính xác output thực tế
> - Nếu output thay đổi, test fail → docs luôn cập nhật

#### Code Coverage — Đo % code được test
```go 
# Generate coverage report
go test -coverprofile=coverage.out ./...

# Xem dạng HTML (màu xanh/đỏ rõ ràng)
go tool cover -html=coverage.out

# Xem coverage theo function
go tool cover -func=coverage.out

# Lấy % total
go tool cover -func=coverage.out | grep total
```

#### Integration Tests — Test với dependency thật
Dùng build tag để tách integration test ra khỏi unit test.
```go 
//go:build integration

package mypackage

func TestDatabaseIntegration(t *testing.T) {
  db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
  if err != nil {
      t.Fatal(err)
  }
  defer db.Close()

  // Test operation thật với DB

  # Chạy riêng integration test
  go test -tags=integration ./...
}
```
> 💡 Setup integration test:
> Dùng Docker Compose để spin up DB/cache/service phụ thuộc
> Cleanup sau mỗi test (truncate table, delete temp file)
> Mock time/network nếu cần deterministic

#### Mocking — Mock interface, không mock concrete type
```go 
// ✅ Đúng: định nghĩa interface nơi dùng, mock implementation
type UserRepository interface {
    FindByID(id string) (*User, error)
}

// Trong test:
type mockUserRepo struct {
    mock.Mock
}

func (m *mockUserRepo) FindByID(id string) (*User, error) {
    args := m.Called(id)
    return args.Get(0).(*User), args.Error(1)
}
```

#### Quick Reference — Lệnh test hữu ích
```go 
# Chạy tất cả test
go test ./...

# Chạy test theo tên chính xác
go test -run TestName ./...

# Chạy subtest trong test
go test -run TestName/subtest ./...

# Chạy nhiều test (regex OR)
go test -run 'Test(Add|Sub)' ./...

# Test bắt đầu bằng chữ hoa
go test -run 'Test[A-Z]' ./...

# Test chứa substring
go test -run '.*Validation.*' ./...

# Chạy tất cả subtest của TestName
go test -run TestName/. ./...

# Lọc subtest theo tên (unit/integration)
go test -run '/(unit|integration)' ./...

# Race detection (phát hiện data race)
go test -race ./...

# Coverage summary
go test -cover ./...

# Benchmark + báo cáo memory alloc
go test -bench=. -benchmem ./...

# Fuzz test
go test -fuzz=FuzzName ./...

# Integration tests (với build tag)
go test -tags=integration ./...
```

#### Liên kết liên quan
- @golang-stretchr-testify: API chi tiết của testify (assert, require, mock, suite)
- @golang-database: Pattern test integration với DB (SQL schema, migration, cleanup)
- @golang-concurrency: Phát hiện goroutine leak, race condition với goleak
- @golang-continuous-integration: Config CI test, GitHub Actions workflow
- @golang-lint: Config testifylint, paralleltest trong golangci-lint