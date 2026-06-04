package repository

import (
	"context"
	"fmt"
	"time"

	"backend/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AnalyticsRepository struct {
	db *pgxpool.Pool
}

func NewAnalyticsRepository(db *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

func (r *AnalyticsRepository) GetPeriodMetrics(ctx context.Context, start, end time.Time, storeID *int) (*domain.PeriodMetrics, error) {
	query := `
		SELECT 
			COALESCE(SUM(o.total_amount), 0.0) as total_sales,
			COUNT(DISTINCT o.id) as total_orders,
			COALESCE(SUM(od.quantity), 0) as items_sold
		FROM orders o
		LEFT JOIN order_details od ON o.id = od.order_id
		JOIN order_status os ON o.order_status_id = os.id
		WHERE o.created_at >= $1 AND o.created_at <= $2
		  AND os.code != 'cancelled'`

	var args []interface{}
	args = append(args, start, end)

	if storeID != nil {
		query += " AND o.store_id = $3"
		args = append(args, *storeID)
	}

	metrics := &domain.PeriodMetrics{}
	err := r.db.QueryRow(ctx, query, args...).Scan(&metrics.TotalSales, &metrics.TotalOrders, &metrics.ItemsSold)
	if err != nil {
		return nil, fmt.Errorf("failed to query period metrics: %w", err)
	}

	if metrics.TotalOrders > 0 {
		metrics.AverageOrderValue = metrics.TotalSales / float64(metrics.TotalOrders)
	}

	return metrics, nil
}

func (r *AnalyticsRepository) GetInventorySummary(ctx context.Context) (*domain.InventorySummary, error) {
	summary := &domain.InventorySummary{}

	// Query 1: Total SKUs
	skuQuery := `SELECT COUNT(id) FROM product_variant WHERE is_deleted = false AND is_active = true`
	err := r.db.QueryRow(ctx, skuQuery).Scan(&summary.TotalSku)
	if err != nil {
		return nil, fmt.Errorf("failed to query total sku count: %w", err)
	}

	// Query 2: Total Quantity in Stock
	qtyQuery := `SELECT COALESCE(SUM(quantity), 0) FROM product_inventory`
	err = r.db.QueryRow(ctx, qtyQuery).Scan(&summary.TotalQuantity)
	if err != nil {
		return nil, fmt.Errorf("failed to query total quantity: %w", err)
	}

	// Query 3: Low Stock Count
	lowStockQuery := `
		SELECT COUNT(DISTINCT pv.id)
		FROM product_variant pv
		JOIN product p ON pv.product_id = p.id
		LEFT JOIN (
			SELECT variant_id, SUM(quantity) as qty
			FROM product_inventory
			GROUP BY variant_id
		) pi ON pv.id = pi.variant_id
		WHERE pv.is_deleted = false AND pv.is_active = true
		  AND COALESCE(pi.qty, 0) <= p.low_stock_threshold`
	err = r.db.QueryRow(ctx, lowStockQuery).Scan(&summary.LowStockCount)
	if err != nil {
		return nil, fmt.Errorf("failed to query low stock count: %w", err)
	}

	// Query 4: Out of Stock Count
	outOfStockQuery := `
		SELECT COUNT(DISTINCT pv.id)
		FROM product_variant pv
		LEFT JOIN (
			SELECT variant_id, SUM(quantity) as qty
			FROM product_inventory
			GROUP BY variant_id
		) pi ON pv.id = pi.variant_id
		WHERE pv.is_deleted = false AND pv.is_active = true
		  AND COALESCE(pi.qty, 0) <= 0`
	err = r.db.QueryRow(ctx, outOfStockQuery).Scan(&summary.OutOfStockCount)
	if err != nil {
		return nil, fmt.Errorf("failed to query out of stock count: %w", err)
	}

	return summary, nil
}

func (r *AnalyticsRepository) GetSalesOverTime(ctx context.Context, start, end time.Time, storeID *int) ([]*domain.SalesOverTime, error) {
	query := `
		SELECT 
			TO_CHAR(o.created_at, 'YYYY-MM-DD') as date_str,
			COALESCE(SUM(o.total_amount), 0.0) as revenue,
			COUNT(o.id) as orders_count
		FROM orders o
		JOIN order_status os ON o.order_status_id = os.id
		WHERE o.created_at >= $1 AND o.created_at <= $2
		  AND os.code != 'cancelled'`

	var args []interface{}
	args = append(args, start, end)

	if storeID != nil {
		query += " AND o.store_id = $3"
		args = append(args, *storeID)
	}

	query += `
		GROUP BY date_str
		ORDER BY date_str ASC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query sales over time: %w", err)
	}
	defer rows.Close()

	var list []*domain.SalesOverTime
	for rows.Next() {
		item := &domain.SalesOverTime{}
		err := rows.Scan(&item.Date, &item.Revenue, &item.OrdersCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sales over time row: %w", err)
		}
		list = append(list, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *AnalyticsRepository) GetTopProducts(ctx context.Context, start, end time.Time, storeID *int, limit int) ([]*domain.TopProductReport, error) {
	query := `
		SELECT 
			p.name as product_name,
			SUM(od.quantity) as sold_qty,
			SUM(od.total_cost) as revenue
		FROM order_details od
		JOIN product_variant pv ON od.variant_id = pv.id
		JOIN product p ON pv.product_id = p.id
		JOIN orders o ON od.order_id = o.id
		JOIN order_status os ON o.order_status_id = os.id
		WHERE o.created_at >= $1 AND o.created_at <= $2
		  AND os.code != 'cancelled'`

	var args []interface{}
	args = append(args, start, end)
	placeholderIdx := 3

	if storeID != nil {
		query += fmt.Sprintf(" AND o.store_id = $%d", placeholderIdx)
		args = append(args, *storeID)
		placeholderIdx++
	}

	query += fmt.Sprintf(`
		GROUP BY p.name
		ORDER BY sold_qty DESC, revenue DESC
		LIMIT $%d`, placeholderIdx)
	args = append(args, limit)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query top products: %w", err)
	}
	defer rows.Close()

	var list []*domain.TopProductReport
	for rows.Next() {
		item := &domain.TopProductReport{}
		err := rows.Scan(&item.ProductName, &item.SoldQty, &item.Revenue)
		if err != nil {
			return nil, fmt.Errorf("failed to scan top products row: %w", err)
		}
		list = append(list, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *AnalyticsRepository) GetTopCategories(ctx context.Context, start, end time.Time, storeID *int, limit int) ([]*domain.TopCategoryReport, error) {
	query := `
		SELECT 
			c.name as category_name,
			SUM(od.quantity) as sold_qty,
			SUM(od.total_cost) as revenue
		FROM order_details od
		JOIN product_variant pv ON od.variant_id = pv.id
		JOIN product p ON pv.product_id = p.id
		JOIN category c ON p.category_id = c.id
		JOIN orders o ON od.order_id = o.id
		JOIN order_status os ON o.order_status_id = os.id
		WHERE o.created_at >= $1 AND o.created_at <= $2
		  AND os.code != 'cancelled'`

	var args []interface{}
	args = append(args, start, end)
	placeholderIdx := 3

	if storeID != nil {
		query += fmt.Sprintf(" AND o.store_id = $%d", placeholderIdx)
		args = append(args, *storeID)
		placeholderIdx++
	}

	query += fmt.Sprintf(`
		GROUP BY c.name
		ORDER BY revenue DESC, sold_qty DESC
		LIMIT $%d`, placeholderIdx)
	args = append(args, limit)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query top categories: %w", err)
	}
	defer rows.Close()

	var list []*domain.TopCategoryReport
	for rows.Next() {
		item := &domain.TopCategoryReport{}
		err := rows.Scan(&item.CategoryName, &item.SoldQty, &item.Revenue)
		if err != nil {
			return nil, fmt.Errorf("failed to scan top categories row: %w", err)
		}
		list = append(list, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *AnalyticsRepository) GetStoreSales(ctx context.Context, start, end time.Time, limit int) ([]*domain.StoreReport, error) {
	query := `
		SELECT 
			o.store_id,
			s.name as store_name,
			COUNT(DISTINCT o.id) as orders_count,
			SUM(o.total_amount) as revenue
		FROM orders o
		JOIN store s ON o.store_id = s.id
		JOIN order_status os ON o.order_status_id = os.id
		WHERE o.created_at >= $1 AND o.created_at <= $2
		  AND os.code != 'cancelled'
		GROUP BY o.store_id, s.name
		ORDER BY revenue DESC
		LIMIT $3`

	rows, err := r.db.Query(ctx, query, start, end, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query store sales: %w", err)
	}
	defer rows.Close()

	var list []*domain.StoreReport
	for rows.Next() {
		item := &domain.StoreReport{}
		err := rows.Scan(&item.StoreID, &item.StoreName, &item.OrdersCount, &item.Revenue)
		if err != nil {
			return nil, fmt.Errorf("failed to scan store sales row: %w", err)
		}
		list = append(list, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *AnalyticsRepository) GetStatusDistribution(ctx context.Context, start, end time.Time, storeID *int) ([]*domain.StatusDistribution, error) {
	query := `
		SELECT 
			os.label as status_label,
			COUNT(o.id) as count
		FROM orders o
		JOIN order_status os ON o.order_status_id = os.id
		WHERE o.created_at >= $1 AND o.created_at <= $2`

	var args []interface{}
	args = append(args, start, end)

	if storeID != nil {
		query += " AND o.store_id = $3"
		args = append(args, *storeID)
	}

	query += `
		GROUP BY os.label
		ORDER BY count DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query status distribution: %w", err)
	}
	defer rows.Close()

	var list []*domain.StatusDistribution
	for rows.Next() {
		item := &domain.StatusDistribution{}
		err := rows.Scan(&item.StatusLabel, &item.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan status distribution row: %w", err)
		}
		list = append(list, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}
