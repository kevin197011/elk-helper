// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package rule

import (
	"fmt"
	"time"

	"github.com/kk/elk-helper/backend/internal/models"
	"github.com/kk/elk-helper/backend/internal/repository/database"
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
	if err := database.DB.Preload("LarkConfig").Preload("ESConfig").Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("failed to get rules: %w", err)
	}
	return rules, nil
}

// GetByID returns a rule by ID
func (s *Service) GetByID(id uint) (*models.Rule, error) {
	var rule models.Rule
	if err := database.DB.Preload("LarkConfig").Preload("ESConfig").First(&rule, id).Error; err != nil {
		return nil, fmt.Errorf("rule not found: %w", err)
	}
	return &rule, nil
}

// Create creates a new rule
func (s *Service) Create(rule *models.Rule) error {
	if err := database.DB.Create(rule).Error; err != nil {
		return fmt.Errorf("failed to create rule: %w", err)
	}
	return nil
}

// Update updates an existing rule
func (s *Service) Update(id uint, rule *models.Rule) error {
	if err := database.DB.Model(&models.Rule{}).Where("id = ?", id).Updates(rule).Error; err != nil {
		return fmt.Errorf("failed to update rule: %w", err)
	}
	return nil
}

// Delete deletes a rule (hard delete - permanently removes from database)
func (s *Service) Delete(id uint) error {
	if err := database.DB.Unscoped().Delete(&models.Rule{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete rule: %w", err)
	}
	return nil
}

// GetEnabled returns all enabled rules
func (s *Service) GetEnabled() ([]models.Rule, error) {
	var rules []models.Rule
	if err := database.DB.Preload("LarkConfig").Preload("ESConfig").Where("enabled = ?", true).Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("failed to get enabled rules: %w", err)
	}
	return rules, nil
}

// UpdateLastRunTime updates the last run time for a rule
func (s *Service) UpdateLastRunTime(id uint, lastRunTime *time.Time) error {
	if err := database.DB.Model(&models.Rule{}).Where("id = ?", id).Update("last_run_time", lastRunTime).Error; err != nil {
		return fmt.Errorf("failed to update last run time: %w", err)
	}
	return nil
}

// EnableRule enables a rule
func (s *Service) EnableRule(id uint) error {
	if err := database.DB.Model(&models.Rule{}).Where("id = ?", id).Update("enabled", true).Error; err != nil {
		return fmt.Errorf("failed to enable rule: %w", err)
	}
	return nil
}

// DisableRule disables a rule
func (s *Service) DisableRule(id uint) error {
	if err := database.DB.Model(&models.Rule{}).Where("id = ?", id).Update("enabled", false).Error; err != nil {
		return fmt.Errorf("failed to disable rule: %w", err)
	}
	return nil
}

// ToggleEnabled toggles the enabled status of a rule
func (s *Service) ToggleEnabled(id uint) error {
	var rule models.Rule
	if err := database.DB.First(&rule, id).Error; err != nil {
		return fmt.Errorf("rule not found: %w", err)
	}
	newStatus := !rule.Enabled
	if err := database.DB.Model(&models.Rule{}).Where("id = ?", id).Update("enabled", newStatus).Error; err != nil {
		return fmt.Errorf("failed to toggle rule status: %w", err)
	}
	return nil
}

// IncrementRunCount increments the run count for a rule
func (s *Service) IncrementRunCount(id uint) error {
	if err := database.DB.Model(&models.Rule{}).Where("id = ?", id).UpdateColumn("run_count", gorm.Expr("run_count + ?", 1)).Error; err != nil {
		return fmt.Errorf("failed to increment run count: %w", err)
	}
	return nil
}

// IncrementAlertCount increments the alert count for a rule
func (s *Service) IncrementAlertCount(id uint, count int64) error {
	if err := database.DB.Model(&models.Rule{}).Where("id = ?", id).UpdateColumn("alert_count", gorm.Expr("alert_count + ?", count)).Error; err != nil {
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
		Enabled:      false, // Start with disabled status for safety
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
	if err := database.DB.Create(&clonedRule).Error; err != nil {
		return nil, fmt.Errorf("failed to create cloned rule: %w", err)
	}

	// Reload with associations
	return s.GetByID(clonedRule.ID)
}

