package slm

import (
	config "single-agent/config"
)

var appCfg, err = config.LoadAppConfig("./config.yaml")

// SystemPromptGuard is the main system prompt for the SLM guard
// It instructs the model to analyze text for security issues
var SystemPromptGuard = appCfg.Prompts.GuardsPrompt.System

// PromptAnalyzeInput is used for analyzing user input/prompt
var PromptAnalyzeInput = appCfg.Prompts.GuardsPrompt.AnalyzeInput

// PromptAnalyzeOutput is used for analyzing model output/response
var PromptAnalyzeOutput = appCfg.Prompts.GuardsPrompt.AnalyzeOutput

// PromptAnalyzeBoth is used for analyzing full conversation context
var PromptAnalyzeBoth = appCfg.Prompts.GuardsPrompt.AnalyzeBoth

// GuardBucket represents a security guard bucket/category
type GuardBucket string

const (
	BucketPromptInjection    GuardBucket = "prompt_injection"
	BucketPromptLeakage      GuardBucket = "prompt_leakage"
	BucketAccessControlAbuse GuardBucket = "access_control_abuse"
	BucketTechnicalExploits  GuardBucket = "technical_exploits"
	BucketTopicPolicyAbuse   GuardBucket = "topic_policy_abuse"
)

// SecurityViolation represents a detected security violation
type SecurityViolation struct {
	Category  string  `json:"category"`
	Severity  string  `json:"severity"`
	Reason    string  `json:"reason"`
	RiskScore float64 `json:"risk_score,omitempty"`
}

// GuardAnalysisResult represents the parsed response from the SLM guard
type GuardAnalysisResult struct {
	Verdict            string              `json:"verdict"`
	RiskLevel          float64             `json:"risk_level"`
	Violations         []SecurityViolation `json:"violations"`
	IsJailbreakAttempt bool                `json:"is_jailbreak_attempt"`
	Recommendations    string              `json:"recommendations"`
}
