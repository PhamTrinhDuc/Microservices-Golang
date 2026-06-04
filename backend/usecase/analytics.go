package usecase

import (
	"context"
	"fmt"
	"time"

	"backend/domain"
)

type AnalyticsUsecase struct {
	repo domain.AnalyticsRepository
}

func NewAnalyticsUsecase(repo domain.AnalyticsRepository) *AnalyticsUsecase {
	return &AnalyticsUsecase{repo: repo}
}

func (u *AnalyticsUsecase) GetSummary(ctx context.Context, start, end time.Time, storeID *int) (*domain.AnalyticsSummary, error) {
	// 1. Calculate duration and previous period bounds
	duration := end.Sub(start)
	prevStart := start.Add(-duration)
	prevEnd := start

	// 2. Fetch current and previous period KPI metrics
	currMetrics, err := u.repo.GetPeriodMetrics(ctx, start, end, storeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current metrics: %w", err)
	}

	prevMetrics, err := u.repo.GetPeriodMetrics(ctx, prevStart, prevEnd, storeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get previous metrics: %w", err)
	}

	// 3. Calculate growth percentages
	var salesGrowth, ordersGrowth, aovGrowth float64

	if prevMetrics.TotalSales > 0 {
		salesGrowth = ((currMetrics.TotalSales - prevMetrics.TotalSales) / prevMetrics.TotalSales) * 100
	} else if currMetrics.TotalSales > 0 {
		salesGrowth = 100.0
	}

	if prevMetrics.TotalOrders > 0 {
		ordersGrowth = (float64(currMetrics.TotalOrders - prevMetrics.TotalOrders) / float64(prevMetrics.TotalOrders)) * 100
	} else if currMetrics.TotalOrders > 0 {
		ordersGrowth = 100.0
	}

	if prevMetrics.AverageOrderValue > 0 {
		aovGrowth = ((currMetrics.AverageOrderValue - prevMetrics.AverageOrderValue) / prevMetrics.AverageOrderValue) * 100
	} else if currMetrics.AverageOrderValue > 0 {
		aovGrowth = 100.0
	}

	// 4. Fetch sub-aggregates
	inventory, err := u.repo.GetInventorySummary(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory summary: %w", err)
	}

	salesOverTime, err := u.repo.GetSalesOverTime(ctx, start, end, storeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales trends: %w", err)
	}

	topProducts, err := u.repo.GetTopProducts(ctx, start, end, storeID, 5)
	if err != nil {
		return nil, fmt.Errorf("failed to get top products: %w", err)
	}

	topCategories, err := u.repo.GetTopCategories(ctx, start, end, storeID, 5)
	if err != nil {
		return nil, fmt.Errorf("failed to get top categories: %w", err)
	}

	storeSales, err := u.repo.GetStoreSales(ctx, start, end, 5)
	if err != nil {
		return nil, fmt.Errorf("failed to get store sales: %w", err)
	}

	statusDist, err := u.repo.GetStatusDistribution(ctx, start, end, storeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order status distribution: %w", err)
	}

	// 5. Assemble summary
	summary := &domain.AnalyticsSummary{
		TotalSales:            currMetrics.TotalSales,
		TotalOrders:           currMetrics.TotalOrders,
		AverageOrderValue:     currMetrics.AverageOrderValue,
		ItemsSold:             currMetrics.ItemsSold,
		PrevTotalSales:        prevMetrics.TotalSales,
		PrevTotalOrders:       prevMetrics.TotalOrders,
		PrevAverageOrderValue: prevMetrics.AverageOrderValue,
		SalesGrowth:           salesGrowth,
		OrdersGrowth:          ordersGrowth,
		AovGrowth:             aovGrowth,
		Inventory:             *inventory,
		SalesOverTime:         salesOverTime,
		TopProducts:           topProducts,
		TopCategories:         topCategories,
		StoreSales:            storeSales,
		StatusDistribution:    statusDist,
	}

	return summary, nil
}
