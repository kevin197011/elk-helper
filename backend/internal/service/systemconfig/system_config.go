// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package system_config

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

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
			Enabled:             true,
			Hour:                3,
			Minute:              0,
			RetentionDays:       90,
			LastExecutionStatus: "never",
		}, nil
	}

	var cleanupConfig models.CleanupConfig
	if err := json.Unmarshal([]byte(config.Value), &cleanupConfig); err != nil {
		return nil, fmt.Errorf("failed to parse cleanup config: %w", err)
	}

	// Set default execution status if not set (for existing configs without this field)
	if cleanupConfig.LastExecutionStatus == "" {
		cleanupConfig.LastExecutionStatus = "never"
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

	// Get existing config to preserve execution status fields
	// These fields should only be updated by UpdateCleanupExecutionStatus, not by regular config updates
	existingConfig, err := s.GetCleanupConfig()
	if err != nil {
		return fmt.Errorf("failed to get existing cleanup config: %w", err)
	}

	// Preserve execution status fields from existing config
	// Frontend only sends enabled, hour, minute, retention_days, so execution status fields will be zero values
	// We preserve them unless they are explicitly set (non-zero values)
	if existingConfig != nil {
		// Preserve LastExecutionStatus if new config doesn't have it (empty string means not provided)
		// But if it's explicitly set to a non-empty value, use the new value (for UpdateCleanupExecutionStatus)
		if config.LastExecutionStatus == "" {
			config.LastExecutionStatus = existingConfig.LastExecutionStatus
		}
		// Preserve LastExecutionTime if new config doesn't have it
		if config.LastExecutionTime == nil {
			config.LastExecutionTime = existingConfig.LastExecutionTime
		}
		// Preserve LastExecutionResult if new config doesn't have it
		if config.LastExecutionResult == "" {
			config.LastExecutionResult = existingConfig.LastExecutionResult
		}
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

// UpdateCleanupExecutionStatus updates the cleanup task execution status
func (s *Service) UpdateCleanupExecutionStatus(status string, result string) error {
	config, err := s.GetCleanupConfig()
	if err != nil {
		return fmt.Errorf("failed to get cleanup config: %w", err)
	}

	now := time.Now()
	config.LastExecutionStatus = status
	config.LastExecutionTime = &now
	config.LastExecutionResult = result

	return s.UpdateCleanupConfig(config)
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

