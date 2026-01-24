// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config represents application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	ES       ElasticsearchConfig
	Worker   WorkerConfig
	Auth     AuthConfig
	Security SecurityConfig
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
	// QueryTimeoutSeconds controls per-operation DB timeout (best-effort guardrail).
	QueryTimeoutSeconds int
}

// ElasticsearchConfig represents Elasticsearch connection configuration
type ElasticsearchConfig struct {
	URL                 string
	Username            string
	Password            string
	UseSSL              bool
	SkipVerify          bool
	CACertificate       string
	QueryTimeoutSeconds int
}

// WorkerConfig represents worker configuration
type WorkerConfig struct {
	Enabled        bool
	CheckInterval  int // seconds - how often to check for rule changes
	RetryTimes     int
	BatchSize      int
	MaxConcurrency int // max concurrent rule executions
	// AlertSendTimeoutSeconds controls the max duration for sending a single alert notification.
	AlertSendTimeoutSeconds int
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	JWTSecret               string
	LoginRateLimitPerMinute int
	LoginRateLimitBurst     int
}

// SecurityConfig represents security related settings.
type SecurityConfig struct {
	EncryptionKeyBase64 string
	EncryptionKey       []byte
}

var AppConfig *Config

// Load loads configuration from environment variables
func Load() error {
	_ = godotenv.Load()

	mode := getEnv("GIN_MODE", "release")
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" && mode == "debug" {
		// 开发模式下提供默认值，避免本地启动卡住；生产环境会在 Validate 中强校验并拒绝默认值
		jwtSecret = "elk-helper-dev-secret-not-for-production"
	}

	AppConfig = &Config{
		Server: ServerConfig{
			Port:        getEnv("SERVER_PORT", "8080"),
			Host:        getEnv("SERVER_HOST", "0.0.0.0"),
			Mode:        mode,
			CORSOrigins: getEnvSlice("CORS_ORIGINS", []string{"http://localhost:3000", "http://localhost:5173", "http://localhost"}),
		},
		Database: DatabaseConfig{
			Host:                getEnv("DB_HOST", "localhost"),
			Port:                getEnv("DB_PORT", "5432"),
			User:                getEnv("DB_USER", "postgres"),
			Password:            getEnv("DB_PASSWORD", "postgres"),
			DBName:              getEnv("DB_NAME", "elk_helper"),
			SSLMode:             getEnv("DB_SSLMODE", "disable"),
			QueryTimeoutSeconds: parseIntWithDefault(getEnv("DB_QUERY_TIMEOUT_SECONDS", "5"), 5),
		},
		ES: ElasticsearchConfig{
			URL:                 getEnv("ES_URL", "http://localhost:9200"),
			Username:            getEnv("ES_USERNAME", ""),
			Password:            getEnv("ES_PASSWORD", ""),
			UseSSL:              parseBoolWithDefault(getEnv("ES_USE_SSL", ""), strings.HasPrefix(getEnv("ES_URL", "http://localhost:9200"), "https://")),
			SkipVerify:          parseBoolWithDefault(getEnv("ES_SKIP_VERIFY", ""), false),
			CACertificate:       getEnv("ES_CA_CERTIFICATE", ""),
			QueryTimeoutSeconds: parseIntWithDefault(getEnv("ES_QUERY_TIMEOUT_SECONDS", "30"), 30),
		},
		Worker: WorkerConfig{
			Enabled:                 getEnv("WORKER_ENABLED", "true") == "true",
			CheckInterval:           parseIntWithDefault(getEnv("WORKER_CHECK_INTERVAL", "30"), 30),
			RetryTimes:              parseIntWithDefault(getEnv("WORKER_RETRY_TIMES", "3"), 3),
			BatchSize:               parseIntWithDefault(getEnv("WORKER_BATCH_SIZE", "200"), 200),
			MaxConcurrency:          parseIntWithDefault(getEnv("WORKER_MAX_CONCURRENCY", "10"), 10),
			AlertSendTimeoutSeconds: parseIntWithDefault(getEnv("ALERT_SEND_TIMEOUT_SECONDS", "20"), 20),
		},
		Auth: AuthConfig{
			JWTSecret:               jwtSecret,
			LoginRateLimitPerMinute: parseIntWithDefault(getEnv("LOGIN_RATE_LIMIT_PER_MINUTE", "60"), 60),
			LoginRateLimitBurst:     parseIntWithDefault(getEnv("LOGIN_RATE_LIMIT_BURST", "20"), 20),
		},
		Security: SecurityConfig{
			EncryptionKeyBase64: getEnv("APP_ENCRYPTION_KEY", ""),
		},
	}

	if AppConfig.Security.EncryptionKeyBase64 != "" {
		keyBytes, err := base64.RawStdEncoding.DecodeString(AppConfig.Security.EncryptionKeyBase64)
		if err != nil {
			// fallback to standard base64
			keyBytes, err = base64.StdEncoding.DecodeString(AppConfig.Security.EncryptionKeyBase64)
		}
		if err != nil {
			return fmt.Errorf("invalid APP_ENCRYPTION_KEY (base64 decode failed): %w", err)
		}
		AppConfig.Security.EncryptionKey = keyBytes
	}

	return nil
}

func parseIntWithDefault(s string, defaultValue int) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	if result <= 0 {
		return defaultValue
	}
	return result
}

func parseBoolWithDefault(s string, defaultValue bool) bool {
	if s == "" {
		return defaultValue
	}
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return defaultValue
	}
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
		item := strings.TrimSpace(v)
		if item != "" {
			result = append(result, item)
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

	if err := validateJWTSecret(c.Server.Mode, c.Auth.JWTSecret); err != nil {
		return err
	}

	if c.Security.EncryptionKeyBase64 != "" && len(c.Security.EncryptionKey) != 32 {
		return fmt.Errorf("APP_ENCRYPTION_KEY must decode to 32 bytes (got %d)", len(c.Security.EncryptionKey))
	}

	return nil
}

func validateJWTSecret(mode, secret string) error {
	if secret == "" {
		if mode == "debug" {
			// debug 模式允许为空（Load 中会兜底默认值），这里保持兼容
			return nil
		}
		return errors.New("JWT_SECRET is required")
	}

	// release 模式强校验：最小长度 + 禁止默认值
	if mode == "release" {
		if len(secret) < 32 {
			return fmt.Errorf("JWT_SECRET is too short (minimum 32 characters required)")
		}

		if secret == "elk-helper-secret-key-change-in-production" || secret == "elk-helper-dev-secret-not-for-production" {
			return fmt.Errorf("JWT_SECRET must be changed for production deployments")
		}
	}

	return nil
}
