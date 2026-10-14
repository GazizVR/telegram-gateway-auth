package config

import (
	"errors"
	"fmt"
	"os"
)

type Config struct {
	Token         string
	SecretKey     string
	HTTPAddr      string
	RedisAddr     string
	RedisPassword string
}

var (
	ErrMissingToken     = errors.New("required token is not set")
	ErrMissingHTTPAddr  = errors.New("required http addr is not set")
	ErrMissingRedisAddr = errors.New("required redis addr is not set")
	ErrMissingSecretKey = errors.New("required secret key is not set")
)

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
	redisPassword := os.Getenv("REDIS_PASSWORD")
	return &Config{
		Token:         token,
		SecretKey:     secretKey,
		HTTPAddr:      httpAddr,
		RedisAddr:     redisAddr,
		RedisPassword: redisPassword,
	}, nil
}
