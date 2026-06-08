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
)

// Stylist represents the response from backend
type Stylist struct {
	ID       string `json:"id"`
	BranchID string `json:"branch_id"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	IsActive bool   `json:"is_active"`
}

type StylistTool struct {
	baseURL string
	client  *http.Client
}

func NewStylistTool(baseURL string) *StylistTool {
	return &StylistTool{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// Definition returns the tool definitions for Stylist API
func (t *StylistTool) ListStylistDefinition() *mcp.Tool {
	return &mcp.Tool{
		Name:        "list_stylists",
		Description: "List all stylists from backend API. Supports filtering by branch and pagination.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"branch_id": map[string]any{
					"type":        "string",
					"description": "Optional branch ID to filter stylists",
				},
				"page": map[string]any{
					"type":        "number",
					"description": "Page number (default 1)",
				},
				"limit": map[string]any{
					"type":        "number",
					"description": "Number of items per page (default 10)",
				},
			},
		},
	}
}

func (t *StylistTool) GetStylistDefinition() *mcp.Tool {
	return &mcp.Tool{
		Name:        "get_stylist_by_id",
		Description: "Get details of a specific stylist by their ID.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]any{
					"type":        "string",
					"description": "The unique ID of the stylist",
				},
			},
			"required": []string{"id"},
		},
	}
}

type ListStylistsArgs struct {
	BranchID string  `json:"branch_id"`
	Page     float64 `json:"page"`
	Limit    float64 `json:"limit"`
}

type GetStylistArgs struct {
	ID string `json:"id"`
}

// Handlers
func (t *StylistTool) ListStylistHandler(ctx context.Context, req *mcp.CallToolRequest, args ListStylistsArgs) (*mcp.CallToolResult, any, error) {
	u, _ := url.Parse(fmt.Sprintf("%s/stylists", t.baseURL))
	q := u.Query()

	if args.BranchID != "" {
		u.Path = fmt.Sprintf("/stylists/branch/%s", args.BranchID)
	}

	if args.Page > 0 {
		q.Set("page", fmt.Sprintf("%.0f", args.Page))
	}
	if args.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%.0f", args.Limit))
	}
	u.RawQuery = q.Encode()

	// 1. Tạo request mới
	httpReq, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, nil, err
	}
	// 2. CỰC KỲ QUAN TRỌNG: Inject Trace Context vào Header của request gửi đi
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, propagation.HeaderCarrier(httpReq.Header))

	// 3. Thực hiện gọi API bằng request đã được inject header
	resp, err := t.client.Do(httpReq)

	if err != nil {
		return &mcp.CallToolResult{IsError: true}, mcp.TextContent{Text: fmt.Sprintf("API request failed: %v", err)}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &mcp.CallToolResult{IsError: true}, mcp.TextContent{Text: fmt.Sprintf("API returned error: %s", resp.Status)}, nil
	}

	var apiResp struct {
		Data struct {
			Items []Stylist `json:"items"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return &mcp.CallToolResult{IsError: true}, mcp.TextContent{Text: fmt.Sprintf("failed to decode response: %v", err)}, nil
	}

	resultText := fmt.Sprintf("Found %d stylist(s):\n", len(apiResp.Data.Items))
	for _, s := range apiResp.Data.Items {
		resultText += fmt.Sprintf("- %s (Phone: %s, Branch: %s)\n", s.Name, s.Phone, s.BranchID)
	}

	return &mcp.CallToolResult{}, mcp.TextContent{Text: resultText}, nil
}

func (t *StylistTool) GetStylistHandler(ctx context.Context, req *mcp.CallToolRequest, args GetStylistArgs) (*mcp.CallToolResult, any, error) {
	apiURL := fmt.Sprintf("%s/stylists/%s", t.baseURL, args.ID)

	// 1. Tạo request mới với Context
	httpReq, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, nil, err
	}

	// 2. Inject Trace Context vào Header
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, propagation.HeaderCarrier(httpReq.Header))

	// 3. Thực hiện gọi API
	resp, err := t.client.Do(httpReq)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, mcp.TextContent{Text: fmt.Sprintf("API request failed: %v", err)}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &mcp.CallToolResult{IsError: true}, mcp.TextContent{Text: "Stylist not found"}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return &mcp.CallToolResult{IsError: true}, mcp.TextContent{Text: fmt.Sprintf("API returned error: %s", resp.Status)}, nil
	}

	var apiResp struct {
		Data Stylist `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return &mcp.CallToolResult{IsError: true}, mcp.TextContent{Text: fmt.Sprintf("failed to decode response: %v", err)}, nil
	}

	s := apiResp.Data
	resultText := fmt.Sprintf("Stylist Details:\nID: %s\nName: %s\nPhone: %s\nBranch ID: %s\nActive: %v\n",
		s.ID, s.Name, s.Phone, s.BranchID, s.IsActive)

	return &mcp.CallToolResult{}, mcp.TextContent{Text: resultText}, nil
}
