package limiter

import (
	"testing"
	"time"
)

type MockRedisClient struct {
	requestCount map[string]int64
}

func NewMockRedisClient() RedisClientInterface {
	return &MockRedisClient{
		requestCount: make(map[string]int64),
	}
}

func (m *MockRedisClient) SlidingWindowCheck(userID string, _, _, limit int64) bool {
	if m.requestCount[userID] >= limit {
		return false
	}
	m.requestCount[userID]++
	return true
}

func (m *MockRedisClient) LeakyBucketCheck(userID string, _, _, limit int64) bool {
	if m.requestCount[userID] >= limit {
		return false
	}
	m.requestCount[userID]++
	return true
}

func (m *MockRedisClient) ZRemRangeByScore(_ string, _, _ int64) error {
	return nil
}

func (m *MockRedisClient) ZCard(userID string) (int64, error) {
	return m.requestCount[userID], nil
}

func (m *MockRedisClient) ZAdd(userID string, _ float64, _ int64) error {
	m.requestCount[userID]++
	return nil
}

func (m *MockRedisClient) Expire(_ string, _ time.Duration) error {
	return nil
}

func TestSlidingWindowLimiter(t *testing.T) {
	t.Parallel()
	mockClient := NewMockRedisClient()
	limiter := NewSlidingWindowLimiter(mockClient)

	userID := "test-user"
	limit := 2

	// Two initial requests should succeed
	if !limiter.RateLimit(userID, limit) {
		t.Errorf("Expected request to be allowed, but it was denied")
	}
	if !limiter.RateLimit(userID, limit) {
		t.Errorf("Expected request to be allowed, but it was denied")
	}

	// The next one should exceed the limit
	if limiter.RateLimit(userID, limit) {
		t.Errorf("Expected request to be denied after exceeding the rate limit, but it was allowed")
	}
}
