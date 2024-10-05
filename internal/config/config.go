package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds the configuration for the application.
type Config struct {
	RedisAddr        string
	ServerPort       string
	FailureThreshold uint32
	CBTimeout        time.Duration
	LockTime         time.Duration
}

// LoadConfig loads configuration from environment variables or defaults.
func LoadConfig() *Config {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	ft, err := strconv.Atoi(os.Getenv("FAILURE_THRESHOLD"))
	if err != nil {
		ft = 3
	}

	cbTimeout, err := strconv.Atoi(os.Getenv("CB_TIMEOUT"))
	if err != nil {
		cbTimeout = 5
	}
	lockTime, err := strconv.Atoi(os.Getenv("LOCK_TIME"))
	if err != nil {
		lockTime = 5
	}

	return &Config{
		RedisAddr:        redisAddr,
		ServerPort:       serverPort,
		FailureThreshold: uint32(ft),
		CBTimeout:        time.Duration(cbTimeout) * time.Second,
		LockTime:         time.Duration(lockTime) * time.Second,
	}
}
