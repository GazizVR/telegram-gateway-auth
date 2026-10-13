package domain

type VerificationID string

type Verification struct {
	ID       *VerificationID
	OTP      string
	Attempts int
}
