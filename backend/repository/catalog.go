package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"backend/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CatalogRepository struct {
	db *pgxpool.Pool
}

func NewCatalogRepository(db *pgxpool.Pool) *CatalogRepository {
	return &CatalogRepository{db: db}
}

// --- Category ---

func (r *CatalogRepository) CreateCategory(ctx context.Context, cat *domain.Category) (*domain.Category, error) {
	query := `
		INSERT INTO category (name, parent_id, icon_img_url, slug, sort_order, is_deleted)
		VALUES ($1, $2, $3, $4, $5, false)
		RETURNING id`

	err := r.db.QueryRow(ctx, query,
		cat.Name,
		cat.ParentID,
		cat.Icon,
		cat.Slug,
		cat.SortOrder,
	).Scan(&cat.ID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrDuplicateSlug
		}
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return cat, nil
}

func (r *CatalogRepository) ListCategories(ctx context.Context) ([]*domain.Category, error) {
	query := `
		SELECT id, name, parent_id, icon_img_url, slug, sort_order, is_deleted
		FROM category
		WHERE is_deleted = false
		ORDER BY sort_order ASC, name ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	categories := make([]*domain.Category, 0)
	for rows.Next() {
		cat := &domain.Category{}
		err := rows.Scan(
			&cat.ID,
			&cat.Name,
			&cat.ParentID,
			&cat.Icon,
			&cat.Slug,
			&cat.SortOrder,
			&cat.IsDeleted,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, cat)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("categories rows error: %w", err)
	}

	return categories, nil
}

func (r *CatalogRepository) GetCategoryByID(ctx context.Context, id int) (*domain.Category, error) {
	cat := &domain.Category{}
	query := `
		SELECT id, name, parent_id, icon_img_url, slug, sort_order, is_deleted
		FROM category
		WHERE id = $1 AND is_deleted = false`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&cat.ID,
		&cat.Name,
		&cat.ParentID,
		&cat.Icon,
		&cat.Slug,
		&cat.SortOrder,
		&cat.IsDeleted,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to get category by ID: %w", err)
	}

	return cat, nil
}

func (r *CatalogRepository) UpdateCategory(ctx context.Context, cat *domain.Category) (*domain.Category, error) {
	query := `
		UPDATE category
		SET name = $1, parent_id = $2, icon_img_url = $3, slug = $4, sort_order = $5
		WHERE id = $6 AND is_deleted = false`

	tag, err := r.db.Exec(ctx, query,
		cat.Name,
		cat.ParentID,
		cat.Icon,
		cat.Slug,
		cat.SortOrder,
		cat.ID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrDuplicateSlug
		}
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return nil, domain.ErrCategoryNotFound
	}

	return cat, nil
}

func (r *CatalogRepository) DeleteCategory(ctx context.Context, id int) error {
	// Soft delete category
	query := `UPDATE category SET is_deleted = true WHERE id = $1 AND is_deleted = false`
	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete category: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrCategoryNotFound
	}

	return nil
}

// --- Brand ---

func (r *CatalogRepository) CreateBrand(ctx context.Context, brand *domain.Brand) (*domain.Brand, error) {
	query := `
		INSERT INTO brand (name, slug, logo_url, is_active, is_deleted)
		VALUES ($1, $2, $3, $4, false)
		RETURNING id`

	err := r.db.QueryRow(ctx, query,
		brand.Name,
		brand.Slug,
		brand.LogoURL,
		brand.IsActive,
	).Scan(&brand.ID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrDuplicateSlug
		}
		return nil, fmt.Errorf("failed to create brand: %w", err)
	}

	return brand, nil
}

func (r *CatalogRepository) ListBrands(ctx context.Context) ([]*domain.Brand, error) {
	query := `
		SELECT id, name, slug, logo_url, is_active, is_deleted
		FROM brand
		WHERE is_deleted = false AND is_active = true
		ORDER BY name ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query brands: %w", err)
	}
	defer rows.Close()

	brands := make([]*domain.Brand, 0)
	for rows.Next() {
		brand := &domain.Brand{}
		err := rows.Scan(
			&brand.ID,
			&brand.Name,
			&brand.Slug,
			&brand.LogoURL,
			&brand.IsActive,
			&brand.IsDeleted,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan brand: %w", err)
		}
		brands = append(brands, brand)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("brands rows error: %w", err)
	}

	return brands, nil
}

func (r *CatalogRepository) GetBrandByID(ctx context.Context, id int) (*domain.Brand, error) {
	brand := &domain.Brand{}
	query := `
		SELECT id, name, slug, logo_url, is_active, is_deleted
		FROM brand
		WHERE id = $1 AND is_deleted = false`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&brand.ID,
		&brand.Name,
		&brand.Slug,
		&brand.LogoURL,
		&brand.IsActive,
		&brand.IsDeleted,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrBrandNotFound
		}
		return nil, fmt.Errorf("failed to get brand by ID: %w", err)
	}

	return brand, nil
}

func (r *CatalogRepository) UpdateBrand(ctx context.Context, brand *domain.Brand) (*domain.Brand, error) {
	query := `
		UPDATE brand
		SET name = $1, slug = $2, logo_url = $3, is_active = $4
		WHERE id = $5 AND is_deleted = false`

	tag, err := r.db.Exec(ctx, query,
		brand.Name,
		brand.Slug,
		brand.LogoURL,
		brand.IsActive,
		brand.ID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrDuplicateSlug
		}
		return nil, fmt.Errorf("failed to update brand: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return nil, domain.ErrBrandNotFound
	}

	return brand, nil
}

func (r *CatalogRepository) DeleteBrand(ctx context.Context, id int) error {
	// Soft delete brand
	query := `UPDATE brand SET is_deleted = true WHERE id = $1 AND is_deleted = false`
	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete brand: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrBrandNotFound
	}

	return nil
}

// --- Product & Specs & Options & Variants ---

func (r *CatalogRepository) CreateProduct(ctx context.Context, input *domain.CreateProductInput) (*domain.Product, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Insert product
	prodQuery := `
		INSERT INTO product (id, category_id, brand_id, name, slug, meta_title, meta_description, img_thumb, weight, low_stock_threshold, specs_jsonb, is_active, is_deleted, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, true, false, NOW(), NOW())`

	_, err = tx.Exec(ctx, prodQuery,
		input.Product.ID,
		input.Product.CategoryID,
		input.Product.BrandID,
		input.Product.Name,
		input.Product.Slug,
		input.Product.MetaTitle,
		input.Product.MetaDescription,
		input.Product.ImgThumb,
		input.Product.Weight,
		input.Product.LowStockThreshold,
		input.Product.SpecsJSONB,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrDuplicateSlug
		}
		return nil, fmt.Errorf("failed to insert product: %w", err)
	}

	// 2. Insert specs
	specQuery := `
		INSERT INTO product_spec (product_id, "group", key, value, unit, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6)`

	for _, spec := range input.Specs {
		_, err = tx.Exec(ctx, specQuery,
			input.Product.ID,
			spec.Group,
			spec.Key,
			spec.Value,
			spec.Unit,
			spec.SortOrder,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert product spec: %w", err)
		}
	}

	// 3. Insert Option Types & Option Values, build mapping
	// Map to store: OptionType.Name + ":" + OptionValue.Value -> OptionValue.ID
	valueMap := make(map[string]int)

	optTypeQuery := `
		INSERT INTO product_option_type (product_id, name, sort_order)
		VALUES ($1, $2, $3)
		RETURNING id`

	optValQuery := `
		INSERT INTO product_option_value (option_type_id, value, color_code, sort_order)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	for _, optType := range input.OptionTypes {
		var optTypeID int
		err = tx.QueryRow(ctx, optTypeQuery, input.Product.ID, optType.Name, optType.SortOrder).Scan(&optTypeID)
		if err != nil {
			return nil, fmt.Errorf("failed to insert option type: %w", err)
		}

		for _, optVal := range optType.Values {
			var optValID int
			err = tx.QueryRow(ctx, optValQuery, optTypeID, optVal.Value, optVal.ColorCode, optVal.SortOrder).Scan(&optValID)
			if err != nil {
				return nil, fmt.Errorf("failed to insert option value: %w", err)
			}
			// Map value
			key := fmt.Sprintf("%s:%s", optType.Name, optVal.Value)
			valueMap[key] = optValID
		}
	}

	// 4. Insert Variants and link Variant Options
	variantQuery := `
		INSERT INTO product_variant (product_id, name, sku, sell_price, compare_price, weight, is_active, is_deleted)
		VALUES ($1, $2, $3, $4, $5, $6, true, false)
		RETURNING id`

	varOptionQuery := `
		INSERT INTO product_variant_option (variant_id, option_value_id)
		VALUES ($1, $2)`

	for _, variant := range input.Variants {
		var variantID int
		err = tx.QueryRow(ctx, variantQuery,
			input.Product.ID,
			variant.Name,
			variant.SKU,
			variant.SellPrice,
			variant.ComparePrice,
			variant.Weight,
		).Scan(&variantID)
		if err != nil {
			return nil, fmt.Errorf("failed to insert variant: %w", err)
		}

		for _, optSel := range variant.Options {
			// Find Value ID from option name and option value
			key := fmt.Sprintf("%s:%s", optSel.OptionTypeName, optSel.Value)
			valID, exists := valueMap[key]
			if !exists {
				return nil, fmt.Errorf("option value %s for type %s not found in mapping", optSel.Value, optSel.OptionTypeName)
			}

			_, err = tx.Exec(ctx, varOptionQuery, variantID, valID)
			if err != nil {
				return nil, fmt.Errorf("failed to link variant option: %w", err)
			}
		}
	}

	// 5. Insert secondary product images
	imgQuery := `
		INSERT INTO product_image (product_id, url, is_thumbnail, sort_order)
		VALUES ($1, $2, false, $3)`

	for _, img := range input.Images {
		_, err = tx.Exec(ctx, imgQuery, input.Product.ID, img.URL, img.SortOrder)
		if err != nil {
			return nil, fmt.Errorf("failed to insert product secondary image: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return input.Product, nil
}

func (r *CatalogRepository) GetProductByID(ctx context.Context, id string) (*domain.Product, error) {
	prod := &domain.Product{}
	query := `
		SELECT id, category_id, brand_id, name, slug, meta_title, meta_description, img_thumb, weight, low_stock_threshold, specs_jsonb, is_active, is_deleted, created_at, updated_at
		FROM product
		WHERE id = $1 AND is_deleted = false`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&prod.ID,
		&prod.CategoryID,
		&prod.BrandID,
		&prod.Name,
		&prod.Slug,
		&prod.MetaTitle,
		&prod.MetaDescription,
		&prod.ImgThumb,
		&prod.Weight,
		&prod.LowStockThreshold,
		&prod.SpecsJSONB,
		&prod.IsActive,
		&prod.IsDeleted,
		&prod.CreatedAt,
		&prod.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to get product by ID: %w", err)
	}

	return prod, nil
}

func (r *CatalogRepository) SearchProducts(ctx context.Context, query *domain.ProductSearchQuery) (*domain.ProductSearchResult, error) {
	countQueryStr := `SELECT COUNT(*) FROM product p WHERE p.is_deleted = false AND p.is_active = true`
	selectQueryStr := `
		SELECT p.id, p.category_id, p.brand_id, p.name, p.slug, p.meta_title, p.meta_description, p.img_thumb, p.weight, p.low_stock_threshold, p.specs_jsonb, p.is_active, p.is_deleted, p.created_at, p.updated_at,
		       COALESCE(min_var.compare_price, min_var.sell_price, 0) as price,
		       CASE WHEN min_var.compare_price > min_var.sell_price THEN min_var.sell_price ELSE NULL END as discount_price,
		       CASE WHEN min_var.compare_price > min_var.sell_price THEN ROUND((min_var.compare_price - min_var.sell_price) / min_var.compare_price * 100) ELSE 0 END as discount_percent,
		       COALESCE(inv.total_qty, 0) as stock,
		       COALESCE(avg_rev.rating, 0.0) as rating,
		       COALESCE(avg_rev.review_count, 0) as review_count
		FROM product p
		LEFT JOIN (
			SELECT product_id, MIN(sell_price) as sell_price, MIN(compare_price) as compare_price
			FROM product_variant
			WHERE is_deleted = false AND is_active = true
			GROUP BY product_id
		) min_var ON p.id = min_var.product_id
		LEFT JOIN (
			SELECT pv.product_id, SUM(pi.quantity - pi.reserved) as total_qty
			FROM product_variant pv
			JOIN product_inventory pi ON pv.id = pi.variant_id
			WHERE pv.is_deleted = false AND pv.is_active = true
			GROUP BY pv.product_id
		) inv ON p.id = inv.product_id
		LEFT JOIN (
			SELECT product_id, COALESCE(AVG(rating), 0.0) as rating, COUNT(*) as review_count
			FROM reviews
			GROUP BY product_id
		) avg_rev ON p.id = avg_rev.product_id
		WHERE p.is_deleted = false AND p.is_active = true`

	var conditions []string
	var args []interface{}
	placeholderIdx := 1

	if query.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", placeholderIdx))
		args = append(args, *query.CategoryID)
		placeholderIdx++
	}

	if query.BrandID != nil {
		conditions = append(conditions, fmt.Sprintf("brand_id = $%d", placeholderIdx))
		args = append(args, *query.BrandID)
		placeholderIdx++
	}

	if query.Query != "" {
		conditions = append(conditions, fmt.Sprintf("(p.name ILIKE $%d OR p.slug ILIKE $%d OR p.id ILIKE $%d OR EXISTS (SELECT 1 FROM product_variant pv WHERE pv.product_id = p.id AND pv.sku ILIKE $%d AND pv.is_deleted = false))", placeholderIdx, placeholderIdx, placeholderIdx, placeholderIdx))
		args = append(args, "%"+query.Query+"%")
		placeholderIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " AND " + strings.Join(conditions, " AND ")
	}

	// 1. Get total count
	var totalCount int
	err := r.db.QueryRow(ctx, countQueryStr+whereClause, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count products: %w", err)
	}

	// 2. Add Sort
	sortCol := "created_at"
	sortOrder := "DESC"
	if query.Sort != "" {
		switch query.Sort {
		case "name_asc":
			sortCol = "name"
			sortOrder = "ASC"
		case "name_desc":
			sortCol = "name"
			sortOrder = "DESC"
		case "oldest":
			sortCol = "created_at"
			sortOrder = "ASC"
		default:
			sortCol = "created_at"
			sortOrder = "DESC"
		}
	}
	orderByClause := fmt.Sprintf(" ORDER BY %s %s", sortCol, sortOrder)

	// 3. Add Pagination
	limit := query.Limit
	if limit <= 0 {
		limit = 10
	}
	offset := (query.Page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	limitPlaceholder := placeholderIdx
	placeholderIdx++
	args = append(args, limit)

	offsetPlaceholder := placeholderIdx
	args = append(args, offset)

	paginationClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", limitPlaceholder, offsetPlaceholder)

	// 4. Query products
	rows, err := r.db.Query(ctx, selectQueryStr+whereClause+orderByClause+paginationClause, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query products list: %w", err)
	}
	defer rows.Close()

	products := make([]*domain.Product, 0)
	for rows.Next() {
		prod := &domain.Product{}
		err := rows.Scan(
			&prod.ID,
			&prod.CategoryID,
			&prod.BrandID,
			&prod.Name,
			&prod.Slug,
			&prod.MetaTitle,
			&prod.MetaDescription,
			&prod.ImgThumb,
			&prod.Weight,
			&prod.LowStockThreshold,
			&prod.SpecsJSONB,
			&prod.IsActive,
			&prod.IsDeleted,
			&prod.CreatedAt,
			&prod.UpdatedAt,
			&prod.Price,
			&prod.DiscountPrice,
			&prod.DiscountPercent,
			&prod.Stock,
			&prod.Rating,
			&prod.ReviewCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product row: %w", err)
		}
		products = append(products, prod)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("products list rows error: %w", err)
	}

	return &domain.ProductSearchResult{
		Products:   products,
		TotalCount: totalCount,
		Page:       query.Page,
		Limit:      limit,
	}, nil
}

func (r *CatalogRepository) GetProductDetails(ctx context.Context, id string) (*domain.ProductDetailsResponse, error) {
	// 1. Get Product along with Category & Brand Names
	prod := &domain.Product{}
	var brandName, categoryName string

	prodQuery := `
		SELECT p.id, p.category_id, p.brand_id, p.name, p.slug, p.meta_title, p.meta_description, p.img_thumb, p.weight, p.low_stock_threshold, p.specs_jsonb, p.is_active, p.is_deleted, p.created_at, p.updated_at,
		       b.name as brand_name, c.name as category_name,
		       COALESCE(min_var.compare_price, min_var.sell_price, 0) as price,
		       CASE WHEN min_var.compare_price > min_var.sell_price THEN min_var.sell_price ELSE NULL END as discount_price,
		       CASE WHEN min_var.compare_price > min_var.sell_price THEN ROUND((min_var.compare_price - min_var.sell_price) / min_var.compare_price * 100) ELSE 0 END as discount_percent,
		       COALESCE(inv.total_qty, 0) as stock,
		       COALESCE(avg_rev.rating, 0.0) as rating,
		       COALESCE(avg_rev.review_count, 0) as review_count
		FROM product p
		JOIN brand b ON p.brand_id = b.id
		JOIN category c ON p.category_id = c.id
		LEFT JOIN (
			SELECT product_id, MIN(sell_price) as sell_price, MIN(compare_price) as compare_price
			FROM product_variant
			WHERE is_deleted = false AND is_active = true
			GROUP BY product_id
		) min_var ON p.id = min_var.product_id
		LEFT JOIN (
			SELECT pv.product_id, SUM(pi.quantity - pi.reserved) as total_qty
			FROM product_variant pv
			JOIN product_inventory pi ON pv.id = pi.variant_id
			WHERE pv.is_deleted = false AND pv.is_active = true
			GROUP BY pv.product_id
		) inv ON p.id = inv.product_id
		LEFT JOIN (
			SELECT product_id, COALESCE(AVG(rating), 0.0) as rating, COUNT(*) as review_count
			FROM reviews
			GROUP BY product_id
		) avg_rev ON p.id = avg_rev.product_id
		WHERE (p.id = $1 OR p.slug = $1) AND p.is_deleted = false`

	err := r.db.QueryRow(ctx, prodQuery, id).Scan(
		&prod.ID,
		&prod.CategoryID,
		&prod.BrandID,
		&prod.Name,
		&prod.Slug,
		&prod.MetaTitle,
		&prod.MetaDescription,
		&prod.ImgThumb,
		&prod.Weight,
		&prod.LowStockThreshold,
		&prod.SpecsJSONB,
		&prod.IsActive,
		&prod.IsDeleted,
		&prod.CreatedAt,
		&prod.UpdatedAt,
		&brandName,
		&categoryName,
		&prod.Price,
		&prod.DiscountPrice,
		&prod.DiscountPercent,
		&prod.Stock,
		&prod.Rating,
		&prod.ReviewCount,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to query product: %w", err)
	}

	// 2. Get Specs
	specsQuery := `
		SELECT id, product_id, "group", key, value, unit, sort_order
		FROM product_spec
		WHERE product_id = $1
		ORDER BY sort_order ASC, id ASC`

	specsRows, err := r.db.Query(ctx, specsQuery, prod.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query specs: %w", err)
	}
	defer specsRows.Close()

	specs := make([]*domain.ProductSpec, 0)
	for specsRows.Next() {
		spec := &domain.ProductSpec{}
		err := specsRows.Scan(&spec.ID, &spec.ProductID, &spec.Group, &spec.Key, &spec.Value, &spec.Unit, &spec.SortOrder)
		if err != nil {
			return nil, fmt.Errorf("failed to scan spec: %w", err)
		}
		specs = append(specs, spec)
	}

	// 3. Get Option Types
	optTypesQuery := `
		SELECT id, product_id, name, sort_order
		FROM product_option_type
		WHERE product_id = $1
		ORDER BY sort_order ASC, id ASC`

	optTypesRows, err := r.db.Query(ctx, optTypesQuery, prod.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query option types: %w", err)
	}
	defer optTypesRows.Close()

	optionTypes := make([]*domain.ProductOptionType, 0)
	optionTypesMap := make(map[int]*domain.ProductOptionType)

	for optTypesRows.Next() {
		ot := &domain.ProductOptionType{}
		err := optTypesRows.Scan(&ot.ID, &ot.ProductID, &ot.Name, &ot.SortOrder)
		if err != nil {
			return nil, fmt.Errorf("failed to scan option type: %w", err)
		}
		ot.Values = make([]domain.ProductOptionValue, 0)
		optionTypes = append(optionTypes, ot)
		optionTypesMap[ot.ID] = ot
	}

	// 4. Get Option Values
	optValsQuery := `
		SELECT pov.id, pov.option_type_id, pov.value, pov.color_code, pov.sort_order
		FROM product_option_value pov
		JOIN product_option_type pot ON pov.option_type_id = pot.id
		WHERE pot.product_id = $1
		ORDER BY pov.sort_order ASC, pov.id ASC`

	optValsRows, err := r.db.Query(ctx, optValsQuery, prod.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query option values: %w", err)
	}
	defer optValsRows.Close()

	for optValsRows.Next() {
		val := domain.ProductOptionValue{}
		err := optValsRows.Scan(&val.ID, &val.OptionTypeID, &val.Value, &val.ColorCode, &val.SortOrder)
		if err != nil {
			return nil, fmt.Errorf("failed to scan option value: %w", err)
		}
		if ot, exists := optionTypesMap[val.OptionTypeID]; exists {
			ot.Values = append(ot.Values, val)
		}
	}

	// 5. Get Variants
	variantsQuery := `
		SELECT id, product_id, name, sku, sell_price, compare_price, latest_cost_price, weight, is_active, is_deleted
		FROM product_variant
		WHERE product_id = $1 AND is_deleted = false
		ORDER BY id ASC`

	variantsRows, err := r.db.Query(ctx, variantsQuery, prod.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query variants: %w", err)
	}
	defer variantsRows.Close()

	variants := make([]*domain.ProductVariant, 0)
	variantsMap := make(map[int]*domain.ProductVariant)

	for variantsRows.Next() {
		v := &domain.ProductVariant{}
		err := variantsRows.Scan(&v.ID, &v.ProductID, &v.Name, &v.SKU, &v.SellPrice, &v.ComparePrice, &v.LatestCostPrice, &v.Weight, &v.IsActive, &v.IsDeleted)
		if err != nil {
			return nil, fmt.Errorf("failed to scan variant: %w", err)
		}
		v.Options = make([]domain.VariantOption, 0)
		variants = append(variants, v)
		variantsMap[v.ID] = v
	}

	// 6. Get Variant Options Map
	varOptsQuery := `
		SELECT pvo.variant_id, pvo.option_value_id, pov.value, pov.color_code, pot.id as option_type_id, pot.name as option_type_name
		FROM product_variant_option pvo
		JOIN product_option_value pov ON pvo.option_value_id = pov.id
		JOIN product_option_type pot ON pov.option_type_id = pot.id
		JOIN product_variant pv ON pvo.variant_id = pv.id
		WHERE pv.product_id = $1 AND pv.is_deleted = false`

	varOptsRows, err := r.db.Query(ctx, varOptsQuery, prod.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query variant option mappings: %w", err)
	}
	defer varOptsRows.Close()

	for varOptsRows.Next() {
		var variantID, optValID, optTypeID int
		var optVal, optTypeName string
		var colorCode *string

		err := varOptsRows.Scan(&variantID, &optValID, &optVal, &colorCode, &optTypeID, &optTypeName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan variant option map row: %w", err)
		}

		if v, exists := variantsMap[variantID]; exists {
			v.Options = append(v.Options, domain.VariantOption{
				OptionTypeID:   optTypeID,
				OptionTypeName: optTypeName,
				ValueID:        optValID,
				Value:          optVal,
				ColorCode:      colorCode,
			})
		}
	}

	// 7. Get Images
	imagesQuery := `
		SELECT id, product_id, variant_id, url, alt_text, sort_order, is_thumbnail
		FROM product_image
		WHERE product_id = $1
		ORDER BY sort_order ASC, id ASC`

	imagesRows, err := r.db.Query(ctx, imagesQuery, prod.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query images: %w", err)
	}
	defer imagesRows.Close()

	images := make([]*domain.ProductImage, 0)
	for imagesRows.Next() {
		img := &domain.ProductImage{}
		err := imagesRows.Scan(&img.ID, &img.ProductID, &img.VariantID, &img.URL, &img.AltText, &img.SortOrder, &img.IsThumbnail)
		if err != nil {
			return nil, fmt.Errorf("failed to scan image: %w", err)
		}
		images = append(images, img)
	}

	// 8. Get Reviews
	reviewsQuery := `
		SELECT r.id, r.user_id, u.full_name, r.product_id, r.order_id, r.rating, r.comment, r.images, r.status, r.created_at, r.updated_at
		FROM reviews r
		JOIN users u ON r.user_id = u.id
		WHERE r.product_id = $1 AND r.status = 'approved'
		ORDER BY r.created_at DESC`

	reviewsRows, err := r.db.Query(ctx, reviewsQuery, prod.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query reviews: %w", err)
	}
	defer reviewsRows.Close()

	reviews := make([]*domain.Review, 0)
	for reviewsRows.Next() {
		rev := &domain.Review{}
		err := reviewsRows.Scan(
			&rev.ID,
			&rev.UserID,
			&rev.UserFullName,
			&rev.ProductID,
			&rev.OrderID,
			&rev.Rating,
			&rev.Comment,
			&rev.Images,
			&rev.Status,
			&rev.CreatedAt,
			&rev.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan review: %w", err)
		}
		reviews = append(reviews, rev)
	}

	return &domain.ProductDetailsResponse{
		Product:      prod,
		BrandName:    brandName,
		CategoryName: categoryName,
		Specs:        specs,
		OptionTypes:  optionTypes,
		Variants:     variants,
		Images:       images,
		Reviews:      reviews,
	}, nil
}

func (r *CatalogRepository) UpdateProduct(ctx context.Context, prod *domain.Product, specs []*domain.ProductSpec, images []*domain.ProductImage) (*domain.Product, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Update product
	query := `
		UPDATE product
		SET category_id = $1, brand_id = $2, name = $3, slug = $4, meta_title = $5, meta_description = $6, img_thumb = $7, weight = $8, low_stock_threshold = $9, specs_jsonb = $10, updated_at = NOW()
		WHERE id = $11 AND is_deleted = false`

	tag, err := tx.Exec(ctx, query,
		prod.CategoryID,
		prod.BrandID,
		prod.Name,
		prod.Slug,
		prod.MetaTitle,
		prod.MetaDescription,
		prod.ImgThumb,
		prod.Weight,
		prod.LowStockThreshold,
		prod.SpecsJSONB,
		prod.ID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrDuplicateSlug
		}
		return nil, fmt.Errorf("failed to update product row: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return nil, domain.ErrProductNotFound
	}

	// 2. Delete existing specs
	_, err = tx.Exec(ctx, `DELETE FROM product_spec WHERE product_id = $1`, prod.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to clear old product specs: %w", err)
	}

	// 3. Insert new specs
	specQuery := `
		INSERT INTO product_spec (product_id, "group", key, value, unit, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6)`

	for _, spec := range specs {
		_, err = tx.Exec(ctx, specQuery,
			prod.ID,
			spec.Group,
			spec.Key,
			spec.Value,
			spec.Unit,
			spec.SortOrder,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert product spec during update: %w", err)
		}
	}

	// 4. Delete old secondary product images
	_, err = tx.Exec(ctx, `DELETE FROM product_image WHERE product_id = $1 AND is_thumbnail = false AND variant_id IS NULL`, prod.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to clear old product secondary images: %w", err)
	}

	// 5. Insert new secondary product images
	imgQuery := `
		INSERT INTO product_image (product_id, url, is_thumbnail, sort_order)
		VALUES ($1, $2, false, $3)`

	for _, img := range images {
		_, err = tx.Exec(ctx, imgQuery, prod.ID, img.URL, img.SortOrder)
		if err != nil {
			return nil, fmt.Errorf("failed to insert product secondary image during update: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return prod, nil
}

func (r *CatalogRepository) DeleteProduct(ctx context.Context, id string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Soft delete product
	prodQuery := `UPDATE product SET is_deleted = true, updated_at = NOW() WHERE id = $1 AND is_deleted = false`
	tag, err := tx.Exec(ctx, prodQuery, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete product: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrProductNotFound
	}

	// 2. Soft delete variants of product
	_, err = tx.Exec(ctx, `UPDATE product_variant SET is_deleted = true WHERE product_id = $1 AND is_deleted = false`, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete variants: %w", err)
	}

	return tx.Commit(ctx)
}

// --- Option Types & Values ---

func (r *CatalogRepository) AddOptionValues(ctx context.Context, productID string, optionName string, values []*domain.ProductOptionValue) ([]*domain.ProductOptionValue, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Find option type
	var optTypeID int
	findQuery := `SELECT id FROM product_option_type WHERE product_id = $1 AND name = $2`
	err = tx.QueryRow(ctx, findQuery, productID, optionName).Scan(&optTypeID)

	if err != nil {
		if err == pgx.ErrNoRows {
			// Create new option type
			insertQuery := `
				INSERT INTO product_option_type (product_id, name, sort_order)
				VALUES ($1, $2, (SELECT COALESCE(MAX(sort_order), 0) + 1 FROM product_option_type WHERE product_id = $1))
				RETURNING id`
			err = tx.QueryRow(ctx, insertQuery, productID, optionName).Scan(&optTypeID)
			if err != nil {
				return nil, fmt.Errorf("failed to create product option type: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to query product option type: %w", err)
		}
	}

	// Insert option values
	insertValQuery := `
		INSERT INTO product_option_value (option_type_id, value, color_code, sort_order)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	for _, val := range values {
		val.OptionTypeID = optTypeID
		err = tx.QueryRow(ctx, insertValQuery, optTypeID, val.Value, val.ColorCode, val.SortOrder).Scan(&val.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to insert option value: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return values, nil
}

// --- Variant ---

func (r *CatalogRepository) CreateVariant(ctx context.Context, variant *domain.ProductVariant, optionValueIDs []int) (*domain.ProductVariant, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Insert Variant
	query := `
		INSERT INTO product_variant (product_id, name, sku, sell_price, compare_price, weight, is_active, is_deleted)
		VALUES ($1, $2, $3, $4, $5, $6, true, false)
		RETURNING id`

	err = tx.QueryRow(ctx, query,
		variant.ProductID,
		variant.Name,
		variant.SKU,
		variant.SellPrice,
		variant.ComparePrice,
		variant.Weight,
	).Scan(&variant.ID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, errors.New("sku already in use")
		}
		return nil, fmt.Errorf("failed to create variant: %w", err)
	}

	// 2. Link options
	linkQuery := `INSERT INTO product_variant_option (variant_id, option_value_id) VALUES ($1, $2)`
	for _, valID := range optionValueIDs {
		_, err = tx.Exec(ctx, linkQuery, variant.ID, valID)
		if err != nil {
			return nil, fmt.Errorf("failed to link variant to option value ID %d: %w", valID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return variant, nil
}

func (r *CatalogRepository) UpdateVariant(ctx context.Context, id int, name, sku string, sellPrice float64, comparePrice, weight *float64) (*domain.ProductVariant, error) {
	query := `
		UPDATE product_variant
		SET name = $1, sku = $2, sell_price = $3, compare_price = $4, weight = $5
		WHERE id = $6 AND is_deleted = false
		RETURNING id, product_id, name, sku, sell_price, compare_price, latest_cost_price, weight, is_active, is_deleted`

	v := &domain.ProductVariant{}
	err := r.db.QueryRow(ctx, query, name, sku, sellPrice, comparePrice, weight, id).Scan(
		&v.ID,
		&v.ProductID,
		&v.Name,
		&v.SKU,
		&v.SellPrice,
		&v.ComparePrice,
		&v.LatestCostPrice,
		&v.Weight,
		&v.IsActive,
		&v.IsDeleted,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrVariantNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, errors.New("sku already in use")
		}
		return nil, fmt.Errorf("failed to update variant: %w", err)
	}
	return v, nil
}

func (r *CatalogRepository) DeleteVariant(ctx context.Context, id int) error {
	query := `UPDATE product_variant SET is_deleted = true WHERE id = $1 AND is_deleted = false`
	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete variant: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrVariantNotFound
	}
	return nil
}

