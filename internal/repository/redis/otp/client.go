// Package redis_otp stores OTP state in Redis or Valkey:
// code hashes, verification attempts and send rate limits.
package redis_otp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gazizvr/telegram-gateway-auth/internal/domain"
	"github.com/redis/go-redis/v9"
)

// Store implements the storage and limiter interfaces of the service package.
type Store struct {
	rdb               *redis.Client
	maxHourlyAttempts int
}

// Option configures a Store.
type Option func(*Store)

// WithMaxHourlyAttempts sets the hourly send limit per phone number. Default is 4.
func WithMaxHourlyAttempts(a int) Option {
	return func(s *Store) {
		s.maxHourlyAttempts = a
	}
}

const defaultMaxHourlyAttempts = 4

// NewStore returns a Store that uses rdb.
func NewStore(
	rdb *redis.Client,
	ops ...Option,
) *Store {
	store := &Store{
		rdb:               rdb,
		maxHourlyAttempts: defaultMaxHourlyAttempts,
	}
	for _, op := range ops {
		op(store)
	}
	return store
}

func codeHashKey(phone string) string {
	return fmt.Sprintf("otp:%s", phone)
}

// Get returns the stored code hash, or domain.ErrOTPNotFound if none exists.
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

// Exists reports whether a code hash is stored for phone.
func (s *Store) Exists(ctx context.Context, phone string) (bool, error) {
	count, err := s.rdb.Exists(ctx, codeHashKey(phone)).Result()
	if err != nil {
		return false, fmt.Errorf("check exists: %w", err)
	}
	return count > 0, nil
}

// Set stores codeHash for ttl. It returns domain.ErrMessageAlreadySent
// if a code already exists for phone.
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

// Delete removes the code hash for phone.
func (s *Store) Delete(ctx context.Context, phone string) error {
	if err := s.rdb.Del(ctx, codeHashKey(phone)).Err(); err != nil {
		return fmt.Errorf("delete code hash: %w", err)
	}
	return nil
}

// GetTTL returns the remaining lifetime of the code hash for phone.
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

// Increment counts a verification attempt and returns the new total.
// The counter expires after ttl.
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

// Release resets the verification attempt counter for phone.
func (s *Store) Release(ctx context.Context, phone string) error {
	if err := s.rdb.Del(ctx, codeAttemptsKey(phone)).Err(); err != nil {
		return fmt.Errorf("release attempts: %w", err)
	}
	return nil
}

func cooldownKey(phone string) string {
	return "otp:cooldown:" + phone
}

func hourlyCooldownKey(phone string) string {
	return "otp:hourly:" + phone
}

// CanSend returns a *domain.FloodWaitError if phone is in the one-minute
// cooldown or has exceeded the hourly send limit.
func (s *Store) CanSend(ctx context.Context, phone string) error {
	script := redis.NewScript(`
		local hourlyCooldownKey = KEYS[1]
		local maxHourlyAttempts = tonumber(ARGV[1])
		local attempts = tonumber(redis.call('GET',hourlyCooldownKey))
		if attempts ~= nil and attempts > maxHourlyAttempts then
			return tonumber(redis.call('TTL',hourlyCooldownKey))
		end
		local cooldownKey = KEYS[2]
		local cooldown = redis.call('TTL',cooldownKey)
		return tonumber(cooldown)
	`)
	ttl, err := script.Run(ctx, s.rdb, []string{hourlyCooldownKey(phone), cooldownKey(phone)}, s.maxHourlyAttempts).Int64()
	if err != nil {
		return err
	}
	if ttl > 0 {
		return &domain.FloodWaitError{
			WaitTime: time.Duration(ttl) * time.Second,
		}
	}
	return nil
}

// RecordSend starts the one-minute cooldown and counts the send
// toward the hourly limit.
func (s *Store) RecordSend(ctx context.Context, phone string) error {
	tx := s.rdb.TxPipeline()
	cmd := tx.SetNX(ctx, cooldownKey(phone), "", time.Minute)
	tx.Incr(ctx, hourlyCooldownKey(phone))
	tx.ExpireNX(ctx, hourlyCooldownKey(phone), time.Hour)
	if _, err := tx.Exec(ctx); err != nil {
		return fmt.Errorf("record send: %w", err)
	}
	if !cmd.Val() {
		return domain.ErrOTPNotFound
	}
	return nil
}
