// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package system_config

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kk/elk-helper/backend/internal/models"
	"github.com/kk/elk-helper/backend/internal/repository/database"
	"gorm.io/gorm"
)

// Service provides system configuration management operations
type Service struct{}

// NewService creates a new system config service
func NewService() *Service {
	return &Service{}
}

// GetCleanupConfig returns the cleanup task configuration
func (s *Service) GetCleanupConfig() (*models.CleanupConfig, error) {
	config, err := s.getByKey("cleanup_config")
	if err != nil {
		// Database error (not just not found)
		return nil, fmt.Errorf("failed to get cleanup config: %w", err)
	}

	// If config not found, return default
	if config == nil {
		return &models.CleanupConfig{
			Enabled:       true,
			Hour:          3,
			Minute:        0,
			RetentionDays: 90,
		}, nil
	}

	var cleanupConfig models.CleanupConfig
	if err := json.Unmarshal([]byte(config.Value), &cleanupConfig); err != nil {
		return nil, fmt.Errorf("failed to parse cleanup config: %w", err)
	}

	return &cleanupConfig, nil
}

// UpdateCleanupConfig updates the cleanup task configuration
func (s *Service) UpdateCleanupConfig(config *models.CleanupConfig) error {
	// Validate
	if config.Hour < 0 || config.Hour > 23 {
		return fmt.Errorf("hour must be between 0 and 23")
	}
	if config.Minute < 0 || config.Minute > 59 {
		return fmt.Errorf("minute must be between 0 and 59")
	}
	if config.RetentionDays < 1 {
		return fmt.Errorf("retention_days must be at least 1")
	}

	value, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Update or create
	existing, err := s.getByKey("cleanup_config")
	if err != nil {
		// Database error (not just not found)
		return fmt.Errorf("failed to get cleanup config: %w", err)
	}

	if existing == nil {
		// Create new
		newConfig := &models.SystemConfig{
			Key:         "cleanup_config",
			Value:       string(value),
			Description: "定时清理任务配置：执行时间、保留天数",
		}
		if err := database.DB.Create(newConfig).Error; err != nil {
			return fmt.Errorf("failed to create cleanup config: %w", err)
		}
		return nil
	}

	// Update existing
	existing.Value = string(value)
	if err := database.DB.Save(existing).Error; err != nil {
		return fmt.Errorf("failed to update cleanup config: %w", err)
	}

	return nil
}

// getByKey gets a system config by key
// Returns nil, nil if not found (not an error case)
func (s *Service) getByKey(key string) (*models.SystemConfig, error) {
	var config models.SystemConfig
	if err := database.DB.Where("key = ?", key).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Not found is not an error, return nil config
			return nil, nil
		}
		// Other database errors are real errors
		return nil, err
	}
	return &config, nil
}

