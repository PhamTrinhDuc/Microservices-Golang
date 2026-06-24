package tools

import (
	"context"
	"fmt"
	"mcp-server/internal/database"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type HybridSearchTool struct {
	db database.Store
}

func NewHybridSearchTool(db database.Store) *HybridSearchTool {
	return &HybridSearchTool{db: db}
}

func (t *HybridSearchTool) Definition() *mcp.Tool {
	return &mcp.Tool{
		Name:        "seach_policies",
		Description: "Perform hybrid search combining BM25 lexical search with vector semantic similarity for policies.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "The search query text",
				},
				"limit": map[string]any{
					"type":        "number",
					"description": "Maximum number of results to return",
				},
			},
			"required": []string{"query"},
		},
	}
}

type HybridArgs struct {
	Query string  `json:"query"`
	Limit float64 `json:"limit"`
}

func (t *HybridSearchTool) Handler(ctx context.Context, req *mcp.CallToolRequest, args HybridArgs) (*mcp.CallToolResult, any, error) {

	limit := int(args.Limit)
	if limit <= 0 {
		limit = 10
	}

	documents, err := t.db.SearchDocuments(ctx, args.Query, limit)
	if err != nil {
		res := &mcp.CallToolResult{IsError: true}
		return res, mcp.TextContent{Text: fmt.Sprintf("hybrid search failed: %v", err)}, nil
	}

	var resultText string
	if len(documents) == 0 {
		resultText = "No documents found."
	} else {
		resultText = fmt.Sprintf("Found %d documents\n", len(documents))
		for i, doc := range documents {
			resultText += fmt.Sprintf("%d. %s\n", i+1, doc.Content)
		}
	}

	return &mcp.CallToolResult{}, mcp.TextContent{Text: resultText}, nil
}
