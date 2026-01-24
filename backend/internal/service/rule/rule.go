// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package rule

import (
	"context"
	"fmt"
	"time"

	appconfig "github.com/kk/elk-helper/backend/internal/config"
	"github.com/kk/elk-helper/backend/internal/models"
	"github.com/kk/elk-helper/backend/internal/repository/database"
	"github.com/kk/elk-helper/backend/internal/security"
	"gorm.io/gorm"
)

// Service provides rule management operations
type Service struct{}

// NewService creates a new rule service
func NewService() *Service {
	return &Service{}
}

// GetAll returns all rules
func (s *Service) GetAll() ([]models.Rule, error) {
	var rules []models.Rule
	db, cancel := database.WithTimeout(context.Background())
	defer cancel()

	if err := db.Preload("LarkConfig").Preload("ESConfig").Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("failed to get rules: %w", err)
	}

	if err := decryptRuleSecrets(rules); err != nil {
		return nil, err
	}
	return rules, nil
}

// GetAllPaged returns rules with pagination and total count
func (s *Service) GetAllPaged(page, pageSize int) ([]models.Rule, int64, error) {
	var rules []models.Rule
	var total int64

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}

	offset := (page - 1) * pageSize

	db, cancel := database.WithTimeout(context.Background())
	defer cancel()

	if err := db.Model(&models.Rule{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count rules: %w", err)
	}

	if err := db.Preload("LarkConfig").Preload("ESConfig").
		Order("id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&rules).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get rules: %w", err)
	}

	if err := decryptRuleSecrets(rules); err != nil {
		return nil, 0, err
	}

	return rules, total, nil
}

// GetByID returns a rule by ID
func (s *Service) GetByID(id uint) (*models.Rule, error) {
	var rule models.Rule
	db, cancel := database.WithTimeout(context.Background())
	defer cancel()

	if err := db.Preload("LarkConfig").Preload("ESConfig").First(&rule, id).Error; err != nil {
		return nil, fmt.Errorf("rule not found: %w", err)
	}
	if err := decryptRuleSecretInPlace(&rule); err != nil {
		return nil, err
	}
	return &rule, nil
}

// GetByName returns a rule by name (for deduplication check)
func (s *Service) GetByName(name string) (*models.Rule, error) {
	var rule models.Rule
	db, cancel := database.WithTimeout(context.Background())
	defer cancel()

	if err := db.Preload("LarkConfig").Preload("ESConfig").Where("name = ?", name).First(&rule).Error; err != nil {
		return nil, fmt.Errorf("rule not found: %w", err)
	}
	if err := decryptRuleSecretInPlace(&rule); err != nil {
		return nil, err
	}
	return &rule, nil
}

// Create creates a new rule
func (s *Service) Create(rule *models.Rule) error {
	db, cancel := database.WithTimeout(context.Background())
	defer cancel()

	if rule.LarkWebhook != "" {
		enc, err := security.MaybeEncrypt(rule.LarkWebhook, appconfig.AppConfig.Security.EncryptionKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt rule webhook: %w", err)
		}
		rule.LarkWebhook = enc
	}

	if err := db.Create(rule).Error; err != nil {
		return fmt.Errorf("failed to create rule: %w", err)
	}

	if rule.LarkWebhook != "" {
		plain, err := security.MaybeDecrypt(rule.LarkWebhook, appconfig.AppConfig.Security.EncryptionKey)
		if err == nil {
			rule.LarkWebhook = plain
		}
	}
	return nil
}

// Update updates an existing rule
func (s *Service) Update(id uint, rule *models.Rule) error {
	db, cancel := database.WithTimeout(context.Background())
	defer cancel()

	if rule.LarkWebhook != "" {
		enc, err := security.MaybeEncrypt(rule.LarkWebhook, appconfig.AppConfig.Security.EncryptionKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt rule webhook: %w", err)
		}
		rule.LarkWebhook = enc
	}

	if err := db.Model(&models.Rule{}).Where("id = ?", id).Updates(rule).Error; err != nil {
		return fmt.Errorf("failed to update rule: %w", err)
	}
	return nil
}

// Delete deletes a rule (hard delete - permanently removes from database)
// Also deletes all associated alerts
func (s *Service) Delete(id uint) error {
	// Start a transaction to ensure atomicity
	db, cancel := database.WithTimeout(context.Background())
	defer cancel()

	tx := db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	// Delete all associated alerts first (hard delete)
	if err := tx.Unscoped().Where("rule_id = ?", id).Delete(&models.Alert{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete associated alerts: %w", err)
	}

	// Then delete the rule (hard delete)
	if err := tx.Unscoped().Delete(&models.Rule{}, id).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete rule: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetEnabled returns all enabled rules
func (s *Service) GetEnabled() ([]models.Rule, error) {
	var rules []models.Rule
	db, cancel := database.WithTimeout(context.Background())
	defer cancel()

	if err := db.Preload("LarkConfig").Preload("ESConfig").Where("enabled = ?", true).Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("failed to get enabled rules: %w", err)
	}
	if err := decryptRuleSecrets(rules); err != nil {
		return nil, err
	}
	return rules, nil
}

// UpdateLastRunTime updates the last run time for a rule
func (s *Service) UpdateLastRunTime(id uint, lastRunTime *time.Time) error {
	db, cancel := database.WithTimeout(context.Background())
	defer cancel()

	if err := db.Model(&models.Rule{}).Where("id = ?", id).Update("last_run_time", lastRunTime).Error; err != nil {
		return fmt.Errorf("failed to update last run time: %w", err)
	}
	return nil
}

// EnableRule enables a rule
func (s *Service) EnableRule(id uint) error {
	db, cancel := database.WithTimeout(context.Background())
	defer cancel()

	if err := db.Model(&models.Rule{}).Where("id = ?", id).Update("enabled", true).Error; err != nil {
		return fmt.Errorf("failed to enable rule: %w", err)
	}
	return nil
}

// DisableRule disables a rule
func (s *Service) DisableRule(id uint) error {
	db, cancel := database.WithTimeout(context.Background())
	defer cancel()

	if err := db.Model(&models.Rule{}).Where("id = ?", id).Update("enabled", false).Error; err != nil {
		return fmt.Errorf("failed to disable rule: %w", err)
	}
	return nil
}

// ToggleEnabled toggles the enabled status of a rule
func (s *Service) ToggleEnabled(id uint) error {
	var rule models.Rule
	db, cancel := database.WithTimeout(context.Background())
	defer cancel()

	if err := db.First(&rule, id).Error; err != nil {
		return fmt.Errorf("rule not found: %w", err)
	}
	newStatus := !rule.Enabled
	if err := db.Model(&models.Rule{}).Where("id = ?", id).Update("enabled", newStatus).Error; err != nil {
		return fmt.Errorf("failed to toggle rule status: %w", err)
	}
	return nil
}

// IncrementRunCount increments the run count for a rule
func (s *Service) IncrementRunCount(id uint) error {
	db, cancel := database.WithTimeout(context.Background())
	defer cancel()

	if err := db.Model(&models.Rule{}).Where("id = ?", id).UpdateColumn("run_count", gorm.Expr("run_count + ?", 1)).Error; err != nil {
		return fmt.Errorf("failed to increment run count: %w", err)
	}
	return nil
}

// IncrementAlertCount increments the alert count for a rule
func (s *Service) IncrementAlertCount(id uint, count int64) error {
	db, cancel := database.WithTimeout(context.Background())
	defer cancel()

	if err := db.Model(&models.Rule{}).Where("id = ?", id).UpdateColumn("alert_count", gorm.Expr("alert_count + ?", count)).Error; err != nil {
		return fmt.Errorf("failed to increment alert count: %w", err)
	}
	return nil
}

// Clone clones an existing rule with a new name
func (s *Service) Clone(id uint, newName string) (*models.Rule, error) {
	// Get the original rule
	original, err := s.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get original rule: %w", err)
	}

	// Create a new rule with copied data
	clonedRule := models.Rule{
		Name:         newName,
		IndexPattern: original.IndexPattern,
		Queries:      original.Queries,
		Enabled:      original.Enabled, // Inherit enabled status from original rule
		Interval:     original.Interval,
		ESConfigID:   original.ESConfigID,
		LarkWebhook:  original.LarkWebhook,
		LarkConfigID: original.LarkConfigID,
		Description:  original.Description,
		// Statistics fields are not copied - they start fresh
		LastRunTime: nil,
		RunCount:    0,
		AlertCount:  0,
	}

	// Create the cloned rule
	db, cancel := database.WithTimeout(context.Background())
	defer cancel()

	if clonedRule.LarkWebhook != "" {
		enc, err := security.MaybeEncrypt(clonedRule.LarkWebhook, appconfig.AppConfig.Security.EncryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt rule webhook: %w", err)
		}
		clonedRule.LarkWebhook = enc
	}

	if err := db.Create(&clonedRule).Error; err != nil {
		return nil, fmt.Errorf("failed to create cloned rule: %w", err)
	}

	// Reload with associations
	return s.GetByID(clonedRule.ID)
}

func decryptRuleWebhooks(rules []models.Rule) error {
	return decryptRuleSecrets(rules)
}

func decryptRuleSecrets(rules []models.Rule) error {
	for i := range rules {
		if err := decryptRuleSecretInPlace(&rules[i]); err != nil {
			return err
		}
	}
	return nil
}

func decryptRuleSecretInPlace(rule *models.Rule) error {
	// Rule direct webhook
	if rule.LarkWebhook != "" {
		plain, err := security.MaybeDecrypt(rule.LarkWebhook, appconfig.AppConfig.Security.EncryptionKey)
		if err != nil {
			return fmt.Errorf("failed to decrypt rule webhook: %w", err)
		}
		rule.LarkWebhook = plain
	}

	// LarkConfig webhook (when rules use config rather than direct webhook)
	if rule.LarkConfig != nil && rule.LarkConfig.WebhookURL != "" {
		plain, err := security.MaybeDecrypt(rule.LarkConfig.WebhookURL, appconfig.AppConfig.Security.EncryptionKey)
		if err != nil {
			return fmt.Errorf("failed to decrypt lark config webhook: %w", err)
		}
		rule.LarkConfig.WebhookURL = plain
	}

	// ESConfig password (used by executor/query service)
	if rule.ESConfig != nil && rule.ESConfig.Password != "" {
		plain, err := security.MaybeDecrypt(rule.ESConfig.Password, appconfig.AppConfig.Security.EncryptionKey)
		if err != nil {
			return fmt.Errorf("failed to decrypt es config password: %w", err)
		}
		rule.ESConfig.Password = plain
	}

	return nil
}
