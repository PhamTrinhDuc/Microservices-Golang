package slm

import (
	"context"
	"fmt"

	"google.golang.org/adk/model"

	promy_anthropic "multi-agent/provider/anthropic"
	promy_gemini "multi-agent/provider/gemini"
	promy_openai "multi-agent/provider/openai"
)

// Config represents the configuration for creating an SLM guard
type Config struct {
	// Provider type: "openai", "anthropic", or "gemini"
	Provider string

	// OpenAI-specific config
	OpenAIConfig *OpenAIProviderConfig

	// Anthropic-specific config
	AnthropicConfig *AnthropicProviderConfig

	// Gemini-specific config
	GeminiConfig *GeminiProviderConfig

	// RiskThreshold: if risk_level >= threshold, mark as unsafe (0.0-1.0)
	RiskThreshold float64

	// BehaviorThreshold: if risk_level >= threshold, mark as borderline (0.0-1.0)
	BehaviorThreshold float64

	// Timeout in seconds for LLM API call
	Timeout int
}

// OpenAIProviderConfig holds OpenAI-specific configuration
type OpenAIProviderConfig struct {
	APIKey    string
	BaseURL   string // optional, for compatible providers (Ollama, vLLM, etc.)
	ModelName string
}

// AnthropicProviderConfig holds Anthropic-specific configuration
type AnthropicProviderConfig struct {
	APIKey    string
	BaseURL   string // optional
	ModelName string
}

// GeminiProviderConfig holds Gemini-specific configuration
type GeminiProviderConfig struct {
	APIKey    string
	ModelName string
}

// NewGuard creates a new SLM guard based on the provider configuration
func NewGuard(ctx context.Context, name string, cfg Config) (*Guard, error) {
	// Validate config
	if cfg.RiskThreshold <= 0 || cfg.RiskThreshold > 1.0 {
		cfg.RiskThreshold = 0.8 // default
	}
	if cfg.BehaviorThreshold <= 0 || cfg.BehaviorThreshold > 1.0 {
		cfg.BehaviorThreshold = 0.5 // default
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 // default 30 seconds
	}

	var provider model.LLM
	var err error

	switch cfg.Provider {
	case "openai":
		if cfg.OpenAIConfig == nil {
			return nil, fmt.Errorf("OpenAI config is required for provider 'openai'")
		}
		provider, err = NewOpenAIProvider(cfg.OpenAIConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create OpenAI provider: %w", err)
		}

	case "anthropic":
		if cfg.AnthropicConfig == nil {
			return nil, fmt.Errorf("Anthropic config is required for provider 'anthropic'")
		}
		provider, err = NewAnthropicProvider(cfg.AnthropicConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create Anthropic provider: %w", err)
		}

	case "gemini":
		if cfg.GeminiConfig == nil {
			return nil, fmt.Errorf("Gemini config is required for provider 'gemini'")
		}
		provider, err = NewGeminiProvider(ctx, cfg.GeminiConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create Gemini provider: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}

	return &Guard{
		name:                name,
		provider:            provider,
		riskThreshold:       cfg.RiskThreshold,
		behaviorThreshold:   cfg.BehaviorThreshold,
		timeoutSeconds:      cfg.Timeout,
		systemPrompt:        SystemPromptGuard,
		analyzeInputPrompt:  PromptAnalyzeInput,
		analyzeOutputPrompt: PromptAnalyzeOutput,
	}, nil
}

// NewOpenAIProvider creates an LLM provider for OpenAI using the workspace provider package
func NewOpenAIProvider(cfg *OpenAIProviderConfig) (model.LLM, error) {
	if cfg == nil {
		return nil, fmt.Errorf("OpenAI config cannot be nil")
	}
	if cfg.ModelName == "" {
		cfg.ModelName = "gpt-4o-mini" // default model
	}

	return promy_openai.New(promy_openai.Config{
		APIKey:    cfg.APIKey,
		BaseURL:   cfg.BaseURL,
		ModelName: cfg.ModelName,
	}), nil
}

// NewAnthropicProvider creates an LLM provider for Anthropic using the workspace provider package
func NewAnthropicProvider(cfg *AnthropicProviderConfig) (model.LLM, error) {
	if cfg == nil {
		return nil, fmt.Errorf("Anthropic config cannot be nil")
	}
	if cfg.ModelName == "" {
		cfg.ModelName = "claude-3-5-sonnet-20241022" // default model
	}

	return promy_anthropic.New(promy_anthropic.Config{
		APIKey:          cfg.APIKey,
		BaseURL:         cfg.BaseURL,
		ModelName:       cfg.ModelName,
		MaxOutputTokens: 1024,
	}), nil
}

// NewGeminiProvider creates an LLM provider for Gemini using the workspace provider package
func NewGeminiProvider(ctx context.Context, cfg *GeminiProviderConfig) (model.LLM, error) {
	if cfg == nil {
		return nil, fmt.Errorf("Gemini config cannot be nil")
	}
	if cfg.ModelName == "" {
		cfg.ModelName = "gemini-2.5-flash" // default model
	}

	return promy_gemini.NewGeminiLLM(ctx, cfg.ModelName)
}
