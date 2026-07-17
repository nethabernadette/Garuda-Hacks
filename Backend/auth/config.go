package auth

import (
	"errors"
	"os"
	"strconv"
	"time"
)

const (
	defaultAccessTokenTTL = time.Hour
	jwtIssuer             = "harvestlink"
)

type Config struct {
	JWTSecret      string
	AccessTokenTTL time.Duration
}

func LoadConfigFromEnv() (Config, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return Config{}, ErrJWTSecretNotSet
	}

	ttl := defaultAccessTokenTTL
	if rawTTL := os.Getenv("JWT_ACCESS_TOKEN_TTL_SECONDS"); rawTTL != "" {
		seconds, err := strconv.Atoi(rawTTL)
		if err != nil || seconds <= 0 {
			return Config{}, errors.Join(ErrInvalidJWTConfig, errors.New("JWT_ACCESS_TOKEN_TTL_SECONDS must be a positive integer"))
		}
		ttl = time.Duration(seconds) * time.Second
	}

	return Config{
		JWTSecret:      secret,
		AccessTokenTTL: ttl,
	}, nil
}
