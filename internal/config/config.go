package config

import (
	"os"
)

type Config struct {
	Env      string
	HTTPPort string
	DBDSN    string
}

func Load() Config {
	return Config{
		Env:      getenv("APP_ENV", "dev"),
		HTTPPort: getenv("HTTP_PORT", "8080"),
		DBDSN:    getenv("DB_DSN", "postgres://postgres:postgres@localhost:5432/app?sslmode=disable"),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
