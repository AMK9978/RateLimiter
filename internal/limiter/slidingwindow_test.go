package limiter

import (
	"testing"
	"time"

	"ratelimiter/internal/config"
	"ratelimiter/internal/logger"
)

// nolint
var (
	cfg         *config.Config
	redisClient *RedisClient
)

// TestMain runs before any test in the package and can be used to set up common test data
func TestMain(m *testing.M) {
	logger.Initialize()
	cfg = config.LoadConfig()
	redisClient = NewRedisClient(cfg.RedisAddr)

	m.Run()
}

func TestTimeRateLimit(t *testing.T) {
	t.Parallel()
	limiter := NewSlidingWindowLimiter(redisClient)

	userID := "test-user"
	limit := 2

	for i := 0; i < limit; i++ {
		if !limiter.RateLimit(userID, limit) {
			t.Error("Rate limit failed")
		}
	}
}

func TestTimeRateLimitExceed(t *testing.T) {
	t.Parallel()
	limiter := NewSlidingWindowLimiter(redisClient)

	userID := "test-user2"
	limit := 1

	if limiter.RateLimit(userID, limit-1) {
		t.Error("Rate limit must have returned false when it is exceeded")
	}
}

func TestLeakyRateLimit(t *testing.T) {
	t.Parallel()
	userID := "test-user"
	limiter := NewLeakyBucketLimiter(redisClient, 3, time.Second, userID)

	for i := 0; i < 15; i++ {
		if limiter.RateLimit(userID) {
			logger.Log.Info("Request allowed")
		} else {
			logger.Log.Info("Request denied: Rate limit exceeded")
		}
		// To Simulate time between requests
		time.Sleep(200 * time.Millisecond)
	}
}
