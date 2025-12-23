// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config represents application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	ES       ElasticsearchConfig
	Worker   WorkerConfig
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Port        string
	Host        string
	Mode        string // debug, release
	CORSOrigins []string
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// ElasticsearchConfig represents Elasticsearch connection configuration
type ElasticsearchConfig struct {
	URL      string
	Username string
	Password string
}

// WorkerConfig represents worker configuration
type WorkerConfig struct {
	Enabled        bool
	CheckInterval  int // seconds - how often to check for rule changes
	RetryTimes     int
	BatchSize      int
	MaxConcurrency int // max concurrent rule executions
}

var AppConfig *Config

// Load loads configuration from environment variables
func Load() error {
	_ = godotenv.Load()

	AppConfig = &Config{
		Server: ServerConfig{
			Port:        getEnv("SERVER_PORT", "8080"),
			Host:        getEnv("SERVER_HOST", "0.0.0.0"),
			Mode:        getEnv("GIN_MODE", "release"),
			CORSOrigins: getEnvSlice("CORS_ORIGINS", []string{"http://localhost:3000", "http://localhost:5173", "http://localhost"}),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "elk_helper"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		ES: ElasticsearchConfig{
			URL:      getEnv("ES_URL", "http://localhost:9200"),
			Username: getEnv("ES_USERNAME", ""),
			Password: getEnv("ES_PASSWORD", ""),
		},
		Worker: WorkerConfig{
			Enabled:        getEnv("WORKER_ENABLED", "true") == "true",
			CheckInterval:  parseInt(getEnv("WORKER_CHECK_INTERVAL", "30")),
			RetryTimes:     parseInt(getEnv("WORKER_RETRY_TIMES", "3")),
			BatchSize:      parseInt(getEnv("WORKER_BATCH_SIZE", "200")),
			MaxConcurrency: parseInt(getEnv("WORKER_MAX_CONCURRENCY", "10")),
		},
	}

	return nil
}

func parseInt(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	if result <= 0 {
		return 30 // default
	}
	return result
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvSlice(key string, defaultValue []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	// Split by comma
	result := []string{}
	for _, v := range splitByComma(value) {
		if v != "" {
			result = append(result, v)
		}
	}
	if len(result) == 0 {
		return defaultValue
	}
	return result
}

func splitByComma(s string) []string {
	parts := []string{}
	start := 0
	for i, char := range s {
		if char == ',' {
			if i > start {
				parts = append(parts, s[start:i])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		parts = append(parts, s[start:])
	}
	return parts
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.ES.URL == "" {
		return fmt.Errorf("ES_URL is required")
	}
	return nil
}
