// Package service implements OTP generation, delivery and verification.
package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/gazizvr/telegram-gateway-auth/internal/domain"
)

// CodeHashStore persists code hashes.
type CodeHashStore interface {
	Get(ctx context.Context, phone string) ([]byte, error)
	Exists(ctx context.Context, phone string) (bool, error)
	Set(ctx context.Context, phone string, codeHash []byte, ttl time.Duration) error
	Delete(ctx context.Context, phone string) error
	GetTTL(ctx context.Context, phone string) (*time.Duration, error)
}

// AttemptsLimiter counts failed verification attempts.
type AttemptsLimiter interface {
	Increment(ctx context.Context, phone string, ttl time.Duration) (int, error)
	Release(ctx context.Context, phone string) error
}

// RateLimiter limits how often codes can be sent to one phone number.
type RateLimiter interface {
	CanSend(ctx context.Context, phone string) error
	RecordSend(ctx context.Context, phone string) error
}

// OTPSender delivers a code to a phone number.
type OTPSender interface {
	Send(ctx context.Context, phone, code string, ttl time.Duration) error
}

// OTPService sends and verifies one-time codes.
// Codes are stored only as HMAC-SHA256 hashes.
type OTPService struct {
	sender      OTPSender
	fallback    OTPSender
	store       CodeHashStore
	limiter     AttemptsLimiter
	rateLimiter RateLimiter
	secretKey   []byte
	ttl         time.Duration
	otpLength   int
	maxAttempts int
}

// Option configures an OTPService.
type Option func(*OTPService)

// WithTTL sets the code lifetime. Default is 150 seconds.
func WithTTL(d time.Duration) Option {
	return func(src *OTPService) {
		src.ttl = d
	}
}

// WithOTPLength sets the number of digits in a code. Default is 6.
func WithOTPLength(l int) Option {
	return func(src *OTPService) {
		src.otpLength = l
	}
}

// WithMaxAttempts sets the allowed verification attempts per code. Default is 3.
func WithMaxAttempts(a int) Option {
	return func(src *OTPService) {
		src.maxAttempts = a
	}
}

const (
	defaultTTL             = 150 * time.Second
	defaultOTPLength   int = 6
	defaultMaxAttempts     = 3
)

// NewOTPService returns an OTPService. Codes go through sender first and
// through fallback if sender returns domain.ErrPhoneNotReachable or
// domain.ErrSenderUnavailable. secretKey is the HMAC key for code hashes.
func NewOTPService(
	sender OTPSender,
	fallback OTPSender,
	store CodeHashStore,
	limiter AttemptsLimiter,
	rateLimiter RateLimiter,
	secretKey []byte,
	ops ...Option,
) *OTPService {
	service := &OTPService{
		sender:      sender,
		secretKey:   secretKey,
		fallback:    fallback,
		store:       store,
		limiter:     limiter,
		rateLimiter: rateLimiter,
		ttl:         defaultTTL,
		otpLength:   defaultOTPLength,
		maxAttempts: defaultMaxAttempts,
	}
	for _, opt := range ops {
		opt(service)
	}
	return service
}

func generateOTP(length int) (*string, error) {
	limit := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(length)), nil)
	num, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return nil, fmt.Errorf("generate otp: %w", err)
	}
	otp := fmt.Sprintf("%0*d", length, num.Int64())
	return &otp, nil
}

func signHash(message, key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	hash := mac.Sum(nil)
	return hash
}

// Send generates a code and delivers it to phone.
// It returns domain.ErrMessageAlreadySent if an active code exists,
// or a *domain.FloodWaitError if the send rate limit is hit.
func (s *OTPService) Send(ctx context.Context, phone string) error {
	exists, err := s.store.Exists(ctx, phone)
	if err != nil {
		return err
	}
	if exists {
		return domain.ErrMessageAlreadySent
	}
	if err := s.rateLimiter.CanSend(ctx, phone); err != nil {
		return err
	}
	code, err := generateOTP(s.otpLength)
	if err != nil {
		return err
	}
	if err := s.sender.Send(ctx, phone, *code, s.ttl); err != nil {
		if errors.Is(err, domain.ErrPhoneNotReachable) ||
			errors.Is(err, domain.ErrSenderUnavailable) {
			if err := s.fallback.Send(ctx, phone, *code, s.ttl); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	_ = s.rateLimiter.RecordSend(ctx, phone)
	return s.store.Set(ctx, phone, signHash([]byte(*code), s.secretKey), s.ttl)
}

func (s *OTPService) verifyCodeHash(code string, hash []byte) bool {
	expected := signHash([]byte(code), s.secretKey)
	return hmac.Equal(hash, expected)
}

// Verify checks code against the stored hash and consumes it on success.
// After the maximum number of attempts it discards the code and returns
// domain.ErrMaxAttemptsExceeded.
func (s *OTPService) Verify(ctx context.Context, phone, code string) error {
	codeHash, err := s.store.Get(ctx, phone)
	if err != nil {
		return err
	}
	if codeHash == nil {
		return domain.ErrOTPNotFound
	}
	currentTTL, err := s.store.GetTTL(ctx, phone)
	if err != nil {
		return err
	}
	attempts, err := s.limiter.Increment(ctx, phone, *currentTTL)
	if err != nil {
		return err
	}
	if attempts > s.maxAttempts {
		s.store.Delete(ctx, phone)
		s.limiter.Release(ctx, phone)
		return domain.ErrMaxAttemptsExceeded
	}
	if !s.verifyCodeHash(code, codeHash) {
		return domain.ErrInvalidCode
	}
	return func() error {
		err = s.store.Delete(ctx, phone)
		err = s.limiter.Release(ctx, phone)
		return nil
	}()
}
