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

type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	CategoryID  int     `json:"category_id"`
	BrandID     int     `json:"brand_id"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	Rating      float64 `json:"rating"`
	ReviewCount int     `json:"review_count"`
}

type ProductTool struct {
	baseURL string
	client  *http.Client
}

func NewProductTool(baseURL string) *ProductTool {
	return &ProductTool{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// Definition returns the tool definitions for Product API
func (t *ProductTool) ListProductDefinition() *mcp.Tool {
	return &mcp.Tool{
		Name:        "list_products",
		Description: "Search and list products with pagination and filters (category, brand, price, rating, availability, specification filters).",
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
				"category_id": map[string]any{
					"type":        "number",
					"description": "Filter by Category ID",
				},
				"brand_id": map[string]any{
					"type":        "number",
					"description": "Filter by Brand ID",
				},
				"q": map[string]any{
					"type":        "string",
					"description": "Search keyword matching name, slug, ID or variant SKU",
				},
				"sort": map[string]any{
					"type":        "string",
					"description": "Sort products: 'price_asc', 'price_desc', 'rating_desc', 'newest'",
				},
				"min_price": map[string]any{
					"type":        "number",
					"description": "Minimum product price",
				},
				"max_price": map[string]any{
					"type":        "number",
					"description": "Maximum product price",
				},
				"min_rating": map[string]any{
					"type":        "number",
					"description": "Minimum rating",
				},
				"in_stock_only": map[string]any{
					"type":        "boolean",
					"description": "Only show products that are currently in stock (quantity > 0)",
				},
				"spec_filter": map[string]any{
					"type":        "object",
					"description": "Key-value pair object of specific specs to filter (e.g. {'Screen': '6.1', 'Resolution': 'Retina'})",
				},
			},
		},
	}
}

func (t *ProductTool) GetProductDefinition() *mcp.Tool {
	return &mcp.Tool{
		Name:        "get_product_by_id",
		Description: "Get details of a specific product by their ID (including brand, category, specs, and variants).",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]any{
					"type":        "string",
					"description": "The unique ID of the product",
				},
			},
			"required": []string{"id"},
		},
	}
}

type ListProductsArgs struct {
	Page        float64                `json:"page"`
	Limit       float64                `json:"limit"`
	CategoryID  *float64               `json:"category_id"`
	BrandID     *float64               `json:"brand_id"`
	Q           string                 `json:"q"`
	Sort        string                 `json:"sort"`
	MinPrice    *float64               `json:"min_price"`
	MaxPrice    *float64               `json:"max_price"`
	MinRating   *float64               `json:"min_rating"`
	InStockOnly bool                   `json:"in_stock_only"`
	SpecFilter  map[string]interface{} `json:"spec_filter"`
}

type GetProductArgs struct {
	ID string `json:"id"`
}

// Handlers
func (t *ProductTool) ListProductHandler(ctx context.Context, req *mcp.CallToolRequest, args ListProductsArgs) (*mcp.CallToolResult, any, error) {
	u, _ := url.Parse(fmt.Sprintf("%s/products", t.baseURL))
	q := u.Query()

	if args.Page > 0 {
		q.Set("page", fmt.Sprintf("%.0f", args.Page))
	}
	if args.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%.0f", args.Limit))
	}
	if args.CategoryID != nil {
		q.Set("category_id", fmt.Sprintf("%.0f", *args.CategoryID))
	}
	if args.BrandID != nil {
		q.Set("brand_id", fmt.Sprintf("%.0f", *args.BrandID))
	}
	if args.Q != "" {
		q.Set("q", args.Q)
	}
	if args.Sort != "" {
		q.Set("sort", args.Sort)
	}
	if args.MinPrice != nil {
		q.Set("min_price", fmt.Sprintf("%.2f", *args.MinPrice))
	}
	if args.MaxPrice != nil {
		q.Set("max_price", fmt.Sprintf("%.2f", *args.MaxPrice))
	}
	if args.MinRating != nil {
		q.Set("min_rating", fmt.Sprintf("%.2f", *args.MinRating))
	}
	if args.InStockOnly {
		q.Set("in_stock_only", "true")
	}
	if args.SpecFilter != nil {
		for key, val := range args.SpecFilter {
			q.Set(fmt.Sprintf("spec_filter[%s]", key), fmt.Sprintf("%v", val))
		}
	}
	u.RawQuery = q.Encode()
	apiURL := u.String()

	// 1. Tạo request mới với Context
	httpReq, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, nil, err
	}

	// 2. Inject Trace Context
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, propagation.HeaderCarrier(httpReq.Header))

	// 3. Thực hiện gọi API
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
			Products       []Product      `json:"products"`
			TotalCount     int            `json:"total_count"`
			Page           int            `json:"page"`
			Limit          int            `json:"limit"`
			HasMore        bool           `json:"has_more"`
			AppliedFilters []string       `json:"applied_filters"`
			Suggestions    map[string]any `json:"suggestions"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return &mcp.CallToolResult{IsError: true}, mcp.TextContent{Text: fmt.Sprintf("failed to decode response: %v", err)}, nil
	}

	resultText := fmt.Sprintf("Found %d product(s) (Total: %d, Page: %d, Limit: %d):\n", 
		len(apiResp.Data.Products), apiResp.Data.TotalCount, apiResp.Data.Page, apiResp.Data.Limit)
	for _, s := range apiResp.Data.Products {
		resultText += fmt.Sprintf("- %s (ID: %s, Price: %.2f, Stock: %d, Rating: %.1f, Reviews: %d)\n",
			s.Name, s.ID, s.Price, s.Stock, s.Rating, s.ReviewCount)
	}

	if len(apiResp.Data.AppliedFilters) > 0 {
		resultText += fmt.Sprintf("\nApplied Filters: %v\n", apiResp.Data.AppliedFilters)
	}
	if apiResp.Data.HasMore {
		resultText += "\nNote: There are more products available matching these criteria (HasMore: true).\n"
	}
	if len(apiResp.Data.Suggestions) > 0 {
		resultText += "\nNo results were found with the applied filters, but here are some search suggestions:\n"
		sugBytes, _ := json.MarshalIndent(apiResp.Data.Suggestions, "", "  ")
		resultText += string(sugBytes) + "\n"
	}

	return &mcp.CallToolResult{}, mcp.TextContent{Text: resultText}, nil
}

func (t *ProductTool) GetProductHandler(ctx context.Context, req *mcp.CallToolRequest, args GetProductArgs) (*mcp.CallToolResult, any, error) {
	apiURL := fmt.Sprintf("%s/products/%s", t.baseURL, args.ID)

	// 1. Tạo request mới với Context
	httpReq, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, nil, err
	}

	// 2. Inject Trace Context
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, propagation.HeaderCarrier(httpReq.Header))

	// 3. Thực hiện gọi API
	resp, err := t.client.Do(httpReq)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, mcp.TextContent{Text: fmt.Sprintf("API request failed: %v", err)}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &mcp.CallToolResult{IsError: true}, mcp.TextContent{Text: "Product not found"}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return &mcp.CallToolResult{IsError: true}, mcp.TextContent{Text: fmt.Sprintf("API returned error: %s", resp.Status)}, nil
	}

	var apiResp struct {
		Data struct {
			Product      Product `json:"product"`
			BrandName    string  `json:"brand_name"`
			CategoryName string  `json:"category_name"`
			Specs        []struct {
				Group string  `json:"group"`
				Key   string  `json:"key"`
				Value string  `json:"value"`
				Unit  *string `json:"unit"`
			} `json:"specs"`
			Variants []struct {
				ID        int     `json:"id"`
				Name      string  `json:"name"`
				SKU       string  `json:"sku"`
				SellPrice float64 `json:"sell_price"`
			} `json:"variants"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return &mcp.CallToolResult{IsError: true}, mcp.TextContent{Text: fmt.Sprintf("failed to decode response: %v", err)}, nil
	}

	p := apiResp.Data.Product
	resultText := fmt.Sprintf("Product Details:\n")
	resultText += fmt.Sprintf("ID: %s\nName: %s\nCategory: %s (ID: %d)\nBrand: %s (ID: %d)\nPrice: %.2f\nStock: %d\nRating: %.1f (%d reviews)\n",
		p.ID, p.Name, apiResp.Data.CategoryName, p.CategoryID, apiResp.Data.BrandName, p.BrandID, p.Price, p.Stock, p.Rating, p.ReviewCount)

	if len(apiResp.Data.Specs) > 0 {
		resultText += "\nSpecifications:\n"
		for _, sp := range apiResp.Data.Specs {
			unitStr := ""
			if sp.Unit != nil {
				unitStr = " " + *sp.Unit
			}
			resultText += fmt.Sprintf("- [%s] %s: %s%s\n", sp.Group, sp.Key, sp.Value, unitStr)
		}
	}

	if len(apiResp.Data.Variants) > 0 {
		resultText += "\nVariants:\n"
		for _, v := range apiResp.Data.Variants {
			resultText += fmt.Sprintf("- %s (SKU: %s, Price: %.2f)\n", v.Name, v.SKU, v.SellPrice)
		}
	}

	return &mcp.CallToolResult{}, mcp.TextContent{Text: resultText}, nil
}
