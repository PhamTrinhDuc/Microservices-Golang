package mcp

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

type InvocationContextParams struct {
	Artifacts agent.Artifacts
	Memory    agent.Memory
	Session   session.Session

	Branch string
	Agent  agent.Agent

	UserContent   *genai.Content
	RunConfig     *agent.RunConfig
	EndInvocation bool
	InvocationID  string
}

func TestNewMCP(t *testing.T) {
	tests := []struct {
		name          string
		mcpServerURL  string
		allowedTools  []string
		approvalTools []string
	}{
		{
			name:          "",
			mcpServerURL:  "",
			allowedTools:  []string{},
			approvalTools: []string{},
		},
		{
			name:          "with valid url",
			mcpServerURL:  "http://localhost:8081",
			allowedTools:  []string{},
			approvalTools: []string{},
		},
		{
			name:          "with empty allowed tools",
			mcpServerURL:  "http://localhost:8081",
			allowedTools:  []string{},
			approvalTools: []string{},
		},
		{
			name:          "with valid tools and approval tools",
			mcpServerURL:  "http://localhost:8081",
			allowedTools:  []string{"test"},
			approvalTools: []string{"test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := &mcp.SSEClientTransport{Endpoint: "htpp://localhost:8001"}
			got, err := NewMCPTool(transport, tt.allowedTools, tt.approvalTools)

			if err != nil {
				t.Errorf("NewMCPTool() error = %v, wantErr %v", err, false)
			} else {
				t.Log(got.Name())
			}
		})
	}
}
