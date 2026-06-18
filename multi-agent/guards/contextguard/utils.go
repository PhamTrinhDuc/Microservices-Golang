package contextguard

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

// ========================================== SUMMARY HELPER FUNCTIOn ==========================================
const summarizeSystemPrompt = ``

// loadSummary loads the summary from the agent's state, return "" if not found.
func loadSummary(ctx agent.CallbackContext) string {
	key := stateKeyPrefixSummary + ctx.AgentName()
	value, err := ctx.State().Get(key)
	if err != nil {
		return ""
	}
	return value.(string)
}

// persistSummary writes the summary and a diagnostic token count to session
// state. Errors are logged but not propagated.
func persistSummary(ctx agent.CallbackContext, summary string, tokenEstimate int) {
	keySummary := stateKeyPrefixSummary + ctx.AgentName()
	keySummaryTokenEstimate := stateKeyPrefixSummaryTokenEstimate + ctx.AgentName()

	if err := ctx.State().Set(keySummary, summary); err != nil {
		slog.Warn(fmt.Sprintf("%s [%s]: failed to persist summary", PackageName, StrategySlidingWindow),
			"error", err.Error(),
			"agent", ctx.AgentName(),
			"session", ctx.SessionID(),
		)
	}

	if err := ctx.State().Set(keySummaryTokenEstimate, tokenEstimate); err != nil {
		slog.Warn(fmt.Sprintf("%s [%s]: failed to persist summary token estimate", PackageName, StrategySlidingWindow),
			"error", err.Error(),
			"agent", ctx.AgentName(),
			"session", ctx.SessionID(),
		)
	}
}

func injectSummary(req *model.LLMRequest, existingSummary string, lastCompactIdx int) {
	if len(req.Contents) > 0 && req.Contents[0] != nil {
		first := req.Contents[0]
		if first.Parts[0] != nil && first.Parts[0].Text != "" {
			return
		}
	}

	contentSummary := &genai.Content{
		Parts: []*genai.Part{
			{Text: existingSummary},
		},
		Role: "user",
	}

	if lastCompactIdx > 0 && lastCompactIdx <= len(req.Contents) {
		req.Contents = append(req.Contents[:lastCompactIdx], req.Contents[lastCompactIdx])
	} else {
		req.Contents = append(req.Contents, contentSummary)
	}
}

// truncateForSummarizer trims the conversation contents so that the
// summarization prompt itself doesn't exceed the summarizer LLM's context
// window. It keeps the most recent messages (freshest context) and drops
// the oldest ones when the total exceeds 80% of contextWindow. The 80%
// budget leaves room for the system prompt, previous summary, and output.
func truncateForSummarizer(contents []*genai.Content, contextWindow int) []*genai.Content {
	totalTokens := estimateContentTokens(contents)
	buget := int(float64(contextWindow) * 0.8)
	if totalTokens <= buget {
		return contents // no need to truncate
	}

	for len(contents) > 2 && totalTokens > buget {
		totalTokens -= estimateContentTokens([]*genai.Content{contents[0]})
		contents = contents[1:]
	}
	return contents
}

// loadContentsAtCompaction reads the Content count recorded at the last
// sliding-window compaction. Returns 0 if no compaction has happened yet.
func loadContentsAtCompaction(ctx agent.CallbackContext) int {
	key := stateKeyPrefixContentsAtCompaction + ctx.AgentName()
	val, err := ctx.State().Get(key)
	if err != nil {
		return 0
	}
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case int:
		return v
	case float64:
		return int(v)
	}
	return 0
}

func persistContentAtCompaction(ctx agent.CallbackContext, totalContents int) {
	key := stateKeyPrefixContentsAtCompaction + ctx.AgentName()
	if err := ctx.State().Set(key, totalContents); err != nil {
		slog.Warn(fmt.Sprintf("%s [%s]: failed to persist contents at compaction", PackageName, StrategySlidingWindow),
			"error", err.Error(),
			"agent", ctx.AgentName(),
			"session", ctx.SessionID(),
		)
	}
}

// computeBuffer compute buffer size for sliding window strategy.
func computeBuffer(contextWindow int) int {
	if contextWindow >= largeContextWindowThreshold {
		return largeContextWindowBuffer
	}
	return int(float64(contextWindow) * smallContextWindowRatio)
}

// summarize compact old messages and todos to summary content
func summarize(ctx agent.CallbackContext, llm model.LLM, oldMessages []*genai.Content, existingSummary string, todos []TodoItem, buffer int) (string, error) {
	maxOutputTokens := int32(float64(buffer) * 0.5)
	maxWords := int32(float64(maxOutputTokens) * 0.75)

	systemPrompt := summarizeSystemPrompt + fmt.Sprintf("\n\nKeep the summary under %d words", maxWords)
	userPrompt := buildSummarizePrompt(oldMessages, existingSummary, todos)

	req := &model.LLMRequest{
		Model: llm.Name(),
		Contents: []*genai.Content{
			{
				Role:  "user",
				Parts: []*genai.Part{{Text: userPrompt}},
			},
		},
		Config: &genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{Parts: []*genai.Part{{Text: systemPrompt}}},
			MaxOutputTokens:   maxOutputTokens,
		},
	}
	var resultSummrized string
	for resp, err := range llm.GenerateContent(ctx, req, false) {
		if err != nil {
			slog.Error(
				fmt.Sprintf("%s [%s]: summarization LLM call failed", PackageName, StrategySlidingWindow),
				"error", err.Error(),
				"agent", ctx.AgentName(),
				"session", ctx.SessionID(),
			)
			return "", fmt.Errorf("summarization LLM call failed: %w", err)
		}
		if resp != nil && resp.Content != nil {
			for _, part := range resp.Content.Parts {
				if part != nil && part.Text != "" {
					resultSummrized += part.Text
				}
			}
		}
	}
	if resultSummrized == "" {
		return buildFallbackSummary(oldMessages, existingSummary), nil
	}
	return resultSummrized, nil
}

// replaceSummary replace old contents with new summary and recent messages (summary = oldMessages, recentMessages = newMessages splited)
func replaceSummary(req *model.LLMRequest, summary string, recentMessages []*genai.Content) {
	summaryContent := &genai.Content{
		Role: "user",
		Parts: []*genai.Part{
			{Text: fmt.Sprintf("[Summary of previous conversation]\n%s\n[End of previous summary]", summary)},
		},
	}
	req.Contents = append([]*genai.Content{summaryContent}, recentMessages...)
}

// injectContinuation appends a continuation instruction to req.Contents so
// the agent knows to resume work without re-asking the user. If userContent
// is available, the original user request is included for reference.
func injectContinuation(req *model.LLMRequest, userContent *genai.Content) {
	var userRequestText string
	if userContent != nil {
		for _, part := range userContent.Parts {
			if part != nil && part.Text != "" {
				userRequestText = part.Text
				break
			}
		}
	}
	var msgContent string
	if userRequestText != "" {
		msgContent = fmt.Sprintf(
			"[System: The conversation was compacted because it exceeded the context window. "+
				"The summary above contains all prior context. The user's current request is: `%s`. "+
				"Continue working on this request without asking the user to repeat anything.]", userRequestText)
	} else {
		msgContent = "[System: The conversation was compacted because it exceeded the context window. " +
			"The summary above contains all prior context. " +
			"Continue working without asking the user to repeat anything.]"
	}

	req.Contents = append(req.Contents, &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{{Text: msgContent}},
	})

}

// BuildSummrizePrompt
func buildSummarizePrompt(oldMessages []*genai.Content, existingSummary string, todos []TodoItem) string {
	var sb strings.Builder
	sb.WriteString("Provide a detailed summary of the following conversation\n\n")
	if existingSummary != "" {
		sb.WriteString("[Previous summary of context]\n")
		sb.WriteString(existingSummary)
		sb.WriteString("\n[End of previous summary]\n")
		sb.WriteString("Incorporate the previous summary into your new summary, updating any information that has changed, correcting any inaccuracies, and adding any new details from the intervening conversation.\n")
	}
	sb.WriteString("[Conversation to summarize]\n")
	for _, content := range oldMessages {
		if content == nil {
			continue
		}
		role := content.Role
		parts := content.Parts
		for _, part := range parts {
			if part == nil {
				continue
			}

			if part.Text != "" {
				sb.WriteString(fmt.Sprintf("%s: %s\n", role, part.Text))
			}

			if part.FunctionCall != nil {
				sb.WriteString(fmt.Sprintf("%s : [called tool] %s \n", role, part.FunctionCall.Name))
			}

			if part.FunctionResponse != nil {
				sb.WriteString(fmt.Sprintf("%s : [tool] %s returned a result \n", role, part.FunctionResponse.Name))
			}
		}
	}
	sb.WriteString("[End of conversation]\n")

	if len(todos) > 0 {
		sb.WriteString("[Current todos list]\n")
		for _, t := range todos {
			fmt.Fprintf(&sb, "- [%s] %s \n", t.Status, t.Content)
		}

		sb.WriteString("[End of todos list]\n")
		sb.WriteString("Include these tasks and their statuses in your summary under a dedicated \"## Todo List\" section. ")
		sb.WriteString("Instruct the resuming assistant to restore them using the `todos` tool to continue tracking progress.\n")
	}
	return sb.String()
}

// buildFallbackSummary build fallback summary when llm summarization failed
func buildFallbackSummary(oldMessages []*genai.Content, existingSummary string) string {
	var sb strings.Builder
	if existingSummary != "" {
		sb.WriteString(existingSummary)
		sb.WriteString("\n\n")
	}

	for _, content := range oldMessages {
		if content == nil {
			continue
		}
		role := content.Role
		if role == "" {
			role = "unknown"
		}
		for _, part := range content.Parts {
			if part != nil && part.Text != "" {
				sb.WriteString(fmt.Sprintf("%s: ", role))
				if len(part.Text) > 200 {
					sb.WriteString(part.Text[:200])
				} else {
					sb.WriteString(part.Text)
				}
				sb.WriteString("\n")
			}
		}
	}
	return sb.String()
}

// ===================================== TODOS HELPER FUNCTION ==================================================

// TotoItem present for single task
type TodoItem struct {
	Content    string `json:"content"`
	Status     string `json:"status"`
	ActiveForm string `json:"active_form,omitempty"`
}

// loadTodos reads the todo list from session state. Returns nil if no todos
// are stored. Supports both []TodoItem and []any (from JSON deserialization).
func loadTodos(ctx agent.CallbackContext) []TodoItem {
	val, err := ctx.State().Get("todos")
	if err != nil || val == nil {
		return nil
	}

	switch v := val.(type) {
	case []TodoItem:
		return v
	case []any:
		var items []TodoItem
		for _, raw := range v {
			m, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			item := TodoItem{}
			if c, ok := m["content"].(string); ok {
				item.Content = c
			}
			if s, ok := m["status"].(string); ok {
				item.Status = s
			}
			if a, ok := m["active_form"].(string); ok {
				item.ActiveForm = a
			}
			if item.Content != "" {
				items = append(items, item)
			}
		}
		return items
	}
	return nil
}

// ====================================== SPLIT CONTENTS HELPER FUNCTION ==========================================

// safeSplitIndex computes the largest index <= splitPoint that lies on a turn boundary.
// If splitPoint would split a turn, it is adjusted to the end of the previous turn.
func safeSplitIndex(contents []*genai.Content, splitPoint int) int {
	if splitPoint <= 0 || splitPoint >= len(contents) {
		return splitPoint
	}

	orgIdx := splitPoint

	splitPoint = walkBackToPairBoundary(contents, splitPoint)
	if splitPoint <= 0 {
		splitPoint = walkForwardToPairBoundary(contents, orgIdx)
	}

	if splitPoint <= 0 {
		splitPoint = 1
	}
	if splitPoint >= len(contents) {
		splitPoint = len(contents) - 1
	}
	return splitPoint
}

// walkBackToPairBoundary walk backward to tool_call/tool_response pairs boundary
func walkBackToPairBoundary(contents []*genai.Content, idx int) int {
	for idx > 0 {
		c := contents[idx]
		if c == nil {
			return idx
		}

		if c.Role == "model" && contentHasFunctionCall(c) {
			idx++
			continue
		}

		if c.Role == "user" && contentHasFunctionResponse(c) {
			idx++
			continue
		}
		break
	}
	return idx
}

// walkForwardToPairBoundary walk forward to tool_call/tool_response pairs boundary
func walkForwardToPairBoundary(contents []*genai.Content, idx int) int {
	for idx > 0 {
		c := contents[idx]
		if c == nil {
			return idx
		}

		if c.Role == "model" && contentHasFunctionCall(c) {
			idx--
			continue
		}

		if c.Role == "user" && contentHasFunctionResponse(c) {
			idx--
			continue
		}
		break
	}
	return idx
}

// contentHashFunctionCall return true if content has function call.
func contentHasFunctionCall(c *genai.Content) bool {
	for _, part := range c.Parts {
		if part != nil && part.FunctionCall != nil {
			return true
		}
	}
	return false
}

// contentHashFunctionResponse return true if content has function response.
func contentHasFunctionResponse(c *genai.Content) bool {
	for _, part := range c.Parts {
		if part != nil && part.FunctionResponse != nil {
			return true
		}
	}
	return false
}

// =============================================== ESTIMATE TOKENS ===============================================

// estimatePartTokens returns a rough token count for a single Part using
// the ~4 chars per token heuristic. It accounts for Text, FunctionCall
// (name + args), and FunctionResponse (name + response).
func estimatePartTokens(part *genai.Part) int {
	if part == nil {
		return 0
	}
	totalTokens := 0
	// caculate tokens for text
	if part.Text != "" {
		totalTokens += len(part.Text) / 4
	}
	// caculate tokens for function call (function_name + function_args)
	if part.FunctionCall != nil {
		totalTokens += len(part.FunctionCall.Name) / 4
		for k, v := range part.FunctionCall.Args {
			totalTokens += len(k) / 4
			totalTokens += len(fmt.Sprintf("%v", v)) / 4 // value must be string
		}
	}
	// caculate tokens for function response (function_name + function_response)
	if part.FunctionResponse != nil {
		totalTokens += len(part.FunctionResponse.Name) / 4
		totalTokens += len(part.FunctionResponse.Response) / 4
	}
	// caculate tokens for inline data (mime_type + data)
	if part.InlineData != nil {
		totalTokens += len(part.InlineData.MIMEType) / 4
		totalTokens += len(part.InlineData.Data) / 4
	}
	return totalTokens
}

// estimateToolTokens returns a rough token count for a Tool slice.
func estimateToolTokens(tools []*genai.Tool) int {
	totalTokens := 0
	for _, tool := range tools {
		if tool == nil {
			continue
		}
		for _, fd := range tool.FunctionDeclarations {
			totalTokens += len(fd.Name) / 4
			totalTokens += len(fd.Description) / 4
			if fd.ParametersJsonSchema != nil {
				data, err := json.Marshal(fd.ParametersJsonSchema)
				if err == nil {
					totalTokens += len(data) / 4
				}
			} else if (fd.Parameters) != nil {
				data, err := json.Marshal(fd.Parameters)
				if err == nil {
					totalTokens += len(data) / 4
				}
			}
		}
	}
	return totalTokens

}

// estimateContentTokens returns a rough token count for a Content slice by summing
// the token estimates for each Part.
func estimateContentTokens(contents []*genai.Content) int {
	totalTokens := 0
	for _, content := range contents {
		if content != nil {
			for _, part := range content.Parts {
				totalTokens += estimatePartTokens(part)
			}
		}
	}
	return totalTokens
}

// estimateTokens returns a rough token count for a LLMRequest by summing
// the token estimates for each Content, SystemInstruction and Tools.
func estimateTokens(req *model.LLMRequest) int {
	totalTokens := 0
	totalTokens += estimateContentTokens(req.Contents)
	if req.Tools != nil {
		if req.Config.SystemInstruction != nil {
			for _, part := range req.Config.SystemInstruction.Parts {
				totalTokens += estimatePartTokens(part)
			}
		}
		totalTokens += estimateToolTokens(req.Config.Tools)
	}
	return totalTokens
}

// persistRealTokens persists the actual number of tokens used for the last LLM request.
func persistRealTokens(ctx agent.CallbackContext, promptTokens int) {
	key := stateKeyPrefixRealTokens + ctx.AgentName()
	if err := ctx.State().Set(key, promptTokens); err != nil {
		slog.Warn(
			fmt.Sprintf("%s: failed to persist real tokens", PackageName),
			"agent", ctx.AgentName(),
			"session", ctx.SessionID(),
			"error", err.Error(),
		)
	}
}

// loadRealTokens loads the actual number of tokens used for the last LLM request.
// Returns the number of tokens and true if found, 0 and false otherwise.
func loadRealTokens(ctx agent.CallbackContext) (int, bool) {
	key := stateKeyPrefixRealTokens + ctx.AgentName()
	val, err := ctx.State().Get(key)
	if err != nil || val == nil {
		return 0, false
	}
	return val.(int), true
}
