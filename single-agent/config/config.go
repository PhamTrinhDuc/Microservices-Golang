package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type AgentConfig struct {
	Name          string   `yaml:"name"`
	Description   string   `yaml:"description"`
	Instruction   string   `yaml:"instruction"`
	AllowedTools  []string `yaml:"allowedTools"`
	ApprovedTools []string `yaml:"approvedTools"`
}

type PromptGuardConfig struct {
	System        string `yaml:"system"`
	AnalyzeInput  string `yaml:"analyze_input"`
	AnalyzeOutput string `yaml:"analyze_output"`
	AnalyzeBoth   string `yaml:"analyze_both"`
}

type PromptsConfig struct {
	GuardsPrompt   PromptGuardConfig `yaml:"guard_prompt"`
	GuardContext   string            `yaml:"guard_context"`
	SystemPrompt   string            `yaml:"system_prompt"`
	ResponseFormat string            `yaml:"response_format"`
}

type AppConfig struct {
	Prompts   PromptsConfig          `yaml:"prompts"`
	Agents    map[string]AgentConfig `yaml:"agents"`
	Models    map[string]string      `yaml:"models"`
	McpServer string                 `yaml:"mcp_server"`
}

func LoadAgentConfig(filePath string, agentName string) (*AgentConfig, error) {
	// 1. Read File yaml
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 2. Parse yaml
	var config AppConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	agent, ok := config.Agents[agentName]
	if !ok {
		return nil, fmt.Errorf("agent %s not found", agentName)
	}
	return &agent, nil
}

func LoadPromptsConfig(filePath string) (*PromptsConfig, error) {
	// 1. Read File yaml
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 2. Parse yaml
	var config AppConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config.Prompts, nil
}

func LoadAppConfig(filePath string) (*AppConfig, error) {
	// 1. Read File yaml
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 2. Parse yaml
	var config AppConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
