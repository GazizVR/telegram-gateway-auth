package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/gazizvr/telegram-gateway-auth/internal/domain"
)

type OTPStore interface {
	Get(ctx context.Context, phone string) (*domain.Verification, error)
	Exists(ctx context.Context, phone string) (bool, error)
	Set(ctx context.Context, phone string, verification domain.Verification, ttl time.Duration) error
	IncrAttempts(ctx context.Context, phone string) (int, error)
	Delete(ctx context.Context, phone string) error
}

type OTPSender interface {
	Send(ctx context.Context, phone, code string, ttl time.Duration) error
}

type OTPService struct {
	sender   OTPSender
	fallback OTPSender
	store    OTPStore
}

func NewOTPService(
	sender OTPSender,
	fallback OTPSender,
	store OTPStore,
) *OTPService {
	return &OTPService{
		sender:   sender,
		fallback: fallback,
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
	if err := s.sender.Send(ctx, phone, *otp, defaultTTL); err != nil {
		if errors.Is(err, domain.ErrPhoneNotReachable) ||
			errors.Is(err, domain.ErrSenderUnavailable) {
			if err := s.fallback.Send(ctx, phone, *otp, defaultTTL); err != nil {
				return err
			}
		} else {
			return err
		}
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
		s.store.Delete(ctx, phone)
		return domain.ErrMaxAttemptsExceeded
	}
	attempts, err := s.store.IncrAttempts(ctx, phone)
	if err != nil {
		return err
	}
	if attempts > maxAttempts {
		s.store.Delete(ctx, phone)
		return domain.ErrMaxAttemptsExceeded
	}
	if verification.OTP != code {
		return domain.ErrInvalidCode
	}
	return s.store.Delete(ctx, phone)
}
