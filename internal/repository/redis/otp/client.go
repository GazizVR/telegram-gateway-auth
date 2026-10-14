package redis_otp

import (
	"context"
	"fmt"
	"time"

	"github.com/gazizvr/telegram-gateway-auth/internal/domain"
	"github.com/redis/go-redis/v9"
)

type Store struct {
	rdb *redis.Client
}

func NewStore(rdb *redis.Client) *Store {
	return &Store{
		rdb: rdb,
	}
}

func key(phone string) string {
	return fmt.Sprintf("otp:%s", phone)
}

func (s *Store) Get(ctx context.Context, phone string) (*domain.Verification, error) {
	var verification verificationRedisDTO
	cmd := s.rdb.HGetAll(ctx, key(phone))
	if err := cmd.Err(); err != nil {
		return nil, fmt.Errorf("get verification: %w", err)
	}
	if len(cmd.Val()) < 1 {
		return nil, domain.ErrOTPNotFound
	}
	if err := cmd.Scan(&verification); err != nil {
		return nil, fmt.Errorf("get verification: %w", err)
	}
	return verification.toDomain(), nil
}

func (s *Store) Exists(ctx context.Context, phone string) (bool, error) {
	count, err := s.rdb.Exists(ctx, key(phone)).Result()
	if err != nil {
		return false, fmt.Errorf("check exists: %w", err)
	}
	return count > 0, nil
}

func (s *Store) Set(
	ctx context.Context,
	phone string,
	verification domain.Verification,
	ttl time.Duration,
) error {
	verificationDTO := toRedisDTO(verification)
	_, err := s.rdb.TxPipelined(ctx, func(p redis.Pipeliner) error {
		p.HSet(ctx, key(phone), verificationDTO)
		p.Expire(ctx, key(phone), ttl)
		return nil
	})
	if err != nil {
		return fmt.Errorf("set verification: %w", err)
	}
	return nil
}

func (s *Store) IncrAttempts(ctx context.Context, phone string) (int, error) {
	attempts, err := s.rdb.HIncrBy(ctx, key(phone), "attempts", 1).Result()
	if err != nil {
		return 0, fmt.Errorf("increment attempts: %w", err)
	}
	return int(attempts), err
}

func (s *Store) Delete(ctx context.Context, phone string) error {
	if err := s.rdb.Del(ctx, key(phone)).Err(); err != nil {
		return fmt.Errorf("delete verification: %w", err)
	}
	return nil
}
