package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"log"

	"github.com/redis/go-redis/v9"
	"google.golang.org/adk/session"
)

// redisEvents implements session.Events with live Redis reads.
// When filtered is true, the cached slice is the authoritative source (e.g.
// after Get applied NumRecentEvents / After filters) and loadFromRedis returns
// it directly without re-fetching.

type redisEvents struct {
	client   *redis.Client
	key      string
	cached   []*session.Event
	isFilter bool
}

func newRedisEvents(events []*session.Event, key string, client *redis.Client) *redisEvents {
	if events == nil {
		events = make([]*session.Event, 0)
	}
	return &redisEvents{
		client: client,
		key:    key,
		cached: events,
	}
}

func newRedisEventsWithFilter(events []*session.Event, key string, client *redis.Client) *redisEvents {
	if events == nil {
		events = make([]*session.Event, 0)
	}
	return &redisEvents{
		client:   client,
		key:      key,
		cached:   events,
		isFilter: true,
	}
}

// loadFromRedis load events from redis based on key
// If client is nil or key is empty, return cached events
func (e *redisEvents) loadFromRedis() []*session.Event {
	if e.client == nil || e.key == "" {
		return e.cached
	}

	eventsData, err := e.client.LRange(context.Background(), e.key, 0, -1).Result()
	// fmt.Println("debug", eventsData)
	if err != nil {
		log.Printf("failed to load events from redis: %v", err)
		return e.cached
	}

	var events []*session.Event
	for _, ed := range eventsData {
		var evt session.Event
		if err := json.Unmarshal([]byte(ed), &evt); err != nil {
			log.Printf("failed to unmarshal event data: %v", err)
			return e.cached
		}
		events = append(events, &evt)
	}
	fmt.Println("event", events)
	return events
}

// All return all events
func (s *redisEvents) All() iter.Seq[*session.Event] {
	events := s.loadFromRedis()
	return func(yield func(*session.Event) bool) {
		for _, evt := range events {
			if !yield(evt) {
				return
			}
		}
	}
}

// Len return the number of events
func (s *redisEvents) Len() int {
	return len(s.loadFromRedis())
}

// At return the event at index i
func (s *redisEvents) At(i int) *session.Event {
	events := s.loadFromRedis()
	return events[i]
}

var _ session.Events = (*redisEvents)(nil)
