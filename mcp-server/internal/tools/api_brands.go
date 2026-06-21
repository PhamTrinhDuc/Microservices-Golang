package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"mcp-server/internal/utils"
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

type ListBrandsArgs struct {
	Page       float64              `json:"page"`
	Limit      float64              `json:"limit"`
	IsPopular  *utils.FlexibleBool  `json:"is_popular"`
	SearchTerm utils.FlexibleString `json:"search_term"`
}

// Definition returns the tool definitions for Brand API
func (t *BrandTool) ListBrandDefinition() *mcp.Tool {
	return &mcp.Tool{
		Name:        "list_brands",
		Description: "List all product brands.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"page": map[string]any{
					"type":        "number",
					"description": "Page number",
					"default":     1,
				},
				"limit": map[string]any{
					"type":        "number",
					"description": "Number of items per page",
					"default":     10,
				},
				"is_popular": map[string]any{
					"type":        []string{"boolean", "string"},
					"description": "true if you need popular brands only, false otherwise. Accepts true/false/True/False",
					"default":     true,
				},
				"search_term": map[string]any{
					"type":        []string{"string", "number"},
					"description": "Keyword to search for brands",
				},
			},
		},
	}
}

// Handler
func (t *BrandTool) ListBrandHandler(ctx context.Context, req *mcp.CallToolRequest, args ListBrandsArgs) (*mcp.CallToolResult, any, error) {
	u, _ := url.Parse(fmt.Sprintf("%s/brands", t.baseURL))
	q := u.Query()
	if args.Page > 0 {
		q.Set("page", fmt.Sprintf("%.0f", args.Page))
	}
	if args.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%.0f", args.Limit))
	}
	if args.IsPopular != nil && bool(*args.IsPopular) {
		q.Set("is_popular", "true")
	}
	if args.SearchTerm != "" {
		q.Set("search_term", string(args.SearchTerm))
	}
	
	u.RawQuery = q.Encode()
	apiURL := u.String()

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
