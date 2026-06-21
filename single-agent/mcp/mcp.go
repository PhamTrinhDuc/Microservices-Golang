package mcp

import (
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/mcptoolset"
)

func NewMCPTool(transport *mcp.SSEClientTransport, allowedTools []string, approvalTools []string) (tool.Toolset, error) {
	// 1. Validate approvalTools phải có ở trong allowedTools
	allowedSet := make(map[string]bool)
	for _, t := range allowedTools {
		allowedSet[t] = true
	}

	for _, t := range approvalTools {
		if !allowedSet[t] {
			return nil, fmt.Errorf("approvalTool %q must be in allowedTools", t)
		}
	}

	// 2. Khởi tạo Toolset từ MCP
	// mcptoolset sẽ fetch toàn bộ các tool từ mcp-server
	mcpTools, err := mcptoolset.New(mcptoolset.Config{
		Transport: transport,
		// Chỉ cho phép các tools được định nghĩa trong allowedTools
		ToolFilter: tool.StringPredicate(allowedTools),
		// Human-in-the-loop: Yêu cầu xác nhận (approve) cho một số action nhạy cảm
		RequireConfirmationProvider: func(toolName string, toolInput any) bool {
			for _, tool := range approvalTools {
				if tool == toolName {
					// log.Printf("[DEBUG] → CONFIRM REQUIRED for %s", toolName)
					return true
				}
			}
			// log.Printf("[DEBUG] → no confirm needed for %s", toolName)
			return false
		},
	})
	if err != nil {
		return nil, err
	}
	return mcpTools, nil
}
