// Package config loads service configuration from environment variables.
package config

import (
	"errors"
	"fmt"
	"os"
)

// Config holds the service settings read from the environment.
//
// Required fields are TOKEN, SECRET_KEY and REDIS_ADDR.
// SMS fallback is disabled when either Eskiz field is empty.
type Config struct {
	Token         string
	SecretKey     string
	HTTPAddr      string
	RedisAddr     string
	RedisPassword string
	EskizEmail    string
	EskizPassword string
	WebDemo       bool
	GRPCAddr      string
}

// Errors returned by Load for a missing required variable.
var (
	ErrMissingToken     = errors.New("required token is not set")
	ErrMissingRedisAddr = errors.New("required redis addr is not set")
	ErrMissingSecretKey = errors.New("required secret key is not set")
)

// Load reads TOKEN, SECRET_KEY and REDIS_ADDR, which are required,
// and others are optional.
// The returned error wraps the matching ErrMissing* value.
func Load() (*Config, error) {
	token := os.Getenv("TOKEN")
	if token == "" {
		return nil, fmt.Errorf("TOKEN: %w", ErrMissingToken)
	}
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		return nil, fmt.Errorf("SECRET_KEY: %w", ErrMissingSecretKey)
	}
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		return nil, fmt.Errorf("REDIS_ADDR: %w", ErrMissingRedisAddr)
	}
	var webDemo bool
	envWebDemo := os.Getenv("WEB_DEMO")
	if envWebDemo == "true" {
		webDemo = true
	}
	return &Config{
		Token:         token,
		SecretKey:     secretKey,
		HTTPAddr:      os.Getenv("HTTP_ADDR"),
		RedisAddr:     redisAddr,
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		EskizEmail:    os.Getenv("ESKIZ_EMAIL"),
		EskizPassword: os.Getenv("ESKIZ_PASSWORD"),
		WebDemo:       webDemo,
		GRPCAddr:      os.Getenv("GRPC_ADDR"),
	}, nil
}
