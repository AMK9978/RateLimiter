package routes

import (
	"net/http"
	"strconv"

	"ratelimiter/internal/limiter"
)

// RedisClient reference should be injected via this package.
var redisClient *limiter.RedisClient // nolint

// SetRedisClient allows injecting the Redis client into this package.
func SetRedisClient(client *limiter.RedisClient) {
	redisClient = client
}

// LimiterHandler handles rate-limiting requests and can work with both sliding window and leaky bucket algorithms.
func LimiterHandler(handler limiter.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("userID")
		if userID == "" {
			http.Error(w, "userID is required", http.StatusBadRequest)
			return
		}

		windowDuration, err := strconv.Atoi(r.URL.Query().Get("window_duration"))
		if err != nil {
			http.Error(w, "window_duration is required", http.StatusBadRequest)
			return
		}

		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil {
			http.Error(w, "limit is required", http.StatusBadRequest)
			return
		}

		// Perform rate limiting using the injected handler (Sliding or Leaky)
		result := handler.RateLimiter(userID, int64(windowDuration), int64(limit), redisClient)
		if result {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Request allowed"))
		} else {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("Rate limit exceeded"))
		}
	}
}

// RegisterRoutes registers all routes and handlers for the app.
func RegisterRoutes(mux *http.ServeMux) {
	// Create instances of the handlers
	slidingHandler := &limiter.SlidingHandler{}
	leakyHandler := &limiter.LeakyHandler{}

	// Register routes using the generic LimiterHandler
	mux.HandleFunc("/limit", LimiterHandler(slidingHandler))
	mux.HandleFunc("/leaky-limit", LimiterHandler(leakyHandler))
}
