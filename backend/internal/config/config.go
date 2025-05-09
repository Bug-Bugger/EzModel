package config

import (
	"os"
	"time"
)

type Config struct {
	Port     string
	Env      string
	Database struct {
		Host     string
		Port     string
		User     string
		Password string
		DBName   string
		SSLMode  string
	}
	JWT struct {
		Secret          string
		AccessTokenExp  time.Duration
		RefreshTokenExp time.Duration
	}
}

func New() *Config {
	cfg := &Config{
		Port: getEnv("PORT", ":8080"),
		Env:  getEnv("ENV", "development"),
	}

	cfg.Database.Host = getEnv("DB_HOST", "localhost")
	cfg.Database.Port = getEnv("DB_PORT", "5432")
	cfg.Database.User = getEnv("DB_USER", "postgres")
	cfg.Database.Password = getEnv("DB_PASSWORD", "")
	cfg.Database.DBName = getEnv("DB_NAME", "ezmodel_backend")
	cfg.Database.SSLMode = getEnv("DB_SSL_MODE", "disable")

	// JWT Configuration
	cfg.JWT.Secret = getEnv("JWT_SECRET", "")
	accessExp, _ := time.ParseDuration(getEnv("JWT_ACCESS_EXP", "15m"))
	refreshExp, _ := time.ParseDuration(getEnv("JWT_REFRESH_EXP", "7d"))
	cfg.JWT.AccessTokenExp = accessExp
	cfg.JWT.RefreshTokenExp = refreshExp

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
