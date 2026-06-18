package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/adk/session"
)

// RedisService: implement session.Service adk using Redis backend
type RedisService struct {
	client       *redis.Client
	ttl          time.Duration
	appStateTTL  time.Duration
	userStateTTL time.Duration
	mu           sync.Mutex
}

// redisSession implements session.Session.
type redisSession struct {
	id             string // ID Session
	appName        string
	userID         string
	state          *redisState
	events         *redisEvents
	lastUpdateTime time.Time
}

func (s *redisSession) ID() string                { return s.id }
func (s *redisSession) AppName() string           { return s.appName }
func (s *redisSession) UserID() string            { return s.userID }
func (s *redisSession) State() session.State      { return s.state }
func (s *redisSession) Events() session.Events    { return s.events }
func (s *redisSession) LastUpdateTime() time.Time { return s.lastUpdateTime }

// storableSession is the JSON-serializable representation of a session.
// State only contains session-scoped keys (no app: or user: prefixed keys).
type storableSession struct {
	ID             string         `json:"id"`
	AppName        string         `json:"app_name"`
	UserID         string         `json:"user_id"`
	State          map[string]any `json:"state"`
	LastUpdateTime time.Time      `json:"last_update_time"`
}

// RedisSessionServiceConfig holds configuration for RedisSessionService.
type RedisConfig struct {
	Host     string
	Port     int
	Username string
	// Password for Redis authentication (optional)
	Password string
	// DB is the Redis database number
	DB int
	// TTL is the session expiration time (default: 24 hours)
	TTL time.Duration
	// Timeout is the time connect to Redis
	Timeout time.Duration
	// AppStateTTL is the expiration time for app-scoped state.
	// Defaults to 0 (no expiration), matching the canonical ADK behaviour
	// where app state outlives individual sessions.
	AppStateTTL time.Duration
	// UserStateTTL is the expiration time for user-scoped state.
	// Defaults to 0 (no expiration), matching the canonical ADK behaviour
	// where user state outlives individual sessions.
	UserStateTTL time.Duration
}

func NewRedisService(cfg RedisConfig) (*RedisService, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("Missing host field")
	}
	if cfg.Password == "" {
		return nil, fmt.Errorf("Missing password field")
	}
	if cfg.Port <= 0 {
		return nil, fmt.Errorf("Missing port field")
	}
	if cfg.Username == "" {
		return nil, fmt.Errorf("Missing username field")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second // 30s timeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	if cfg.TTL == 0 {
		cfg.TTL = 24 * time.Hour
	}

	return &RedisService{
		client:       client,
		ttl:          cfg.TTL,
		appStateTTL:  cfg.AppStateTTL,
		userStateTTL: cfg.UserStateTTL,
		mu:           sync.Mutex{},
	}, nil
}

// Close close connect to Redis
func (s *RedisService) Close() error {
	return s.client.Close()
}

func (s *RedisService) FlushDB(ctx context.Context) error {
	return s.client.FlushDB(ctx).Err()
}

// appStateKey return specific app key
func (s *RedisService) appStateKey(appName string) string {
	return fmt.Sprintf("appstate:%s", appName)
}

// userStateKey return specific user key
func (s *RedisService) userStateKey(appName string, userID string) string {
	return fmt.Sprintf("users:%s:%s", appName, userID)
}

// sesionStateKey return key specific session key
func (s *RedisService) sessionKey(appName string, userID string, sessionKey string) string {
	return fmt.Sprintf("session:%s:%s:%s", appName, userID, sessionKey)
}

// sessionKeys return list session keys
func (s *RedisService) sessionKeys(appName string, userID string) string {
	return fmt.Sprintf("sessions:%s:%s", appName, userID)
}

// eventKey return key of event (same session key but diff prefix)
func (s *RedisService) eventsKey(appName, userID, sessionID string) string {
	return fmt.Sprintf("events:%s:%s:%s", appName, userID, sessionID)
}

// Create creates a new session. It returns an error if a session with the
// same ID already exists, matching the canonical ADK behaviour.
func (s *RedisService) Create(ctx context.Context, req *session.CreateRequest) (*session.CreateResponse, error) {
	// Hàm tạo 1 session mới, nếu giống CURD thông thường thì chỉ return redisSession là xong.
	// Nhưng tại sao lại phải có những chỗ updateAppState và updateAppUser?
	// Xét TH sau với hàm updateAppUser: khách hàng ấn vào 1 sản phẩm giày nike trên web và chọn chat ngay:
	// Lúc này việc update có nhiệm vụ cập nhật các state như: {"user:last_viewed_product": "Giày Nike", "current_page": "Promotion"}
	// Các thông tin này sẽ có ích cho AI khi chat

	s.mu.Lock() // lock avoid race condition on 1 instance
	defer s.mu.Unlock()

	// Coding start:
	// 1. check exists session key
	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	sessKey := s.sessionKey(req.AppName, req.UserID, sessionID) // key of specific session
	eventsKey := s.eventsKey(req.AppName, req.UserID, sessionID)
	if exists, err := s.client.Exists(ctx, sessKey).Result(); err != nil || exists > 0 {
		if err != nil {
			return nil, fmt.Errorf("failed to check session existence: %w", err)
		}
		return nil, fmt.Errorf("session %s already exists", sessionID)
	}
	// 2. Update appState and UserState
	appDelta, userDelta, sessionDelta := extractStateDeltas(req.State)
	appState, err := s.updateAppState(ctx, req.AppName, appDelta)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	userState, err := s.updateUserState(ctx, req.AppName, req.UserID, userDelta)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	// 3. Merge appDelta, userDelta, sessionDelta
	mergedState := s.mergeState(appState, userState, sessionDelta)

	// 4. Init redis Session implemented session.Session
	redisSession := &redisSession{
		id:             sessionID,
		appName:        req.AppName,
		userID:         req.UserID,
		events:         newRedisEvents(nil, eventsKey, s.client),
		state:          newRedisState(mergedState, s.client, sessKey, s.ttl, s, req.AppName, req.UserID),
		lastUpdateTime: time.Now(),
	}

	// 5. Init storableSession hold data Session
	storableSess := storableSession{
		ID:             sessionID,
		AppName:        req.AppName,
		UserID:         req.UserID,
		State:          mergedState,
		LastUpdateTime: redisSession.lastUpdateTime,
	}

	data, err := json.Marshal(storableSess)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session data: %w", err)
	}
	// 6. Add storableSession to Redis
	if err := s.client.Set(ctx, sessKey, data, s.ttl).Err(); err != nil {
		return nil, fmt.Errorf("failed to store session to redis: %w", err)
	}
	// 7. Store key session to Set contain sessions of a user SAdd(key_member, [member1, member2, v.v.])
	idxSessKey := s.sessionKeys(req.AppName, req.UserID)
	if err := s.client.SAdd(ctx, idxSessKey, sessionID).Err(); err != nil {
		return nil, fmt.Errorf("failed to update session index: %w", err)
	}
	// set expire
	s.client.Expire(ctx, idxSessKey, s.ttl)
	return &session.CreateResponse{Session: redisSession}, nil
}

func (s *RedisService) Get(ctx context.Context, req *session.GetRequest) (*session.GetResponse, error) {
	// 1. Get session data
	sessionKey := s.sessionKey(req.AppName, req.UserID, req.SessionID)
	sessionData, err := s.client.Get(ctx, sessionKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("session not found: %s", req.SessionID)
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// unmarshal data to storableSession
	var storable storableSession
	if err := json.Unmarshal(sessionData, &storable); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	// 2. Get event data
	eventKey := s.eventsKey(req.AppName, req.UserID, req.SessionID)
	eventsData, err := s.client.LRange(ctx, eventKey, 0, -1).Result() // get all events
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	// parse to list events have type session.Event
	var events []*session.Event
	for _, item := range eventsData {
		event := &session.Event{}
		if err := json.Unmarshal([]byte(item), event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
		}
		events = append(events, event)
	}

	// 3. Get userDelta and appDelta => inject to state in session => context for agent
	userDelta := s.loadUserState(ctx, req.AppName, req.UserID)
	appDelta := s.loadAppState(ctx, req.AppName)
	mergedState := s.mergeState(appDelta, userDelta, storable.State)

	// init redis Session (implemented session.Session)
	sess := &redisSession{
		id:             req.SessionID,
		appName:        req.AppName,
		userID:         req.UserID,
		lastUpdateTime: storable.LastUpdateTime,
		state:          newRedisState(mergedState, s.client, sessionKey, s.ttl, s, req.AppName, req.UserID),
	}

	// 4. Filter event (optional)
	if req.NumRecentEvents > 0 && len(events) > req.NumRecentEvents {
		events = events[len(events)-req.NumRecentEvents:]
	}
	if !req.After.IsZero() {
		var filtered []*session.Event
		for _, evt := range events {
			if !evt.Timestamp.Before(req.After) {
				filtered = append(filtered, evt)
			}
		}
		events = filtered
	}
	sess.events = newRedisEvents(events, eventKey, s.client)
	return &session.GetResponse{Session: sess}, nil
}

// List returns all sessions for a user.
func (s *RedisService) List(ctx context.Context, req *session.ListRequest) (*session.ListResponse, error) {
	// 1. Get all session from key member (sessionKeys hold sessions for a user)
	sessionKeys := s.sessionKeys(req.AppName, req.UserID)

	data, err := s.client.SMembers(ctx, sessionKeys).Result()
	if err != nil {
		return nil, fmt.Errorf("Failed to get sessions for a user: %w", err)
	}
	// 2. Get one-by-one session from list members (sessions for a user)
	var sessions []session.Session
	for _, sessionKey := range data {
		respSess, err := s.Get(ctx, &session.GetRequest{
			AppName:   req.AppName,
			UserID:    req.UserID,
			SessionID: sessionKey,
		})
		if err != nil {
			fmt.Printf("failed to get session have key: %s", sessionKey)
			continue
		}
		sessions = append(sessions, respSess.Session)
	}

	return &session.ListResponse{Sessions: sessions}, nil
}

// Delete removes a session
// Sadd(key_member, [member1, member2, ...])
// SRem(key_member, [member1, member2])
func (s *RedisService) Delete(ctx context.Context, req *session.DeleteRequest) error {
	sessionKeys := s.sessionKeys(req.AppName, req.UserID)
	sesionKey := s.sessionKey(req.AppName, req.UserID, req.SessionID)
	eventKey := s.eventsKey(req.AppName, req.UserID, req.SessionID)

	pipe := s.client.Pipeline()
	pipe.Del(ctx, sesionKey)                   // delete a session key
	pipe.Del(ctx, eventKey)                    // delete event beloging to that session
	pipe.SRem(ctx, sessionKeys, req.SessionID) // delete sessionID beloging sessionKeys contain it

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// AppendEvent append event to session
// lock RedisService avoid race condition on 1 instance (must be implemented for session.Service)
func (s *RedisService) AppendEvent(ctx context.Context, sess session.Session, evt *session.Event) error {
	// 1. If chunk streaming => do nothing
	if evt.Partial {
		return nil
	}

	// lock RedisService avoid race condition on 1 instance
	s.mu.Lock()
	defer s.mu.Unlock()

	// 2. Add event to Redis
	// trim key have prefix temp:
	if len(evt.Actions.StateDelta) > 0 {
		filtered := make(map[string]any, len(evt.Actions.StateDelta))
		for k, v := range evt.Actions.StateDelta {
			if !strings.HasPrefix(k, session.KeyPrefixTemp) {
				filtered[k] = v
			}
		}
		evt.Actions.StateDelta = filtered
	}

	evt.Timestamp = time.Now()
	if evt.ID == "" {
		evt.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}

	evtData, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	evtKey := s.eventsKey(sess.AppName(), sess.UserID(), sess.ID())
	if err := s.client.RPush(ctx, evtKey, evtData).Err(); err != nil {
		return fmt.Errorf("failed to push event to Redis: %w", err)
	}

	s.client.Expire(ctx, evtKey, s.ttl)
	// 3. Update State from session.Session
	// a. get session from redis and update state data from session.Session
	sessKey := s.sessionKey(sess.AppName(), sess.UserID(), sess.ID())
	sessFromRedis, err := s.client.Get(ctx, sessKey).Bytes()
	if err != nil {
		return fmt.Errorf("failed to get session for update: %w", err)
	}

	var storable storableSession
	if err := json.Unmarshal([]byte(sessFromRedis), &storable); err != nil {
		return fmt.Errorf("failed to unmarshal session to storable: %w", err)
	}

	for k, v := range sess.State().All() {
		// if key no prefix => session state
		if strings.HasPrefix(k, session.KeyPrefixApp) || strings.HasPrefix(k, session.KeyPrefixUser) || strings.HasPrefix(k, session.KeyPrefixTemp) {
			continue
		}
		storable.State[k] = v
	}

	// b. Update State from evt.Actions.StateDelta
	if len(evt.Actions.StateDelta) > 0 {
		appDelta, userDelta, stateDelta := extractStateDeltas(evt.Actions.StateDelta)
		s.updateAppState(ctx, sess.AppName(), appDelta)
		s.updateUserState(ctx, sess.AppName(), sess.UserID(), userDelta)
		// add state delta to session state
		for k, v := range stateDelta {
			storable.State[k] = v
		}
	}

	if len(storable.State) > 0 {
		storable.LastUpdateTime = time.Now()
	}
	updatedData, err := json.Marshal(storable)
	if err != nil {
		return fmt.Errorf("failed to marshal update State data: %w", err)
	}
	if err := s.client.Set(ctx, sessKey, updatedData, s.ttl).Err(); err != nil {
		return fmt.Errorf("failed to store update State: %w", err)
	}
	return nil
}

func (s *RedisService) mergeState(appState map[string]any, userState map[string]any, sessionState map[string]any) map[string]any {
	data := make(map[string]any)
	for k, v := range sessionState {
		data[k] = v
	}
	for k, v := range userState {
		data[session.KeyPrefixUser+k] = v
	}
	for k, v := range appState {
		data[session.KeyPrefixApp+k] = v
	}
	return data
}

// updateAppState update data State of User
func (s *RedisService) updateUserState(ctx context.Context, appName string, userID string, userDelta map[string]any) (map[string]any, error) {

	if len(userDelta) == 0 {
		return s.loadUserState(ctx, appName, userID), nil
	}
	key := s.userStateKey(appName, userID)
	fields := marshalHashFields(userDelta)
	if _, err := s.client.HSet(ctx, key, fields).Result(); err != nil {
		return nil, fmt.Errorf("failed to update user State: %w", err)
	}
	if s.userStateTTL > 0 {
		s.client.Expire(ctx, key, s.userStateTTL)
	} else {
		s.client.Persist(ctx, key)
	}
	return s.loadUserState(ctx, appName, userID), nil
}

// updateAppState update data State of App
func (s *RedisService) updateAppState(ctx context.Context, appName string, appDelta map[string]any) (map[string]any, error) {
	// if not available data => return old data store in key
	if appDelta == nil {
		return s.loadAppState(ctx, appName), nil
	}

	if len(appDelta) == 0 {
		return s.loadAppState(ctx, appName), nil
	}

	key := s.appStateKey(appName)
	fields := marshalHashFields(appDelta)

	if _, err := s.client.HSet(ctx, key, fields).Result(); err != nil {
		return nil, fmt.Errorf("failed to update app state: %w", err)
	}
	if s.appStateTTL > 0 {
		s.client.Expire(ctx, key, s.appStateTTL)
	} else {
		s.client.Persist(ctx, key)
	}
	return s.loadAppState(ctx, appName), nil
}

// loadAppState return Hash data of App inherit loadHashState
func (s *RedisService) loadAppState(ctx context.Context, appName string) map[string]any {
	return s.loadHashState(ctx, s.appStateKey(appName))
}

// loadAppState return Hash data of User inherit loadHashState
func (s *RedisService) loadUserState(ctx context.Context, appName string, userID string) map[string]any {
	return s.loadHashState(ctx, s.userStateKey(appName, userID))
}

// loadHashState load Hash data from Redis
func (s *RedisService) loadHashState(ctx context.Context, key string) map[string]any {
	result, err := s.client.HGetAll(ctx, key).Result()
	if err != nil {
		log.Printf("failed to get hash data for key %s: %v", key, err)
		return make(map[string]any)
	}
	return unmarshalHashFields(result)
}

var _ session.Session = (*redisSession)(nil)
