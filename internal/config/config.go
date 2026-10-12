package config

import "os"

type Config struct {
	Token    string
	HTTPAddr string
}

var (
    ErrMissingToken = errors.New("required token is not set")
    ErrMissingHTTPAddr = errors.New("required http addr is not set")
)

func Load() (*Config,error) {
    token := os.Getenv("TOKEN")
    if token == "" {
        return nil, fmt.Errorf("TOKEN: %w",ErrMissingToken)
    }
    httpAddr := os.Getenv("HTTP_ADDR")
    if httpAddr == "" {
        return nil, fmt.Errorf("HTTP_ADDR: %w",ErrMissingHTTPAddr)
    }
	return &Config{
		Token:    token,
		HTTPAddr: httpAddr,
	}, nil
}
