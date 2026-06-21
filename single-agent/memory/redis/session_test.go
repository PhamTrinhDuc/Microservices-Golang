package redis

import (
	"context"
	"fmt"
	"single-agent/utils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/adk/model"
	"google.golang.org/adk/session"
)

const (
	theme = "dark"
	lang  = "vi"

	appName   = "app_001"
	userID    = "user_001"
	sessionID = "session_001"

	userName = "Jiyuu"
	account  = "Jiyuu@gmail.com"
)

func GetConfigRedis() RedisConfig {
	return RedisConfig{
		Host:         utils.GetEnvString("REDIS_HOST", "localhost"),
		Port:         utils.GetEnvInt("REDIS_PORT", 6379),
		Username:     utils.GetEnvString("REDIS_USERNAME", "jiyuu"),
		Password:     utils.GetEnvString("REDIS_PASSWORD", "a2amcpgo"),
		AppStateTTL:  1 * time.Second,
		UserStateTTL: 1 * time.Second,
		TTL:          15 * time.Second,
	}
}

func SetupRedis() *RedisService {
	cfg := GetConfigRedis()
	client, _ := NewRedisService(cfg)
	return client
}

func getName(name string) string {
	return fmt.Sprintf(" %d.🧪%s", countFuncTest, name)
}

var countFuncTest = 0
var redisSvc = SetupRedis()

func TestNewRedis(t *testing.T) {
	env := GetConfigRedis()
	countFuncTest += 1
	tests := []struct {
		name    string
		config  RedisConfig
		wantErr bool
	}{
		// TC0: Missing 1 số trường quan trọng
		{
			name:    getName("Missing host"),
			config:  RedisConfig{Port: 6379, DB: 5, Password: env.Password, Username: env.Username},
			wantErr: true,
		},
		{
			name:    getName("Invalid port"),
			config:  RedisConfig{Host: env.Host, Port: 0, Password: env.Password, Username: env.Username},
			wantErr: true,
		},
		{
			name:    getName("Invalid username"),
			config:  RedisConfig{Host: env.Host, Port: env.Port, Username: "", Password: env.Password},
			wantErr: true,
		},
		{
			name:    getName("Invalid password"),
			config:  RedisConfig{Host: env.Host, Port: env.Port, Username: env.Username, Password: ""},
			wantErr: true,
		},
		{
			name:    getName("Connect success"),
			config:  RedisConfig{Host: env.Host, Port: env.Port, Username: env.Username, Password: env.Password},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewRedisService(tc.config)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateKeys(t *testing.T) {
	countFuncTest += 1
	tests := []struct {
		name     string
		expected string
		call     func() string
	}{
		{
			name:     getName("Create app key"),
			expected: fmt.Sprintf("appstate:%s", appName),
			call:     func() string { return redisSvc.appStateKey(appName) },
		},
		{
			name:     getName("Create user key"),
			expected: fmt.Sprintf("users:%s:%s", appName, userID),
			call:     func() string { return redisSvc.userStateKey(appName, userID) },
		},
		{
			name:     getName("Create session key"),
			expected: fmt.Sprintf("session:%s:%s:%s", appName, userID, sessionID),
			call:     func() string { return redisSvc.sessionKey(appName, userID, sessionID) },
		},
		{
			name:     getName("Create session list key"),
			expected: fmt.Sprintf("sessions:%s:%s", appName, userID),
			call:     func() string { return redisSvc.sessionKeys(appName, userID) },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.call())
		})
	}
}

func TestMergeState(t *testing.T) {
	countFuncTest += 1
	appState := map[string]any{
		"theme": "dark",
		"lang":  "vi",
	}

	userState := map[string]any{
		"user_id": userID,
		"name":    userName,
		"account": account,
	}

	sessionState := map[string]any{
		"sesion_id": sessionID,
	}

	mergedStateExpected := map[string]any{
		"app:theme":    theme,
		"app:lang":     lang,
		"user:user_id": userID,
		"user:name":    userName,
		"user:account": account,
		"sesion_id":    sessionID,
	}
	redis := SetupRedis()
	t.Run(getName("Merge state"), func(t *testing.T) {
		mergedStateActual := redis.mergeState(appState, userState, sessionState)
		assert.Equal(t, mergedStateExpected, mergedStateActual)
	})
}

func TestUpdateState(t *testing.T) {
	countFuncTest += 1
	ctx := context.Background()
	redis := SetupRedis()
	tests := []struct {
		name     string
		expected map[string]any
		call     func() map[string]any
	}{
		{
			name: getName("Update user state"),
			expected: map[string]any{
				"user_id": userID,
				"name":    userName,
				"account": account,
			},
			call: func() map[string]any {
				key := redis.userStateKey(appName, userID)

				redis.client.HSet(
					ctx, key,
					map[string]any{
						"user_id": userID,
						"name":    "Eren",
						"account": "Eren@gmail.com",
					},
				)

				result, err := redis.updateUserState(
					ctx, appName, userID,
					map[string]any{
						"user_id": userID,
						"name":    userName,
						"account": account,
					},
				)
				assert.NoError(t, err)
				return result
			},
		},

		{
			name: getName("Update app state"),
			expected: map[string]any{
				"theme": theme,
				"lang":  lang,
			},
			call: func() map[string]any {
				key := redis.appStateKey(appName)
				redis.client.HSet(ctx, key,
					map[string]any{
						"theme": "light",
						"lang":  "en",
					})

				result, err := redis.updateAppState(
					ctx, appName,
					map[string]any{
						"theme": theme,
						"lang":  lang,
					},
				)
				assert.NoError(t, err)
				return result
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.call())
			redisSvc.FlushDB(ctx)
		})
	}
}

func TestCreateAndGetSession(t *testing.T) {
	// Prefix user: => user delta
	// Prefix app: => app delta
	// Prefix tmp: => pass through
	// No Prefix => session delta
	countFuncTest += 1
	ctx := context.Background()
	tests := []struct {
		name     string
		request  session.CreateRequest
		expected map[string]any
	}{
		{
			name: getName("Missing state"),
			request: session.CreateRequest{
				AppName: appName,
				UserID:  userID,
				State:   nil,
			},
			expected: map[string]any{},
		},
		{
			name: getName("Missing prefix data"),
			request: session.CreateRequest{
				AppName: appName,
				UserID:  userID,
				State:   map[string]any{"session_id": sessionID},
			},
			expected: map[string]any{"session_id": sessionID},
		},
		{
			name: getName("Perfect data"),
			request: session.CreateRequest{
				AppName: appName,
				UserID:  userID,
				State: map[string]any{
					session.KeyPrefixUser + "user_id": userID,
					session.KeyPrefixUser + "name":    userName,
					session.KeyPrefixUser + "email":   account,
					session.KeyPrefixApp + "theme":    theme,
					session.KeyPrefixApp + "lang":     lang,
					"sesion_id":                       sessionID,
				},
			},
			expected: map[string]any{
				"user:user_id": userID,
				"user:name":    userName,
				"user:email":   account,
				"app:theme":    theme,
				"app:lang":     lang,
				"sesion_id":    sessionID,
			},
		},
		{
			name: getName("Hybrid prefix data and no prefix data"),
			request: session.CreateRequest{
				AppName: appName,
				UserID:  userID,
				State: map[string]any{
					session.KeyPrefixUser + "user_id": userID,
					session.KeyPrefixUser + "name":    userName,
					session.KeyPrefixApp + "theme":    theme,
					session.KeyPrefixApp + "lang":     lang,
					"session_id":                      sessionID,
				},
			},
			expected: map[string]any{
				"user:user_id": userID,
				"user:name":    userName,
				"app:theme":    theme,
				"app:lang":     lang,
				"session_id":   sessionID,
			},
		},
	}
	for _, tc := range tests {
		request := tc.request

		creResp, err := redisSvc.Create(ctx, &request)
		assert.NoError(t, err)
		if err != nil {
			continue
		}

		getResp, err := redisSvc.Get(ctx, &session.GetRequest{
			AppName:   request.AppName,
			UserID:    request.UserID,
			SessionID: creResp.Session.ID(),
		})
		assert.NoError(t, err)

		t.Run(tc.name, func(t *testing.T) {
			assert.NoError(t, err)
			assert.NotNil(t, creResp)
			assert.Equal(t, appName, creResp.Session.AppName())

			// Convert session.State interface to *redisState struct to access Data()
			actualState := getResp.Session.State().(*redisState).Data()
			assert.Equal(t, tc.expected, actualState)
		})
		redisSvc.FlushDB(ctx) // Clean state for each test case
	}
}

func TestListSession(t *testing.T) {
	countFuncTest += 1
	ctx := context.Background()

	tests := []struct {
		name     string
		reqs     []session.CreateRequest
		expected []string
	}{
		{
			name: getName("One member"),
			reqs: []session.CreateRequest{
				{
					AppName:   appName,
					UserID:    userID,
					SessionID: sessionID,
					State: map[string]any{
						session.KeyPrefixUser + "user_id": userID,
						session.KeyPrefixUser + "email":   account,
						session.KeyPrefixApp + "lang":     lang,
					},
				},
			},
			expected: []string{sessionID},
		},
		{
			name: getName("Two member"),
			reqs: []session.CreateRequest{
				{
					AppName:   appName,
					UserID:    userID,
					SessionID: sessionID,
					State: map[string]any{
						session.KeyPrefixUser + "user_id": userID,
						session.KeyPrefixUser + "email":   account,
						session.KeyPrefixApp + "lang":     lang,
					},
				},
				{
					AppName:   appName,
					UserID:    userID,
					SessionID: sessionID + "2",
					State: map[string]any{
						session.KeyPrefixUser + "user_id": userID,
						session.KeyPrefixUser + "email":   account,
						session.KeyPrefixApp + "lang":     lang,
					},
				},
			},
			expected: []string{sessionID, sessionID + "2"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for i := range tc.reqs {
				_, err := redisSvc.Create(ctx, &tc.reqs[i])
				assert.NoError(t, err)
			}

			resp, err := redisSvc.List(ctx, &session.ListRequest{
				AppName: appName,
				UserID:  userID,
			})
			assert.NoError(t, err)

			var actualIDs []string
			for _, s := range resp.Sessions {
				actualIDs = append(actualIDs, s.ID())
			}
			assert.ElementsMatch(t, tc.expected, actualIDs)
		})
		redisSvc.FlushDB(ctx)
	}
}

func TestDeleteSession(t *testing.T) {
	ctx := context.Background()

	reqCreate := session.CreateRequest{
		AppName:   appName,
		UserID:    userID,
		SessionID: sessionID,
		State: map[string]any{
			session.KeyPrefixUser + "user_id": userID,
			session.KeyPrefixUser + "email":   account,
			session.KeyPrefixApp + "lang":     lang,
		},
	}

	reqDel := session.DeleteRequest{
		AppName:   appName,
		UserID:    userID,
		SessionID: sessionID,
	}

	reqGet := session.GetRequest{
		AppName:   appName,
		UserID:    userID,
		SessionID: sessionID,
	}
	t.Run(getName("delete session"), func(t *testing.T) {
		_, err := redisSvc.Create(ctx, &reqCreate)
		assert.NoError(t, err)

		err = redisSvc.Delete(ctx, &reqDel)
		assert.NoError(t, err)

		resp, err := redisSvc.Get(ctx, &reqGet)
		assert.Nil(t, resp)
	})
	redisSvc.FlushDB(ctx)
}

func TestAddEvents(t *testing.T) {
	ctx := context.Background()

	// 1. Create session
	sess, err := redisSvc.Create(ctx, &session.CreateRequest{
		AppName:   appName,
		UserID:    userID,
		SessionID: sessionID,
		State: map[string]any{
			"user:user_name": userName,
			"user:email":     account,
			"last_updated":   "2026-13-5",
		},
	})
	assert.NoError(t, err)

	// 2. Add event
	err = redisSvc.AppendEvent(ctx, sess.Session, &session.Event{
		ID:     redisSvc.eventsKey(appName, userID, sessionID),
		Author: "user",
		Actions: session.EventActions{
			StateDelta: map[string]any{"temp:count": 1, "assistant": "I can..."},
		},
	})
	assert.NoError(t, err)

	gotSess, err := redisSvc.Get(ctx, &session.GetRequest{
		AppName:   appName,
		UserID:    userID,
		SessionID: sessionID,
	})

	t.Run(getName("test add events"), func(t *testing.T) {
		assert.Equal(t, gotSess.Session.State().(*redisState).data, map[string]any{
			"user:user_name": userName,
			"user:email":     account,
			"last_updated":   "2026-13-5",
			"assistant":      "I can...",
		})
	})

	redisSvc.FlushDB(ctx)
}

func TestAddEventsWithPartial(t *testing.T) {
	ctx := context.Background()

	// 1. Create session
	sess, err := redisSvc.Create(ctx, &session.CreateRequest{
		AppName:   appName,
		UserID:    userID,
		SessionID: sessionID,
		State: map[string]any{
			"user:user_name": userName,
			"user:email":     account,
			"last_updated":   "2026-13-5",
		},
	})
	assert.NoError(t, err)

	// 2. Add event
	err = redisSvc.AppendEvent(ctx, sess.Session, &session.Event{
		ID:          redisSvc.eventsKey(appName, userID, sessionID),
		Author:      "user",
		LLMResponse: model.LLMResponse{Partial: true},
		Actions: session.EventActions{
			StateDelta: map[string]any{"temp:count": 1, "assistant": "I can..."},
		},
	})
	assert.NoError(t, err)

	gotSess, err := redisSvc.Get(ctx, &session.GetRequest{
		AppName:   appName,
		UserID:    userID,
		SessionID: sessionID,
	})
	t.Run(getName("test add event with partial"), func(t *testing.T) {
		assert.Equal(t, gotSess.Session.State().(*redisState).data, map[string]any{
			"user:user_name": userName,
			"user:email":     account,
			"last_updated":   "2026-13-5",
		})
		redisSvc.FlushDB(ctx)
	})
}
