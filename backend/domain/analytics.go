package domain

import (
	"context"
	"time"
)

type InventorySummary struct {
	TotalSku        int `json:"total_sku"`
	TotalQuantity   int `json:"total_quantity"`
	LowStockCount   int `json:"low_stock_count"`
	OutOfStockCount int `json:"out_of_stock_count"`
}

type TopCategoryReport struct {
	CategoryName string  `json:"category_name"`
	SoldQty      int     `json:"sold_qty"`
	Revenue      float64 `json:"revenue"`
}

type PeriodMetrics struct {
	TotalSales        float64 `json:"total_sales"`
	TotalOrders        int     `json:"total_orders"`
	AverageOrderValue float64 `json:"average_order_value"`
	ItemsSold         int     `json:"items_sold"`
}

type SalesOverTime struct {
	Date        string  `json:"date"`
	Revenue     float64 `json:"revenue"`
	OrdersCount int     `json:"orders_count"`
}

type TopProductReport struct {
	ProductName string  `json:"product_name"`
	SoldQty     int     `json:"sold_qty"`
	Revenue     float64 `json:"revenue"`
}

type StoreReport struct {
	StoreID     int     `json:"store_id"`
	StoreName   string  `json:"store_name"`
	OrdersCount int     `json:"orders_count"`
	Revenue     float64 `json:"revenue"`
}

type StatusDistribution struct {
	StatusLabel string `json:"status_label"`
	Count       int    `json:"count"`
}

type AnalyticsSummary struct {
	TotalSales            float64               `json:"total_sales"`
	TotalOrders           int                   `json:"total_orders"`
	AverageOrderValue     float64               `json:"average_order_value"`
	ItemsSold             int                   `json:"items_sold"`
	PrevTotalSales        float64               `json:"prev_total_sales"`
	PrevTotalOrders       int                   `json:"prev_total_orders"`
	PrevAverageOrderValue float64               `json:"prev_average_order_value"`
	SalesGrowth           float64               `json:"sales_growth"`
	OrdersGrowth          float64               `json:"orders_growth"`
	AovGrowth             float64               `json:"aov_growth"`
	Inventory             InventorySummary      `json:"inventory"`
	SalesOverTime         []*SalesOverTime      `json:"sales_over_time"`
	TopProducts           []*TopProductReport   `json:"top_products"`
	TopCategories         []*TopCategoryReport  `json:"top_categories"`
	StoreSales            []*StoreReport        `json:"store_sales"`
	StatusDistribution    []*StatusDistribution `json:"status_distribution"`
}

type AnalyticsRepository interface {
	GetPeriodMetrics(ctx context.Context, start, end time.Time, storeID *int) (*PeriodMetrics, error)
	GetInventorySummary(ctx context.Context) (*InventorySummary, error)
	GetSalesOverTime(ctx context.Context, start, end time.Time, storeID *int) ([]*SalesOverTime, error)
	GetTopProducts(ctx context.Context, start, end time.Time, storeID *int, limit int) ([]*TopProductReport, error)
	GetTopCategories(ctx context.Context, start, end time.Time, storeID *int, limit int) ([]*TopCategoryReport, error)
	GetStoreSales(ctx context.Context, start, end time.Time, limit int) ([]*StoreReport, error)
	GetStatusDistribution(ctx context.Context, start, end time.Time, storeID *int) ([]*StatusDistribution, error)
}

type AnalyticsUsecase interface {
	GetSummary(ctx context.Context, start, end time.Time, storeID *int) (*AnalyticsSummary, error)
}
