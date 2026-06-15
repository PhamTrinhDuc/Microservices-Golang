package database

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *DB {
	db, err := NewDB(context.Background(), NewDBConfig())
	require.NoError(t, err, "Failed to connect to test database")

	// model, err := ollama.NewClient(getTestModelConfig())
	// require.NoError(t, err, "failed to to setup Ollama model")
	return db
}

func TestVectorSearch_SkipsNullEmbeddings(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create query embedding
	queryEmbedding := make([]float32, 1535)
	for i := range queryEmbedding {
		queryEmbedding[i] = float32(i) * 0.001
	}

	// Vector search should only return documents with embeddings
	results, err := db.VectorSearch(ctx, queryEmbedding, 5)
	require.NoError(t, err, "Vector search should not fail")

	// All returned documents should have embeddings
	for _, result := range results {
		assert.NotNil(t, result.Document.Embedding, "Vector search should only return docs with embeddings")
	}
}

func TestHybridSearch_HandlesNullEmbeddings(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create query embedding
	queryEmbedding := make([]float32, 1024)
	for i := range queryEmbedding {
		queryEmbedding[i] = 0.1
	}

	params := HybridSearchParams{
		Query:        "security policy",
		Embedding:    queryEmbedding,
		Limit:        10,
		BM25Weight:   0.5,
		VectorWeight: 0.5,
		MinBM25Score: 0.0,
		MinVectorSim: 0.0,
	}

	// Hybrid search should handle documents without embeddings gracefully
	results, err := db.HybridSearch(ctx, params)
	require.NoError(t, err, "Hybrid search should not fail with NULL embeddings")
	assert.NotNil(t, results, "Results should not be nil")
	assert.Condition(t, func() (success bool) {
		return len(results) > 0
	}, "Hybrid search results should still appear")
}

func TestConcurrentRetrievals(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	var wg sync.WaitGroup

	// Get sample documents
	docs, err := db.ListDocuments(ctx, 20, 0)
	require.NoError(t, err)
	require.NotEmpty(t, docs)

	// Perform concurrent retrievals
	numWorkers := 10
	sem := make(chan struct{}, numWorkers)

	for i := range len(docs) {
		wg.Add(1) // Đếm thêm 1 việc cần làm
		// Đợi lấy "vé" từ semaphore, nếu đủ 10 người rồi thì sẽ kẹt lại ở đây
		sem <- struct{}{}

		go func(workerID int) {
			defer wg.Done() // Đảm bảo dù lỗi hay không cũng sẽ trừ máy đếm khi xong

			defer func() {
				<-sem
			}() // Xong việc thì trả lại vé cho người sau. Không để ngoài func vì phải chờ for xong mới thực thi => không ai trả vé

			retrieved, err := db.GetDocument(ctx, docs[i].ID)
			// t.Logf("received ID %s for document", retrieved.ID)
			if err != nil {
				t.Errorf("Worker %d failed to retrieve document: %v", workerID, err)
				return
			}
			if retrieved == nil {
				t.Errorf("Worker %d got nil document", workerID)
				return
			}
		}(i)
	}
	// 4. Đợi tất cả hoàn thành
	wg.Wait()
	t.Log("✓ All concurrent retrievals completed successfully")
}
