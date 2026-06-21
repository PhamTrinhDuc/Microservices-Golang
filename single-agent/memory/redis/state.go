package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/adk/session"
)

// redisState implements session.State with Redis persistence.
// It holds the merged (all tiers) state in memory and routes writes to the
// correct Redis key based on the key prefix.
type redisState struct {
	data    map[string]any
	client  *redis.Client
	key     string
	ttl     time.Duration
	service *RedisService
	appName string
	userID  string
}

func newRedisState(initial map[string]any, client *redis.Client, key string, ttl time.Duration, service *RedisService, appName, userID string) *redisState {
	data := make(map[string]any)
	for k, v := range initial {
		data[k] = v
	}
	return &redisState{
		data:    data,
		client:  client,
		key:     key,
		ttl:     ttl,
		service: service,
		appName: appName,
		userID:  userID,
	}
}

func (s *redisState) Data() map[string]any {
	return s.data
}

func (s *redisState) Get(key string) (any, error) {
	v, ok := s.data[key]
	if ok {
		return v, nil
	}
	return nil, session.ErrStateKeyNotExist
}

func (s *redisState) Set(key string, value any) error {
	ctx := context.Background()
	if cleaned, found := strings.CutPrefix(key, session.KeyPrefixApp); found {
		appKey := s.service.appStateKey(s.appName)
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal data: %w", err)
		}
		s.client.HSet(ctx, appKey, cleaned, string(data))
		if s.service.appStateTTL > 0 {
			s.client.Expire(ctx, appKey, s.service.appStateTTL)
		} else {
			s.client.Persist(ctx, appKey)
		}
	}

	if cleaned, found := strings.CutPrefix(key, session.KeyPrefixUser); found {
		userKey := s.service.userStateKey(s.appName, s.userID)

		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal data: %w", err)
		}
		s.client.HSet(ctx, userKey, cleaned, string(data))

		if s.service.userStateTTL > 0 {
			s.client.Expire(ctx, userKey, s.service.userStateTTL)
		} else {
			s.client.Persist(ctx, userKey)
		}
	}
	if strings.HasPrefix(key, session.KeyPrefixTemp) {
		return nil
	}
	return s.UpdateSessionState()
}

func (s *redisState) UpdateSessionState() error {
	ctx := context.Background()
	data, err := s.client.Get(ctx, s.key).Result()
	if err != nil {
		return fmt.Errorf("failed to get session for state update: %w", err)
	}

	var storable storableSession
	if err := json.Unmarshal([]byte(data), &storable); err != nil {
		return fmt.Errorf("failed to unmarshal data for storable: %w", err)
	}

	for k, v := range s.data {
		if !strings.Contains(k, session.KeyPrefixApp) || !strings.Contains(k, session.KeyPrefixUser) || !strings.Contains(k, session.KeyPrefixTemp) {
			storable.State[k] = v
		}
	}
	storable.LastUpdateTime = time.Now()

	dataUpdated, err := json.Marshal(storable)
	if err != nil {
		return fmt.Errorf("failed to marshal data updated: %w", err)
	}
	if err := s.client.Set(ctx, s.key, dataUpdated, s.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set data updated: %w", err)
	}
	return nil
}

func (s *redisState) All() iter.Seq2[string, any] {
	return func(yield func(string, any) bool) {
		for k, v := range s.data {
			if !yield(k, v) {
				return
			}
		}
	}
}

var _ session.State = (*redisState)(nil)
