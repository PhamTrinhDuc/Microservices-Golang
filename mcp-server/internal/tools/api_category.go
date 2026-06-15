package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type Category struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ParentID  *int   `json:"parent_id"`
	Slug      string `json:"slug"`
	SortOrder int    `json:"sort_order"`
}

type CategoryTool struct {
	baseURL string
	client  *http.Client
}

func NewCategoryTool(baseURL string) *CategoryTool {
	return &CategoryTool{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// Definition returns the tool definitions for Category API
func (t *CategoryTool) ListCategoryDefinition() *mcp.Tool {
	return &mcp.Tool{
		Name:        "list_categories",
		Description: "List all product categories available in the catalog.",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}
}

// Handler
func (t *CategoryTool) ListCategoryHandler(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	apiURL := fmt.Sprintf("%s/categories", t.baseURL)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, nil, err
	}

	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, propagation.HeaderCarrier(httpReq.Header))

	resp, err := t.client.Do(httpReq)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, mcp.TextContent{Text: fmt.Sprintf("API request failed: %v", err)}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &mcp.CallToolResult{IsError: true}, mcp.TextContent{Text: fmt.Sprintf("API returned error: %s", resp.Status)}, nil
	}

	var apiResp struct {
		Data []Category `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return &mcp.CallToolResult{IsError: true}, mcp.TextContent{Text: fmt.Sprintf("failed to decode response: %v", err)}, nil
	}

	resultText := fmt.Sprintf("Found %d category(ies):\n", len(apiResp.Data))
	for _, c := range apiResp.Data {
		parentStr := "None"
		if c.ParentID != nil {
			parentStr = fmt.Sprintf("%d", *c.ParentID)
		}
		resultText += fmt.Sprintf("- %s (ID: %d, Parent ID: %s, Slug: %s, Order: %d)\n", c.Name, c.ID, parentStr, c.Slug, c.SortOrder)
	}

	return &mcp.CallToolResult{}, mcp.TextContent{Text: resultText}, nil
}
