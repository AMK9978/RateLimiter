package limiter

import (
	"time"
)

type RateLimiter interface {
	RateLimit(userID string, limit int) bool
}

type SlidingWindowLimiter struct {
	redisClient RedisClientInterface
}

func NewSlidingWindowLimiter(redisClient RedisClientInterface) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		redisClient: redisClient,
	}
}

func (limiter *SlidingWindowLimiter) RateLimit(userID string, limit int) bool {
	currentTime := time.Now().Unix()
	windowDuration := 1

	return limiter.redisClient.SlidingWindowCheck(userID, currentTime, int64(windowDuration), int64(limit))
}
