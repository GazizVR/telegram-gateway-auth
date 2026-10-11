package config

import "os"

type Config struct {
	Token    string
	HTTPAddr string
}

func Load() *Config {
	return &Config{
		Token:    os.Getenv("TOKEN"),
		HTTPAddr: os.Getenv("HTTP_ADDR"),
	}
}
