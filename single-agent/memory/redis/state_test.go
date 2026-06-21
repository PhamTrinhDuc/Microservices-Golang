package redis

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/adk/session"
)

func setUpRedisState() redisState {
	return redisState{
		service: redisSvc,
		client:  redisSvc.client,
		key:     redisSvc.sessionKey(appName, userID, sessionID),
		ttl:     redisSvc.ttl,
		appName: appName,
		userID:  userID,
		data: map[string]any{
			"userName": userName,
			"userId":   userID,
		},
	}
}

func TestMethodsState(t *testing.T) {
	ctx := context.Background()
	redisState := setUpRedisState()

	t.Run(getName("check data state"), func(t *testing.T) {
		assert.Equal(t, redisState.data, map[string]any{
			"userName": userName,
			"userId":   userID,
		})
	})

	t.Run(getName("check app state"), func(t *testing.T) {
		keyApp := redisState.service.appStateKey(appName)

		// create request
		req := session.CreateRequest{
			AppName:   appName,
			UserID:    userID,
			SessionID: sessionID,
			State: map[string]any{
				session.KeyPrefixUser + "name": userName,
				session.KeyPrefixApp + "theme": theme,
				"session_id":                   sessionID,
			},
		}
		_, err := redisState.service.Create(ctx, &req)
		assert.NoError(t, err)

		// set state
		err = redisState.Set(
			"app:theme", "dark",
		)
		assert.NoError(t, err)

		// get state
		data, err := redisState.client.HGet(ctx, keyApp, "theme").Result()
		assert.NoError(t, err)

		assert.Equal(t, "\"dark\"", data)

		redisState.service.FlushDB(ctx)
	})
}
