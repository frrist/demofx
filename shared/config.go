package shared

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config holds application-wide configuration
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	App      AppConfig      `json:"app"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	MaxConnections int    `json:"max_connections"`
	Timeout        int    `json:"timeout_seconds"`
	CacheSize      int    `json:"cache_size"`
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Environment string `json:"environment"`
	LogLevel    string `json:"log_level"`
	Features    map[string]bool `json:"features"`
}

// LoadConfig loads configuration from file or returns defaults
func LoadConfig(path string) (*Config, error) {
	// Default configuration
	config := &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: "8080",
		},
		Database: DatabaseConfig{
			MaxConnections: 10,
			Timeout:        30,
			CacheSize:      100,
		},
		App: AppConfig{
			Environment: "development",
			LogLevel:    "info",
			Features: map[string]bool{
				"cache_enabled": true,
				"rate_limiting": false,
			},
		},
	}

	// Override from environment variables if present
	if port := os.Getenv("PORT"); port != "" {
		config.Server.Port = port
	}
	if host := os.Getenv("HOST"); host != "" {
		config.Server.Host = host
	}

	// Try to load from file if path provided
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			// Return default config if file doesn't exist
			if os.IsNotExist(err) {
				return config, nil
			}
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := json.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	return config, nil
}