Hàm này là “trái tim” của cơ chế **memory compaction / sliding window** cho chatbot AI.

Nó làm nhiệm vụ:

```txt id="c4n6hh"
Khi chat history quá dài:
→ summarize phần cũ
→ giữ lại phần mới
→ nhét summary vào context
```

để tránh vượt context window của model.

---

# Ý tưởng tổng quát

Thay vì gửi:

```txt id="j1tzpq"
1000 messages
```

cho model mỗi lần chat…

nó sẽ biến thành:

```txt id="msupv0"
[Tóm tắt lịch sử cũ]
+
[vài message gần đây]
+
[user message mới]
```

=> tiết kiệm token cực mạnh.

---

# Flow tổng quát

Hàm này:

1. Kiểm tra có vượt `maxTurns` chưa
2. Nếu chưa:

   * chỉ inject summary cũ
3. Nếu vượt:

   * summarize phần cũ
   * giữ phần recent
   * replace context
4. Nếu vẫn quá dài:

   * compact mạnh hơn
   * retry nhiều lần

---

# Phân tích từng đoạn

---

# 1. Load state cũ

```go id="9n01s2"
existingSummary := loadSummary(ctx)
contentsAtLastCompaction := loadContentsAtCompaction(ctx)
```

Load:

* summary cũ
* vị trí lần compact trước

Ví dụ:

```txt id="ocjlwm"
Summary:
"User đang build AI agent bằng Go..."
```

---

# 2. Tính số turn từ lần compact trước

```go id="vljb5j"
totalContents := len(req.Contents)
turnsSinceCompaction := totalContents - contentsAtLastCompaction
```

Ví dụ:

| Value               | Meaning                        |
| ------------------- | ------------------------------ |
| totalContents = 120 | hiện có 120 messages           |
| lastCompact = 90    | lần trước compact ở message 90 |
| turnsSince = 30     | có thêm 30 messages            |

---

# 3. Nếu chưa vượt maxTurns

```go id="jq6vl0"
if turnsSinceCompaction <= s.maxTurns {
```

Ví dụ:

```txt id="4aqf8s"
maxTurns = 40
current delta = 30
```

=> chưa cần summarize.

---

## Nhưng vẫn inject summary

```go id="9mvl9j"
injectSummary(req, existingSummary, contentsAtLastCompaction)
```

Tức là:

```txt id="mjlwmj"
system prompt:
"Conversation summary: ..."
```

được thêm vào request.

---

# 4. Nếu vượt limit → bắt đầu compact

```go id="yuyf0d"
turn limit exceeded, summarizing
```

---

# 5. Lock mutex

```go id="r0qqfi"
s.mu.Lock()
defer s.mu.Unlock()
```

Tránh:

```txt id="bwswli"
2 goroutine compact cùng lúc
```

làm corrupt state.

---

# 6. Tính token threshold

```go id="ucv7g7"
func computeBuffer(contextWindow int) int {
	if contextWindow >= largeContextWindowThreshold {
		return largeContextWindowBuffer
	}
	return int(float64(contextWindow) * smallContextWindowRatio)
}

Tính buffer dựa vào contextWindow của model 
- Nếu contextWindow lớn như các model ChatGPT, Claude, v.v... > `largeContextWindowThreshold=200k token` => trả về lượng buffer cố định là `largeContextWindowBuffer=20k token`. Bởi vì các model này có lượng contextWindow lớn nên trả mặc định về `20k` token 
- Còn với các model nhỏ hơn như `llama3-8b`, `gemma-9b-it`, ... thì sẽ tính theo phần trăm context window còn lại với `smallContextWindowRatio=0.2`, tức tùy vào lượng contextWindow của từng model mà sẽ giữ lại 20% contextWindow làm buffer, model có contextWindow `8K token` mà lại buffer `20k token` là không hợp lý
- Mục đích của việc chừa ra 1 lượng buffer mà không dùng hết toàn bộ lượng contextWindow của LLM: 
    - **tokenizer mismatch**: tokenizer mà LLM sử dụng có thể không giống tokenizer mà SDK dùng để đếm token, dẫn đến việc tính toán sai lệch lượng contextWindow
    - **hidden system tokens**: LLM có thể có các hidden system tokens mà SDK không tính toán được, dẫn đến việc tính toán sai lệch lượng contextWindow
    - **tool call overhead**: tool call có thể làm tăng lượng contextWindow, dẫn đến việc tính toán sai lệch lượng contextWindow
    - **response vượt limit**: response từ LLM có thể vượt quá lượng contextWindow, dẫn đến việc tính toán sai lệch lượng contextWindow
    - **provider tính token khác mình**: provider LLM có thể tính token theo cách khác với SDK, dẫn đến việc tính toán sai lệch lượng contextWindow

contextWindow := s.registry.ContextWindow(req.Model)
buffer := computeBuffer(contextWindow)
threshold := contextWindow - buffer
```

Ví dụ:

| Value         | Meaning |
| ------------- | ------- |
| contextWindow | 128k    |
| buffer        | 20k     |
| threshold     | 108k    |

Mục tiêu:

```txt id="m9rkzr"
sau compact phải < threshold
```

---

# 7. recentKeep

```go id="a6n7rl"
userContent := ctx.UserContent() 
todos := loadTodos(ctx)
recentKeep := max(3, s.maxTurns*30/100)
```
- `userContent` nội dung chat hiện tại của user 

- `todos` load các kế hoạch của LLM nếu có 

- `maxTurns` thường dùng để giới hạn số lượng lượt hội thoại (conversation turns) được giữ lại trong context khi áp dụng chiến lược kiểu sliding window. Ví dụ: `maxTurns=20` nghĩa là: sau khi compact, agent sẽ cố gắng giữ lại tối đa 20 lượt hội thoại gần nhất.

- `recentKeep` là số lượng lượt hội thoại gần nhất được giữ lại trong context. Giá trị này được tính bằng: `max(3, s.maxTurns*30/100)`.

---

# 8. Loop retry compaction

```go id="92nsgp"
for attempt := range s.maxCompactionAttempts {
```

Nếu summarize xong vẫn quá dài:

```txt id="vwuc2r"
compact tiếp mạnh hơn
```

---

# 9. Chia old/recent messages

```go id="jlwm6f"
splitIdx := safeSplitIndex(...)
oldContents := ...
recentContents := ...
```

Ví dụ:

```txt id="rtggoj"
[ old old old ][ recent recent ]
```

```go 
func safeSplitIndex(contents []*genai.Content, idx int) int {
	if idx <= 0 || idx >= len(contents) {
		return idx
	}
	origIdx := idx
	idx = walkBackToPairBoundary(contents, idx)
	if idx <= 0 {
		idx = walkForwardToPairBoundary(contents, origIdx)
	}
	if idx <= 0 {
		idx = 1
	}
	if idx >= len(contents) {
		idx = len(contents) - 1
	}
	return idx
}

func walkBackToPairBoundary(contents []*genai.Content, idx int) int {
	for idx > 0 {
		c := contents[idx]
		if c == nil {
			break
		}
		if c.Role == "user" && contentHasFunctionResponse(c) {
			idx--
			continue
		}
		if c.Role == "model" && contentHasFunctionCall(c) {
			idx--
			continue
		}
		break
	}
	return idx
}

func walkForwardToPairBoundary(contents []*genai.Content, idx int) int {
	for idx < len(contents) {
		c := contents[idx]
		if c == nil {
			break
		}
		if c.Role == "model" && contentHasFunctionCall(c) {
			idx++
			continue
		}
		if c.Role == "user" && contentHasFunctionResponse(c) {
			idx++
			break
		}
		break
	}
	return idx
}
```

Ở đây việc xử lý tool call và tool response rất quan trọng vì nó có thể làm hỏng việc chia old/recent messages. Nếu index nằm giữa `tool_call/tool_response` khi chia contents sẽ gây mất context cho LLM, ở đây code xử lý 2 vấn đề: 
1. Trước tiên `walkBackToPairBoundary` tức lùi về phía trước của index hiện tại cho đến khi gặp được  content không phải `tool_call/tool_response` 
- Kiểm tra content của index hiện tại có phải là role `user` và có  `tool_response` không. Nếu phải thì lùi lại 1 bước
- Kiểm tra tiếp content của index hiện tại có phải là role `model` và có  `tool_calls` không. Nếu phải thì lùi lại 1 bước
- Nếu LLM gọi tools nhiều lần thì vòng lặp while sẽ lặp liên tục cho đến khi gặp content khác `tool_call/tool_response`

2. Nếu `walkBackToPairBoundary` không tìm được content không phải `tool_call/tool_response` thì sẽ dùng `walkForwardToPairBoundary` tức lùi về phía sau của index hiện tại cho đến khi gặp được  content không phải `tool_call/tool_response` 
- Kiểm tra content của index hiện tại có phải là role `model` và có  `tool_calls` không. Nếu phải thì lùi lại 1 bước
- Kiểm tra tiếp content của index hiện tại có phải là role `user` và có  `tool_response` không. Nếu phải thì lùi lại 1 bước
- Nếu LLM gọi tools nhiều lần thì vòng lặp while sẽ lặp liên tục cho đến khi gặp content khác `tool_call/tool_response`
---

# 10. Summarize phần cũ

```go id="nrlw4t"
summary, err := summarize(ctx, s.llm, oldContents, existingSummary, buffer, todos)
```

Thực hiện summary từ content cũ, summary trước đó, không vượt quá `maxOutputTokens = buffer*0.5 token`

```txt id="m8h7t4"
- User đang làm AI infra bằng Go
- Đang implement context compression
- TODO: optimize token estimation
```

---

# 11. Persist summary

```go id="r7kjvu"
persistSummary(...)
persistContentsAtCompaction(...)
```

Lưu summary xuống session/memory/database.

---

# 12. Replace request contents

```go id="qor8xe"
replaceSummary(req, summary, recentContents)
```

Biến request từ:

```txt id="p0b4jx"
100 messages
```

thành:

```txt id="kszkfu"
[summary]
+
[last 6 messages]
```

---

# 13. Inject continuation

```go id="hjlwm5"
injectContinuation(req, userContent)
```

Có thể thêm instruction kiểu:

```txt id="lms0jp"
"Continue conversation naturally..."
```

---

# 14. Estimate token mới

```go id="jlwmc6"
newTokens := estimateTokens(req)

func estimatePartTokens(part *genai.Part) int {
	if part == nil {
		return 0
	}
	total := 0
	if part.Text != "" {
		total += len(part.Text) / 4
	}
	if part.FunctionCall != nil {
		total += len(part.FunctionCall.Name) / 4
		for k, v := range part.FunctionCall.Args {
			total += len(k) / 4
			total += len(fmt.Sprintf("%v", v)) / 4
		}
	}
	if part.FunctionResponse != nil {
		total += len(part.FunctionResponse.Name) / 4
		total += len(fmt.Sprintf("%v", part.FunctionResponse.Response)) / 4
	}
	if part.InlineData != nil {
		total += len(part.InlineData.MIMEType) / 4
		total += len(part.InlineData.Data) / 4
	}
	return total
}
```

Function này thực tokenizer theo ý tưởng ~1 token ~ 4 ký tự `len(text)/4` để nhanh chóng estimate token, không phải tokenizer thật vì:
- Tokenizer thật chậm hơn
- Cần Provider-specific
- Cần SDK riêng
- Đôi khi phải gọi API

1. Thực hiện estimate cho nội dung content `part.Text`:  `total += len(part.Text) / 4`
**Ví dụ**: 
```go
part.Text = "Hello, how are you?"
len(part.Text) = 20
total += 20 / 4 = 5 token
```

2. Thực hiện estimate cho nội dung Content và Args của `part.FunctionCall`:  
**Ví dụ**:
```go
{
  "name": "search_weather",
  "args": {
    "city": "Hanoi",
    "country": "Vietnam"
  }
}
len(part.FunctionCall.Name) / 4 => 1 token
len(part.FunctionCall.Args) / 4 => 3 token (convert args sang string rồi mới tính)
```

3. Thực hiện estimate cho nội dung Content và Args của `part.FunctionResponse`:
**Ví dụ**:
```go
{
  "name": "search_weather",
  "response": {
    "temp": "25°C",
    "city": "Hanoi"
  }
}
len(part.FunctionResponse.Name) / 4 => 1 token
len(part.FunctionResponse.Response) / 4 => 3 token (convert response sang string rồi mới tính)
```

4. Thực hiện estimate cho nội dung InlineData của `part.InlineData`:  
**Ví dụ**:
```go
part.InlineData = &genai.InlineData{
	MIMEType: "image/png",
	Data:     []byte("..."),
}
len(part.InlineData.MIMEType) / 4 => 2 token
len(part.InlineData.Data) / 4 => 5 token
```
---

# 15. Nếu vẫn quá dài → compact mạnh hơn

```go id="30kbs6"
recentKeep = max(3, recentKeep/2)
```

Ví dụ:

| Attempt | recentKeep |
| ------- | ---------- |
| 1       | 12         |
| 2       | 6          |
| 3       | 3          |

Tức là:

```txt id="09u8do"
giữ ít recent messages hơn
```

để giảm token.

---

# 16. Kết thúc

Nếu:

```txt id="l45z3u"
newTokens < threshold
```

=> success.

---

# Đây là kỹ thuật gì?

Đây là pattern rất phổ biến:

## Conversation Compaction

hay:

## Hierarchical Memory

hay:

## Recursive Summarization

dùng trong:

* ChatGPT
* Claude
* AI agents
* AutoGPT
* LangChain
* memory systems

---

# Kiến trúc tổng quát

```txt id="b03c4u"
Raw chat history
        ↓
Sliding window detects overflow
        ↓
Summarize old messages
        ↓
Keep recent messages
        ↓
Inject summary into prompt
        ↓
Continue chatting
```

---

# Điểm hay của implementation này

Nó khá production-grade vì có:

✅ retry compaction
✅ token threshold
✅ context-aware buffer
✅ mutex safety
✅ persisted summaries
✅ adaptive shrinking
✅ watermark tracking
✅ continuation injection

---

# Tóm gọn cực ngắn

Hàm này:

```txt id="kq8w4l"
Tự động nén chat history bằng AI summarization
```

để:

* tránh vượt context window
* giữ continuity conversation
* giảm token cost
* scale chat lâu dài.
