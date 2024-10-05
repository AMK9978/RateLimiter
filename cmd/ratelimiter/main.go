package main

import (
	"net/http"

	"ratelimiter/internal/config"
	"ratelimiter/internal/limiter"
	"ratelimiter/internal/logger"
	"ratelimiter/internal/routes"
)

func main() {
	// Initialize the logger
	logger.Initialize()

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize Redis client
	redisClient := limiter.NewRedisClient(cfg.RedisAddr)

	// Inject Redis client into routes package
	routes.SetRedisClient(redisClient)

	// Set up the routes
	mux := http.NewServeMux()
	routes.RegisterRoutes(mux)

	// Start the server
	logger.Log.Infof("Server started on port %s", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, mux); err != nil {
		logger.Log.Fatalf("Server failed to start: %v", err)
	}
}
