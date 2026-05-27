package domain

import (
	"context"
	"time"
)

// Brand maps to brand table
type Brand struct {
	ID        int    `json:"id" db:"id"`
	Name      string `json:"name" db:"name" validate:"required"`
	Slug      string `json:"slug" db:"slug"`
	LogoURL   *string `json:"logo_url,omitempty" db:"logo_url"`
	IsActive  bool   `json:"is_active" db:"is_active"`
	IsDeleted bool   `json:"is_deleted" db:"is_deleted"`
}

// Category maps to category table
type Category struct {
	ID        int    `json:"id" db:"id"`
	Name      string `json:"name" db:"name" validate:"required"`
	ParentID  *int   `json:"parent_id,omitempty" db:"parent_id"`
	Icon      *string `json:"icon,omitempty" db:"icon"`
	Slug      string `json:"slug" db:"slug"`
	SortOrder int    `json:"sort_order" db:"sort_order"`
	IsDeleted bool   `json:"is_deleted" db:"is_deleted"`
}

// CategoryNode is used for hierarchical trees
type CategoryNode struct {
	*Category
	Children []*CategoryNode `json:"children,omitempty"`
}

// Product maps to product table
type Product struct {
	ID                string     `json:"id" db:"id" validate:"required"`
	CategoryID        int        `json:"category_id" db:"category_id" validate:"required"`
	BrandID           int        `json:"brand_id" db:"brand_id" validate:"required"`
	Name              string     `json:"name" db:"name" validate:"required"`
	Slug              string     `json:"slug" db:"slug"`
	MetaTitle         *string    `json:"meta_title,omitempty" db:"meta_title"`
	MetaDescription   *string    `json:"meta_description,omitempty" db:"meta_description"`
	ImgThumb          *string    `json:"img_thumb,omitempty" db:"img_thumb"`
	Weight            *float64   `json:"weight,omitempty" db:"weight"`
	LowStockThreshold int        `json:"low_stock_threshold" db:"low_stock_threshold"`
	SpecsJSONB        *any       `json:"specs_jsonb,omitempty" db:"specs_jsonb"` // snapshot json
	IsActive          bool       `json:"is_active" db:"is_active"`
	IsDeleted         bool       `json:"is_deleted" db:"is_deleted"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

// ProductSpec maps to product_spec table
type ProductSpec struct {
	ID        int    `json:"id" db:"id"`
	ProductID string `json:"product_id" db:"product_id"`
	Group     string `json:"group" db:"group" validate:"required"`
	Key       string `json:"key" db:"key" validate:"required"`
	Value     string `json:"value" db:"value" validate:"required"`
	Unit      *string `json:"unit,omitempty" db:"unit"`
	SortOrder int    `json:"sort_order" db:"sort_order"`
}

// ProductOptionType maps to product_option_type table
type ProductOptionType struct {
	ID        int    `json:"id" db:"id"`
	ProductID string `json:"product_id" db:"product_id"`
	Name      string `json:"name" db:"name" validate:"required"` // e.g. Color, Storage
	SortOrder int    `json:"sort_order" db:"sort_order"`
}

// ProductOptionValue maps to product_option_value table
type ProductOptionValue struct {
	ID           int     `json:"id" db:"id"`
	OptionTypeID int     `json:"option_type_id" db:"option_type_id"`
	Value        string  `json:"value" db:"value" validate:"required"` // e.g. Red, 256GB
	ColorCode    *string `json:"color_code,omitempty" db:"color_code"`
	SortOrder    int     `json:"sort_order" db:"sort_order"`
}

// ProductVariant maps to product_variant table
type ProductVariant struct {
	ID        int      `json:"id" db:"id"`
	ProductID string   `json:"product_id" db:"product_id"`
	Name      string   `json:"name" db:"name" validate:"required"`
	SKU       string   `json:"sku" db:"sku" validate:"required"`
	Price     float64  `json:"price" db:"price" validate:"required,gt=0"`
	PriceBase *float64 `json:"price_base,omitempty" db:"price_base"`
	Weight    *float64 `json:"weight,omitempty" db:"weight"`
	IsActive  bool     `json:"is_active" db:"is_active"`
	IsDeleted bool     `json:"is_deleted" db:"is_deleted"`
}

// ProductImage maps to product_image table
type ProductImage struct {
	ID          int    `json:"id" db:"id"`
	ProductID   string `json:"product_id" db:"product_id"`
	VariantID   *int   `json:"variant_id,omitempty" db:"variant_id"`
	URL         string `json:"url" db:"url" validate:"required,url"`
	AltText     *string `json:"alt_text,omitempty" db:"alt_text"`
	SortOrder   int    `json:"sort_order" db:"sort_order"`
	IsThumbnail bool   `json:"is_thumbnail" db:"is_thumbnail"`
}

// Complex aggregation structs
type OptionTypeDetail struct {
	ProductOptionType
	Values []ProductOptionValue `json:"values"`
}

type VariantDetail struct {
	ProductVariant
	Options []ProductOptionValue `json:"options"`
}

type ProductDetail struct {
	Product     *Product           `json:"product"`
	Brand       *Brand             `json:"brand"`
	Category    *Category          `json:"category"`
	Specs       []ProductSpec      `json:"specs"`
	OptionTypes []OptionTypeDetail `json:"option_types"`
	Variants    []VariantDetail    `json:"variants"`
	Images      []ProductImage     `json:"images"`
}

type CreateProductRequest struct {
	Product     *Product            `json:"product" validate:"required"`
	Specs       []ProductSpec       `json:"specs,omitempty"`
	OptionTypes []ProductOptionType `json:"option_types,omitempty"`
}

type CreateOptionValueRequest struct {
	OptionTypeID int                  `json:"option_type_id" validate:"required"`
	Values       []ProductOptionValue `json:"values" validate:"required,dive"`
}

type GenerateVariantRequest struct {
	Variant       *ProductVariant `json:"variant" validate:"required"`
	OptionValueIDs []int          `json:"option_value_ids" validate:"required"`
	Images        []ProductImage  `json:"images,omitempty"`
}

// CatalogRepository defines DB capabilities
type CatalogRepository interface {
	// Brand & Category writes
	CreateCategory(ctx context.Context, c *Category) (*Category, error)
	ListCategories(ctx context.Context) ([]*Category, error)
	GetCategoryByID(ctx context.Context, id int) (*Category, error)

	CreateBrand(ctx context.Context, b *Brand) (*Brand, error)
	ListBrands(ctx context.Context) ([]*Brand, error)
	GetBrandByID(ctx context.Context, id int) (*Brand, error)

	// Product writes
	CreateProduct(ctx context.Context, p *Product, specs []ProductSpec, opts []ProductOptionType) (*Product, error)
	GetProductByID(ctx context.Context, id string) (*Product, error)
	ListProducts(ctx context.Context, query string, categoryID *int, brandID *int, specsQuery map[string]interface{}, page, limit int) ([]*Product, int64, error)

	// Variant & Options writes
	CreateOptionValues(ctx context.Context, optionTypeID int, vals []ProductOptionValue) ([]ProductOptionValue, error)
	CreateVariant(ctx context.Context, v *ProductVariant, optionValueIDs []int, images []ProductImage) (*ProductVariant, error)

	// Detailed mapping checks
	GetSpecs(ctx context.Context, productID string) ([]ProductSpec, error)
	GetOptionTypes(ctx context.Context, productID string) ([]OptionTypeDetail, error)
	GetVariants(ctx context.Context, productID string) ([]VariantDetail, error)
	GetImages(ctx context.Context, productID string) ([]ProductImage, error)
}

// CatalogUsecase defines business rules
type CatalogUsecase interface {
	CreateCategory(ctx context.Context, c *Category) (*Category, error)
	ListCategoryTree(ctx context.Context) ([]*CategoryNode, error)
	CreateBrand(ctx context.Context, b *Brand) (*Brand, error)
	ListBrands(ctx context.Context) ([]*Brand, error)

	CreateProduct(ctx context.Context, req *CreateProductRequest) (*Product, error)
	SearchProducts(ctx context.Context, query string, categoryID *int, brandID *int, specsQuery map[string]interface{}, page, limit int) ([]*Product, int64, error)
	GetProductDetails(ctx context.Context, id string) (*ProductDetail, error)

	AddOptionValues(ctx context.Context, optionTypeID int, values []ProductOptionValue) ([]ProductOptionValue, error)
	GenerateVariant(ctx context.Context, req *GenerateVariantRequest) (*ProductVariant, error)
}
