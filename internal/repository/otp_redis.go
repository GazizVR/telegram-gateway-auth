package repository

import (
	"context"
	"errors"
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

func (r *OTPRedis) Get(ctx context.Context, phone string) (*string, error) {
	val, err := r.rdb.Get(ctx, phone).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, domain.ErrOTPNotFound
		}
		return nil, err
	}
	return &val, nil
}

func (r *OTPRedis) Set(
	ctx context.Context,
	phone, reqID string,
	expAt time.Duration,
) error {
	return r.rdb.Set(ctx, phone, reqID, expAt).Err()
}

func (r *OTPRedis) Delete(ctx context.Context, phone string) error {
	return r.rdb.Del(ctx, phone).Err()
}
