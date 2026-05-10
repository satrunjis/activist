package config

import (
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	HTTPAddr              string
	PublicBaseURL         string
	DatabaseURL           string
	SessionCookieName     string
	SessionIdleTTL        time.Duration
	SessionAbsoluteTTL    time.Duration
	LogLevel              string
	DevAuthAssumeAdmin    bool
	DevAuthAdminLogin     string
	SeedSuperuserLogin    string
	SeedSuperuserPassword string
}

func Load() (Config, error) {
	databaseURL, err := mustEnv("DATABASE_URL")
	if err != nil {
		return Config{}, err
	}
	if _, err := pgxpool.ParseConfig(databaseURL); err != nil {
		return Config{}, fmt.Errorf("DATABASE_URL is invalid: %w", err)
	}

	sessionIdleTTL, err := mustDurationEnv("SESSION_IDLE_TTL")
	if err != nil {
		return Config{}, err
	}

	sessionAbsoluteTTL, err := mustDurationEnv("SESSION_ABSOLUTE_TTL")
	if err != nil {
		return Config{}, err
	}

	httpAddr, err := mustEnv("HTTP_ADDR")
	if err != nil {
		return Config{}, err
	}
	publicBaseURL, err := mustEnv("PUBLIC_BASE_URL")
	if err != nil {
		return Config{}, err
	}
	sessionCookieName, err := mustEnv("SESSION_COOKIE_NAME")
	if err != nil {
		return Config{}, err
	}
	logLevel, err := mustEnv("LOG_LEVEL")
	if err != nil {
		return Config{}, err
	}
	devAuthAssumeAdmin, err := boolEnv("DEV_AUTH_ASSUME_ADMIN", false)
	if err != nil {
		return Config{}, err
	}
	devAuthAdminLogin := stringEnv("DEV_AUTH_ADMIN_LOGIN", "admin")
	seedSuperuserLogin, err := mustEnv("SEED_SUPERUSER_LOGIN")
	if err != nil {
		return Config{}, err
	}
	seedSuperuserPassword, err := mustEnv("SEED_SUPERUSER_PASSWORD")
	if err != nil {
		return Config{}, err
	}

	return Config{
		HTTPAddr:              httpAddr,
		PublicBaseURL:         publicBaseURL,
		DatabaseURL:           databaseURL,
		SessionCookieName:     sessionCookieName,
		SessionIdleTTL:        sessionIdleTTL,
		SessionAbsoluteTTL:    sessionAbsoluteTTL,
		LogLevel:              logLevel,
		DevAuthAssumeAdmin:    devAuthAssumeAdmin,
		DevAuthAdminLogin:     devAuthAdminLogin,
		SeedSuperuserLogin:    seedSuperuserLogin,
		SeedSuperuserPassword: seedSuperuserPassword,
	}, nil
}

func mustEnv(key string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return value, nil
}

func mustDurationEnv(key string) (time.Duration, error) {
	rawValue, err := mustEnv(key)
	if err != nil {
		return 0, err
	}
	parsedValue, err := time.ParseDuration(rawValue)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", key, err)
	}
	return parsedValue, nil
}

func boolEnv(key string, fallback bool) (bool, error) {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback, nil
	}
	switch raw {
	case "1", "true", "TRUE", "True", "yes", "YES", "on", "ON":
		return true, nil
	case "0", "false", "FALSE", "False", "no", "NO", "off", "OFF":
		return false, nil
	default:
		return false, fmt.Errorf("%s must be a boolean value", key)
	}
}

func stringEnv(key string, fallback string) string {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback
	}
	return raw
}
