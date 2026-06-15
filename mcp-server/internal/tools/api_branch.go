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

type Brand struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	IsActive bool   `json:"is_active"`
}

type BrandTool struct {
	baseURL string
	client  *http.Client
}

func NewBrandTool(baseURL string) *BrandTool {
	return &BrandTool{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// Definition returns the tool definitions for Brand API
func (t *BrandTool) ListBrandDefinition() *mcp.Tool {
	return &mcp.Tool{
		Name:        "list_brands",
		Description: "List all product brands / manufacturers available in the catalog.",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}
}

// Handler
func (t *BrandTool) ListBrandHandler(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	apiURL := fmt.Sprintf("%s/brands", t.baseURL)

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
		Data []Brand `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return &mcp.CallToolResult{IsError: true}, mcp.TextContent{Text: fmt.Sprintf("failed to decode response: %v", err)}, nil
	}

	resultText := fmt.Sprintf("Found %d brand(s):\n", len(apiResp.Data))
	for _, b := range apiResp.Data {
		resultText += fmt.Sprintf("- %s (ID: %d, Slug: %s, Active: %v)\n", b.Name, b.ID, b.Slug, b.IsActive)
	}

	return &mcp.CallToolResult{}, mcp.TextContent{Text: resultText}, nil
}
