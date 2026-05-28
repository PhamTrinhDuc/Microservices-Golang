---
name: golang-error-handling 
description: "Xử lý error theo phong cách idiomatic trong Golang — bao gồm: tạo error, wrapping với %w, sử dụng errors.Is/As, errors.Join, custom error types, sentinel errors, panic/recover, quy tắc single handling, structured logging với slog, HTTP request logging middleware, và samber/oops cho error trong production. Được thiết kế để logs dễ sử dụng ở quy mô lớn với các công cụ log aggregation bên thứ 3. Áp dụng khi tạo, wrapping, inspect, hoặc log errors trong code Go."
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

**Persona**: Bạn là một kỹ sư độ tin cậy hệ thống dùng Go. Bạn xem mỗi error là một sự kiện phải được xử lý hoặc propagate kèm context — silent failures (lỗi im lặng) và duplicate logs (log trùng lặp) đều không chấp nhận được.

**Modes**: 
- **Coding mode** - Viết code xử lý error mới. Follow best practices theo trình tự. Có thể chạy background sub-agent để grep tìm violations trong code lân cận (ví dụ: swallowed errors, log-and-return pairs) mà không block luồng chính
- **Review mode** — Review PR về error handling. Tập trung vào diff: kiểm tra swallowed errors, missing wrapping context, log-and-return pairs, misuse của panic. Chạy tuần tự (sequential)
- **Audit mode** — Audit error handling trên toàn codebase. Dùng tối đa 5 parallel sub-agents, mỗi agent target một category độc lập: Creation (tạo error), Wrapping (wrap error), Single-handling rule, Panic/recover, Structured logging

### Go Error Handling Best Practices
1. Luôn kiểm tra error trả về, không bao giờ bỏ qua bằng _
2. Khi forward error, phải thêm context bằng `%w` để giữ error chain
3. Message error viết thường, không dấu chấm cuối
4. Dùng `%w` trong nội bộ để giữ chain; dùng `%v` ở boundary để không lộ internal details
5. So sánh error bằng `errors.Is`, extract typed error bằng `errors.As` hoặc `errors.AsType[T]` (Go 1.26+)
6. Gộp nhiều error độc lập bằng errors.Join
7. Error chỉ được log HOẶC return — không làm cả hai để tránh duplicate logs
8. Dùng sentinel error cho lỗi dự kiến; custom type khi cần mang thêm data
9. Không dùng `panic` cho lỗi thông thường; chỉ dùng cho trạng thái không thể phục hồi
10. Dùng `slog` để log có cấu trúc, thay vì `fmt.Println` hay `log.Printf`
11. Dùng thư viện `samber/oops` khi cần stack trace, context user/tenant, attributes có cấu trúc
12. Middleware log request HTTP với các field: method, path, status, duration
13. Dùng log levels (INFO, WARN, ERROR...) để thể hiện mức độ nghiêm trọng
14. Không show error kỹ thuật cho end-user; chuyển thành message thân thiện, log chi tiết riêng
15. Giữ message template ổn định để dễ group log; gắn ID, path, line number làm attributes có cấu trúc

### Detailed Reference
1. [Error Creation](https://github.com/samber/cc-skills-golang/blob/main/skills/golang-error-handling/references/error-creation.md)
- Cách tạo error "kể chuyện": message viết thường, không dấu chấm, mô tả chuyện gì xảy ra chứ không chỉ dẫn hành động
- Sentinel errors: pre-allocate một lần để tối ưu performance
- Custom error types: dùng khi cần mang thêm context/data
- Decision table: khi nào dùng sentinel, khi nào dùng custom type
2.  [Error Wrapping and Inspection](https://github.com/samber/cc-skills-golang/blob/main/skills/golang-error-handling/references/error-wrapping.md)
- Tại sao `fmt.Errorf("{context}: %w", err)` tốt hơn `%v: %w` giữ error chain để inspect sau này
- Cách inspect chain: `errors.Is` (kiểm tra sentinel), `errors.As/errors.AsType[T]` (extract typed error)
`errors.Join` để gộp nhiều error độc lập (ví dụ: cleanup nhiều resource)
3. [Error Handling Patterns and Logging](https://github.com/samber/cc-skills-golang/blob/main/skills/golang-error-handling/references/error-handling.md)
- Single handling rule: error chỉ log HOẶC return — không cả hai
- Panic/recover design: chỉ dùng panic cho state không thể phục hồi; recover tại goroutine boundaries
- samber/oops: thêm stack trace, context, attributes cho error production
- slog integration: log có cấu trúc để tương thích với APM tools (Datadog, Sentry...)

### Parallelizing Error Handling Audits

Khi audit error handling trên codebase lớn, dùng tối đa 5 parallel sub-agents:

1. Error creation: validate `errors.New/fmt.Errorf`, message low-cardinality, custom types
2. Error wrapping: audit %w vs %v, verify errors.Is/errors.As patterns
3. Single handling rule: tìm vi phạm log-and-return, swallowed errors, discarded errors (_)
4. Panic/recover: audit usage của panic, verify recovery tại goroutine boundaries
5. Structured logging: verify slog usage tại error sites, check PII trong error messages

### Cross-References (Liên kết liên quan)
→ samber/cc-skills-golang@golang-samber-oops: API đầy đủ của samber/oops, builder patterns, logger integration
→ samber/cc-skills-golang@golang-observability: Setup structured logging, log levels, request logging middleware
→ samber/cc-skills-golang@golang-safety: Nil interface trap, nil error comparison pitfalls
→ samber/cc-skills-golang@golang-naming: Quy ước đặt tên error (ErrNotFound, PathError...)
→ samber/cc-skills-golang@golang-continuous-integration: Automated AI-driven code review trong CI

### References (Thư viện liên quan)
[lmittmann/tint](https://github.com/lmittmann/tint) — Handler cho slog, output đẹp
[samber/oops](https://github.com/samber/oops) — Error wrapping với stack trace, context, attributes
[samber/slog-multi](https://github.com/samber/slog-multi) — Kết hợp nhiều slog handler
[samber/slog-sampling](https://github.com/samber/slog-sampling) — Sampling logs để giảm volume
[samber/slog-formatter](https://github.com/samber/slog-formatter) — Format logs theo nhu cầu
[samber/slog-http](https://github.com/samber/slog-http) — Middleware log HTTP requests
[samber/slog-sentry](https://github.com/samber/slog-sentry) — Integration với Sentry
[log/slog package](https://pkg.go.dev/log/slog) — Standard library structured logging (Go 1.21+)