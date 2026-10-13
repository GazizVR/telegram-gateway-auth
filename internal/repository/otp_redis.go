package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gazizvr/telegram-gateway-auth/internal/domain"
	"github.com/redis/go-redis/v9"
)

type OTPRedis struct {
	rdb *redis.Client
}

func NewOTPRedis(rdb *redis.Client) *OTPRedis {
	return &OTPRedis{
		rdb: rdb,
	}
}

func (r *OTPRedis) Get(ctx context.Context, phone string) (*domain.Challenge, error) {
	val, err := r.rdb.Get(ctx, phone).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, domain.ErrOTPNotFound
		}
		return nil, fmt.Errorf("get challenge failed: %w", err)
	}
	challenge := domain.Challenge(val)
	return &challenge, nil
}

func (r *OTPRedis) Set(
	ctx context.Context,
	phone string,
	challenge domain.Challenge,
	expAt time.Duration,
) error {
	if err := r.rdb.Set(ctx, phone, string(challenge), expAt).Err(); err != nil {
		return fmt.Errorf("set challenge failed: %w", err)
	}
	return nil
}

func (r *OTPRedis) Delete(ctx context.Context, phone string) error {
	if err := r.rdb.Del(ctx, phone).Err(); err != nil {
		return fmt.Errorf("delete challenge failed: %w", err)
	}
	return nil
}
