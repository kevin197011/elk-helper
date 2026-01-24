// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package lark_config

import (
	"context"
	"fmt"
	"time"

	appconfig "github.com/kk/elk-helper/backend/internal/config"
	"github.com/kk/elk-helper/backend/internal/models"
	"github.com/kk/elk-helper/backend/internal/repository/database"
	"github.com/kk/elk-helper/backend/internal/security"
)

// Service provides Lark configuration management operations
type Service struct{}

// NewService creates a new Lark config service
func NewService() *Service {
	return &Service{}
}

// GetAll returns all Lark configurations
func (s *Service) GetAll() ([]models.LarkConfig, error) {
	var configs []models.LarkConfig
	db, cancel := database.WithTimeout(context.Background())
	defer cancel()

	if err := db.Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("failed to get Lark configs: %w", err)
	}

	for i := range configs {
		plain, err := security.MaybeDecrypt(configs[i].WebhookURL, appconfig.AppConfig.Security.EncryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt webhook url: %w", err)
		}
		configs[i].WebhookURL = plain
	}

	return configs, nil
}

// GetByID returns a Lark config by ID
func (s *Service) GetByID(id uint) (*models.LarkConfig, error) {
	var cfg models.LarkConfig
	db, cancel := database.WithTimeout(context.Background())
	defer cancel()

	if err := db.First(&cfg, id).Error; err != nil {
		return nil, fmt.Errorf("Lark config not found: %w", err)
	}
	plain, err := security.MaybeDecrypt(cfg.WebhookURL, appconfig.AppConfig.Security.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt webhook url: %w", err)
	}
	cfg.WebhookURL = plain
	return &cfg, nil
}

// GetDefault returns the default Lark configuration
func (s *Service) GetDefault() (*models.LarkConfig, error) {
	var cfg models.LarkConfig
	db, cancel := database.WithTimeout(context.Background())
	defer cancel()

	if err := db.Where("is_default = ?", true).First(&cfg).Error; err != nil {
		// If no default config, return the first enabled one
		if err := db.Where("enabled = ?", true).First(&cfg).Error; err != nil {
			return nil, fmt.Errorf("no Lark config found: %w", err)
		}
	}
	plain, err := security.MaybeDecrypt(cfg.WebhookURL, appconfig.AppConfig.Security.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt webhook url: %w", err)
	}
	cfg.WebhookURL = plain
	return &cfg, nil
}

// GetByName returns a Lark config by name
func (s *Service) GetByName(name string) (*models.LarkConfig, error) {
	var cfg models.LarkConfig
	db, cancel := database.WithTimeout(context.Background())
	defer cancel()

	if err := db.Where("name = ?", name).First(&cfg).Error; err != nil {
		return nil, fmt.Errorf("Lark config not found: %w", err)
	}
	plain, err := security.MaybeDecrypt(cfg.WebhookURL, appconfig.AppConfig.Security.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt webhook url: %w", err)
	}
	cfg.WebhookURL = plain
	return &cfg, nil
}

// Create creates a new Lark configuration
func (s *Service) Create(config *models.LarkConfig) error {
	db, cancel := database.WithTimeout(context.Background())
	defer cancel()

	enc, err := security.MaybeEncrypt(config.WebhookURL, appconfig.AppConfig.Security.EncryptionKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt webhook url: %w", err)
	}
	config.WebhookURL = enc

	// If this is set as default, unset other defaults
	if config.IsDefault {
		if err := db.Model(&models.LarkConfig{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
			return fmt.Errorf("failed to unset other defaults: %w", err)
		}
	}

	if err := db.Create(config).Error; err != nil {
		return fmt.Errorf("failed to create Lark config: %w", err)
	}

	plain, err := security.MaybeDecrypt(config.WebhookURL, appconfig.AppConfig.Security.EncryptionKey)
	if err == nil {
		config.WebhookURL = plain
	}
	return nil
}

// Update updates an existing Lark configuration
func (s *Service) Update(id uint, config *models.LarkConfig) error {
	db, cancel := database.WithTimeout(context.Background())
	defer cancel()

	enc, err := security.MaybeEncrypt(config.WebhookURL, appconfig.AppConfig.Security.EncryptionKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt webhook url: %w", err)
	}
	config.WebhookURL = enc

	// If this is set as default, unset other defaults
	if config.IsDefault {
		if err := db.Model(&models.LarkConfig{}).Where("is_default = ? AND id != ?", true, id).Update("is_default", false).Error; err != nil {
			return fmt.Errorf("failed to unset other defaults: %w", err)
		}
	}

	if err := db.Model(&models.LarkConfig{}).Where("id = ?", id).Updates(config).Error; err != nil {
		return fmt.Errorf("failed to update Lark config: %w", err)
	}
	return nil
}

// Delete deletes a Lark configuration
func (s *Service) Delete(id uint) error {
	db, cancel := database.WithTimeout(context.Background())
	defer cancel()

	// Check if any rules are using this config
	var count int64
	if err := db.Model(&models.Rule{}).Where("lark_config_id = ?", id).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check rule usage: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("cannot delete: %d rules are using this config", count)
	}

	// Hard delete - permanently removes from database to free disk space
	if err := db.Unscoped().Delete(&models.LarkConfig{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete Lark config: %w", err)
	}
	return nil
}

// UpdateTestResult updates the test result for a Lark configuration
func (s *Service) UpdateTestResult(id uint, status string, errMsg string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"last_test_at": &now,
		"test_status":  status,
	}
	if errMsg != "" {
		updates["test_error"] = errMsg
	}

	db, cancel := database.WithTimeout(context.Background())
	defer cancel()

	if err := db.Model(&models.LarkConfig{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update test result: %w", err)
	}
	return nil
}

// SetDefault sets a configuration as the default one
func (s *Service) SetDefault(id uint) error {
	db, cancel := database.WithTimeout(context.Background())
	defer cancel()

	// Unset all defaults
	if err := db.Model(&models.LarkConfig{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
		return fmt.Errorf("failed to unset other defaults: %w", err)
	}

	// Set this one as default
	if err := db.Model(&models.LarkConfig{}).Where("id = ?", id).Update("is_default", true).Error; err != nil {
		return fmt.Errorf("failed to set default: %w", err)
	}

	return nil
}
