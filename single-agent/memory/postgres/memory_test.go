package postgres

import (
	"context"
	"iter"
	"single-agent/utils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/adk/memory"
	"google.golang.org/adk/model"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

const (
	sessionID = "session_001"
	appName   = "multi_agent"
	userID    = "user_001"
)

func GetConfigPGMem() PostgresMemoryServiceConfig {
	return PostgresMemoryServiceConfig{
		ConnString: utils.GetEnvString("POSTGRES_URL", "postgres://jiyuu_user:jiyuu_password@localhost:5433/ecommerce_db"),
		EmbeddingModel: NewOpenAICompatibleEmbedding(
			OpenAICompatibleEmbeddingConfig{
				// BaseURL:   utils.GetEnvString("BASE_URL_OLLAMA", "http://localhost:11434/v1"),
				// Model:     utils.GetEnvString("EMBEDDING_MODEL", "qwen3-embedding:0.6b"),
				BaseURL:   utils.GetEnvString("OPENAI_BASE_URL", "https://api.openai.com/v1"),
				Model:     utils.GetEnvString("OPENAI_EMBEDDING_MODEL", "text-embedding-3-large"),
				Dimension: utils.GetEnvInt("OPENAI_EMBEDDING_DIM", 675),
				APIKey:    utils.GetEnvString("OPENAI_API_KEY", ""),
			},
		),
		TopKBM25:     50,
		TopKVector:   50,
		TopKHybrid:   5,
		WeightBM25:   0.5,
		WeightVector: 0.5,
	}
}

var pgSvc, _ = NewPostgresMemoryService(
	context.Background(),
	GetConfigPGMem(),
)

// mockSession implements session.Session for testing
type mockSession struct {
	id      string
	appName string
	userID  string
	events  *mockEvents
}

func (s *mockSession) ID() string                { return s.id }
func (s *mockSession) AppName() string           { return s.appName }
func (s *mockSession) UserID() string            { return s.userID }
func (s *mockSession) State() session.State      { return nil }
func (s *mockSession) Events() session.Events    { return s.events }
func (s *mockSession) LastUpdateTime() time.Time { return time.Now() }

type mockEvents struct {
	events []*session.Event
}

func (e *mockEvents) All() iter.Seq[*session.Event] {
	return func(yield func(*session.Event) bool) {
		for _, evt := range e.events {
			if !yield(evt) {
				return
			}
		}
	}
}

func (e *mockEvents) Len() int {
	return len(e.events)
}

func (e *mockEvents) At(i int) *session.Event {
	if i < 0 || i >= len(e.events) {
		return nil
	}
	return e.events[i]
}

func createTestSession(id, appName, userID string, messages []struct{ author, text string }) *mockSession {
	var events []*session.Event
	for i, msg := range messages {
		events = append(events, &session.Event{
			ID:        id + "-" + string(rune('a'+i)),
			Author:    msg.author,
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
			LLMResponse: model.LLMResponse{
				Content: &genai.Content{
					Parts: []*genai.Part{genai.NewPartFromText(msg.text)},
					Role:  msg.author,
				},
			},
		})
	}
	return &mockSession{
		id:      id,
		appName: appName,
		userID:  userID,
		events:  &mockEvents{events: events},
	}
}

func clearTable(ctx context.Context, t *testing.T) {
	_, err := pgSvc.pool.Exec(ctx, "TRUNCATE TABLE memory_entries RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("Failed to clear table: %v", err)
	}
}

// Test AddSession perform testing insert session to db and verify it
func TestAddSession(t *testing.T) {
	// add to database
	ctx := context.Background()
	sess := createTestSession(
		sessionID, appName, userID,
		[]struct{ author, text string }{
			{"user", "Hello"},
			{"assistant", "How can I help you"},
		},
	)

	err := pgSvc.AddSessionToMemory(ctx, sess)
	assert.NoError(t, err)

	// verify inserted
	rows, err := pgSvc.pool.Query(ctx,
		"SELECT content, author, timestamp FROM memory_entries WHERE app_name=$1 AND user_id=$2 AND session_id=$3",
		appName, userID, sessionID,
	)

	assert.NoError(t, err)
	memories, err := scanMemories(rows)

	assert.NoError(t, err)
	assert.Equal(t, len(memories), 2)

	defer clearTable(ctx, t)
}

func TestSearchMemory(t *testing.T) {
	// add to database
	ctx := context.Background()
	sess := createTestSession(
		sessionID, appName, userID,
		[]struct{ author, text string }{
			{"user", "Hello"},
			{"assistant", "How can I help you"},
		},
	)

	err := pgSvc.AddSessionToMemory(ctx, sess)
	assert.NoError(t, err)

	// search memory
	searchReq := memory.SearchRequest{
		Query:   "Hello",
		AppName: appName,
		UserID:  userID,
	}
	t.Run("search by text", func(t *testing.T) {
		searchResp, err := pgSvc.SearchByText(ctx, &searchReq, 5)
		// for _, entry := range searchResp {
		// 	role := entry.Content.Role
		// 	fmt.Println(role)
		// 	for _, part := range entry.Content.Parts {
		// 		fmt.Println(part.Text)
		// 	}
		// }
		assert.NoError(t, err)
		assert.Condition(t, func() (success bool) {
			return len(searchResp) > 0
		}, "Result must be > 0")
	})

	t.Run("search by vector", func(t *testing.T) {
		embdding, err := pgSvc.embeddingModel.Embed(ctx, searchReq.Query)
		assert.NoError(t, err)

		searchResp, err := pgSvc.SearchByVector(ctx, &searchReq, embdding, 1)
		assert.NoError(t, err)

		// for _, entry := range searchResp {
		// 	role := entry.Content.Role
		// 	fmt.Println(role)
		// 	for _, part := range entry.Content.Parts {
		// 		fmt.Println(part.Text)
		// 	}
		// }
		assert.NoError(t, err)
		assert.Condition(t, func() (success bool) {
			return len(searchResp) > 0
		}, "Result must be > 0")
	})

	t.Run("hybrid search", func(t *testing.T) {
		// Dùng SearchMemory (wrapper đã có sẵn default values)
		searchResp, err := pgSvc.SearchMemory(ctx, &searchReq)
		assert.NoError(t, err)

		// Hoặc nếu muốn gọi trực tiếp HybridSearch để test với thông số tùy chỉnh:
		// embedding, _ := pgSvc.embeddingModel.Embed(ctx, searchReq.Query)
		// searchResp, err := pgSvc.HybridSearch(ctx, &searchReq, embedding, 50, 50, 5, 0.5, 0.5)

		assert.NoError(t, err)
		assert.Condition(t, func() (success bool) {
			return len(searchResp.Memories) > 0
		}, "Result must be > 0")
	})

	defer clearTable(ctx, t)
}

func TestUpdateMemory(t *testing.T) {
	// add to database
	ctx := context.Background()
	sess := createTestSession(
		sessionID, appName, userID,
		[]struct{ author, text string }{
			{"user", "Hello"},
			{"assistant", "How can I help you"},
		},
	)

	err := pgSvc.AddSessionToMemory(ctx, sess)
	assert.NoError(t, err)

	newContent := "Hello. How can I help you?"
	err = pgSvc.UpdateMemory(ctx, appName, userID, 2, newContent)
	assert.NoError(t, err)

	entries, err := pgSvc.SearchWithID(ctx, &memory.SearchRequest{UserID: userID, AppName: appName, Query: ""})
	assert.Condition(t, func() (success bool) {
		for _, entry := range entries {
			for _, part := range entry.Content.Parts {
				if part.Text == newContent {
					return true
				}
			}
		}
		return false
	}, "Content no updated")

	defer clearTable(ctx, t)
}

func TestDeleteMemory(t *testing.T) {
	// add to database
	ctx := context.Background()
	sess := createTestSession(
		sessionID, appName, userID,
		[]struct{ author, text string }{
			{"user", "Hello"},
			{"assistant", "How can I help you"},
		},
	)

	err := pgSvc.AddSessionToMemory(ctx, sess)
	assert.NoError(t, err)

	err = pgSvc.DeleteMemory(ctx, appName, userID, 1)
	assert.NoError(t, err)

	clearTable(ctx, t)
}
