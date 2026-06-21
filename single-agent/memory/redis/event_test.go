package redis

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/adk/session"
)

func setUpEvents() *redisEvents {
	return newRedisEvents(
		nil,
		redisSvc.eventsKey(appName, userID, sessionID),
		redisSvc.client,
	)
}

func TestMethodsEvent(t *testing.T) {
	ctx := context.Background()
	redisEvt := setUpEvents()

	// 1. Create session
	reqCreate := session.CreateRequest{
		AppName:   appName,
		UserID:    userID,
		SessionID: sessionID,
		State:     map[string]any{"session_id": sessionID},
	}
	sess, err := redisSvc.Create(ctx, &reqCreate)
	assert.NoError(t, err)

	// 2. Create event
	evt := session.Event{
		ID:     redisSvc.eventsKey(appName, userID, sessionID),
		Author: "user",
		Actions: session.EventActions{
			StateDelta: map[string]any{"temp:count": 1, "user:username": userName},
		},
	}

	err = redisSvc.AppendEvent(ctx, sess.Session, &evt)
	assert.NoError(t, err)

	t.Run(getName("check len event"), func(t *testing.T) {
		assert.Equal(t, redisEvt.Len(), 1) // event have delta: user:user_name
	})

	t.Run(getName("check at"), func(t *testing.T) {
		assert.Equal(t, redisEvt.At(0).Actions.StateDelta, map[string]any{"user:username": "Jiyuu"})
	})

	redisSvc.FlushDB(ctx)
}
