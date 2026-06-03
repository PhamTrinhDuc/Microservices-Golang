package utils

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// GetEnvString lấy giá trị string từ biến môi trường, nếu không có thì dùng mặc định
func GetEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvInt lấy giá trị int từ biến môi trường
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

// GetEnvBool lấy giá trị boolean từ biến môi trường (hỗ trợ 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False)
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

func GetEnvFloat(key string, defaultValue float64) float64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseFloat(valueStr, 32)
	if err != nil {
		return defaultValue
	}
	return float64(value)
}

// Sửa GetListFiles để trả về danh sách file []string
func GetListFiles(data_path string) ([]string, error) {
	var data_paths []string

	// Callback chỉ được phép trả về duy nhất 1 giá trị là error
	err := filepath.WalkDir(data_path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.ToLower(filepath.Ext(path)) == ".md" {
			fmt.Printf("Đang đọc file: %s\n", path)
			data_paths = append(data_paths, path)
		}

		return nil // Trả về nil nếu thành công
	})

	if err != nil {
		return nil, err
	}

	return data_paths, nil
}
