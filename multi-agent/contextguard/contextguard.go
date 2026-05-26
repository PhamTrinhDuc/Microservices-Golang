package contextguard

import (
	"fmt"
	"log/slog"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/plugin"
	"google.golang.org/adk/runner"
)

var (
	PackageName           = "Contextguard"
	StrategySlidingWindow = "sliding_window"
	StrategyThreshold     = "threshold"

	stateKeyPrefixSummary              = "__context_guard_summary_"
	stateKeyPrefixSummaryTokenEstimate = "__context_guard_summary_token_estimate_"

	stateKeyPrefixContentsAtCompaction = "__context_guard_contents_at_compaction_"
	stateKeyPrefixRealTokens           = "__context_guard_real_tokens_"

	largeContextWindowThreshold = 200_000
	largeContextWindowBuffer    = 20_000
	smallContextWindowRatio     = 0.20

	defaultMaxCompactionAttempts = 3
	defaultMaxTurns              = 20
)

// Strategy defines how a compaction algorithm decides whether and how to
// compact conversation history before an LLM call.
type Strategy interface {
	Name() string
	Compact(ctx agent.CallbackContext, req *model.LLMRequest) error
}

// ContextGuard accumulates per-agent strategies and produces a single
// runner.PluginConfig. Use New to create one, Add to register agents,
// and PluginConfig to get the final configuration.
type ContextGuard struct {
	registry   ModelRegistry
	strategies map[string]Strategy
}

type agentConfig struct {
	strategy          string
	maxTurns          int
	maxTokens         int
	maxCompactAttemps int
}

// WithSlidingWindow selects the sliding-window strategy with the given
// maximum number of Content entries before compaction.
func withSlidingWindow(maxTurns int) func(*agentConfig) {
	return func(c *agentConfig) {
		c.strategy = StrategySlidingWindow
		c.maxTurns = maxTurns
	}
}

// WithMaxTokens sets a manual context window size override (in tokens).
// Only used by the threshold strategy. When set, the ModelRegistry is
// bypassed for this agent.
func WithMaxTokens(maxTokens int) func(*agentConfig) {
	return func(c *agentConfig) {
		c.maxTokens = maxTokens
	}
}

// WithMaxCompactionAttempts sets the maximum number of summarization retries
// when a single compaction pass still exceeds the threshold. Applies to both
// strategies. Defaults to 3 when not set or when n <= 0.
func WithMaxCompactAttempts(maxCompactAttempts int) func(*agentConfig) {
	return func(c *agentConfig) {
		c.maxCompactAttemps = maxCompactAttempts
	}
}

// Add registers an agent with its LLM for summarization. Without options,
// the threshold strategy is used with limits from the ModelRegistry.
func (g *ContextGuard) Add(agentID string, llm model.LLM, opts ...func(*agentConfig)) {
	cfg := &agentConfig{
		strategy: StrategyThreshold,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.maxCompactAttemps < 0 {
		cfg.maxCompactAttemps = defaultMaxCompactionAttempts
	}

	switch cfg.strategy {
	case StrategySlidingWindow:
		maxTurns := cfg.maxTurns
		if maxTurns < 0 {
			maxTurns = defaultMaxTurns
		}
		g.strategies[agentID] = newSlidingWindowStrategy(g.registry, llm, maxTurns, cfg.maxCompactAttemps)
	default:
		g.strategies[agentID] = newThresholdStrategy(g.registry, llm, cfg.maxTokens, cfg.maxCompactAttemps)

	}
	slog.Info(
		fmt.Sprintf("%s: strategy configured", PackageName),
		"agent", agentID,
		"strategy", g.strategies[agentID].Name(),
	)
}

// PluginConfig returns a runner.PluginConfig ready to pass to the ADK
// launcher or runner.
func (g *ContextGuard) PluginConfig() runner.PluginConfig {
	p, _ := plugin.New(
		plugin.Config{
			Name:                PackageName,
			BeforeModelCallback: nil,
			AfterModelCallback:  nil,
		},
	)

	return runner.PluginConfig{
		Plugins: []*plugin.Plugin{p},
	}
}

func (g *ContextGuard) beforeModel(ctx agent.CallbackContext, req *model.LLMRequest) (*model.LLMResponse, error) {
	if req == nil || len(req.Contents) == 0 {
		slog.Warn(
			fmt.Sprintf("%s: beforeModel called with nil or empty request", PackageName),
			"agent", ctx.AgentName(),
			"session", ctx.SessionID(),
		)
		return nil, nil
	}

	strategy, ok := g.strategies[ctx.AgentName()]
	if !ok {
		return nil, nil
	}

	if err := strategy.Compact(ctx, req); err != nil {
		slog.Error(
			fmt.Sprintf("%s: compaction failed", PackageName),
			"agent", ctx.AgentName(),
			"session", ctx.SessionID(),
			"error", err.Error(),
		)
		return nil, err
	}

	return nil, nil
}

func (g *ContextGuard) afterModel(ctx agent.CallbackContext, resp *model.LLMResponse, _ error) (*model.LLMResponse, error) {
	if resp == nil || resp.Partial {
		slog.Warn(
			fmt.Sprintf("%s: afterModel called with nil or empty response", PackageName),
			"agent", ctx.AgentName(),
			"session", ctx.SessionID(),
		)
		return nil, nil
	}

	if resp.UsageMetadata == nil {
		return nil, nil
	}

	if _, ok := g.strategies[ctx.AgentName()]; !ok {
		return nil, nil
	}

	promptTokens := int(resp.UsageMetadata.PromptTokenCount)
	if promptTokens > 0 {
		persistRealTokens(ctx, promptTokens)
	}
	return nil, nil

}
