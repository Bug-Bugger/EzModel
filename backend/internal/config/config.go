package config

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"os"
	"strings"
	"time"
)

type Config struct {
	Port           string
	Env            string
	Region         string
	AllowedOrigins []string
	Database       struct {
		Host     string
		Port     string
		User     string
		Password string
		DBName   string
		SSLMode  string
	}
	DatabaseReplica struct {
		Enabled  bool
		Host     string
		Port     string
		User     string
		Password string
		DBName   string
		SSLMode  string
	}
	Redis struct {
		Enabled  bool
		Host     string
		Port     string
		Password string
		DB       int
		TLS      bool
	}
	JWT struct {
		Secret          string
		AccessTokenExp  time.Duration
		RefreshTokenExp time.Duration
	}
}

func New() *Config {
	port := getEnv("PORT", "8080")
	// Add colon if not present
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	cfg := &Config{
		Port:   port,
		Env:    getEnv("ENV", "development"),
		Region: getEnv("REGION", "region1"),
	}

	// CORS Configuration - parse comma-separated origins
	originsStr := getEnv("ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:4173")
	cfg.AllowedOrigins = strings.Split(originsStr, ",")
	// Trim whitespace from each origin
	for i, origin := range cfg.AllowedOrigins {
		cfg.AllowedOrigins[i] = strings.TrimSpace(origin)
	}

	// Primary Database Configuration
	cfg.Database.Host = getEnv("DB_HOST", "localhost")
	cfg.Database.Port = getEnv("DB_PORT", "5432")
	cfg.Database.User = getEnv("DB_USER", "postgres")
	cfg.Database.Password = getEnv("DB_PASSWORD", "")
	cfg.Database.DBName = getEnv("DB_NAME", "ezmodel_backend")
	cfg.Database.SSLMode = getEnv("DB_SSL_MODE", "disable")

	// Read Replica Configuration
	cfg.DatabaseReplica.Enabled = getEnv("DB_REPLICA_ENABLED", "false") == "true"
	cfg.DatabaseReplica.Host = getEnv("DB_REPLICA_HOST", "")
	cfg.DatabaseReplica.Port = getEnv("DB_REPLICA_PORT", "5432")
	cfg.DatabaseReplica.User = getEnv("DB_REPLICA_USER", cfg.Database.User)
	cfg.DatabaseReplica.Password = getEnv("DB_REPLICA_PASSWORD", cfg.Database.Password)
	cfg.DatabaseReplica.DBName = getEnv("DB_REPLICA_NAME", cfg.Database.DBName)
	cfg.DatabaseReplica.SSLMode = getEnv("DB_REPLICA_SSL_MODE", cfg.Database.SSLMode)

	// Redis Configuration
	cfg.Redis.Enabled = getEnv("REDIS_ENABLED", "false") == "true"
	cfg.Redis.Host = getEnv("REDIS_HOST", "localhost")
	cfg.Redis.Port = getEnv("REDIS_PORT", "6379")
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")
	cfg.Redis.TLS = getEnv("REDIS_TLS", "false") == "true"
	cfg.Redis.DB = 0
	if dbStr := getEnv("REDIS_DB", "0"); dbStr != "0" {
		if db, err := time.ParseDuration(dbStr + "s"); err == nil {
			cfg.Redis.DB = int(db.Seconds())
		}
	}

	// JWT Configuration
	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		if cfg.Env == "production" {
			log.Fatal("JWT_SECRET environment variable is required in production")
		}
		// Generate random secret for development
		randomBytes := make([]byte, 32)
		if _, err := rand.Read(randomBytes); err != nil {
			log.Fatal("Failed to generate random JWT secret:", err)
		}
		jwtSecret = base64.StdEncoding.EncodeToString(randomBytes)
		log.Println("WARNING: Using randomly generated JWT secret. Set JWT_SECRET environment variable for production!")
	}
	cfg.JWT.Secret = jwtSecret

	accessExp, _ := time.ParseDuration(getEnv("JWT_ACCESS_TOKEN_EXP", "15m"))
	refreshExp, _ := time.ParseDuration(getEnv("JWT_REFRESH_TOKEN_EXP", "7d"))
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
