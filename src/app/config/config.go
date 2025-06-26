package config

import (
	"errors"
	"go-mma/shared/common/env"
	"time"
)

var (
	ErrInvalidHTTPPort = errors.New("HTTP_PORT must be a positive integer")
	ErrGracefulTimeout = errors.New("GRACEFUL_TIMEOUT must be a positive duration")
	ErrDSN             = errors.New("DB_DSN must be set")
)

// รวมการโหลดค่าคอนฟิกทั้งหมดไว้ในจุดเดียว
type Config struct {
	HTTPPort        int
	GracefulTimeout time.Duration
	DSN             string
	GatewayHost     string
	GatewayBasePath string
}

func Load() (*Config, error) {
	config := &Config{
		HTTPPort:        env.GetIntDefault("HTTP_PORT", 8090),
		GracefulTimeout: env.GetDurationDefault("GRACEFUL_TIMEOUT", 5*time.Second),
		DSN:             env.Get("DB_DSN"),
		GatewayHost:     env.Get("GATEWAY_HOST"),
		GatewayBasePath: env.GetDefault("GATEWAY_BASEURL", "/api/v1"),
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
	if len(c.DSN) == 0 {
		return ErrDSN
	}

	return nil
}
