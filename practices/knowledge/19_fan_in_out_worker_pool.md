## Case 1: Fan-out/Fan-in

```go
workers -> ch(result)
```

Channel ở đây chứa:

```text
Kết quả
```

Các goroutine đang là **sender**:

```go
ch <- result
```

Muốn close channel thì phải đảm bảo:

```text
Không còn ai gửi nữa
```

nên:

```go
wg.Wait()
close(ch)
```

Tức là:

```text
Đợi toàn bộ sender chết
        │
        ▼
     close(ch)
```

---

## Case 2: Worker Pool

```go
jobs channel -> workers
```

Channel ở đây chứa:

```text
Công việc
```

Worker đang là **receiver**:

```go
job := <-jobs
```

Muốn worker chết thì phải báo:

```text
Không còn job nữa
```

nên:

```go
close(jobs)
wg.Wait()
```

Tức là:

```text
Đóng nguồn job
      │
      ▼
worker thoát
      │
      ▼
Done()
      │
      ▼
Wait()
```

---

### Mẹo nhớ

Đừng nhớ:

```text
Wait trước close?
hay close trước Wait?
```

Hãy nhớ:

> **Close khi bạn muốn báo "không còn dữ liệu nữa".**

Sau đó tự hỏi:

### Fan-out

```text
Dữ liệu = result
```

Ai tạo result?

```text
goroutines
```

=> phải đợi chúng xong

```text
Wait -> Close
```

---

### Worker Pool

```text
Dữ liệu = jobs
```

Ai tạo jobs?

```text
Producer
```

Khi shutdown:

```text
Không còn job mới nữa
```

=> đóng ngay

```text
Close -> Wait
```

---

Một câu phân biệt cực nhanh:

**Fan-in/Fan-out**

```text
Tôi đợi worker xong để đóng channel.
```

**Worker Pool**

```text
Tôi đóng channel để worker xong.
```

Đó là lý do thứ tự bị đảo ngược. Đây không phải mẹo nhớ vặt đâu, mà là do channel đang đóng hai vai trò hoàn toàn khác nhau trong hai pattern. 😄
