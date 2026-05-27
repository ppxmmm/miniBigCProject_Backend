package util

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config contains runtime configuration loaded from environment variables.
type Config struct {
	AppName     string
	Port        string
	CORSOrigins string
	DB          DBConfig
	AI          AIConfig
}

// DBConfig contains database connection settings.
type DBConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
}

// AIConfig contains Gemini API settings.
type AIConfig struct {
	GeminiAPIKey string
	GeminiModel  string
	Timeout      time.Duration
	MCP          MCPConfig
}

// MCPConfig contains settings for the local MCP stdio server.
type MCPConfig struct {
	Enabled        bool
	Command        string
	Args           []string
	BackendBaseURL string
	Timeout        time.Duration
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
	mcpTimeoutSeconds, err := strconv.Atoi(Env("MCP_TIMEOUT_SECONDS", "10"))
	if err != nil || mcpTimeoutSeconds <= 0 {
		mcpTimeoutSeconds = 10
	}
	aiTimeoutSeconds, err := strconv.Atoi(Env("AI_TIMEOUT_SECONDS", "90"))
	if err != nil || aiTimeoutSeconds <= 0 {
		aiTimeoutSeconds = 90
	}

	return Config{
		AppName: Env("APP_NAME", "Mini BigC API"),
		Port:    Env("PORT", "5001"),
		CORSOrigins: Env(
			"CORS_ORIGINS",
			"http://localhost:3000,http://localhost:3001,http://localhost:5173,http://127.0.0.1:3000,http://127.0.0.1:5173",
		),
		DB: DBConfig{
			User:     Env("DB_USER", "admin"),
			Password: Env("DB_PASSWORD", "root"),
			Host:     Env("DB_HOST", "localhost"),
			Port:     Env("DB_PORT", "5433"),
			Name:     Env("DB_NAME", "test_db"),
		},
		AI: AIConfig{
			GeminiAPIKey: Env("GEMINI_API_KEY", ""),
			GeminiModel:  Env("GEMINI_MODEL", "gemini-2.5-flash"),
			Timeout:      time.Duration(aiTimeoutSeconds) * time.Second,
			MCP: MCPConfig{
				Enabled:        strings.ToLower(Env("MCP_ENABLED", "true")) != "false",
				Command:        Env("MCP_SERVER_COMMAND", "mcp_server/venv/bin/python"),
				Args:           strings.Fields(Env("MCP_SERVER_ARGS", "mcp_server/server.py")),
				BackendBaseURL: Env("MINIBIGC_API_BASE_URL", "http://localhost:"+Env("PORT", "5001")),
				Timeout:        time.Duration(mcpTimeoutSeconds) * time.Second,
			},
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
