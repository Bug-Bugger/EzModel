package config

import (
	"os"
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

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
