package config

import "os"

type Config struct {
	DSN  string
	Port string
}

func New() *Config {
	return &Config{
		DSN:  getEnv("DB_DSN", "postgres://postgres:postgres@localhost:5432/pr_service?sslmode=disable"),
		Port: getEnv("PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
