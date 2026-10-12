package config

import (
    "os"
    "fmt"
    "errors"
)
    
type Config struct {
	Token    string
	HTTPAddr string
    RedisAddr string
    RedisPassword string
}

var (
    ErrMissingToken = errors.New("required token is not set")
    ErrMissingHTTPAddr = errors.New("required http addr is not set")
    ErrMissingRedisAddr = errors.New("required redis addr is not set")
)

func Load() (*Config,error) {
    token := os.Getenv("TOKEN")
    if token == "" {
        return nil, fmt.Errorf("TOKEN: %w",ErrMissingToken)
    }
    httpAddr := os.Getenv("HTTP_ADDR")
    if httpAddr == "" {
        return nil, fmt.Errorf("HTTP_ADDR: %w",ErrMissingHTTPAddr)
    }
    redisAddr := os.Getenv("REDIS_ADDR")
    if redisAddr == "" {
        return nil, fmt.Errorf("REDIS_ADDR: %w",ErrMissingRedisAddr)
    }
    redisPassword := os.Getenv("REDIS_PASSWORD")
	return &Config{
		Token:    token,
		HTTPAddr: httpAddr,
        RedisAddr: redisAddr,
        RedisPassword: redisPassword,
	}, nil
}
