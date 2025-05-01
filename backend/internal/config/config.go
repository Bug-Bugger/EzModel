package config

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
		Port: ":8080",
		Env:  "development",
	}

	cfg.Database.Host = "localhost"
	cfg.Database.Port = "5432"
	cfg.Database.User = "hoopoe"
	cfg.Database.Password = ""
	cfg.Database.DBName = "ezmodel_backend"
	cfg.Database.SSLMode = "disable"

	return cfg
}
