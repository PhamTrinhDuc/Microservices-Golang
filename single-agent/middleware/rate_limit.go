package middleware

import (
	"fmt"
	"net/http"
	"single-agent/internal/auth"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Redis Scripts
const (
	FixedWindowScript = `
local curr = redis.call("INCR", KEYS[1])
if curr == 1 then 
	redis.call("EXPIRE", KEYS[1], ARGV[1])
end
return curr
`

	TokenBucketScript = `
-- KEYS[1]: Key định danh (ví dụ: "ratelimit:user_123")
-- ARGV[1]: Capacity (Sức chứa tối đa của xô)
-- ARGV[2]: Refill Rate (Tốc độ nạp thẻ: số thẻ/giây)
-- ARGV[3]: Now (Thời điểm hiện tại - Unix timestamp)

local bucket = redis.call('hmget', KEYS[1], 'tokens', 'last_updated')
local last_tokens = tonumber(bucket[1]) or tonumber(ARGV[1])
local last_updated = tonumber(bucket[2]) or tonumber(ARGV[3]) 

local duration = math.max(0, ARGV[3] - last_updated)
local added_tokens = duration * ARGV[2]

local tokens_cu = math.min(added_tokens + last_tokens, ARGV[1])

if tokens_cu >= 1 then 
	tokens_cu = tokens_cu - 1
	redis.call('hmset', KEYS[1], 'tokens', tokens_cu, 'last_updated', ARGV[3])
	redis.call('expire', KEYS[1], 600)
	return 1
else 
	return 0
end
`
)

type FixedWindowLimiter struct {
	redis  *redis.Client
	limit  int
	window time.Duration
}

type TokenBucketLimiter struct {
	redis    *redis.Client
	capacity int
	rate     float64
}

func NewFixedWindowLimiter(redisClient *redis.Client, limit int, window time.Duration) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		redis:  redisClient,
		limit:  limit,
		window: window,
	}
}

func NewTokenBucketLimiter(redisClient *redis.Client, capacity int, rate float64) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		redis:    redisClient,
		capacity: capacity,
		rate:     rate,
	}
}

// Handler returns a Gin middleware for Fixed Window rate limiting
func (rl *FixedWindowLimiter) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get(string(auth.ContextKeyUserID))
		if !exists {
			c.Next()
			return
		}

		key := fmt.Sprintf("ratelimit:fixed:%s:%d", userID, time.Now().Unix()/60)
		result, err := rl.redis.Eval(c.Request.Context(), FixedWindowScript, []string{key}, int(rl.window.Seconds())).Result()
		if err != nil {
			fmt.Printf("Rate limit error: %v\n", err)
			c.Next()
			return
		}

		count, ok := result.(int64)
		if !ok || int(count) > rl.limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			// c.Abort()
			return
		}

		c.Next()
	}
}

// Handler returns a Gin middleware for Token Bucket rate limiting
func (rl *TokenBucketLimiter) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get(string(auth.ContextKeyUserID))
		if !exists {
			c.Next()
			return
		}

		key := fmt.Sprintf("ratelimit:bucket:%s", userID)
		now := time.Now().Unix()

		result, err := rl.redis.Eval(c.Request.Context(), TokenBucketScript, []string{key}, rl.capacity, rl.rate, now).Int()
		if err != nil {
			fmt.Printf("Rate limit error: %v\n", err)
			c.Next()
			return
		}

		if result == 0 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			// c.Abort()
			return
		}

		c.Next()
	}
}
