// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package es_config

import (
	"fmt"
	"time"

	"github.com/kk/elk-helper/backend/internal/models"
	"github.com/kk/elk-helper/backend/internal/repository/database"
)

// Service provides ES configuration management operations
type Service struct{}

// NewService creates a new ES config service
func NewService() *Service {
	return &Service{}
}

// GetAll returns all ES configurations
func (s *Service) GetAll() ([]models.ESConfig, error) {
	var configs []models.ESConfig
	if err := database.DB.Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("failed to get ES configs: %w", err)
	}
	return configs, nil
}

// GetByID returns an ES config by ID
func (s *Service) GetByID(id uint) (*models.ESConfig, error) {
	var config models.ESConfig
	if err := database.DB.First(&config, id).Error; err != nil {
		return nil, fmt.Errorf("ES config not found: %w", err)
	}
	return &config, nil
}

// GetDefault returns the default ES configuration
func (s *Service) GetDefault() (*models.ESConfig, error) {
	var config models.ESConfig
	if err := database.DB.Where("is_default = ?", true).First(&config).Error; err != nil {
		// If no default config, return the first enabled one
		if err := database.DB.Where("enabled = ?", true).First(&config).Error; err != nil {
			return nil, fmt.Errorf("no ES config found: %w", err)
		}
	}
	return &config, nil
}

// GetByName returns an ES config by name
func (s *Service) GetByName(name string) (*models.ESConfig, error) {
	var config models.ESConfig
	if err := database.DB.Where("name = ?", name).First(&config).Error; err != nil {
		return nil, fmt.Errorf("ES config not found: %w", err)
	}
	return &config, nil
}

// Create creates a new ES configuration
func (s *Service) Create(config *models.ESConfig) error {
	// Check if a config with the same name exists (now we use hard delete, so only check active configs)
	var existingConfig models.ESConfig
	err := database.DB.Where("name = ?", config.Name).First(&existingConfig).Error
	if err == nil {
		// Config with same name already exists
		return fmt.Errorf("configuration name already exists")
	}

	// If this is set as default, unset other defaults
	if config.IsDefault {
		if err := database.DB.Model(&models.ESConfig{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
			return fmt.Errorf("failed to unset other defaults: %w", err)
		}
	}

	if err := database.DB.Create(config).Error; err != nil {
		return fmt.Errorf("failed to create ES config: %w", err)
	}
	return nil
}

// Update updates an existing ES configuration
func (s *Service) Update(id uint, config *models.ESConfig) error {
	// If this is set as default, unset other defaults
	if config.IsDefault {
		if err := database.DB.Model(&models.ESConfig{}).Where("is_default = ? AND id != ?", true, id).Update("is_default", false).Error; err != nil {
			return fmt.Errorf("failed to unset other defaults: %w", err)
		}
	}

	// Build update map, excluding password if it's empty
	updateData := map[string]interface{}{
		"name":         config.Name,
		"url":          config.URL,
		"username":     config.Username,
		"use_ssl":      config.UseSSL,
		"skip_verify":  config.SkipVerify,
		"is_default":   config.IsDefault,
		"description":  config.Description,
		"enabled":      config.Enabled,
	}

	// Only update password if it's provided (not empty)
	if config.Password != "" {
		updateData["password"] = config.Password
	}

	// Only update CA certificate if it's provided
	if config.CACertificate != "" {
		updateData["ca_certificate"] = config.CACertificate
	}

	if err := database.DB.Model(&models.ESConfig{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return fmt.Errorf("failed to update ES config: %w", err)
	}
	return nil
}

// Delete deletes an ES configuration (hard delete - permanently removes from database)
func (s *Service) Delete(id uint) error {
	if err := database.DB.Unscoped().Delete(&models.ESConfig{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete ES config: %w", err)
	}
	return nil
}

// UpdateTestResult updates the test result for an ES configuration
func (s *Service) UpdateTestResult(id uint, status string, errMsg string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"last_test_at": &now,
		"test_status":  status,
	}
	if errMsg != "" {
		updates["test_error"] = errMsg
	}

	if err := database.DB.Model(&models.ESConfig{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update test result: %w", err)
	}
	return nil
}

// SetDefault sets a configuration as the default one
func (s *Service) SetDefault(id uint) error {
	// Unset all defaults
	if err := database.DB.Model(&models.ESConfig{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
		return fmt.Errorf("failed to unset other defaults: %w", err)
	}

	// Set this one as default
	if err := database.DB.Model(&models.ESConfig{}).Where("id = ?", id).Update("is_default", true).Error; err != nil {
		return fmt.Errorf("failed to set default: %w", err)
	}

	return nil
}

