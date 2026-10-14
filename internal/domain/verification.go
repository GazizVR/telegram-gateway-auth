package domain

type Verification struct {
	CodeHash string
	Attempts int
}
