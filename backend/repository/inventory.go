package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"backend/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InventoryRepository struct {
	db *pgxpool.Pool
}

func NewInventoryRepository(db *pgxpool.Pool) *InventoryRepository {
	return &InventoryRepository{db: db}
}

// --- Store CRUD ---

func (r *InventoryRepository) CreateStore(ctx context.Context, store *domain.Store) (*domain.Store, error) {
	query := `
		INSERT INTO store (name, hotline, district, province, ward, road, email, lat, lng, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, true, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, query,
		store.Name,
		store.Hotline,
		store.District,
		store.Province,
		store.Ward,
		store.Road,
		store.Email,
		store.Lat,
		store.Lng,
	).Scan(&store.ID, &store.CreatedAt, &store.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create store: %w", err)
	}

	store.IsActive = true
	return store, nil
}

func (r *InventoryRepository) ListStores(ctx context.Context, province string, district string) ([]*domain.Store, error) {
	query := `
		SELECT id, name, hotline, district, province, ward, road, email, lat, lng, is_active, created_at, updated_at
		FROM store
		WHERE is_active = true`

	var args []interface{}
	placeholderIdx := 1

	if province != "" {
		query += fmt.Sprintf(" AND province ILIKE $%d", placeholderIdx)
		args = append(args, "%"+province+"%")
		placeholderIdx++
	}

	if district != "" {
		query += fmt.Sprintf(" AND district ILIKE $%d", placeholderIdx)
		args = append(args, "%"+district+"%")
		placeholderIdx++
	}

	query += " ORDER BY id ASC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query stores: %w", err)
	}
	defer rows.Close()

	stores := make([]*domain.Store, 0)
	for rows.Next() {
		s := &domain.Store{}
		err := rows.Scan(
			&s.ID,
			&s.Name,
			&s.Hotline,
			&s.District,
			&s.Province,
			&s.Ward,
			&s.Road,
			&s.Email,
			&s.Lat,
			&s.Lng,
			&s.IsActive,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan store: %w", err)
		}
		stores = append(stores, s)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("stores list rows error: %w", err)
	}

	return stores, nil
}

func (r *InventoryRepository) GetStoreByID(ctx context.Context, id int) (*domain.Store, error) {
	s := &domain.Store{}
	query := `
		SELECT id, name, hotline, district, province, ward, road, email, lat, lng, is_active, created_at, updated_at
		FROM store
		WHERE id = $1`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&s.ID,
		&s.Name,
		&s.Hotline,
		&s.District,
		&s.Province,
		&s.Ward,
		&s.Road,
		&s.Email,
		&s.Lat,
		&s.Lng,
		&s.IsActive,
		&s.CreatedAt,
		&s.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrStoreNotFound
		}
		return nil, fmt.Errorf("failed to get store by ID: %w", err)
	}

	return s, nil
}

func (r *InventoryRepository) UpdateStore(ctx context.Context, store *domain.Store) (*domain.Store, error) {
	query := `
		UPDATE store
		SET name = $1, hotline = $2, district = $3, province = $4, ward = $5, road = $6, email = $7, lat = $8, lng = $9, is_active = $10, updated_at = NOW()
		WHERE id = $11
		RETURNING updated_at`

	err := r.db.QueryRow(ctx, query,
		store.Name,
		store.Hotline,
		store.District,
		store.Province,
		store.Ward,
		store.Road,
		store.Email,
		store.Lat,
		store.Lng,
		store.IsActive,
		store.ID,
	).Scan(&store.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrStoreNotFound
		}
		return nil, fmt.Errorf("failed to update store: %w", err)
	}

	return store, nil
}

func (r *InventoryRepository) DeactivateStore(ctx context.Context, id int) error {
	query := `UPDATE store SET is_active = false, updated_at = NOW() WHERE id = $1`
	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to deactivate store: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrStoreNotFound
	}

	return nil
}

// --- Supplier CRUD ---

func (r *InventoryRepository) CreateSupplier(ctx context.Context, supplier *domain.Supplier) (*domain.Supplier, error) {
	query := `
		INSERT INTO suppliers (name, address, phone, email, contact_name, contact_phone, is_deleted)
		VALUES ($1, $2, $3, $4, $5, $6, false)
		RETURNING id`

	err := r.db.QueryRow(ctx, query,
		supplier.Name,
		supplier.Address,
		supplier.Phone,
		supplier.Email,
		supplier.ContactName,
		supplier.ContactPhone,
	).Scan(&supplier.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to create supplier: %w", err)
	}

	supplier.IsDeleted = false
	return supplier, nil
}

func (r *InventoryRepository) ListSuppliers(ctx context.Context) ([]*domain.Supplier, error) {
	query := `
		SELECT s.id, s.name, s.address, s.phone, s.email, s.contact_name, s.contact_phone, s.is_deleted,
		       COALESCE(COUNT(DISTINCT ii.id), 0) as total_imports,
		       MAX(ii.created_at) as last_imported_at,
		       COALESCE(SUM(iid.quantity * iid.price_import), 0) as total_import_value
		FROM suppliers s
		LEFT JOIN import_invoices ii ON s.id = ii.supplier_id
		LEFT JOIN import_invoice_details iid ON ii.id = iid.invoice_id
		GROUP BY s.id
		ORDER BY s.id ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query suppliers: %w", err)
	}
	defer rows.Close()

	suppliers := make([]*domain.Supplier, 0)
	for rows.Next() {
		s := &domain.Supplier{}
		err := rows.Scan(
			&s.ID, &s.Name, &s.Address, &s.Phone, &s.Email, &s.ContactName, &s.ContactPhone, &s.IsDeleted,
			&s.TotalImports, &s.LastImportedAt, &s.TotalImportValue,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan supplier: %w", err)
		}
		suppliers = append(suppliers, s)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("suppliers list rows error: %w", err)
	}

	return suppliers, nil
}

func (r *InventoryRepository) GetSupplierByID(ctx context.Context, id int) (*domain.Supplier, error) {
	s := &domain.Supplier{}
	query := `
		SELECT s.id, s.name, s.address, s.phone, s.email, s.contact_name, s.contact_phone, s.is_deleted,
		       COALESCE(COUNT(DISTINCT ii.id), 0) as total_imports,
		       MAX(ii.created_at) as last_imported_at,
		       COALESCE(SUM(iid.quantity * iid.price_import), 0) as total_import_value
		FROM suppliers s
		LEFT JOIN import_invoices ii ON s.id = ii.supplier_id
		LEFT JOIN import_invoice_details iid ON ii.id = iid.invoice_id
		WHERE s.id = $1
		GROUP BY s.id`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.Name, &s.Address, &s.Phone, &s.Email, &s.ContactName, &s.ContactPhone, &s.IsDeleted,
		&s.TotalImports, &s.LastImportedAt, &s.TotalImportValue,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrSupplierNotFound
		}
		return nil, fmt.Errorf("failed to get supplier by ID: %w", err)
	}

	return s, nil
}

func (r *InventoryRepository) UpdateSupplier(ctx context.Context, supplier *domain.Supplier) (*domain.Supplier, error) {
	query := `
		UPDATE suppliers
		SET name = $1, address = $2, phone = $3, email = $4, contact_name = $5, contact_phone = $6, is_deleted = $7
		WHERE id = $8`

	tag, err := r.db.Exec(ctx, query,
		supplier.Name,
		supplier.Address,
		supplier.Phone,
		supplier.Email,
		supplier.ContactName,
		supplier.ContactPhone,
		supplier.IsDeleted,
		supplier.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update supplier: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return nil, domain.ErrSupplierNotFound
	}

	return supplier, nil
}

func (r *InventoryRepository) DeleteSupplier(ctx context.Context, id int) error {
	query := `UPDATE suppliers SET is_deleted = true WHERE id = $1`

	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete supplier: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrSupplierNotFound
	}

	return nil
}

// --- Inventory Transaction Operations ---

func (r *InventoryRepository) CreateImportInvoice(ctx context.Context, creatorID int, invoice *domain.ImportInvoice, details []*domain.ImportInvoiceDetail) (*domain.ImportInvoice, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Ensure status is valid, default to published
	if invoice.Status == "" {
		invoice.Status = "published"
	}

	// 1. Insert import invoice
	invoiceQuery := `
		INSERT INTO import_invoices (supplier_id, store_id, created_by, total_items, note, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING id, created_at`

	err = tx.QueryRow(ctx, invoiceQuery,
		invoice.SupplierID,
		invoice.StoreID,
		creatorID,
		invoice.TotalItems,
		invoice.Note,
		invoice.Status,
	).Scan(&invoice.ID, &invoice.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to insert import invoice: %w", err)
	}

	invoice.CreatedBy = creatorID

	// 2. Loop details to insert details (and optionally update inventory/logs)
	detailQuery := `
		INSERT INTO import_invoice_details (invoice_id, variant_id, quantity, stock_before, price_import)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	upsertInvQuery := `
		INSERT INTO product_inventory (variant_id, store_id, quantity, reserved, last_updated)
		VALUES ($1, $2, $3, 0, NOW())
		ON CONFLICT (variant_id, store_id)
		DO UPDATE SET quantity = EXCLUDED.quantity, last_updated = NOW()`

	logQuery := `
		INSERT INTO inventory_log (variant_id, store_id, change_qty, qty_after, reason, ref_id, created_by, created_at)
		VALUES ($1, $2, $3, $4, 'import', $5, $6, NOW())`

	refIDStr := strconv.Itoa(invoice.ID)

	for _, d := range details {
		var stockBefore int
		if invoice.Status == "published" {
			// Fetch stock before (FOR UPDATE to prevent concurrent adjustments)
			selectForUpdate := `SELECT quantity FROM product_inventory WHERE variant_id = $1 AND store_id = $2 FOR UPDATE`
			err = tx.QueryRow(ctx, selectForUpdate, d.VariantID, invoice.StoreID).Scan(&stockBefore)
			if err != nil {
				if err == pgx.ErrNoRows {
					stockBefore = 0
				} else {
					var pgErr *pgconn.PgError
					if errors.As(err, &pgErr) && pgErr.Code == "23503" {
						return nil, domain.ErrVariantNotFound
					}
					return nil, fmt.Errorf("failed to select current quantity FOR UPDATE: %w", err)
				}
			}
		} else {
			// For drafts, we can just check if variant exists first
			var exists bool
			err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM product_variant WHERE id = $1 AND is_deleted = false)`, d.VariantID).Scan(&exists)
			if err != nil {
				return nil, fmt.Errorf("failed to verify variant existence: %w", err)
			}
			if !exists {
				return nil, domain.ErrVariantNotFound
			}
			stockBefore = 0
		}

		qtyAfter := stockBefore + d.Quantity
		d.InvoiceID = invoice.ID
		d.StockBefore = stockBefore

		// Insert detail row
		err = tx.QueryRow(ctx, detailQuery,
			invoice.ID,
			d.VariantID,
			d.Quantity,
			d.StockBefore,
			d.PriceImport,
		).Scan(&d.ID)

		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				return nil, domain.ErrVariantNotFound
			}
			return nil, fmt.Errorf("failed to insert import invoice detail: %w", err)
		}

		// Only modify inventory & log if NOT draft
		if invoice.Status == "published" {
			// Upsert inventory quantity
			_, err = tx.Exec(ctx, upsertInvQuery, d.VariantID, invoice.StoreID, qtyAfter)
			if err != nil {
				return nil, fmt.Errorf("failed to upsert inventory: %w", err)
			}

			// Insert inventory log
			_, err = tx.Exec(ctx, logQuery,
				d.VariantID,
				invoice.StoreID,
				d.Quantity,
				qtyAfter,
				refIDStr,
				creatorID,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to insert inventory log: %w", err)
			}
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return invoice, nil
}

func (r *InventoryRepository) ListImportInvoices(ctx context.Context, storeID *int, page, limit int) ([]*domain.ImportInvoiceResponse, int, error) {
	countQuery := `SELECT COUNT(*) FROM import_invoices ii`
	selectQuery := `
		SELECT ii.id, ii.supplier_id, sup.name as supplier_name, ii.store_id, s.name as store_name,
		       ii.created_by, u.full_name as creator_name, ii.total_items, ii.note, ii.status, ii.created_at
		FROM import_invoices ii
		JOIN suppliers sup ON ii.supplier_id = sup.id
		JOIN store s ON ii.store_id = s.id
		JOIN users u ON ii.created_by = u.id`

	var conditions []string
	var args []interface{}
	placeholderIdx := 1

	if storeID != nil {
		conditions = append(conditions, fmt.Sprintf("ii.store_id = $%d", placeholderIdx))
		args = append(args, *storeID)
		placeholderIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	var totalCount int
	err := r.db.QueryRow(ctx, countQuery+whereClause, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count import invoices: %w", err)
	}

	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	orderByAndPagination := fmt.Sprintf(" ORDER BY ii.created_at DESC, ii.id DESC LIMIT $%d OFFSET $%d", placeholderIdx, placeholderIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, selectQuery+whereClause+orderByAndPagination, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query import invoices: %w", err)
	}
	defer rows.Close()

	invoices := make([]*domain.ImportInvoiceResponse, 0)
	for rows.Next() {
		ii := &domain.ImportInvoiceResponse{}
		err := rows.Scan(
			&ii.ID,
			&ii.SupplierID,
			&ii.SupplierName,
			&ii.StoreID,
			&ii.StoreName,
			&ii.CreatedBy,
			&ii.CreatorName,
			&ii.TotalItems,
			&ii.Note,
			&ii.Status,
			&ii.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan import invoice row: %w", err)
		}
		invoices = append(invoices, ii)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("import invoices list rows error: %w", err)
	}

	return invoices, totalCount, nil
}

func (r *InventoryRepository) GetImportInvoiceDetails(ctx context.Context, invoiceID int) (*domain.ImportInvoiceDetailsResponse, error) {
	// 1. Fetch invoice header
	ii := &domain.ImportInvoiceResponse{}
	invoiceQuery := `
		SELECT ii.id, ii.supplier_id, sup.name as supplier_name, ii.store_id, s.name as store_name,
		       ii.created_by, u.full_name as creator_name, ii.total_items, ii.note, ii.status, ii.created_at
		FROM import_invoices ii
		JOIN suppliers sup ON ii.supplier_id = sup.id
		JOIN store s ON ii.store_id = s.id
		JOIN users u ON ii.created_by = u.id
		WHERE ii.id = $1`

	err := r.db.QueryRow(ctx, invoiceQuery, invoiceID).Scan(
		&ii.ID,
		&ii.SupplierID,
		&ii.SupplierName,
		&ii.StoreID,
		&ii.StoreName,
		&ii.CreatedBy,
		&ii.CreatorName,
		&ii.TotalItems,
		&ii.Note,
		&ii.Status,
		&ii.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrImportInvoiceNotFound
		}
		return nil, fmt.Errorf("failed to query import invoice header: %w", err)
	}

	// 2. Fetch detail rows
	detailsQuery := `
		SELECT iid.id, iid.variant_id, pv.name as variant_name, pv.sku, iid.quantity, iid.stock_before, iid.price_import
		FROM import_invoice_details iid
		JOIN product_variant pv ON iid.variant_id = pv.id
		WHERE iid.invoice_id = $1
		ORDER BY iid.id ASC`

	rows, err := r.db.Query(ctx, detailsQuery, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query import details: %w", err)
	}
	defer rows.Close()

	details := make([]*domain.ImportInvoiceDetailResponse, 0)
	for rows.Next() {
		d := &domain.ImportInvoiceDetailResponse{}
		err := rows.Scan(
			&d.ID,
			&d.VariantID,
			&d.VariantName,
			&d.SKU,
			&d.Quantity,
			&d.StockBefore,
			&d.PriceImport,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan detail row: %w", err)
		}
		details = append(details, d)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("import details rows error: %w", err)
	}

	return &domain.ImportInvoiceDetailsResponse{
		Invoice: ii,
		Details: details,
	}, nil
}

func (r *InventoryRepository) AdjustInventory(ctx context.Context, storeID int, creatorID int, adjustments []*domain.AdjustItemDTO) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	upsertInvQuery := `
		INSERT INTO product_inventory (variant_id, store_id, quantity, reserved, last_updated)
		VALUES ($1, $2, $3, 0, NOW())
		ON CONFLICT (variant_id, store_id)
		DO UPDATE SET quantity = EXCLUDED.quantity, last_updated = NOW()`

	logQuery := `
		INSERT INTO inventory_log (variant_id, store_id, change_qty, qty_after, reason, created_by, created_at)
		VALUES ($1, $2, $3, $4, 'manual_adjust', $5, NOW())`

	for _, adj := range adjustments {
		// Fetch stock before (FOR UPDATE to prevent concurrent modifications)
		var stockBefore int
		selectForUpdate := `SELECT quantity FROM product_inventory WHERE variant_id = $1 AND store_id = $2 FOR UPDATE`
		err = tx.QueryRow(ctx, selectForUpdate, adj.VariantID, storeID).Scan(&stockBefore)
		if err != nil {
			if err == pgx.ErrNoRows {
				stockBefore = 0
			} else {
				return fmt.Errorf("failed to query inventory quantity FOR UPDATE: %w", err)
			}
		}

		changeQty := adj.NewQuantity - stockBefore

		// If nothing changed, we skip logging and upserting for this variant to be efficient
		if changeQty == 0 {
			continue
		}

		// Verify variant exists first by performing updates (Fails if foreign key violation)
		// Or run a check. To prevent foreign key errors with useful domain messages, we can check variant existence.
		var exists bool
		err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM product_variant WHERE id = $1 AND is_deleted = false)`, adj.VariantID).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to verify variant existence: %w", err)
		}
		if !exists {
			return domain.ErrVariantNotFound
		}

		// Upsert product inventory
		_, err = tx.Exec(ctx, upsertInvQuery, adj.VariantID, storeID, adj.NewQuantity)
		if err != nil {
			return fmt.Errorf("failed to adjust product inventory: %w", err)
		}

		// Insert log
		_, err = tx.Exec(ctx, logQuery,
			adj.VariantID,
			storeID,
			changeQty,
			adj.NewQuantity,
			creatorID,
		)
		if err != nil {
			return fmt.Errorf("failed to log inventory adjustment: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *InventoryRepository) ListStoreInventory(ctx context.Context, storeID int) ([]*domain.ProductInventory, error) {
	query := `
		SELECT pi.variant_id, pi.store_id, pi.quantity, pi.reserved, pi.last_updated,
		       COALESCE(pv.name, '') as variant_name, COALESCE(p.name, '') as product_name, COALESCE(pv.sku, '') as sku
		FROM product_inventory pi
		LEFT JOIN product_variant pv ON pi.variant_id = pv.id
		LEFT JOIN product p ON pv.product_id = p.id
		WHERE pi.store_id = $1
		ORDER BY pi.variant_id ASC`

	rows, err := r.db.Query(ctx, query, storeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query store inventory: %w", err)
	}
	defer rows.Close()

	inventoryList := make([]*domain.ProductInventory, 0)
	for rows.Next() {
		pi := &domain.ProductInventory{}
		err := rows.Scan(
			&pi.VariantID,
			&pi.StoreID,
			&pi.Quantity,
			&pi.Reserved,
			&pi.LastUpdated,
			&pi.VariantName,
			&pi.ProductName,
			&pi.SKU,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product inventory row: %w", err)
		}
		inventoryList = append(inventoryList, pi)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("product inventory list rows error: %w", err)
	}

	return inventoryList, nil
}

func (r *InventoryRepository) GetLowStockAlerts(ctx context.Context, storeID *int) ([]*domain.LowStockAlertResponse, error) {
	query := `
		SELECT pi.variant_id, pv.sku, pv.name as variant_name, p.id as product_id, p.name as product_name,
		       pi.store_id, s.name as store_name, pi.quantity, p.low_stock_threshold
		FROM product_inventory pi
		JOIN product_variant pv ON pi.variant_id = pv.id
		JOIN product p ON pv.product_id = p.id
		JOIN store s ON pi.store_id = s.id
		WHERE pi.quantity <= p.low_stock_threshold
		  AND p.is_deleted = false AND pv.is_deleted = false AND s.is_active = true`

	var args []interface{}
	placeholderIdx := 1

	if storeID != nil {
		query += fmt.Sprintf(" AND pi.store_id = $%d", placeholderIdx)
		args = append(args, *storeID)
	}

	query += " ORDER BY pi.quantity ASC, s.name ASC, pv.sku ASC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query low stock alerts: %w", err)
	}
	defer rows.Close()

	alerts := make([]*domain.LowStockAlertResponse, 0)
	for rows.Next() {
		a := &domain.LowStockAlertResponse{}
		err := rows.Scan(
			&a.VariantID,
			&a.SKU,
			&a.VariantName,
			&a.ProductID,
			&a.ProductName,
			&a.StoreID,
			&a.StoreName,
			&a.Quantity,
			&a.LowStockThreshold,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan low stock alert row: %w", err)
		}
		alerts = append(alerts, a)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("low stock alerts rows error: %w", err)
	}

	return alerts, nil
}

func (r *InventoryRepository) GetInventoryLogs(ctx context.Context, q *domain.InventoryLogsQuery) (*domain.InventoryLogsResult, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM inventory_log il
		JOIN product_variant pv ON il.variant_id = pv.id
		JOIN store s ON il.store_id = s.id
		LEFT JOIN users u ON il.created_by = u.id`

	selectQuery := `
		SELECT il.id, il.variant_id, pv.sku, pv.name as variant_name,
		       il.store_id, s.name as store_name, il.change_qty, il.qty_after,
		       il.reason, il.ref_id, il.created_by, COALESCE(u.full_name, 'Hệ thống') as creator_name, il.created_at
		FROM inventory_log il
		JOIN product_variant pv ON il.variant_id = pv.id
		JOIN store s ON il.store_id = s.id
		LEFT JOIN users u ON il.created_by = u.id`

	var conditions []string
	var args []interface{}
	placeholderIdx := 1

	if q.StoreID != nil {
		conditions = append(conditions, fmt.Sprintf("il.store_id = $%d", placeholderIdx))
		args = append(args, *q.StoreID)
		placeholderIdx++
	}

	if q.VariantID != nil {
		conditions = append(conditions, fmt.Sprintf("il.variant_id = $%d", placeholderIdx))
		args = append(args, *q.VariantID)
		placeholderIdx++
	}

	if q.Reason != "" {
		conditions = append(conditions, fmt.Sprintf("il.reason = $%d", placeholderIdx))
		args = append(args, q.Reason)
		placeholderIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	var totalCount int
	err := r.db.QueryRow(ctx, countQuery+whereClause, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count inventory logs: %w", err)
	}

	offset := (q.Page - 1) * q.Limit
	if offset < 0 {
		offset = 0
	}

	orderByPagination := fmt.Sprintf(" ORDER BY il.created_at DESC, il.id DESC LIMIT $%d OFFSET $%d", placeholderIdx, placeholderIdx+1)
	args = append(args, q.Limit, offset)

	rows, err := r.db.Query(ctx, selectQuery+whereClause+orderByPagination, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query inventory logs: %w", err)
	}
	defer rows.Close()

	logs := make([]*domain.InventoryLogResponse, 0)
	for rows.Next() {
		l := &domain.InventoryLogResponse{}
		err := rows.Scan(
			&l.ID,
			&l.VariantID,
			&l.SKU,
			&l.VariantName,
			&l.StoreID,
			&l.StoreName,
			&l.ChangeQty,
			&l.QtyAfter,
			&l.Reason,
			&l.RefID,
			&l.CreatedBy,
			&l.CreatorName,
			&l.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan inventory log row: %w", err)
		}
		logs = append(logs, l)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("inventory logs rows error: %w", err)
	}

	return &domain.InventoryLogsResult{
		Logs:       logs,
		TotalCount: totalCount,
		Page:       q.Page,
		Limit:      q.Limit,
	}, nil
}

func (r *InventoryRepository) ConfirmImportInvoice(ctx context.Context, invoiceID int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Get invoice and check status
	var status string
	var storeID, creatorID int
	err = tx.QueryRow(ctx, "SELECT store_id, created_by, status FROM import_invoices WHERE id = $1 FOR UPDATE", invoiceID).Scan(&storeID, &creatorID, &status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.ErrImportInvoiceNotFound
		}
		return fmt.Errorf("failed to query invoice: %w", err)
	}

	if status != "draft" {
		return fmt.Errorf("invoice is already confirmed or processed")
	}

	// 2. Fetch all details for this invoice
	rows, err := tx.Query(ctx, "SELECT variant_id, quantity, price_import FROM import_invoice_details WHERE invoice_id = $1", invoiceID)
	if err != nil {
		return fmt.Errorf("failed to query invoice details: %w", err)
	}
	defer rows.Close()

	type itemDetail struct {
		VariantID   int
		Quantity    int
		PriceImport float64
	}
	var items []itemDetail

	for rows.Next() {
		var it itemDetail
		if err := rows.Scan(&it.VariantID, &it.Quantity, &it.PriceImport); err != nil {
			return fmt.Errorf("failed to scan invoice detail: %w", err)
		}
		items = append(items, it)
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error reading invoice details: %w", err)
	}

	// 3. Loop details to upsert quantity, insert logs and update stock_before
	upsertInvQuery := `
		INSERT INTO product_inventory (variant_id, store_id, quantity, reserved, last_updated)
		VALUES ($1, $2, $3, 0, NOW())
		ON CONFLICT (variant_id, store_id)
		DO UPDATE SET quantity = EXCLUDED.quantity, last_updated = NOW()`

	logQuery := `
		INSERT INTO inventory_log (variant_id, store_id, change_qty, qty_after, reason, ref_id, created_by, created_at)
		VALUES ($1, $2, $3, $4, 'import', $5, $6, NOW())`

	refIDStr := strconv.Itoa(invoiceID)

	for _, item := range items {
		// Fetch stock before (FOR UPDATE to prevent concurrent adjustments)
		var stockBefore int
		selectForUpdate := `SELECT quantity FROM product_inventory WHERE variant_id = $1 AND store_id = $2 FOR UPDATE`
		err = tx.QueryRow(ctx, selectForUpdate, item.VariantID, storeID).Scan(&stockBefore)
		if err != nil {
			if err == pgx.ErrNoRows {
				stockBefore = 0
			} else {
				return fmt.Errorf("failed to select current quantity FOR UPDATE: %w", err)
			}
		}

		qtyAfter := stockBefore + item.Quantity

		// Update stock_before in details table
		_, err = tx.Exec(ctx, "UPDATE import_invoice_details SET stock_before = $1 WHERE invoice_id = $2 AND variant_id = $3", stockBefore, invoiceID, item.VariantID)
		if err != nil {
			return fmt.Errorf("failed to update detail stock_before: %w", err)
		}

		// Upsert product inventory
		_, err = tx.Exec(ctx, upsertInvQuery, item.VariantID, storeID, qtyAfter)
		if err != nil {
			return fmt.Errorf("failed to adjust product inventory: %w", err)
		}

		// Insert inventory log
		_, err = tx.Exec(ctx, logQuery,
			item.VariantID,
			storeID,
			item.Quantity,
			qtyAfter,
			refIDStr,
			creatorID,
		)
		if err != nil {
			return fmt.Errorf("failed to insert inventory log: %w", err)
		}

		// Update latest_cost_price in product_variant table
		_, err = tx.Exec(ctx, "UPDATE product_variant SET latest_cost_price = $1 WHERE id = $2", item.PriceImport, item.VariantID)
		if err != nil {
			return fmt.Errorf("failed to update variant latest_cost_price: %w", err)
		}
	}

	// 4. Update status to published
	_, err = tx.Exec(ctx, "UPDATE import_invoices SET status = 'published' WHERE id = $1", invoiceID)
	if err != nil {
		return fmt.Errorf("failed to update invoice status: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *InventoryRepository) GetLastImportPrice(ctx context.Context, variantID int) (float64, error) {
	query := `SELECT latest_cost_price FROM product_variant WHERE id = $1`

	var price float64
	err := r.db.QueryRow(ctx, query, variantID).Scan(&price)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get last import price: %w", err)
	}

	return price, nil
}
