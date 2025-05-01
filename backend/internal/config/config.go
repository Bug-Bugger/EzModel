package config

type Config struct {
	Port string
	Env  string
}

func New() *Config {
	return &Config{
		Port: ":8080",
		Env:  "development",
	}
}
