package config

import (
	"os"
)

type Config struct {
	AppPort     string
	DatabaseURL string
}

func Load() Config {
	return Config{
		AppPort:     getenv("APP_PORT", "8080"),
		DatabaseURL: getenv("DATABASE_URL", ""),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
