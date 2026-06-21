package slm

import (
	"context"
	"encoding/json"
	"fmt"
	"single-agent/guards/promptguard"
	"strings"
	"time"

	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

// Guard represents an SLM-based security guard
type Guard struct {
	name                string
	provider            model.LLM
	riskThreshold       float64
	behaviorThreshold   float64
	timeoutSeconds      int
	systemPrompt        string
	analyzeInputPrompt  string
	analyzeOutputPrompt string
}

type Option func(*Guard)

func WithRiskThreshold(riskThreshold float64) Option {
	return func(g *Guard) {
		g.riskThreshold = riskThreshold
	}
}

func WithBehaviorThreshold(behaviorThreshold float64) Option {
	return func(g *Guard) {
		g.behaviorThreshold = behaviorThreshold
	}
}

func WithTimeout(timeoutSeconds int) Option {
	return func(g *Guard) {
		g.timeoutSeconds = timeoutSeconds
	}
}

func WithSystemPrompt(systemPrompt string) Option {
	return func(g *Guard) {
		g.systemPrompt = systemPrompt
	}
}

func WithAnalyzeInputPrompt(analyzeInputPrompt string) Option {
	return func(g *Guard) {
		g.analyzeInputPrompt = analyzeInputPrompt
	}
}

func WithAnalyzeOutputPrompt(analyzeOutputPrompt string) Option {
	return func(g *Guard) {
		g.analyzeOutputPrompt = analyzeOutputPrompt
	}
}

// NewGuard creates a new SLM guard based on the provider configuration
func NewGuard(ctx context.Context, name string, llm model.LLM, opts ...Option) (*Guard, error) {
	g := &Guard{
		name:                name,
		provider:            llm,
		riskThreshold:       0.8,
		behaviorThreshold:   0.5,
		timeoutSeconds:      30,
		systemPrompt:        SystemPromptGuard,
		analyzeInputPrompt:  PromptAnalyzeInput,
		analyzeOutputPrompt: PromptAnalyzeOutput,
	}

	for _, opt := range opts {
		opt(g)
	}

	return g, nil
}

// --- Implement Guard interface ---

func (g *Guard) Name() string {
	return g.name
}

func (g *Guard) SupportInput() bool {
	return true
}

func (g *Guard) SupportOutput() bool {
	return true
}

func (g *Guard) GuardInput(ctx context.Context, input string) (promptguard.GuardVerdict, error) {
	return g.analyze(ctx, input, "", true)
}

func (g *Guard) GuardOutput(ctx context.Context, input, output string) (promptguard.GuardVerdict, error) {
	return g.analyze(ctx, input, output, false)
}

// analyze is the core method that calls the LLM and parses the response
func (g *Guard) analyze(ctx context.Context, input, output string, isInput bool) (promptguard.GuardVerdict, error) {
	// Add timeout to context
	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Duration(g.timeoutSeconds)*time.Second)
	defer cancel()

	// Prepare the prompt based on what we're analyzing
	var userPrompt string
	if isInput {
		userPrompt = strings.ReplaceAll(g.analyzeInputPrompt, "{{CONTENT}}", input)
	} else {
		userPrompt = g.analyzeOutputPrompt
		userPrompt = strings.ReplaceAll(userPrompt, "{{USER_INPUT}}", input)
		userPrompt = strings.ReplaceAll(userPrompt, "{{MODEL_OUTPUT}}", output)
	}

	// Prepare LLM request
	req := &model.LLMRequest{
		Contents: []*genai.Content{
			{
				Role: genai.RoleUser,
				Parts: []*genai.Part{
					{Text: userPrompt},
				},
			},
		},
		Config: &genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{
				Parts: []*genai.Part{
					{Text: g.systemPrompt},
				},
			},
			ResponseMIMEType: "application/json",
		},
	}

	var responseText string

	// Call LLM provider using the abstraction
	for resp, err := range g.provider.GenerateContent(ctxWithTimeout, req, false) {
		if err != nil {
			return promptguard.GuardVerdict{
				GuardName: g.name,
				Verdict:   promptguard.VerdictUncertain,
				Reason:    fmt.Sprintf("LLM analysis failed: %v", err),
				Layer:     "slm",
				Metadata:  nil,
			}, err
		}
		if resp != nil && resp.Content != nil {
			for _, p := range resp.Content.Parts {
				if p.Text != "" {
					responseText += p.Text
				}
			}
		}
	}

	if responseText == "" {
		return promptguard.GuardVerdict{
			GuardName: g.name,
			Verdict:   promptguard.VerdictUncertain,
			Reason:    "empty LLM response",
			Layer:     "slm",
			Metadata:  nil,
		}, fmt.Errorf("empty response")
	}

	// Parse the response
	result, err := g.parseResponse(responseText)
	if err != nil {
		return promptguard.GuardVerdict{
			GuardName: g.name,
			Verdict:   promptguard.VerdictUncertain,
			Reason:    fmt.Sprintf("failed to parse LLM response: %v", err),
			Layer:     "slm",
			Metadata: map[string]any{
				"raw_response": responseText,
			},
		}, err
	}

	// Convert result to GuardVerdict
	return g.resultToVerdict(result, isInput), nil
}

// parseResponse parses the JSON response from the LLM
func (g *Guard) parseResponse(response string) (*GuardAnalysisResult, error) {
	var result GuardAnalysisResult

	// Extract JSON from response (in case there's additional text)
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")
	if jsonStart == -1 || jsonEnd == -1 || jsonStart > jsonEnd {
		return nil, fmt.Errorf("no valid JSON found in response")
	}

	jsonStr := response[jsonStart : jsonEnd+1]
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// resultToVerdict converts GuardAnalysisResult to GuardVerdict
func (g *Guard) resultToVerdict(result *GuardAnalysisResult, isInput bool) promptguard.GuardVerdict {
	var verdict promptguard.Verdict

	switch result.Verdict {
	case "unsafe":
		verdict = promptguard.VerdictUnsafe
	case "borderline":
		verdict = promptguard.VerdictBorderline
	case "safe":
		verdict = promptguard.VerdictSafe
	default:
		verdict = promptguard.VerdictUncertain
	}

	// If risk level is high, override verdict to unsafe
	if result.RiskLevel >= g.riskThreshold && verdict != promptguard.VerdictUnsafe {
		verdict = promptguard.VerdictUnsafe
	}
	// If risk level is medium, mark as borderline if not already unsafe
	if result.RiskLevel >= g.behaviorThreshold && result.RiskLevel < g.riskThreshold && verdict == promptguard.VerdictSafe {
		verdict = promptguard.VerdictBorderline
	}

	// Build reason
	reason := result.Verdict
	if result.IsJailbreakAttempt {
		reason += " (jailbreak attempt detected)"
	}
	if len(result.Violations) > 0 {
		categories := make([]string, 0, len(result.Violations))
		for _, v := range result.Violations {
			categories = append(categories, v.Category)
		}
		reason += fmt.Sprintf(" - violations: %s", strings.Join(categories, ", "))
	}

	metadata := map[string]any{
		"risk_level":           result.RiskLevel,
		"is_jailbreak_attempt": result.IsJailbreakAttempt,
		"violations":           result.Violations,
		"recommendations":      result.Recommendations,
	}

	if isInput {
		metadata["analysis_type"] = "input"
	} else {
		metadata["analysis_type"] = "output"
	}

	return promptguard.GuardVerdict{
		GuardName: g.name,
		Verdict:   verdict,
		Reason:    reason,
		Layer:     "slm",
		Metadata:  metadata,
	}
}
