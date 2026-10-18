package redis_otp

import (
	"context"
	"errors"
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

func codeHashKey(phone string) string {
	return fmt.Sprintf("otp:%s", phone)
}

func (s *Store) Get(ctx context.Context, phone string) ([]byte, error) {
	codeHash, err := s.rdb.Get(ctx, codeHashKey(phone)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, domain.ErrOTPNotFound
		}
		return nil, fmt.Errorf("get code hash: %w", err)
	}
	return codeHash, nil
}

func (s *Store) Exists(ctx context.Context, phone string) (bool, error) {
	count, err := s.rdb.Exists(ctx, codeHashKey(phone)).Result()
	if err != nil {
		return false, fmt.Errorf("check exists: %w", err)
	}
	return count > 0, nil
}

func (s *Store) Set(
	ctx context.Context,
	phone string,
	codeHash []byte,
	ttl time.Duration,
) error {
	ok, err := s.rdb.SetNX(ctx, codeHashKey(phone), codeHash, ttl).Result()
	if err != nil {
		return fmt.Errorf("set code hash: %w", err)
	}
	if !ok {
		return domain.ErrMessageAlreadySent
	}
	return nil
}

func (s *Store) Delete(ctx context.Context, phone string) error {
	if err := s.rdb.Del(ctx, codeHashKey(phone)).Err(); err != nil {
		return fmt.Errorf("delete code hash: %w", err)
	}
	return nil
}

func (s *Store) GetTTL(ctx context.Context, phone string) (*time.Duration, error) {
	ttl, err := s.rdb.TTL(ctx, codeHashKey(phone)).Result()
	if err != nil {
		return nil, fmt.Errorf("get code hash ttl: %w", err)
	}
	return &ttl, nil
}

func codeAttemptsKey(phone string) string {
	return "otp:attempts:" + phone
}

func (s *Store) Increment(ctx context.Context, phone string, ttl time.Duration) (int, error) {
	pipe := s.rdb.TxPipeline()
	key := codeAttemptsKey(phone)
	cmd := pipe.Incr(ctx, key)
	pipe.ExpireNX(ctx, key, ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, fmt.Errorf("increment attempts: %w", err)
	}
	return int(cmd.Val()), nil
}
func (s *Store) Release(ctx context.Context, phone string) error {
	if err := s.rdb.Del(ctx, codeAttemptsKey(phone)).Err(); err != nil {
		return fmt.Errorf("release attempts: %w", err)
	}
	return nil
}
