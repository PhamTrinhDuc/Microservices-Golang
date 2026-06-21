package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadAgentConfig(t *testing.T) {
	// Load config
	configFile := "../config.yaml"
	ecommerceAgent, err := LoadAgentConfig(configFile, "ecommerce_agent")
	assert.NoError(t, err)

	// Kiểm tra ecommerce_agent
	assert.Equal(t, "ecommerce_agent", ecommerceAgent.Name)
	assert.Contains(t, ecommerceAgent.AllowedTools, "hybrid_search_documents")
	assert.Contains(t, ecommerceAgent.AllowedTools, "list_products")
}

func TestLoadAppConfig(t *testing.T) {
	configFile := "../config.yaml"
	config, err := LoadAppConfig(configFile)
	assert.NoError(t, err)
	assert.Equal(t, "gemini-flash-2.5", config.Models["gemini"])
	assert.Equal(t, "openai-4o-mini", config.Models["openai"])
	assert.Equal(t, "http://localhost:8081/mcp", config.McpServer)
}
