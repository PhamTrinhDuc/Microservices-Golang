package contextguard

import (
	"fmt"
	"log/slog"
	"sync"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/model"
)

// thresholdStrategy implements token-based compaction. It estimates total
// tokens before every LLM call and summarizes the entire conversation
// when remaining capacity drops below a safety buffer.
//
// When real token counts are available from the provider (persisted by the
// AfterModelCallback), they are preferred over the len/4 heuristic. A
// calibrated heuristic bridges the timing gap between callbacks so that
// tool results added after the last LLM call are accounted for.
//
// Compaction always produces a full summary (no recent tail preserved),
// matching Crush CLI behaviour. The result is [summary] + [continuation].
type thresholdStrategy struct {
	registry              ModelRegistry
	llm                   model.LLM
	maxTokens             int
	maxCompactionAttempts int
	mu                    sync.Mutex
}

func newThresholdStrategy(registry ModelRegistry, llm model.LLM, maxTokens int, maxCompactAttempts int) *thresholdStrategy {
	return &thresholdStrategy{
		registry:              registry,
		llm:                   llm,
		maxTokens:             maxTokens,
		maxCompactionAttempts: maxCompactAttempts,
	}
}

// Name returns the strategy identifier for logging.
func (s *thresholdStrategy) Name() string {
	return StrategyThreshold
}

func (s *thresholdStrategy) Compact(ctx agent.CallbackContext, req *model.LLMRequest) error {
	var contextWindow int
	if s.maxTokens > 0 {
		contextWindow = s.maxTokens
	} else {
		contextWindow = s.registry.ContextWindow(req.Model)
	}

	buffer := computeBuffer(contextWindow)
	threshold := contextWindow - buffer

	existingSummary := loadSummary(ctx)
	contentsAtLastCompaction := loadContentsAtCompaction(ctx)
	totalContents := len(req.Contents)
	if existingSummary != "" {
		injectSummary(req, existingSummary, contentsAtLastCompaction)
	}

	totalTokens := estimateTokens(req)
	if totalTokens < threshold {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	userContent := ctx.UserContent()
	todos := loadTodos(ctx)

	for attempt := range s.maxCompactionAttempts {
		slog.Info(
			fmt.Sprintf("%s [%s]: attempt %d: before compaction", PackageName, StrategyThreshold, attempt),
			"agent", ctx.AgentName(),
			"session", ctx.SessionID(),
			"attempt", attempt+1,
			"tokens", totalTokens,
			"threshold", threshold,
			"contextWindow", contextWindow,
			"buffer", buffer,
			"maxSummaryWords", int(float64(buffer)*0.50*0.75),
		)

		contentsForSummary := truncateForSummarizer(req.Contents, contextWindow)
		summary, err := summarize(ctx, s.llm, contentsForSummary, existingSummary, todos, buffer)
		if err != nil {
			slog.Warn(
				fmt.Sprintf("%s [%s]: summarization FAILED", PackageName, StrategyThreshold),
				"agent", ctx.AgentName(),
				"session", ctx.SessionID(),
				"error", err,
			)
			summary = buildFallbackSummary(contentsForSummary, existingSummary)
		}

		existingSummary = summary
		persistSummary(ctx, summary, totalTokens)
		persistContentAtCompaction(ctx, totalContents)
		replaceSummary(req, summary, nil)
		injectContinuation(req, userContent)

		// resetCalibration(ctx)

		newTokens := estimateTokens(req)

		slog.Info("ContextGuard [threshold]: compaction pass completed",
			"agent", ctx.AgentName(),
			"session", ctx.SessionID(),
			"attempt", attempt+1,
			"oldMessages", len(req.Contents),
			"newTokenEstimate", newTokens,
			"threshold", threshold,
		)

		if newTokens < threshold {
			break
		}

		if attempt < s.maxCompactionAttempts-1 {
			slog.Warn(
				fmt.Sprintf("%s [%s]: still above threshold after compaction, retrying", PackageName, StrategyThreshold),
				"agent", ctx.AgentName(),
				"attempt", attempt+1,
				"tokens", newTokens,
				"threshold", threshold,
			)
		}
	}
	return nil
}
