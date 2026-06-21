package agents

import (
	"context"
	"strings"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"

	config "single-agent/config"
)

func buildInstruction(promptCfg *config.PromptsConfig, agentCfg *config.AgentConfig) string {
	var sb strings.Builder
	if promptCfg.SystemPrompt != "" {
		sb.WriteString("### SYSTEM PROMPT: \n")
		sb.WriteString(promptCfg.SystemPrompt)
	}
	if agentCfg.Instruction != "" {
		sb.WriteString("### INSTRUCTION LLM: \n")
		sb.WriteString(agentCfg.Instruction)
	}
	if promptCfg.ResponseFormat != "" {
		sb.WriteString("### RESPONSE FORMAT: \n")
		sb.WriteString(promptCfg.ResponseFormat)
	}
	return sb.String()
}

// NewSingleAgent creates a single LLM Agent with a given config and toolset.
func NewSingleAgent(ctx context.Context, agentCfg *config.AgentConfig, promptCfg *config.PromptsConfig, model model.LLM, mcpTools tool.Toolset) (agent.Agent, error) {
	// Combine base instruction with global system prompt and response format guidelines
	instruction := buildInstruction(promptCfg, agentCfg)

	// Create the single LLM Agent with its name, description, instruction and tools
	singleAgent, err := llmagent.New(llmagent.Config{
		Name:        agentCfg.Name,
		Model:       model,
		Description: agentCfg.Description,
		Instruction: instruction,
		Toolsets:    []tool.Toolset{mcpTools},
	})
	if err != nil {
		return nil, err
	}

	return singleAgent, nil
}
