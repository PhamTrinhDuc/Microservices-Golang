package gemini

import (
	"context"
	"fmt"
	"single-agent/utils"

	"google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/genai"
)

// NewGeminiLLM khởi tạo LLM từ Google Gemini
// Hỗ trợ các model: gemini-1.5-flash, gemini-2.5-flash
func NewGeminiLLM(ctx context.Context, modelName string) (model.LLM, error) {
	// Lấy Google API Key từ biến môi trường
	apiKey := utils.GetEnvString("GEMINI_API_KEY", "")

	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is not set")
	}

	model, err := gemini.NewModel(ctx, modelName, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, err
	}

	return model, nil
}
