package limiter

import (
	"context"
	"fmt"
	"log"
	"time"

	"ratelimiter/internal/config"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker"
)

var cfg = config.LoadConfig() // nolint

// RedisClientInterface is an interface to be used by the real RedisClient and MockRedisClient.
type RedisClientInterface interface {
	SlidingWindowCheck(userID string, currentTime, windowDuration, limit int64) bool
	LeakyBucketCheck(userID string, currentTime, windowDuration, limit int64) bool
	ZRemRangeByScore(userID string, start, end int64) error
	ZCard(userID string) (int64, error)
	ZAdd(userID string, score float64, member int64) error
	Expire(userID string, duration time.Duration) error
}

// RedisClient handles Redis operations for a single Redis node.
type RedisClient struct {
	client         *redis.Client
	locker         *redislock.Client
	circuitBreaker *gobreaker.CircuitBreaker
}

// NewRedisClient initializes a Redis client for a single node setup. It uses circuit breaker pattern to improve
// resiliency of the app.
func NewRedisClient(addr string) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	locker := redislock.New(client)

	failureThreshold := cfg.FailureThreshold
	cbTimeout := cfg.CBTimeout

	breakerSettings := gobreaker.Settings{
		Name:    "RedisCircuitBreaker",
		Timeout: cbTimeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > failureThreshold
		},
	}

	cb := gobreaker.NewCircuitBreaker(breakerSettings)

	return &RedisClient{
		client:         client,
		circuitBreaker: cb,
		locker:         locker,
	}
}

// AcquireDistributedLock acquires a distributed lock based on userID. TTL is used to avoid deadlocks.
func (r *RedisClient) AcquireDistributedLock(userID string, ttl time.Duration) (*redislock.Lock, error) {
	ctx := context.Background()
	lockKey := fmt.Sprintf("lock:%s", userID)

	lock, err := r.locker.Obtain(ctx, lockKey, ttl, nil)
	if err == redislock.ErrNotObtained {
		return nil, fmt.Errorf("could not obtain lock for userID %s", userID)
	}
	if err != nil {
		return nil, err
	}
	return lock, nil
}

// ZRemRangeByScore removes items from a sorted set based on score.
func (r *RedisClient) ZRemRangeByScore(userID string, start, end int64) error {
	ctx := context.Background()
	_, err := r.circuitBreaker.Execute(func() (interface{}, error) {
		return r.client.ZRemRangeByScore(ctx, userID, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end)).Result()
	})
	return err
}

// ZCard returns the number of elements in a sorted set.
func (r *RedisClient) ZCard(userID string) (int64, error) {
	ctx := context.Background()
	result, err := r.circuitBreaker.Execute(func() (interface{}, error) {
		return r.client.ZCard(ctx, userID).Result()
	})

	if err != nil {
		return 0, err
	}

	return result.(int64), nil
}

// ZAdd adds a new item to a sorted set.
func (r *RedisClient) ZAdd(userID string, score float64, member int64) error {
	ctx := context.Background()
	_, err := r.circuitBreaker.Execute(func() (interface{}, error) {
		return r.client.ZAdd(ctx, userID, redis.Z{
			Score:  score,
			Member: member,
		}).Result()
	})

	return err
}

// Expire sets an expiration time on a key.
func (r *RedisClient) Expire(userID string, duration time.Duration) error {
	ctx := context.Background()
	_, err := r.circuitBreaker.Execute(func() (interface{}, error) {
		return r.client.Expire(ctx, userID, duration).Result()
	})
	return err
}

// SlidingWindowCheck performs the sliding window algorithm using Redis commands.
func (r *RedisClient) SlidingWindowCheck(userID string, currentTime, windowDuration, limit int64) bool {
	lock, err := r.AcquireDistributedLock(userID, cfg.LockTime)
	if err != nil {
		log.Printf("Error acquiring lock for userID %s: %v", userID, err)
		return false
	}
	defer lock.Release(context.Background())

	startTime := currentTime - windowDuration

	// Removing outdated requests
	if err := r.ZRemRangeByScore(userID, 0, startTime); err != nil {
		log.Printf("Error removing outdated entries for user %s: %v", userID, err)
		return false
	}

	// Getting the current number of requests
	requestCount, err := r.ZCard(userID)
	if err != nil {
		log.Printf("Error getting request count for user %s: %v", userID, err)
		return false
	}

	if requestCount >= limit {
		return false
	}

	if err := r.ZAdd(userID, float64(currentTime), currentTime); err != nil {
		log.Printf("Error adding new request for user %s: %v", userID, err)
		return false
	}

	if err := r.Expire(userID, time.Duration(windowDuration)*time.Second); err != nil {
		log.Printf("Error setting expiration for user %s: %v", userID, err)
		return false
	}

	return true
}

// LeakyBucketCheck performs the leaky bucket algorithm using Redis commands.
func (r *RedisClient) LeakyBucketCheck(userID string, currentTime, windowDuration, limit int64) bool {
	ctx := context.Background()
	lock, err := r.AcquireDistributedLock(userID, cfg.LockTime)
	if err != nil {
		log.Printf("Error acquiring lock for userID %s: %v", userID, err)
		return false
	}
	defer lock.Release(context.Background())

	windowKey := "leaky:" + userID
	leakRate := windowDuration / limit

	earliestRequestTime := currentTime - leakRate*limit
	_, err = r.client.ZRemRangeByScore(ctx, windowKey, "0", fmt.Sprintf("%d", earliestRequestTime)).Result()
	if err != nil {
		log.Printf("Error removing outdated entries for user %s: %v", userID, err)
		return false
	}

	requestCount, err := r.client.ZCard(ctx, windowKey).Result()
	if err != nil {
		log.Printf("Error getting request count for user %s: %v", userID, err)
		return false
	}

	if requestCount >= limit {
		return false
	}

	_, err = r.client.ZAdd(ctx, windowKey, redis.Z{
		Score:  float64(currentTime),
		Member: currentTime,
	}).Result()
	if err != nil {
		log.Printf("Error adding new request for user %s: %v", userID, err)
		return false
	}

	_, err = r.client.Expire(ctx, windowKey, time.Duration(windowDuration)*time.Second).Result()
	if err != nil {
		log.Printf("Error setting expiration for user %s: %v", userID, err)
		return false
	}

	return true
}
