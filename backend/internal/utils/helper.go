package utils

import (
	"backend/domain"
	"os"
	"strconv"
)

// GetEnvString retrieves env string with fallback.
func GetEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvInt retrieves env int with fallback.
func GetEnvInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetEnvBool retrieves env boolean with fallback.
func GetEnvBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetEnvFloat retrieves env float with fallback.
func GetEnvFloat(key string, defaultValue float64) float64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return defaultValue
	}
	return value
}

func CloneQuery(query *domain.ProductSearchQuery) *domain.ProductSearchQuery {
	return &domain.ProductSearchQuery{
		CategoryID:  query.CategoryID,
		BrandID:     query.BrandID,
		MinPrice:    query.MinPrice,
		MaxPrice:    query.MaxPrice,
		Query:       query.Query,
		Sort:        query.Sort,
		Page:        query.Page,
		Limit:       query.Limit,
		InStockOnly: query.InStockOnly,
		MinRating:   query.MinRating,
		SpecFilter:  query.SpecFilter,
	}
}
