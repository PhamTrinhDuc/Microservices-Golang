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
		Name: "list_categories",
		Description: `List all product categories available in the catalog. Returns category name, ID, and slug.
		USE when: you need a category_id or category name but don't have it yet, or when the user asks what product types are available.
		DO NOT USE when: you already know the category_id, or the user's query is specific enough to search directly by keyword.`,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"page": map[string]any{
					"type":        "number",
					"description": "Page number (default 1)",
				},
				"limit": map[string]any{
					"type":        "number",
					"description": "Number of items per page (default 10)",
				},
				"is_popular": map[string]any{
					"type":        []string{"boolean", "string"},
					"description": "true if you want popular categories only, false otherwise. Accepts true/false/True/False",
					"default":     true,
				},
				"search_term": map[string]any{
					"type":        []string{"string", "number"},
					"description": "Keyword to search for a specific category by name (e.g. 'laptop', 'smartwatch'). Use this to resolve category_id without fetching all categories.",
				},
			},
		},
	}
}

type ListCategoriesArgs struct {
	Page       float64              `json:"page"`
	Limit      float64              `json:"limit"`
	IsPopular  *utils.FlexibleBool  `json:"is_popular"`
	SearchTerm utils.FlexibleString `json:"search_term"`
}

// Handler
func (t *CategoryTool) ListCategoryHandler(ctx context.Context, req *mcp.CallToolRequest, args ListCategoriesArgs) (*mcp.CallToolResult, any, error) {
	u, _ := url.Parse(fmt.Sprintf("%s/categories", t.baseURL))
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
		Data []Category `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return &mcp.CallToolResult{IsError: true}, mcp.TextContent{Text: fmt.Sprintf("failed to decode response: %v", err)}, nil
	}

	resultText := fmt.Sprintf("Found %d category(ies):\n", len(apiResp.Data))
	for _, c := range apiResp.Data {
		resultText += fmt.Sprintf("- %s (ID: %d, Slug: %s)\n", c.Name, c.ID, c.Slug)
	}

	return &mcp.CallToolResult{}, mcp.TextContent{Text: resultText}, nil
}
