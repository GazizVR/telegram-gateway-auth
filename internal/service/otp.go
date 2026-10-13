package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/gazizvr/telegram-gateway-auth/internal/domain"
)

type OTPStore interface {
	Get(ctx context.Context, phone string) (*domain.Verification, error)
	Set(ctx context.Context, phone string, verification domain.Verification, ttl time.Duration) error
	Delete(ctx context.Context, phone string) error
}

type OTPProvider interface {
	Send(ctx context.Context, phone, code string, ttl time.Duration) (*domain.VerificationID, error)
	Verify(ctx context.Context, id domain.VerificationID, code string) error
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
	otp, err := generateOTP(defaultOTPLength)
	if err != nil {
		return err
	}
	verificationID, err := s.provider.Send(ctx, phone, *otp, defaultTTL)
	if err != nil {
		return err
	}
	verification := domain.Verification{
		ID:  verificationID,
		OTP: *otp,
	}
	return s.store.Set(ctx, phone, verification, defaultTTL)
}

func (s *OTPService) Verify(ctx context.Context, phone, code string) error {
	verification, err := s.store.Get(ctx, phone)
	if err != nil {
		return err
	}
	if verification == nil || verification.ID == nil {
		return domain.ErrOTPNotFound
	}
	if err := s.provider.Verify(ctx, *verification.ID, code); err != nil {
		slog.ErrorContext(ctx, err.Error())
		return err
	}
	if err := s.store.Delete(ctx, phone); err != nil {
		slog.ErrorContext(ctx, err.Error())
		return err
	}
	return nil
}
