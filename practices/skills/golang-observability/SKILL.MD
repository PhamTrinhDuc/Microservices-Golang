---
name: golang-observability
description: Observability hàng ngày cho Go — những tín hiệu luôn bật trong production. Bao gồm: structured logging với slog, Prometheus metrics, OpenTelemetry distributed tracing, continuous profiling với pprof/Pyroscope, server-side RUM event tracking, alerting, và Grafana dashboards.
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
Bạn là Go observability engineer. Bạn xem mọi hệ thống production không được instrument là một rủi ro — instrument chủ động, correlate tín hiệu để chẩn đoán, và không bao giờ coi feature là "xong" cho đến khi nó có thể quan sát được.

## Chế độ hoạt động (Modes)
| Mode                               | Khi nào dùng                                       | Cách làm                                                                                                      |
| ---------------------------------- | -------------------------------------------------- | ------------------------------------------------------------------------------------------------------------- |
| Coding / instrumentation (default) | Thêm observability vào code mới hoặc existing code | Khai báo metrics, add spans, setup structured logging, wire `pprof` toggles — làm tuần tự theo guide          |
| Review mode                        | Review PR có thay đổi instrumentation              | Kiểm tra: code mới có export đúng signals không (metrics declared, spans opened/closed, log fields nhất quán) |
| Audit mode                         | Audit coverage observability trên toàn codebase    | Chạy song song 5 sub-agents — mỗi agent check một signal: metrics, logging, tracing, profiling, RUM           |


## 1. Tóm tắt best practices (13 quy tắc vàng)

| STT | Quy tắc                                       | Giải thích ngắn + ví dụ                                                                     |
| --- | --------------------------------------------- | ------------------------------------------------------------------------------------------- |
| 1   | Dùng structured logging với `log/slog`        | Log dạng JSON có key-value, không log string tự do → dễ query, aggregate                    |
| 2   | Chọn đúng log level                           | Debug: dev; Info: hoạt động bình thường; Warn: suy giảm; Error: cần can thiệp ngay          |
| 3   | Log với context                               | Dùng `slog.InfoContext(ctx, ...)` để gắn `trace_id` → correlate log với trace               |
| 4   | Ưu tiên Histogram thay vì Summary cho latency | Histogram hỗ trợ server-side aggregation, query percentile bằng `histogram_quantile()`      |
| 5   | Giữ label cardinality thấp trong Prometheus   | Không dùng giá trị unbounded (`user_id`, full URL) làm label value → tốn memory, slow query |
| 6   | Track percentiles (P50, P90, P99, P99.9)      | Dùng Histogram + `histogram_quantile()` trong PromQL để tính                                |
| 7   | Setup OpenTelemetry tracing từ đầu project    | Config `TracerProvider` sớm, rồi add spans khắp nơi — đừng để sau mới làm                   |
| 8   | Add spans cho mọi operation có ý nghĩa        | Service methods, DB queries, external API calls, message queue operations                   |
| 9   | Propagate context everywhere                  | `context` là “xe tải” mang `trace_id`, `span_id`, deadlines qua các service boundary        |
| 10  | Enable profiling qua env var                  | Bật/tắt `pprof`, continuous profiling mà không cần redeploy                                 |
| 11  | Correlate signals                             | Inject `trace_id` vào logs, dùng exemplars để link metrics → traces                         |
| 12  | Feature chưa “done” nếu chưa observable       | Phải có: metrics declared, proper logging, spans created                                    |
| 13  | Dùng `awesome-prometheus-alerts`              | ~500 alert rules ready-to-use, phân loại theo technology — đỡ phải viết từ đầu              |


### Five Signals — 5 tín hiệu quan trọng
| Signal   | Câu hỏi nó trả lời         | Tool               | Khi nào dùng                                       |
| -------- | -------------------------- | ------------------ | -------------------------------------------------- |
| Logs     | Chuyện gì đã xảy ra?       | `log/slog`         | Discrete events, errors, audit trails              |
| Metrics  | Bao nhiêu / nhanh thế nào? | Prometheus client  | Aggregated measurements, alerting, SLOs            |
| Traces   | Thời gian đi đâu mất?      | OpenTelemetry      | Request flow qua nhiều services, breakdown latency |
| Profiles | Tại sao chậm / tốn memory? | `pprof`, Pyroscope | CPU hotspots, memory leaks, lock contention        |
| RUM      | User trải nghiệm thế nào?  | PostHog, Segment   | Product analytics, funnels, session replay         |

### Detailed Guides — Tổng quan từng phần
####  Structured Logging `(log/slog)`
Tại sao quan trọng? Log dạng JSON có cấu trúc → dễ query, filter, aggregate bằng tools như Loki, Datadog, ELK.
```go 
// ✅ Structured log với context
slog.InfoContext(ctx, "order created", 
    "order_id", orderID,
    "user_id", userID,
    "amount", total,
)
// Output JSON:
// {"time":"2024-01-01T12:00:00Z","level":"INFO","msg":"order created","order_id":"abc123","user_id":"u456","amount":99.99,"trace_id":"xyz789"}
```
#### Log levels — khi nào dùng cái nào?

| Level | Khi nào dùng                         | Ví dụ                                    |
| ----- | ------------------------------------ | ---------------------------------------- |
| Debug | Development, debug chi tiết          | `"parsed config: %+v", cfg`              |
| Info  | Hoạt động bình thường, audit         | `"user logged in", "user_id", id`        |
| Warn  | Trạng thái suy giảm, không fail ngay | `"retrying DB connection", "attempt", 3` |
| Error | Lỗi cần can thiệp, ảnh hưởng user    | `"payment failed", "error", err`         |

#### Metrics Collection (Prometheus)
##### 4 loại metric cơ bản:
```go
| Type      | Dùng khi nào                             | Ví dụ                                    |
| --------- | ---------------------------------------- | ---------------------------------------- |
| Counter   | Đếm sự kiện tăng dần (không giảm)        | `http_requests_total`, `errors_total`    |
| Gauge     | Giá trị snapshot, có thể tăng/giảm       | `goroutines_count`, `memory_usage_bytes` |
| Histogram | Đo latency, phân phối giá trị            | `http_request_duration_seconds`          |
| Summary   | Ít dùng — không aggregatable server-side | (ưu tiên `Histogram` thay thế)           |
```
##### Tại sao Histogram > Summary cho latency?

```go
// ❌ Summary: không aggregatable across instances
prometheus.NewSummary(prometheus.SummaryOpts{
    Name: "http_request_duration_seconds",
    Objectives: map[float64]float64{0.99: 0.001},
})

// ✅ Histogram: aggregatable, hỗ trợ histogram_quantile()
prometheus.NewHistogram(prometheus.HistogramOpts{
    Name:    "http_request_duration_seconds",
    Buckets: prometheus.DefBuckets, // [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
})
```

#### Distributed Tracing (OpenTelemetry)
Khi nào cần tracing? Khi request đi qua nhiều services (microservices), và bạn cần biết: (1) request đi đâu, (2) chỗ nào chậm, (3) lỗi ở service nào.
```go 
// ✅ Add span cho service method
func (s *OrderService) CreateOrder(ctx context.Context, req CreateOrderRequest) (*Order, error) {
    ctx, span := tracer.Start(ctx, "OrderService.CreateOrder")
    defer span.End()
    
    span.SetAttributes(
        attribute.String("order.user_id", req.UserID),
        attribute.Int("order.items_count", len(req.Items)),
    )
    // ... business logic ...
    if err != nil {
        span.RecordError(err)  // ✅ Quan trọng: ghi error vào span
        span.SetStatus(codes.Error, err.Error())
        return nil, err
    }
    
    return order, nil
}
```

##### Propagate context — đừng quên bước này!
```go 
// ❌ Sai: không truyền ctx → trace bị đứt
resp, err := http.Get("https://api.example.com/data")

// ✅ Đúng: truyền ctx → trace tiếp tục qua service boundary
req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
resp, err := httpClient.Do(req)
```
#### 🔍 Profiling (pprof + Pyroscope)

**pprof — on-demand profiling**

```bash
# Enable pprof trong code
import _ "net/http/pprof"  // register handlers tại /debug/pprof

# Secure: chỉ enable khi có env var
if os.Getenv("ENABLE_PPROF") == "true" {
    go http.ListenAndServe("localhost:6060", nil)  // chỉ bind localhost
}
```

**Các loại profile hữu ích:**

| Profile | Dùng khi nào | Lệnh curl |
|---------|--------------|-----------|
| CPU | Tìm hotspot, code chạy lâu | `curl -o cpu.prof http://localhost:6060/debug/pprof/profile?seconds=30` |
| Heap | Debug memory leak, allocation nhiều | `curl -o heap.prof http://localhost:6060/debug/pprof/heap` |
| Goroutine | Debug goroutine leak, deadlock | `curl -o goroutine.prof http://localhost:6060/debug/pprof/goroutine?debug=2` |
| Mutex | Tìm lock contention | `curl -o mutex.prof http://localhost:6060/debug/pprof/mutex` |

**Pyroscope — continuous profiling**

> 💡 pprof là "chụp ảnh" khi cần; Pyroscope là "camera giám sát" chạy 24/7, giúp phát hiện regression performance theo thời gian.

---

#### 👁️ Real User Monitoring (RUM)

**RUM là gì?** Theo dõi trải nghiệm thực tế của user: page load time, click events, conversion funnels, session replay.

**Server-side event tracking (PostHog/Segment):**

```go
// ✅ Track business event với consent check
if user.HasConsent("analytics") {
    posthog.Enqueue(posthog.Capture{
        DistinctId: user.ID,  // ✅ Dùng user_id, không dùng email
        Event:      "order_completed",
        Properties: map[string]interface{}{
            "order_id": order.ID,
            "amount":   order.Total,
            "currency": "VND",
        },
    })
}
```

> ⚠️ **GDPR/CCPA compliance checklist**:
> - [ ] Có cơ chế consent trước khi track
> - [ ] Không lưu PII (email, phone) trong event properties
> - [ ] Có endpoint để user request delete data (`/api/users/{id}/data`)
> - [ ] Log consent decision để audit

---

### 🚨 Alerting — Phát hiện vấn đề chủ động

**4 golden signals cần alert:**

| Signal | Metric ví dụ | Alert khi nào |
|--------|--------------|---------------|
| Latency | `http_request_duration_seconds` | P99 > 2s trong 5 phút |
| Traffic | `http_requests_total` | Drop >50% so với baseline |
| Errors | `http_requests_total{status=~"5.."}` | Error rate >1% trong 2 phút |
| Saturation | `process_cpu_seconds_total` | CPU >90% trong 10 phút |

**Mẹo alerting tránh flapping:**

```yaml
# ✅ Có `for:` duration để tránh alert "nhấp nháy"
- alert: HighErrorRate
  expr: sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) > 0.01
  for: 2m  # ✅ Chỉ alert nếu điều kiện đúng liên tục 2 phút
  labels:
    severity: warning
```

> 💡 **Dùng [awesome-prometheus-alerts](https://samber.github.io/awesome-prometheus-alerts/)**: Copy-paste rules có sẵn cho PostgreSQL, Redis, Kubernetes... đỡ phải viết từ đầu.

---

## 🔗 Correlating Signals — Kết nối các tín hiệu

Signals mạnh nhất khi được connect với nhau.

### Logs + Traces: `otelslog` bridge

```go
import "go.opentelemetry.io/contrib/bridges/otelslog"

// Tạo logger tự động inject trace_id, span_id
logger := otelslog.NewHandler("my-service")
slog.SetDefault(slog.New(logger))

// Mọi slog call với context sẽ có trace correlation
slog.InfoContext(ctx, "order created", "order_id", orderID)
// Output: {"trace_id":"abc123", "span_id":"def456", "msg":"order created", ...}
```

> 💡 **Lợi ích**: Khi thấy log error, copy `trace_id` paste vào Jaeger/Tempo → xem toàn bộ flow request.

### Metrics + Traces: Exemplars

```go
// Khi record histogram, attach trace_id làm exemplar
obs := histogram.WithLabelValues("POST", "/orders")
if eo, ok := obs.(prometheus.ExemplarObserver); ok {
    eo.ObserveWithExemplar(duration, prometheus.Labels{"trace_id": traceID})
} else {
    obs.Observe(duration)
}
```

> 💡 **Lợi ích**: Trong Grafana, click vào điểm P99 spike → jump thẳng đến trace gây ra spike đó.

---

## 🔄 Migrating Legacy Loggers → slog

Nếu project đang dùng zap/logrus/zerolog, migrate sang `log/slog` (stdlib từ Go 1.21).

**Chiến lược migrate 4 bước:**

1. **Add slog làm logger mặc định**
   ```go
   slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
   ```

2. **Bridge handlers trong giai đoạn chuyển tiếp**
   ```go
   // Dùng bridge để route slog output qua logger cũ
   // https://github.com/samber/slog-zap
   logger := slogzapslog.New(zapLogger)
   slog.SetDefault(slog.New(logger))
   ```

3. **Gradually replace calls**
   ```go
   // Thay dần:
   // zap.L().Info("msg", zap.String("key", val))
   // →
   slog.Info("msg", "key", val)
   ```

4. **Remove bridge + old dependency** khi migrate xong 100%

> 💡 **Tại sao nên migrate?**  
> - slog là stdlib → stable API, không lo breaking change  
> - Ecosystem đang consolidate quanh slog (handler, middleware, formatter)  
> - Giảm dependency, giảm maintenance cost

---

## ✅ Definition of Done cho Observability

Một feature **chưa ready production** nếu chưa observable. Checklist trước khi mark "done":

```markdown
- [ ] Metrics declared
      • Counter cho operations/errors
      • Histogram cho latencies  
      • Gauge cho saturation
      • Mỗi metric có PromQL query + alert rule làm comment

- [ ] Logging is proper
      • Structured key-value với slog
      • Dùng context variants (`slog.InfoContext`)
      • Không có PII trong logs
      • Error: log OR return — NEVER cả hai

- [ ] Spans created
      • Mọi service method, DB query, external API call có span
      • Attributes relevant, errors ghi bằng `span.RecordError()`

- [ ] Dashboards and alerts exist
      • PromQL từ comments đã wired vào Grafana + Alertmanager
      • Dùng rules từ awesome-prometheus-alerts cho common dependencies

- [ ] RUM events tracked (nếu có user-facing feature)
      • Key business events tracked server-side (PostHog/Segment)
      • Identity key là `user_id` (không dùng email)
      • Consent checked trước khi track
```

---

## ❌ Common Mistakes — Lỗi thường gặp

```go
// ❌ Sai: log AND return → error bị log nhiều lần khi propagate up
if err != nil {
    slog.Error("query failed", "error", err)
    return fmt.Errorf("query: %w", err)  // Error này lại được log ở layer trên
}

// ✅ Đúng: return với context, log một lần ở top-level handler
if err != nil {
    return fmt.Errorf("querying users: %w", err)
}
// Ở handler:
if err := service.DoSomething(); err != nil {
    slog.Error("request failed", "error", err)  // Log duy nhất
    return err
}
```

```go
// ❌ Sai: high-cardinality label (user_id là unbounded)
httpRequests.WithLabelValues(r.Method, r.URL.Path, userID).Inc()
// → 1 triệu user = 1 triệu series trong Prometheus → tốn memory, slow query

// ✅ Đúng: chỉ dùng bounded label values
httpRequests.WithLabelValues(r.Method, routePattern).Inc()
// routePattern = "/api/users/{id}" → cardinality thấp
```

```go
// ❌ Sai: không truyền context → trace bị đứt, không correlate được
result, err := db.Query("SELECT ...")

// ✅ Đúng: context flows through → trace tiếp tục, deadline được tôn trọng
result, err := db.QueryContext(ctx, "SELECT ...")
```

```go
// ❌ Sai: dùng Summary cho latency (không aggregatable)
prometheus.NewSummary(prometheus.SummaryOpts{
    Name: "http_request_duration_seconds",
    Objectives: map[float64]float64{0.99: 0.001},
})

// ✅ Đúng: dùng Histogram (aggregatable, hỗ trợ histogram_quantile)
prometheus.NewHistogram(prometheus.HistogramOpts{
    Name:    "http_request_duration_seconds",
    Buckets: prometheus.DefBuckets,
})
```

---

## 🔗 Liên kết liên quan

- `@golang-error-handling`: Quy tắc single handling (log OR return)
- `@golang-troubleshooting`: Dùng observability signals để diagnose production issues
- `@golang-security`: Bảo vệ pprof endpoints, tránh PII trong logs
- `@golang-context`: Propagate trace context qua service boundaries
- `@promql-cli`: Query và explore PromQL expressions từ CLI