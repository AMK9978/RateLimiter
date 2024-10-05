package limiter

import (
	"log"
	"time"
)

// LeakyBucketLimiter is the leaky bucket struct based on the duration, capacity, and key.
type LeakyBucketLimiter struct {
	redisClient RedisClientInterface
	capacity    int
	leakRate    time.Duration
	bucketKey   string
}

// NewLeakyBucketLimiter creates a new instance of the LeakyBucketLimiter using Redis.
func NewLeakyBucketLimiter(redisClient RedisClientInterface, capacity int, leakRate time.Duration,
	userID string,
) *LeakyBucketLimiter {
	return &LeakyBucketLimiter{
		redisClient: redisClient,
		capacity:    capacity,
		leakRate:    leakRate,
		bucketKey:   "bucket:" + userID,
	}
}

// RateLimit checks if the request should be passed using the leaky bucket algorithm.
func (lb *LeakyBucketLimiter) RateLimit(userID string) bool {
	currentTime := time.Now().Unix()

	// Remove expired requests from the bucket
	startTime := currentTime - int64(lb.leakRate.Seconds())
	_ = lb.redisClient.ZRemRangeByScore(lb.bucketKey, startTime, currentTime)

	// Get the current number of requests in the bucket
	requestCount, err := lb.redisClient.ZCard(lb.bucketKey)
	if err != nil {
		log.Printf("Error getting bucket length for user %s: %v\n", userID, err)
		return false
	}

	// If the bucket is full, reject the request
	if requestCount >= int64(lb.capacity) {
		return false
	}

	// Add the new request with the current timestamp
	err = lb.redisClient.ZAdd(lb.bucketKey, float64(currentTime), currentTime)
	if err != nil {
		log.Printf("Error adding request for user %s: %v\n", userID, err)
		return false
	}

	// Set expiration for the key based on the leak rate to ensure memory cleanup
	_ = lb.redisClient.Expire(lb.bucketKey, lb.leakRate)
	return true
}
