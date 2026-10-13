package redis_otp

import "github.com/gazizvr/telegram-gateway-auth/internal/domain"

type verificationRedisDTO struct {
	ID       *string `redis:"id"`
	OTP      string  `redis:"otp"`
	Attempts int     `redis:"attempts"`
}

func (v verificationRedisDTO) toDomain() domain.Verification {
	var id *domain.VerificationID
	if v.ID != nil {
		rawId := domain.VerificationID(*v.ID)
		id = &rawId
	}
	return domain.Verification{
		ID:       id,
		OTP:      v.OTP,
		Attempts: v.Attempts,
	}
}

func toRedisDTO(v domain.Verification) verificationRedisDTO {
	var id *string
	if v.ID != nil {
		rawId := string(*v.ID)
		id = &rawId
	}
	return verificationRedisDTO{
		ID:       id,
		OTP:      v.OTP,
		Attempts: v.Attempts,
	}
}
