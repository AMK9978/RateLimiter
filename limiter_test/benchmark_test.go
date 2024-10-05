package limiter_test

import (
	"testing"

	"ratelimiter/internal/limiter"
)

func BenchmarkRateLimit(b *testing.B) {
	redisClient := limiter.NewRedisClient("localhost:6379")
	slidingLimiter := limiter.NewSlidingWindowLimiter(redisClient)
	userID := "test-user"
	limit := 2

	// Warm up
	for i := 0; i < limit; i++ {
		slidingLimiter.RateLimit(userID, limit)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		slidingLimiter.RateLimit(userID, limit)
	}
}
