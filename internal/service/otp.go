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
	sender    OTPSender
	fallback  OTPSender
	store     OTPStore
	secretKey []byte
}

func NewOTPService(
	sender OTPSender,
	fallback OTPSender,
	store OTPStore,
	secretKey []byte,
) *OTPService {
	return &OTPService{
		sender:    sender,
		secretKey: secretKey,
		fallback:  fallback,
		store:     store,
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

func signHash(message, key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	hash := mac.Sum(nil)
	return hash
}

func (s *OTPService) generateCodeHash(length int) (*string, error) {
	otp, err := generateOTP(length)
	if err != nil {
		return nil, err
	}
	codeHash := signHash([]byte(*otp), s.secretKey)
	code := string(codeHash)
	return &code, nil
}

const ttl = 150 * time.Second
const otpLength int = 6

func (s *OTPService) Send(ctx context.Context, phone string) error {
	exists, err := s.store.Exists(ctx, phone)
	if err != nil {
		return err
	}
	if exists {
		return domain.ErrMessageAlreadySent
	}
	codeHash, err := s.generateCodeHash(otpLength)
	if err != nil {
		return err
	}
	if err := s.sender.Send(ctx, phone, *codeHash, ttl); err != nil {
		if errors.Is(err, domain.ErrPhoneNotReachable) ||
			errors.Is(err, domain.ErrSenderUnavailable) {
			if err := s.fallback.Send(ctx, phone, *codeHash, ttl); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	verification := domain.Verification{CodeHash: *codeHash}
	return s.store.Set(ctx, phone, verification, ttl)
}

const maxAttempts = 3

func (s *OTPService) verifyCodeHash(code string, hash []byte) bool {
	expected := signHash([]byte(code), s.secretKey)
	return hmac.Equal(hash, expected)
}

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
	if s.verifyCodeHash(code, []byte(verification.CodeHash)) {
		return domain.ErrInvalidCode
	}
	return s.store.Delete(ctx, phone)
}
