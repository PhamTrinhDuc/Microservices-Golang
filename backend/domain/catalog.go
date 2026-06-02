package domain

import (
	"context"
	"time"
)

// --- models ---

type Category struct {
	ID        int     `json:"id" db:"id"`
	Name      string  `json:"name" db:"name"`
	ParentID  *int    `json:"parent_id" db:"parent_id"`
	Icon      *string `json:"icon" db:"icon"`
	Slug      string  `json:"slug" db:"slug"`
	SortOrder int     `json:"sort_order" db:"sort_order"`
	IsDeleted bool    `json:"is_deleted" db:"is_deleted"`
}

type Brand struct {
	ID        int     `json:"id" db:"id"`
	Name      string  `json:"name" db:"name"`
	Slug      string  `json:"slug" db:"slug"`
	LogoURL   *string `json:"logo_url" db:"logo_url"`
	IsActive  bool    `json:"is_active" db:"is_active"`
	IsDeleted bool    `json:"is_deleted" db:"is_deleted"`
}

type Product struct {
	ID                string      `json:"id" db:"id"`
	CategoryID        int         `json:"category_id" db:"category_id"`
	BrandID           int         `json:"brand_id" db:"brand_id"`
	Name              string      `json:"name" db:"name"`
	Slug              string      `json:"slug" db:"slug"`
	MetaTitle         *string     `json:"meta_title" db:"meta_title"`
	MetaDescription   *string     `json:"meta_description" db:"meta_description"`
	ImgThumb          *string     `json:"img_thumb" db:"img_thumb"`
	Weight            *float64    `json:"weight" db:"weight"`
	LowStockThreshold int         `json:"low_stock_threshold" db:"low_stock_threshold"`
	SpecsJSONB        interface{} `json:"specs_jsonb,omitempty" db:"specs_jsonb"`
	IsActive          bool        `json:"is_active" db:"is_active"`
	IsDeleted         bool        `json:"is_deleted" db:"is_deleted"`
	Price             float64     `json:"price"`
	DiscountPrice     *float64    `json:"discount_price"`
	DiscountPercent   int         `json:"discount_percent"`
	Stock             int         `json:"stock"`
	Rating            float64     `json:"rating"`
	ReviewCount       int         `json:"review_count"`
	CreatedAt         time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at" db:"updated_at"`
}

type ProductSpec struct {
	ID        int     `json:"id" db:"id"`
	ProductID string  `json:"product_id" db:"product_id"`
	Group     string  `json:"group" db:"group"`
	Key       string  `json:"key" db:"key"`
	Value     string  `json:"value" db:"value"`
	Unit      *string `json:"unit" db:"unit"`
	SortOrder int     `json:"sort_order" db:"sort_order"`
}

type ProductOptionValue struct {
	ID           int     `json:"id" db:"id"`
	OptionTypeID int     `json:"option_type_id" db:"option_type_id"`
	Value        string  `json:"value" db:"value"`
	ColorCode    *string `json:"color_code" db:"color_code"`
	SortOrder    int     `json:"sort_order" db:"sort_order"`
}

type ProductOptionType struct {
	ID        int                  `json:"id" db:"id"`
	ProductID string               `json:"product_id" db:"product_id"`
	Name      string               `json:"name" db:"name"`
	SortOrder int                  `json:"sort_order" db:"sort_order"`
	Values    []ProductOptionValue `json:"values,omitempty"`
}

type VariantOption struct {
	OptionTypeID   int     `json:"option_type_id"`
	OptionTypeName string  `json:"option_type_name"`
	ValueID        int     `json:"value_id"`
	Value          string  `json:"value"`
	ColorCode      *string `json:"color_code"`
}

type ProductVariant struct {
	ID        int             `json:"id" db:"id"`
	ProductID string          `json:"product_id" db:"product_id"`
	Name      string          `json:"name" db:"name"`
	SKU       string          `json:"sku" db:"sku"`
	Price     float64         `json:"price" db:"price"`
	PriceBase *float64        `json:"price_base" db:"price_base"`
	Weight    *float64        `json:"weight" db:"weight"`
	IsActive  bool            `json:"is_active" db:"is_active"`
	IsDeleted bool            `json:"is_deleted" db:"is_deleted"`
	Options   []VariantOption `json:"options,omitempty"`
}

type ProductImage struct {
	ID          int     `json:"id" db:"id"`
	ProductID   string  `json:"product_id" db:"product_id"`
	VariantID   *int    `json:"variant_id" db:"variant_id"`
	URL         string  `json:"url" db:"url"`
	AltText     *string `json:"alt_text" db:"alt_text"`
	SortOrder   int     `json:"sort_order" db:"sort_order"`
	IsThumbnail bool    `json:"is_thumbnail" db:"is_thumbnail"`
}

type Review struct {
	ID           int       `json:"id" db:"id"`
	UserID       int       `json:"user_id" db:"user_id"`
	UserFullName string    `json:"user_full_name" db:"user_full_name"`
	ProductID    string    `json:"product_id" db:"product_id"`
	OrderID      int       `json:"order_id" db:"order_id"`
	Rating       int       `json:"rating" db:"rating"`
	Comment      *string   `json:"comment" db:"comment"`
	Images       interface{} `json:"images" db:"images"` // JSONB
	Status       string    `json:"status" db:"status"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// --- request & response dtos ---

type CreateCategoryRequest struct {
	Name      string  `json:"name" validate:"required"`
	ParentID  *int    `json:"parent_id"`
	Icon      *string `json:"icon"`
	Slug      string  `json:"slug"` // optional
	SortOrder int     `json:"sort_order"`
}

type UpdateCategoryRequest struct {
	Name      string  `json:"name" validate:"required"`
	ParentID  *int    `json:"parent_id"`
	Icon      *string `json:"icon"`
	Slug      string  `json:"slug"`
	SortOrder int     `json:"sort_order"`
}

type CreateBrandRequest struct {
	Name     string  `json:"name" validate:"required"`
	Slug     string  `json:"slug"` // optional
	LogoURL  *string `json:"logo_url"`
	IsActive bool    `json:"is_active"`
}

type UpdateBrandRequest struct {
	Name     string  `json:"name" validate:"required"`
	Slug     string  `json:"slug"`
	LogoURL  *string `json:"logo_url"`
	IsActive bool    `json:"is_active"`
}

type ProductSpecDTO struct {
	Group     string  `json:"group" validate:"required"`
	Key       string  `json:"key" validate:"required"`
	Value     string  `json:"value" validate:"required"`
	Unit      *string `json:"unit"`
	SortOrder int     `json:"sort_order"`
}

type OptionValueDTO struct {
	Value     string  `json:"value" validate:"required"`
	ColorCode *string `json:"color_code"`
	SortOrder int     `json:"sort_order"`
}

type OptionDTO struct {
	Name   string           `json:"name" validate:"required"`
	Values []OptionValueDTO `json:"values" validate:"required,min=1,dive"`
}

type VariantDTO struct {
	Name        string   `json:"name" validate:"required"`
	SKU         string   `json:"sku" validate:"required"`
	Price       float64  `json:"price" validate:"required,gt=0"`
	PriceBase   *float64 `json:"price_base"`
	Weight      *float64 `json:"weight"`
	OptionValue string   `json:"option_value" validate:"required"`
}

type CreateProductRequest struct {
	ID                string           `json:"id" validate:"required"`
	CategoryID        int              `json:"category_id" validate:"required"`
	BrandID           int              `json:"brand_id" validate:"required"`
	Name              string           `json:"name" validate:"required"`
	Slug              string           `json:"slug"`
	MetaTitle         *string          `json:"meta_title"`
	MetaDescription   *string          `json:"meta_description"`
	ImgThumb          *string          `json:"img_thumb"`
	Weight            *float64         `json:"weight"`
	LowStockThreshold int              `json:"low_stock_threshold"`
	SpecsJSONB        interface{}      `json:"specs_jsonb"`
	Specs             []ProductSpecDTO `json:"specs"`
	Options           []OptionDTO      `json:"options"`
	Variants          []VariantDTO     `json:"variants"`
	Images            []string         `json:"images"`
}

type UpdateProductRequest struct {
	CategoryID        int              `json:"category_id" validate:"required"`
	BrandID           int              `json:"brand_id" validate:"required"`
	Name              string           `json:"name" validate:"required"`
	Slug              string           `json:"slug"`
	MetaTitle         *string          `json:"meta_title"`
	MetaDescription   *string          `json:"meta_description"`
	ImgThumb          *string          `json:"img_thumb"`
	Weight            *float64         `json:"weight"`
	LowStockThreshold int              `json:"low_stock_threshold"`
	SpecsJSONB        interface{}      `json:"specs_jsonb"`
	Specs             []ProductSpecDTO `json:"specs"`
	Images            []string         `json:"images"`
}

type OptionValueRequest struct {
	Value     string  `json:"value" validate:"required"`
	ColorCode *string `json:"color_code"`
	SortOrder int     `json:"sort_order"`
}

type AddOptionValuesRequest struct {
	ProductID  string               `json:"product_id" validate:"required"`
	OptionName string               `json:"option_name" validate:"required"`
	Values     []OptionValueRequest `json:"values" validate:"required,min=1,dive"`
}

type GenerateVariantRequest struct {
	Name           string   `json:"name" validate:"required"`
	SKU            string   `json:"sku" validate:"required"`
	Price          float64  `json:"price" validate:"required,gt=0"`
	PriceBase      *float64 `json:"price_base"`
	Weight         *float64 `json:"weight"`
	OptionValueIDs []int    `json:"option_value_ids" validate:"required,min=1"`
}

type UpdateVariantRequest struct {
	Name      string   `json:"name" validate:"required"`
	SKU       string   `json:"sku" validate:"required"`
	Price     float64  `json:"price" validate:"required,gt=0"`
	PriceBase *float64 `json:"price_base"`
	Weight    *float64 `json:"weight"`
}

type ProductSearchQuery struct {
	CategoryID *int   `form:"category_id"`
	BrandID    *int   `form:"brand_id"`
	Query      string `form:"q"`
	Page       int    `form:"page"`
	Limit      int    `form:"limit"`
	Sort       string `form:"sort"`
}

type ProductSearchResult struct {
	Products   []*Product `json:"products"`
	TotalCount int        `json:"total_count"`
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
}

type ProductDetailsResponse struct {
	Product      *Product             `json:"product"`
	BrandName    string               `json:"brand_name"`
	CategoryName string               `json:"category_name"`
	Specs        []*ProductSpec       `json:"specs"`
	OptionTypes  []*ProductOptionType `json:"option_types"`
	Variants     []*ProductVariant    `json:"variants"`
	Images       []*ProductImage      `json:"images"`
	Siblings     []*Product           `json:"siblings,omitempty"` // For capacity switching like TGDD
	Reviews      []*Review            `json:"reviews"`
}

type CreateProductInput struct {
	Product     *Product
	Specs       []*ProductSpec
	OptionTypes []*ProductOptionType
	Variants    []*ProductVariant
	Images      []*ProductImage
}

// --- interfaces ---

type CatalogRepository interface {
	// Category
	CreateCategory(ctx context.Context, cat *Category) (*Category, error)
	ListCategories(ctx context.Context) ([]*Category, error)
	GetCategoryByID(ctx context.Context, id int) (*Category, error)
	UpdateCategory(ctx context.Context, cat *Category) (*Category, error)
	DeleteCategory(ctx context.Context, id int) error

	// Brand
	CreateBrand(ctx context.Context, brand *Brand) (*Brand, error)
	ListBrands(ctx context.Context) ([]*Brand, error)
	GetBrandByID(ctx context.Context, id int) (*Brand, error)
	UpdateBrand(ctx context.Context, brand *Brand) (*Brand, error)
	DeleteBrand(ctx context.Context, id int) error

	// Product
	CreateProduct(ctx context.Context, input *CreateProductInput) (*Product, error)
	GetProductByID(ctx context.Context, id string) (*Product, error)
	SearchProducts(ctx context.Context, query *ProductSearchQuery) (*ProductSearchResult, error)
	GetProductDetails(ctx context.Context, id string) (*ProductDetailsResponse, error)
	UpdateProduct(ctx context.Context, prod *Product, specs []*ProductSpec, images []*ProductImage) (*Product, error)
	DeleteProduct(ctx context.Context, id string) error

	// Options
	AddOptionValues(ctx context.Context, productID string, optionName string, values []*ProductOptionValue) ([]*ProductOptionValue, error)
	
	// Variant
	CreateVariant(ctx context.Context, variant *ProductVariant, optionValueIDs []int) (*ProductVariant, error)
	UpdateVariant(ctx context.Context, id int, name, sku string, price float64, priceBase, weight *float64) (*ProductVariant, error)
	DeleteVariant(ctx context.Context, id int) error
}

type CatalogUsecase interface {
	// Category
	CreateCategory(ctx context.Context, req *CreateCategoryRequest) (*Category, error)
	ListCategories(ctx context.Context) ([]*Category, error)
	UpdateCategory(ctx context.Context, id int, req *UpdateCategoryRequest) (*Category, error)
	DeleteCategory(ctx context.Context, id int) error

	// Brand
	CreateBrand(ctx context.Context, req *CreateBrandRequest) (*Brand, error)
	ListBrands(ctx context.Context) ([]*Brand, error)
	UpdateBrand(ctx context.Context, id int, req *UpdateBrandRequest) (*Brand, error)
	DeleteBrand(ctx context.Context, id int) error

	// Product
	CreateProduct(ctx context.Context, req *CreateProductRequest) (*Product, error)
	SearchProducts(ctx context.Context, query *ProductSearchQuery) (*ProductSearchResult, error)
	GetProductDetails(ctx context.Context, id string) (*ProductDetailsResponse, error)
	UpdateProduct(ctx context.Context, id string, req *UpdateProductRequest) (*Product, error)
	DeleteProduct(ctx context.Context, id string) error

	// Options & Variants
	AddOptionValues(ctx context.Context, req *AddOptionValuesRequest) ([]*ProductOptionValue, error)
	GenerateVariant(ctx context.Context, productID string, req *GenerateVariantRequest) (*ProductVariant, error)
	UpdateVariant(ctx context.Context, id int, req *UpdateVariantRequest) (*ProductVariant, error)
	DeleteVariant(ctx context.Context, id int) error
}
