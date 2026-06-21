package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"backend/domain"
	"backend/internal/utils"
)

type CatalogUsecase struct {
	repo domain.CatalogRepository
}

func NewCatalogUsecase(repo domain.CatalogRepository) *CatalogUsecase {
	return &CatalogUsecase{repo: repo}
}

// --- slug generator helper ---

func generateSlug(src string) string {
	src = strings.ToLower(src)
	vietnameseChars := map[string][]string{
		"a": {"á", "à", "ả", "ã", "ạ", "ă", "ắ", "ằ", "ẳ", "ẵ", "ặ", "â", "ấ", "ầ", "ẩ", "ẫ", "ậ"},
		"d": {"đ"},
		"e": {"é", "è", "ẻ", "ẽ", "ẹ", "ê", "ế", "ề", "ể", "ễ", "ệ"},
		"i": {"í", "ì", "ỉ", "ĩ", "ị"},
		"o": {"ó", "ò", "ỏ", "õ", "ọ", "ô", "ố", "ồ", "ổ", "ỗ", "ộ", "ơ", "ớ", "ờ", "ở", "ỡ", "ợ"},
		"u": {"ú", "ù", "ủ", "ũ", "ụ", "ư", "ứ", "ừ", "ử", "ữ", "ự"},
		"y": {"ý", "ỳ", "ỷ", "ỹ", "ỵ"},
	}
	for replace, list := range vietnameseChars {
		for _, char := range list {
			src = strings.ReplaceAll(src, char, replace)
		}
	}
	var sb strings.Builder
	lastWasHyphen := false
	for _, r := range src {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			sb.WriteRune(r)
			lastWasHyphen = false
		} else {
			if !lastWasHyphen && sb.Len() > 0 {
				sb.WriteRune('-')
				lastWasHyphen = true
			}
		}
	}
	result := sb.String()
	result = strings.Trim(result, "-")
	return result
}

func getBaseProductName(name string) string {
	parts := strings.Split(name, " ")
	if len(parts) <= 1 {
		return name
	}
	var baseParts []string
	for _, part := range parts {
		p := strings.ToLower(part)
		// Stop if we hit a capacity suffix or spec keyword
		if strings.HasSuffix(p, "gb") || strings.HasSuffix(p, "tb") {
			break
		}
		baseParts = append(baseParts, part)
	}
	if len(baseParts) == 0 {
		return parts[0]
	}
	return strings.Join(baseParts, " ")
}

// return the list of filters used to search for products and suspected to be the culprit
func diagnoseEmptyResult(query *domain.ProductSearchQuery) []string {
	var culprits []string
	if query.BrandID != nil {
		culprits = append(culprits, "brand_id")
	}
	if query.CategoryID != nil {
		culprits = append(culprits, "category_id")
	}
	if query.MinPrice != nil || query.MaxPrice != nil {
		culprits = append(culprits, "price_range")
	}
	if query.MinRating != nil {
		culprits = append(culprits, "min_rating")
	}
	if query.InStockOnly {
		culprits = append(culprits, "in_stock_only")
	}
	if query.SpecFilter != nil {
		culprits = append(culprits, "spec_filter")
	}
	if query.Query != "" {
		culprits = append(culprits, "query")
	}
	return culprits
}

func buildRelexHint(culprint string, count int) string {
	switch culprint {
	case "brand_id":
		return fmt.Sprintf("Bao gồm hết tất cả các hãng sẽ có %d kết quả", count)
	case "category_id":
		return fmt.Sprintf("Bao gồm hết tất cả các danh mục sẽ có %d kết quả", count)
	case "in_stock_only":
		return fmt.Sprintf("Bao gồm hết hàng sẽ có %d kết quả", count)
	case "price_range":
		return fmt.Sprintf("Bỏ khoảng giá sẽ có %d kết quả", count)
	case "min_rating":
		return fmt.Sprintf("Bỏ đánh giá tối thiểu sẽ có %d kết quả", count)
	case "spec_filter":
		return fmt.Sprintf("Bỏ lọc thông số sẽ có %d kết quả", count)
	case "query":
		return fmt.Sprintf("Tìm kiếm rộng hơn trong danh mục sẽ có %d kết quả", count)
	default:
		return fmt.Sprintf("Nới lỏng bộ lọc sẽ có %d kết quả", count)
	}
}

// --- Category ---

func (uc *CatalogUsecase) CreateCategory(ctx context.Context, req *domain.CreateCategoryRequest) (*domain.Category, error) {
	if req == nil {
		return nil, errors.New("category request cannot be nil")
	}

	slug := req.Slug
	if slug == "" {
		slug = generateSlug(req.Name)
	}

	if req.ParentID != nil {
		// Validate parent category exists
		_, err := uc.repo.GetCategoryByID(ctx, *req.ParentID)
		if err != nil {
			if errors.Is(err, domain.ErrCategoryNotFound) {
				return nil, fmt.Errorf("parent category ID %d does not exist", *req.ParentID)
			}
			return nil, err
		}
	}

	cat := &domain.Category{
		Name:      req.Name,
		ParentID:  req.ParentID,
		Icon:      req.Icon,
		Slug:      slug,
		SortOrder: req.SortOrder,
	}

	return uc.repo.CreateCategory(ctx, cat)
}

func (uc *CatalogUsecase) ListCategories(ctx context.Context, req *domain.ListCategoriesRequest) ([]*domain.Category, error) {
	return uc.repo.ListCategories(ctx, req)
}

func (uc *CatalogUsecase) UpdateCategory(ctx context.Context, id int, req *domain.UpdateCategoryRequest) (*domain.Category, error) {
	if req == nil {
		return nil, errors.New("category request cannot be nil")
	}

	// 1. Check if category exists
	existing, err := uc.repo.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Validate ParentID
	if req.ParentID != nil {
		if *req.ParentID == id {
			return nil, errors.New("a category cannot be its own parent")
		}
		_, err := uc.repo.GetCategoryByID(ctx, *req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent category ID %d does not exist", *req.ParentID)
		}
	}

	slug := req.Slug
	if slug == "" {
		slug = generateSlug(req.Name)
	}

	existing.Name = req.Name
	existing.ParentID = req.ParentID
	existing.Icon = req.Icon
	existing.Slug = slug
	existing.SortOrder = req.SortOrder

	return uc.repo.UpdateCategory(ctx, existing)
}

func (uc *CatalogUsecase) DeleteCategory(ctx context.Context, id int) error {
	return uc.repo.DeleteCategory(ctx, id)
}

func (uc *CatalogUsecase) GetSpecsByCategoryID(ctx context.Context, categoryID int) ([]*domain.CategorySpec, error) {
	_, err := uc.repo.GetCategoryByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	return uc.repo.GetSpecsByCategoryID(ctx, categoryID)
}

// --- Brand ---

func (uc *CatalogUsecase) CreateBrand(ctx context.Context, req *domain.CreateBrandRequest) (*domain.Brand, error) {
	if req == nil {
		return nil, errors.New("brand request cannot be nil")
	}

	slug := req.Slug
	if slug == "" {
		slug = generateSlug(req.Name)
	}

	brand := &domain.Brand{
		Name:     req.Name,
		Slug:     slug,
		LogoURL:  req.LogoURL,
		IsActive: req.IsActive,
	}

	return uc.repo.CreateBrand(ctx, brand)
}

func (uc *CatalogUsecase) ListBrands(ctx context.Context, req *domain.ListBrandsRequest) ([]*domain.Brand, error) {
	return uc.repo.ListBrands(ctx, req)
}

func (uc *CatalogUsecase) UpdateBrand(ctx context.Context, id int, req *domain.UpdateBrandRequest) (*domain.Brand, error) {
	if req == nil {
		return nil, errors.New("brand request cannot be nil")
	}

	existing, err := uc.repo.GetBrandByID(ctx, id)
	if err != nil {
		return nil, err
	}

	slug := req.Slug
	if slug == "" {
		slug = generateSlug(req.Name)
	}

	existing.Name = req.Name
	existing.Slug = slug
	existing.LogoURL = req.LogoURL
	existing.IsActive = req.IsActive

	return uc.repo.UpdateBrand(ctx, existing)
}

func (uc *CatalogUsecase) DeleteBrand(ctx context.Context, id int) error {
	return uc.repo.DeleteBrand(ctx, id)
}

// --- Product ---

func (uc *CatalogUsecase) CreateProduct(ctx context.Context, req *domain.CreateProductRequest) (*domain.Product, error) {
	if req == nil {
		return nil, errors.New("product request cannot be nil")
	}

	// 1. Validate Category and Brand
	_, err := uc.repo.GetCategoryByID(ctx, req.CategoryID)
	if err != nil {
		return nil, err
	}

	_, err = uc.repo.GetBrandByID(ctx, req.BrandID)
	if err != nil {
		return nil, err
	}

	// 2. Slug generation
	slug := req.Slug
	if slug == "" {
		slug = generateSlug(req.Name)
	}

	// 3. Mapping Product Specs
	specs := make([]*domain.ProductSpec, 0, len(req.Specs))
	for _, specDTO := range req.Specs {
		specs = append(specs, &domain.ProductSpec{
			ProductID: req.ID,
			Group:     specDTO.Group,
			Key:       specDTO.Key,
			Value:     specDTO.Value,
			Unit:      specDTO.Unit,
			SortOrder: specDTO.SortOrder,
		})
	}

	// 4. Mapping Product Options
	optionTypes := make([]*domain.ProductOptionType, 0, len(req.Options))
	for idx, optDTO := range req.Options {
		values := make([]domain.ProductOptionValue, 0, len(optDTO.Values))
		for _, valDTO := range optDTO.Values {
			values = append(values, domain.ProductOptionValue{
				Value:     valDTO.Value,
				ColorCode: valDTO.ColorCode,
				SortOrder: valDTO.SortOrder,
			})
		}

		optionTypes = append(optionTypes, &domain.ProductOptionType{
			ProductID: req.ID,
			Name:      optDTO.Name,
			SortOrder: idx,
			Values:    values,
		})
	}

	// 5. Mapping Variants
	variants := make([]*domain.ProductVariant, 0, len(req.Variants))
	for _, vDTO := range req.Variants {
		// Find which option type this value belongs to
		var optTypeName string
		found := false
		for _, opt := range req.Options {
			for _, val := range opt.Values {
				if val.Value == vDTO.OptionValue {
					optTypeName = opt.Name
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			if len(req.Options) > 0 {
				optTypeName = req.Options[0].Name
			} else {
				optTypeName = "Màu sắc"
			}
		}

		variants = append(variants, &domain.ProductVariant{
			ProductID:    req.ID,
			Name:         vDTO.Name,
			SKU:          vDTO.SKU,
			SellPrice:    vDTO.SellPrice,
			ComparePrice: vDTO.ComparePrice,
			Weight:       vDTO.Weight,
			Options: []domain.VariantOption{
				{
					OptionTypeName: optTypeName,
					Value:          vDTO.OptionValue,
				},
			},
		})
	}

	images := make([]*domain.ProductImage, 0, len(req.Images))
	for idx, imgURL := range req.Images {
		images = append(images, &domain.ProductImage{
			ProductID: req.ID,
			URL:       imgURL,
			SortOrder: idx,
		})
	}

	input := &domain.CreateProductInput{
		Product: &domain.Product{
			ID:                req.ID,
			CategoryID:        req.CategoryID,
			BrandID:           req.BrandID,
			Name:              req.Name,
			Slug:              slug,
			MetaTitle:         req.MetaTitle,
			MetaDescription:   req.MetaDescription,
			ImgThumb:          req.ImgThumb,
			Weight:            req.Weight,
			LowStockThreshold: req.LowStockThreshold,
			SpecsJSONB:        req.SpecsJSONB,
		},
		Specs:       specs,
		OptionTypes: optionTypes,
		Variants:    variants,
		Images:      images,
	}

	return uc.repo.CreateProduct(ctx, input)
}

func (uc *CatalogUsecase) buildSuggestions(ctx context.Context, query *domain.ProductSearchQuery, culprints []string) map[string]any {
	suggestions := map[string]any{}
	// try remove suspicious filters one by one
	relexHints := []map[string]any{}

	for _, culprint := range culprints {
		relexedQuery := utils.CloneQuery(query)
		switch culprint {
		case "brand_id":
			relexedQuery.BrandID = nil
		case "category_id":
			relexedQuery.CategoryID = nil
		case "price_range":
			relexedQuery.MinPrice = nil
			relexedQuery.MaxPrice = nil
		case "min_price":
			relexedQuery.MinPrice = nil
		case "max_price":
			relexedQuery.MaxPrice = nil
		case "min_rating":
			relexedQuery.MinRating = nil
		case "in_stock_only":
			relexedQuery.InStockOnly = false
		case "spec_filter":
			relexedQuery.SpecFilter = nil
		case "query":
			relexedQuery.Query = ""
		default:
			continue
		}
		relexedRes, err := uc.repo.SearchProducts(ctx, relexedQuery)
		if err == nil && len(relexedRes.Products) > 0 {
			relexHints = append(relexHints, map[string]any{
				"remove_filter": culprint,
				"result_count":  relexedRes.TotalCount,
				"hint":          buildRelexHint(culprint, relexedRes.TotalCount),
			})
		}
	}

	if len(relexHints) > 0 {
		suggestions["relex_filters"] = relexHints
	}

	if len(relexHints) == 0 && query.CategoryID != nil {
		porpularQuery := &domain.ProductSearchQuery{
			CategoryID:  query.CategoryID,
			BrandID:     query.BrandID,
			Limit:       5,
			Page:        1,
			InStockOnly: true,
			Sort:        "rating_desc",
		}
		products, err := uc.repo.SearchProducts(ctx, porpularQuery)
		if err == nil && len(products.Products) > 0 {
			suggestions["popular_products"] = products.Products
		}
	}
	return suggestions
}

func (uc *CatalogUsecase) SearchProducts(ctx context.Context, query *domain.ProductSearchQuery) (*domain.ProductSearchResult, error) {
	products, err := uc.repo.SearchProducts(ctx, query)
	if err != nil {
		return nil, err
	}

	hasMore := query.Page*query.Limit < products.TotalCount
	appliedFilter := diagnoseEmptyResult(query) // get filters apply to search
	suggestions := make(map[string]any)
	if len(products.Products) == 0 {
		suggestions = uc.buildSuggestions(ctx, query, appliedFilter)
	}

	products.AppliedFilters = appliedFilter
	products.HasMore = hasMore
	products.Suggestions = suggestions

	return products, nil
}

func (uc *CatalogUsecase) GetProductDetails(ctx context.Context, id string) (*domain.ProductDetailsResponse, error) {
	// 1. Get Product Details
	res, err := uc.repo.GetProductDetails(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. TGDD Feature: Query Sibling Products (e.g. iPhone 17 128GB, iPhone 17 512GB)
	baseName := getBaseProductName(res.Product.Name)
	siblingQuery := &domain.ProductSearchQuery{
		CategoryID: &res.Product.CategoryID,
		BrandID:    &res.Product.BrandID,
		Query:      baseName,
		Page:       1,
		Limit:      50,
	}

	siblingRes, err := uc.repo.SearchProducts(ctx, siblingQuery)
	if err == nil && siblingRes != nil {
		// Filter out the current product from siblings
		siblings := make([]*domain.Product, 0)
		for _, p := range siblingRes.Products {
			if p.ID != res.Product.ID {
				siblings = append(siblings, p)
			}
		}
		res.Siblings = siblings
	}

	return res, nil
}

func (uc *CatalogUsecase) UpdateProduct(ctx context.Context, id string, req *domain.UpdateProductRequest) (*domain.Product, error) {
	if req == nil {
		return nil, errors.New("product update request cannot be nil")
	}

	// 1. Check if product exists
	existing, err := uc.repo.GetProductByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Validate Category and Brand
	_, err = uc.repo.GetCategoryByID(ctx, req.CategoryID)
	if err != nil {
		return nil, err
	}

	_, err = uc.repo.GetBrandByID(ctx, req.BrandID)
	if err != nil {
		return nil, err
	}

	// 3. Slug generation
	slug := req.Slug
	if slug == "" {
		slug = generateSlug(req.Name)
	}

	existing.CategoryID = req.CategoryID
	existing.BrandID = req.BrandID
	existing.Name = req.Name
	existing.Slug = slug
	existing.MetaTitle = req.MetaTitle
	existing.MetaDescription = req.MetaDescription
	existing.ImgThumb = req.ImgThumb
	existing.Weight = req.Weight
	existing.LowStockThreshold = req.LowStockThreshold
	existing.SpecsJSONB = req.SpecsJSONB

	// 4. Map specs
	specs := make([]*domain.ProductSpec, 0, len(req.Specs))
	for _, specDTO := range req.Specs {
		specs = append(specs, &domain.ProductSpec{
			ProductID: id,
			Group:     specDTO.Group,
			Key:       specDTO.Key,
			Value:     specDTO.Value,
			Unit:      specDTO.Unit,
			SortOrder: specDTO.SortOrder,
		})
	}

	images := make([]*domain.ProductImage, 0, len(req.Images))
	for idx, imgURL := range req.Images {
		images = append(images, &domain.ProductImage{
			ProductID: id,
			URL:       imgURL,
			SortOrder: idx,
		})
	}

	return uc.repo.UpdateProduct(ctx, existing, specs, images)
}

func (uc *CatalogUsecase) DeleteProduct(ctx context.Context, id string) error {
	// Check product exists
	_, err := uc.repo.GetProductByID(ctx, id)
	if err != nil {
		return err
	}

	return uc.repo.DeleteProduct(ctx, id)
}

func (uc *CatalogUsecase) AddOptionValues(ctx context.Context, req *domain.AddOptionValuesRequest) ([]*domain.ProductOptionValue, error) {
	if req == nil {
		return nil, errors.New("request payload cannot be nil")
	}

	// Check product exists
	_, err := uc.repo.GetProductByID(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	values := make([]*domain.ProductOptionValue, 0, len(req.Values))
	for _, vReq := range req.Values {
		values = append(values, &domain.ProductOptionValue{
			Value:     vReq.Value,
			ColorCode: vReq.ColorCode,
			SortOrder: vReq.SortOrder,
		})
	}

	return uc.repo.AddOptionValues(ctx, req.ProductID, req.OptionName, values)
}

func (uc *CatalogUsecase) GenerateVariant(ctx context.Context, productID string, req *domain.GenerateVariantRequest) (*domain.ProductVariant, error) {
	if req == nil {
		return nil, errors.New("request payload cannot be nil")
	}

	// Check product exists
	_, err := uc.repo.GetProductByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	variant := &domain.ProductVariant{
		ProductID:    productID,
		Name:         req.Name,
		SKU:          req.SKU,
		SellPrice:    req.SellPrice,
		ComparePrice: req.ComparePrice,
		Weight:       req.Weight,
	}

	return uc.repo.CreateVariant(ctx, variant, req.OptionValueIDs)
}

func (uc *CatalogUsecase) UpdateVariant(ctx context.Context, id int, req *domain.UpdateVariantRequest) (*domain.ProductVariant, error) {
	if req == nil {
		return nil, errors.New("request payload cannot be nil")
	}
	return uc.repo.UpdateVariant(ctx, id, req.Name, req.SKU, req.SellPrice, req.ComparePrice, req.Weight)
}

func (uc *CatalogUsecase) DeleteVariant(ctx context.Context, id int) error {
	return uc.repo.DeleteVariant(ctx, id)
}
