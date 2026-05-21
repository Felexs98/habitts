package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppPort string
	AppEnv  string

	JWTSecret        string
	JWTAccessTTL     time.Duration
	JWTRefreshTTL    time.Duration
	MigrationsPath   string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	RateLimitEnabled   bool
	RateLimitAuthRPM   int
	RateLimitAuthBurst int
	RateLimitAPIRPM    int
	RateLimitAPIBurst  int

	ArchivePurgeEnabled  bool
	ArchivePurgeDays     int
	ArchivePurgeInterval time.Duration
}

func Load() *Config {
	return &Config{
		AppPort: getenv("APP_PORT", "8080"),
		AppEnv:  getenv("APP_ENV", "development"),

		JWTSecret:      getenv("JWT_SECRET", "dev-secret-change-in-production"),
		JWTAccessTTL:   getenvDuration("JWT_ACCESS_TTL", 15*time.Minute),
		JWTRefreshTTL:  getenvDuration("JWT_REFRESH_TTL", 168*time.Hour),
		MigrationsPath: getenv("MIGRATIONS_PATH", "migrations"),

		DBHost:     getenv("DB_HOST", "localhost"),
		DBPort:     getenv("DB_PORT", "5432"),
		DBUser:     getenv("DB_USER", "habit"),
		DBPassword: getenv("DB_PASSWORD", "habit_secret"),
		DBName:     getenv("DB_NAME", "habit_tracker"),
		DBSSLMode:  getenv("DB_SSLMODE", "disable"),

		RateLimitEnabled:   getenvBool("RATE_LIMIT_ENABLED", true),
		RateLimitAuthRPM:   getenvInt("RATE_LIMIT_AUTH_RPM", 10),
		RateLimitAuthBurst: getenvInt("RATE_LIMIT_AUTH_BURST", 3),
		RateLimitAPIRPM:    getenvInt("RATE_LIMIT_API_RPM", 120),
		RateLimitAPIBurst:  getenvInt("RATE_LIMIT_API_BURST", 30),

		ArchivePurgeEnabled:  getenvBool("ARCHIVE_PURGE_ENABLED", true),
		ArchivePurgeDays:     getenvInt("ARCHIVE_PURGE_DAYS", 7),
		ArchivePurgeInterval: getenvDuration("ARCHIVE_PURGE_INTERVAL", 24*time.Hour),
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

// MigrateURL — DSN для golang-migrate (postgres://).
func (c *Config) MigrateURL() string {
	user := url.UserPassword(c.DBUser, c.DBPassword)
	return fmt.Sprintf(
		"postgres://%s@%s:%s/%s?sslmode=%s",
		user.String(), c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return fallback
	}
	return n
}

func getenvDuration(key string, fallback time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		return fallback
	}
	return d
}

func getenvBool(key string, fallback bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if v == "" {
		return fallback
	}
	switch v {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
