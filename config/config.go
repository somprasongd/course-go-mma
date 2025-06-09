package config

import (
	"errors"
	"go-mma/util/env"
	"time"
)

var (
	ErrInvalidHTTPPort = errors.New("HTTP_PORT must be a positive integer")
	ErrGracefulTimeout = errors.New("GRACEFUL_TIMEOUT must be a positive duration")
)

// รวมการโหลดค่าคอนฟิกทั้งหมดไว้ในจุดเดียว
type Config struct {
	HTTPPort        int
	GracefulTimeout time.Duration
}

func Load() (*Config, error) {
	config := &Config{
		HTTPPort:        env.GetIntDefault("HTTP_PORT", 8090),
		GracefulTimeout: env.GetDurationDefault("GRACEFUL_TIMEOUT", 5*time.Second),
	}
	err := config.Validate()
	if err != nil {
		return nil, err
	}
	return config, err
}

func (c *Config) Validate() error {
	if c.HTTPPort <= 0 {
		return ErrInvalidHTTPPort
	}
	if c.GracefulTimeout <= 0 {
		return ErrGracefulTimeout
	}

	return nil
}
