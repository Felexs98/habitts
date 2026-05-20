package config

import (
	"fmt"
	"os"
)

type Config struct {
	AppPort string
	AppEnv  string

	JWTSecret string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
}

func Load() *Config {
	return &Config{
		AppPort: getenv("APP_PORT", "8080"),
		AppEnv:  getenv("APP_ENV", "development"),

		JWTSecret: getenv("JWT_SECRET", "dev-secret-change-in-production"),

		DBHost:     getenv("DB_HOST", "localhost"),
		DBPort:     getenv("DB_PORT", "5432"),
		DBUser:     getenv("DB_USER", "habit"),
		DBPassword: getenv("DB_PASSWORD", "habit_secret"),
		DBName:     getenv("DB_NAME", "habit_tracker"),
		DBSSLMode:  getenv("DB_SSLMODE", "disable"),
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
