package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"backend/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CatalogRepository struct {
	db *pgxpool.Pool
}

func NewCatalogRepository(db *pgxpool.Pool) *CatalogRepository {
	return &CatalogRepository{db: db}
}

// ===================== Category operations =====================

func (r *CatalogRepository) CreateCategory(ctx context.Context, c *domain.Category) (*domain.Category, error) {
	query := `
		INSERT INTO category (name, parent_id, icon, slug, sort_order, is_deleted)
		VALUES ($1, $2, $3, $4, $5, false)
		RETURNING id`
	err := r.db.QueryRow(ctx, query, c.Name, c.ParentID, c.Icon, c.Slug, c.SortOrder).Scan(&c.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}
	return c, nil
}

func (r *CatalogRepository) ListCategories(ctx context.Context) ([]*domain.Category, error) {
	query := `
		SELECT id, name, parent_id, icon, slug, sort_order, is_deleted
		FROM category
		WHERE is_deleted = false
		ORDER BY sort_order ASC, id ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	defer rows.Close()

	categories := make([]*domain.Category, 0)
	for rows.Next() {
		c := &domain.Category{}
		err := rows.Scan(&c.ID, &c.Name, &c.ParentID, &c.Icon, &c.Slug, &c.SortOrder, &c.IsDeleted)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, c)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("category rows iteration error: %w", err)
	}

	return categories, nil
}

func (r *CatalogRepository) GetCategoryByID(ctx context.Context, id int) (*domain.Category, error) {
	c := &domain.Category{}
	query := `
		SELECT id, name, parent_id, icon, slug, sort_order, is_deleted
		FROM category
		WHERE id = $1 AND is_deleted = false`

	err := r.db.QueryRow(ctx, query, id).Scan(&c.ID, &c.Name, &c.ParentID, &c.Icon, &c.Slug, &c.SortOrder, &c.IsDeleted)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("category not found")
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	return c, nil
}

// ===================== Brand operations =====================

func (r *CatalogRepository) CreateBrand(ctx context.Context, b *domain.Brand) (*domain.Brand, error) {
	query := `
		INSERT INTO brand (name, slug, logo_url, is_active, is_deleted)
		VALUES ($1, $2, $3, true, false)
		RETURNING id`
	err := r.db.QueryRow(ctx, query, b.Name, b.Slug, b.LogoURL).Scan(&b.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create brand: %w", err)
	}
	return b, nil
}

func (r *CatalogRepository) ListBrands(ctx context.Context) ([]*domain.Brand, error) {
	query := `
		SELECT id, name, slug, logo_url, is_active, is_deleted
		FROM brand
		WHERE is_deleted = false AND is_active = true
		ORDER BY id ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list brands: %w", err)
	}
	defer rows.Close()

	brands := make([]*domain.Brand, 0)
	for rows.Next() {
		b := &domain.Brand{}
		err := rows.Scan(&b.ID, &b.Name, &b.Slug, &b.LogoURL, &b.IsActive, &b.IsDeleted)
		if err != nil {
			return nil, fmt.Errorf("failed to scan brand: %w", err)
		}
		brands = append(brands, b)
	}

	return brands, nil
}

func (r *CatalogRepository) GetBrandByID(ctx context.Context, id int) (*domain.Brand, error) {
	b := &domain.Brand{}
	query := `
		SELECT id, name, slug, logo_url, is_active, is_deleted
		FROM brand
		WHERE id = $1 AND is_deleted = false`

	err := r.db.QueryRow(ctx, query, id).Scan(&b.ID, &b.Name, &b.Slug, &b.LogoURL, &b.IsActive, &b.IsDeleted)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("brand not found")
		}
		return nil, fmt.Errorf("failed to get brand: %w", err)
	}
	return b, nil
}

// ===================== Product operations =====================

func (r *CatalogRepository) CreateProduct(ctx context.Context, p *domain.Product, specs []domain.ProductSpec, opts []domain.ProductOptionType) (*domain.Product, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Serialize specs_jsonb if present
	var specsJSONBytes []byte
	if p.SpecsJSONB != nil {
		specsJSONBytes, err = json.Marshal(p.SpecsJSONB)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal specs_jsonb: %w", err)
		}
	}

	// 1. Insert product
	productQuery := `
		INSERT INTO product (id, category_id, brand_id, name, slug, meta_title, meta_description, img_thumb, weight, low_stock_threshold, specs_jsonb, is_active, is_deleted, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, true, false, NOW(), NOW())
		RETURNING created_at, updated_at`

	err = tx.QueryRow(ctx, productQuery,
		p.ID, p.CategoryID, p.BrandID, p.Name, p.Slug, p.MetaTitle, p.MetaDescription, p.ImgThumb, p.Weight, p.LowStockThreshold, specsJSONBytes,
	).Scan(&p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to insert product: %w", err)
	}

	// 2. Insert specifications
	specQuery := `
		INSERT INTO product_spec (product_id, "group", key, value, unit, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`
	for idx, spec := range specs {
		err = tx.QueryRow(ctx, specQuery, p.ID, spec.Group, spec.Key, spec.Value, spec.Unit, spec.SortOrder).Scan(&specs[idx].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to insert spec %s: %w", spec.Key, err)
		}
	}

	// 3. Insert option types
	optQuery := `
		INSERT INTO product_option_type (product_id, name, sort_order)
		VALUES ($1, $2, $3)
		RETURNING id`
	for idx, opt := range opts {
		err = tx.QueryRow(ctx, optQuery, p.ID, opt.Name, opt.SortOrder).Scan(&opts[idx].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to insert option type %s: %w", opt.Name, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit product: %w", err)
	}

	return p, nil
}

func (r *CatalogRepository) GetProductByID(ctx context.Context, id string) (*domain.Product, error) {
	p := &domain.Product{}
	var specsJSONBytes []byte

	query := `
		SELECT id, category_id, brand_id, name, slug, meta_title, meta_description, img_thumb, weight, low_stock_threshold, specs_jsonb, is_active, is_deleted, created_at, updated_at
		FROM product
		WHERE id = $1 AND is_deleted = false`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.CategoryID, &p.BrandID, &p.Name, &p.Slug, &p.MetaTitle, &p.MetaDescription, &p.ImgThumb, &p.Weight, &p.LowStockThreshold, &specsJSONBytes, &p.IsActive, &p.IsDeleted, &p.CreatedAt, &p.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("product not found")
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if len(specsJSONBytes) > 0 {
		var temp any
		if err := json.Unmarshal(specsJSONBytes, &temp); err == nil {
			p.SpecsJSONB = &temp
		}
	}

	return p, nil
}

func (r *CatalogRepository) ListProducts(ctx context.Context, query string, categoryID *int, brandID *int, specsQuery map[string]interface{}, page, limit int) ([]*domain.Product, int64, error) {
	offset := (page - 1) * limit

	// Dynamic query building
	sqlQuery := `SELECT id, category_id, brand_id, name, slug, img_thumb, specs_jsonb, is_active, created_at, updated_at FROM product WHERE is_deleted = false AND is_active = true`
	countQuery := `SELECT COUNT(*) FROM product WHERE is_deleted = false AND is_active = true`
	args := make([]interface{}, 0)
	argCount := 1

	if query != "" {
		sqlQuery += fmt.Sprintf(" AND (name ILIKE $%d OR meta_description ILIKE $%d)", argCount, argCount)
		countQuery += fmt.Sprintf(" AND (name ILIKE $%d OR meta_description ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+query+"%")
		argCount++
	}

	if categoryID != nil {
		sqlQuery += fmt.Sprintf(" AND category_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND category_id = $%d", argCount)
		args = append(args, *categoryID)
		argCount++
	}

	if brandID != nil {
		sqlQuery += fmt.Sprintf(" AND brand_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND brand_id = $%d", argCount)
		args = append(args, *brandID)
		argCount++
	}

	if len(specsQuery) > 0 {
		jsonBytes, err := json.Marshal(specsQuery)
		if err == nil {
			sqlQuery += fmt.Sprintf(" AND specs_jsonb @> $%d", argCount)
			countQuery += fmt.Sprintf(" AND specs_jsonb @> $%d", argCount)
			args = append(args, jsonBytes)
			argCount++
		}
	}

	// 1. Get total count
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// 2. Query actual list
	sqlQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()

	products := make([]*domain.Product, 0)
	for rows.Next() {
		p := &domain.Product{}
		var specsJSONBytes []byte
		err := rows.Scan(&p.ID, &p.CategoryID, &p.BrandID, &p.Name, &p.Slug, &p.ImgThumb, &specsJSONBytes, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}
		if len(specsJSONBytes) > 0 {
			var temp any
			if err := json.Unmarshal(specsJSONBytes, &temp); err == nil {
				p.SpecsJSONB = &temp
			}
		}
		products = append(products, p)
	}

	return products, total, nil
}

// ===================== Option & Variant operations =====================

func (r *CatalogRepository) CreateOptionValues(ctx context.Context, optionTypeID int, vals []domain.ProductOptionValue) ([]domain.ProductOptionValue, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO product_option_value (option_type_id, value, color_code, sort_order)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	for idx, val := range vals {
		err = tx.QueryRow(ctx, query, optionTypeID, val.Value, val.ColorCode, val.SortOrder).Scan(&vals[idx].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to insert option value %s: %w", val.Value, err)
		}
		vals[idx].OptionTypeID = optionTypeID
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit option values: %w", err)
	}

	return vals, nil
}

func (r *CatalogRepository) CreateVariant(ctx context.Context, v *domain.ProductVariant, optionValueIDs []int, images []domain.ProductImage) (*domain.ProductVariant, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Insert product_variant
	variantQuery := `
		INSERT INTO product_variant (product_id, name, sku, price, price_base, weight, is_active, is_deleted)
		VALUES ($1, $2, $3, $4, $5, $6, true, false)
		RETURNING id`
	err = tx.QueryRow(ctx, variantQuery, v.ProductID, v.Name, v.SKU, v.Price, v.PriceBase, v.Weight).Scan(&v.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert variant: %w", err)
	}

	// 2. Link variants options
	linkQuery := `
		INSERT INTO product_variant_option (variant_id, option_value_id)
		VALUES ($1, $2)`
	for _, valID := range optionValueIDs {
		_, err = tx.Exec(ctx, linkQuery, v.ID, valID)
		if err != nil {
			return nil, fmt.Errorf("failed to link variant option %d: %w", valID, err)
		}
	}

	// 3. Insert images
	imgQuery := `
		INSERT INTO product_image (product_id, variant_id, url, alt_text, sort_order, is_thumbnail)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`
	for idx, img := range images {
		err = tx.QueryRow(ctx, imgQuery, v.ProductID, v.ID, img.URL, img.AltText, img.SortOrder, img.IsThumbnail).Scan(&images[idx].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to insert variant image: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit variant: %w", err)
	}

	return v, nil
}

// ===================== Aggregate Details loaders =====================

func (r *CatalogRepository) GetSpecs(ctx context.Context, productID string) ([]domain.ProductSpec, error) {
	query := `
		SELECT id, product_id, "group", key, value, unit, sort_order
		FROM product_spec
		WHERE product_id = $1
		ORDER BY sort_order ASC, id ASC`

	rows, err := r.db.Query(ctx, query, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product specs: %w", err)
	}
	defer rows.Close()

	specs := make([]domain.ProductSpec, 0)
	for rows.Next() {
		s := domain.ProductSpec{}
		err := rows.Scan(&s.ID, &s.ProductID, &s.Group, &s.Key, &s.Value, &s.Unit, &s.SortOrder)
		if err != nil {
			return nil, fmt.Errorf("failed to scan spec: %w", err)
		}
		specs = append(specs, s)
	}

	return specs, nil
}

func (r *CatalogRepository) GetOptionTypes(ctx context.Context, productID string) ([]domain.OptionTypeDetail, error) {
	query := `
		SELECT id, product_id, name, sort_order
		FROM product_option_type
		WHERE product_id = $1
		ORDER BY sort_order ASC, id ASC`

	rows, err := r.db.Query(ctx, query, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get option types: %w", err)
	}
	defer rows.Close()

	types := make([]domain.OptionTypeDetail, 0)
	for rows.Next() {
		t := domain.OptionTypeDetail{}
		err := rows.Scan(&t.ID, &t.ProductID, &t.Name, &t.SortOrder)
		if err != nil {
			return nil, fmt.Errorf("failed to scan option type: %w", err)
		}
		types = append(types, t)
	}

	// Load values for each type
	valQuery := `
		SELECT id, option_type_id, value, color_code, sort_order
		FROM product_option_value
		WHERE option_type_id = $1
		ORDER BY sort_order ASC, id ASC`

	for idx, t := range types {
		valRows, err := r.db.Query(ctx, valQuery, t.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get option values: %w", err)
		}
		
		vals := make([]domain.ProductOptionValue, 0)
		for valRows.Next() {
			v := domain.ProductOptionValue{}
			err := valRows.Scan(&v.ID, &v.OptionTypeID, &v.Value, &v.ColorCode, &v.SortOrder)
			if err != nil {
				valRows.Close()
				return nil, fmt.Errorf("failed to scan option value: %w", err)
			}
			vals = append(vals, v)
		}
		valRows.Close()
		types[idx].Values = vals
	}

	return types, nil
}

func (r *CatalogRepository) GetVariants(ctx context.Context, productID string) ([]domain.VariantDetail, error) {
	query := `
		SELECT id, product_id, name, sku, price, price_base, weight, is_active, is_deleted
		FROM product_variant
		WHERE product_id = $1 AND is_deleted = false`

	rows, err := r.db.Query(ctx, query, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get variants: %w", err)
	}
	defer rows.Close()

	variants := make([]domain.VariantDetail, 0)
	for rows.Next() {
		v := domain.VariantDetail{}
		err := rows.Scan(&v.ID, &v.ProductID, &v.Name, &v.SKU, &v.Price, &v.PriceBase, &v.Weight, &v.IsActive, &v.IsDeleted)
		if err != nil {
			return nil, fmt.Errorf("failed to scan variant: %w", err)
		}
		variants = append(variants, v)
	}

	// Load linked options for each variant
	optQuery := `
		SELECT pov.id, pov.option_type_id, pov.value, pov.color_code, pov.sort_order
		FROM product_option_value pov
		JOIN product_variant_option pvo ON pov.id = pvo.option_value_id
		WHERE pvo.variant_id = $1`

	for idx, v := range variants {
		optRows, err := r.db.Query(ctx, optQuery, v.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get variant options: %w", err)
		}

		opts := make([]domain.ProductOptionValue, 0)
		for optRows.Next() {
			o := domain.ProductOptionValue{}
			err := optRows.Scan(&o.ID, &o.OptionTypeID, &o.Value, &o.ColorCode, &o.SortOrder)
			if err != nil {
				optRows.Close()
				return nil, fmt.Errorf("failed to scan variant option value: %w", err)
			}
			opts = append(opts, o)
		}
		optRows.Close()
		variants[idx].Options = opts
	}

	return variants, nil
}

func (r *CatalogRepository) GetImages(ctx context.Context, productID string) ([]domain.ProductImage, error) {
	query := `
		SELECT id, product_id, variant_id, url, alt_text, sort_order, is_thumbnail
		FROM product_image
		WHERE product_id = $1
		ORDER BY sort_order ASC, id ASC`

	rows, err := r.db.Query(ctx, query, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product images: %w", err)
	}
	defer rows.Close()

	images := make([]domain.ProductImage, 0)
	for rows.Next() {
		img := domain.ProductImage{}
		err := rows.Scan(&img.ID, &img.ProductID, &img.VariantID, &img.URL, &img.AltText, &img.SortOrder, &img.IsThumbnail)
		if err != nil {
			return nil, fmt.Errorf("failed to scan image: %w", err)
		}
		images = append(images, img)
	}

	return images, nil
}
