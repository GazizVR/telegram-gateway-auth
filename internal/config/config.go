// Package config loads service configuration from environment variables.
package config

import (
	"errors"
	"fmt"
	"os"
)

// Config holds the service settings read from the environment.
//
// Optional fields are WebDemo, RedisPassword, EskizEmail and EskizPassword.
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
}

// Errors returned by Load for a missing required variable.
var (
	ErrMissingToken     = errors.New("required token is not set")
	ErrMissingHTTPAddr  = errors.New("required http addr is not set")
	ErrMissingRedisAddr = errors.New("required redis addr is not set")
	ErrMissingSecretKey = errors.New("required secret key is not set")
)

// Load reads TOKEN, SECRET_KEY, HTTP_ADDR and REDIS_ADDR, which are required,
// and WEB_DEMO, REDIS_PASSWORD, ESKIZ_EMAIL and ESKIZ_PASSWORD, which are optional.
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
	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		return nil, fmt.Errorf("HTTP_ADDR: %w", ErrMissingHTTPAddr)
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
		HTTPAddr:      httpAddr,
		RedisAddr:     redisAddr,
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		EskizEmail:    os.Getenv("ESKIZ_EMAIL"),
		EskizPassword: os.Getenv("ESKIZ_PASSWORD"),
		WebDemo:       webDemo,
	}, nil
}
