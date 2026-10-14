package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/gazizvr/telegram-gateway-auth/internal/domain"
)

type OTPStore interface {
	Get(ctx context.Context, phone string) (*domain.Verification, error)
	Exists(ctx context.Context, phone string) (bool, error)
	Set(ctx context.Context, phone string, verification domain.Verification, ttl time.Duration) error
    IncrAttempts(ctx context.Context, phone string) (int,error)
	Delete(ctx context.Context, phone string) error
}

type OTPProvider interface {
	Send(ctx context.Context, phone, code string, ttl time.Duration) error
}

type OTPService struct {
	provider OTPProvider
	store    OTPStore
}

func NewOTPService(
	provider OTPProvider,
	store OTPStore,
) *OTPService {
	return &OTPService{
		provider: provider,
		store:    store,
	}
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

const (
	defaultTTL           = 150 * time.Second
	defaultOTPLength int = 6
)

func (s *OTPService) Send(ctx context.Context, phone string) error {
	exists, err := s.store.Exists(ctx, phone)
	if err != nil {
		return err
	}
	if exists {
		return domain.ErrMessageAlreadySent
	}
	otp, err := generateOTP(defaultOTPLength)
	if err != nil {
		return err
	}
	if err := s.provider.Send(ctx, phone, *otp, defaultTTL); err != nil {
		return err
	}
	verification := domain.Verification{OTP: *otp}
	return s.store.Set(ctx, phone, verification, defaultTTL)
}

const maxAttempts = 3

func (s *OTPService) Verify(ctx context.Context, phone, code string) error {
	verification, err := s.store.Get(ctx, phone)
	if err != nil {
		return err
	}
	if verification == nil {
		return domain.ErrOTPNotFound
	}
    if verification.Attempts >= maxAttempts {
        return domain.ErrMaxAttemptsExceeded
    }
    attempts, err := s.store.IncrAttempts(ctx,phone)
    if err != nil {
        return err
    }
    if attempts > maxAttempts {
        return domain.ErrMaxAttemptsExceeded
    }
    if verification.OTP != code {
        return domain.ErrInvalidCode
    }
	return s.store.Delete(ctx, phone)
}
