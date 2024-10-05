package limiter

import (
	"time"

	"ratelimiter/internal/logger"
)

type Handler interface {
	RateLimiter(userID string, windowDuration, limit int64, redisClient *RedisClient) bool
}

type SlidingHandler struct{}

type LeakyHandler struct{}

// RateLimiter performs sliding window rate limit checks.
func (s *SlidingHandler) RateLimiter(userID string, windowDuration, limit int64, redisClient *RedisClient) bool {
	currentTime := time.Now().Unix()

	// Perform a sliding window rate limit
	if redisClient.SlidingWindowCheck(userID, currentTime, windowDuration, limit) {
		logger.Log.Info("Sliding Window: Request allowed")
		return true
	}

	logger.Log.Info("Sliding Window: Rate limit exceeded")
	return false
}

// RateLimiter performs leaky bucket rate limit checks.
func (l *LeakyHandler) RateLimiter(userID string, windowDuration, limit int64, redisClient *RedisClient) bool {
	currentTime := time.Now().Unix()

	// Perform a leaky bucket rate limit
	if redisClient.LeakyBucketCheck(userID, currentTime, windowDuration, limit) {
		logger.Log.Info("Leaky Bucket: Request allowed")
		return true
	}

	logger.Log.Info("Leaky Bucket: Rate limit exceeded")
	return false
}
