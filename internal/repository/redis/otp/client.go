package redis_otp

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

func otpKey(phone string) string {
	return fmt.Sprintf("otp:%s", phone)
}

func (r *OTPRedis) Get(ctx context.Context, phone string) (*domain.Verification, error) {
	var verificationDTO verificationRedisDTO
	if err := r.rdb.HGetAll(ctx, otpKey(phone)).Scan(&verificationDTO); err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, domain.ErrOTPNotFound
		}
		return nil, fmt.Errorf("get verification failed: %w", err)
	}
	verification := verificationDTO.toDomain()
	return &verification, nil
}

func (r *OTPRedis) Set(
	ctx context.Context,
	phone string,
	verification domain.Verification,
	ttl time.Duration,
) error {
	verificationDTO := toRedisDTO(verification)
	_, err := r.rdb.TxPipelined(ctx, func(p redis.Pipeliner) error {
		p.HSet(ctx, otpKey(phone), verificationDTO)
		p.Expire(ctx, otpKey(phone), ttl)
		return nil
	})
	if err != nil {
		return fmt.Errorf("set verification: %w", err)
	}
	return nil
}

func (r *OTPRedis) Delete(ctx context.Context, phone string) error {
	if err := r.rdb.Del(ctx, otpKey(phone)).Err(); err != nil {
		return fmt.Errorf("delete verification: %w", err)
	}
	return nil
}
