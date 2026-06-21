package gemini

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

func Test_Response(t *testing.T) {
	modelName := "gemini-2.5-flash" // Đổi về bản stable 2.0
	req := &model.LLMRequest{
		Contents: []*genai.Content{
			genai.NewContentFromText("What is the capital of France? One word.", genai.RoleUser),
		},
	}

	ctx := context.Background()
	llmModel, err := NewGeminiLLM(ctx, modelName)
	require.NoError(t, err)

	fmt.Println("--- Sending Request to Gemini ---")

	// GenerateContent trả về một iterator, ta cần lặp để lấy kết quả (tham số thứ 3 là stream bool)
	for res, err := range llmModel.GenerateContent(ctx, req, false) {
		if err != nil {
			t.Fatalf("Error during generation: %v", err)
		}

		if res != nil && res.Content != nil {
			for _, part := range res.Content.Parts {
				fmt.Print(part.Text)
			}
		}
	}
	fmt.Println("\n--- Generation Finished ---")
}
