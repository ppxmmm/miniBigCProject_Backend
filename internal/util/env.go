package util

import (
	"fmt"
	"os"
)

// Config contains runtime configuration loaded from environment variables.
type Config struct {
	AppName string
	Port    string
	DB      DBConfig
}

// DBConfig contains database connection settings.
type DBConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
}

// Env returns an environment variable or the provided fallback.
func Env(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

// LoadConfig loads application configuration from environment variables.
func LoadConfig() Config {
	return Config{
		AppName: Env("APP_NAME", "My Backend Server"),
		Port:    Env("PORT", "5001"),
		DB: DBConfig{
			User:     Env("DB_USER", "admin"),
			Password: Env("DB_PASSWORD", "root"),
			Host:     Env("DB_HOST", "localhost"),
			Port:     Env("DB_PORT", "5433"),
			Name:     Env("DB_NAME", "test_db"),
		},
	}
}

// DSN builds a PostgreSQL data source name for database/sql.
func (config DBConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Name,
	)
}
