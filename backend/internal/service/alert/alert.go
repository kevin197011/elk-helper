// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package alert

import (
	"fmt"
	"time"

	"github.com/kk/elk-helper/backend/internal/models"
	"github.com/kk/elk-helper/backend/internal/repository/database"
)

// Service provides alert management operations
type Service struct{}

// NewService creates a new alert service
func NewService() *Service {
	return &Service{}
}

// Create creates a new alert record
func (s *Service) Create(alert *models.Alert) error {
	if err := database.DB.Create(alert).Error; err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}
	return nil
}

// GetAll returns all alerts with pagination (without logs for performance)
func (s *Service) GetAll(page, pageSize int) ([]models.Alert, int64, error) {
	var alerts []models.Alert
	var total int64

	offset := (page - 1) * pageSize

	if err := database.DB.Model(&models.Alert{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count alerts: %w", err)
	}

	// Optimize: Don't load logs field in list view for performance
	// Logs can be hundreds of KB or even MBs, causing slow page loads
	// Only load logs when viewing individual alert details
	if err := database.DB.Preload("Rule").
		Select("id", "created_at", "rule_id", "index_name", "log_count", "time_range", "status", "error_msg").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&alerts).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get alerts: %w", err)
	}

	return alerts, total, nil
}

// GetByID returns an alert by ID with limited logs (max 10 for performance)
func (s *Service) GetByID(id uint) (*models.Alert, error) {
	var alert models.Alert
	if err := database.DB.Preload("Rule").First(&alert, id).Error; err != nil {
		return nil, fmt.Errorf("alert not found: %w", err)
	}

	// Limit logs to first 10 entries for performance
	// This prevents loading huge JSON blobs that can be hundreds of KB or even MBs
	if len(alert.Logs) > 10 {
		alert.Logs = alert.Logs[:10]
	}

	return &alert, nil
}

// GetByRuleID returns alerts for a specific rule
func (s *Service) GetByRuleID(ruleID uint, limit int) ([]models.Alert, error) {
	var alerts []models.Alert
	query := database.DB.Where("rule_id = ?", ruleID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&alerts).Error; err != nil {
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}
	return alerts, nil
}

// Delete deletes an alert (hard delete - permanently removes from database)
func (s *Service) Delete(id uint) error {
	if err := database.DB.Unscoped().Delete(&models.Alert{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete alert: %w", err)
	}
	return nil
}

// GetStats returns alert statistics
func (s *Service) GetStats(duration time.Duration) (map[string]interface{}, error) {
	var totalCount int64
	var sentCount int64
	var failedCount int64

	since := time.Now().Add(-duration)

	if err := database.DB.Model(&models.Alert{}).
		Where("created_at >= ?", since).
		Count(&totalCount).Error; err != nil {
		return nil, err
	}

	if err := database.DB.Model(&models.Alert{}).
		Where("created_at >= ? AND status = ?", since, models.AlertStatusSent).
		Count(&sentCount).Error; err != nil {
		return nil, err
	}

	if err := database.DB.Model(&models.Alert{}).
		Where("created_at >= ? AND status = ?", since, models.AlertStatusFailed).
		Count(&failedCount).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total":  totalCount,
		"sent":   sentCount,
		"failed": failedCount,
	}, nil
}

// BatchDelete deletes multiple alerts (hard delete - permanently removes from database)
func (s *Service) BatchDelete(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	if err := database.DB.Unscoped().Where("id IN ?", ids).Delete(&models.Alert{}).Error; err != nil {
		return fmt.Errorf("failed to batch delete alerts: %w", err)
	}
	return nil
}

// CleanupOldData deletes alerts older than the specified duration (hard delete - permanently removes from database)
func (s *Service) CleanupOldData(olderThan time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-olderThan)

	result := database.DB.Unscoped().Where("created_at < ?", cutoffTime).Delete(&models.Alert{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup old alerts: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// RuleAlertStats represents alert statistics for a single rule
type RuleAlertStats struct {
	RuleID    uint       `json:"rule_id"`
	RuleName  string     `json:"rule_name"`
	Total     int64      `json:"total"`
	Sent      int64      `json:"sent"`
	Failed    int64      `json:"failed"`
	LastAlert *time.Time `json:"last_alert"`
}

// TimeSeriesDataPoint represents a single data point in time series
type TimeSeriesDataPoint struct {
	Time  string `json:"time"`
	Value int64  `json:"value"`
}

// RuleTimeSeriesStats represents time series statistics for a single rule
type RuleTimeSeriesStats struct {
	RuleID     uint                  `json:"rule_id"`
	RuleName   string                `json:"rule_name"`
	Total      int64                 `json:"total"`
	DataPoints []TimeSeriesDataPoint `json:"data_points"`
}

// GetRuleAlertStats returns alert statistics grouped by rule
func (s *Service) GetRuleAlertStats(duration time.Duration) ([]RuleAlertStats, error) {
	since := time.Now().Add(-duration)

	type QueryResult struct {
		RuleID    uint
		RuleName  string
		Total     int64
		Sent      int64
		Failed    int64
		LastAlert *time.Time
	}

	var results []QueryResult

	// Query to get stats per rule
	err := database.DB.Model(&models.Alert{}).
		Select(`
			alerts.rule_id,
			rules.name as rule_name,
			COUNT(*) as total,
			SUM(CASE WHEN alerts.status = 'sent' THEN 1 ELSE 0 END) as sent,
			SUM(CASE WHEN alerts.status = 'failed' THEN 1 ELSE 0 END) as failed,
			MAX(alerts.created_at) as last_alert
		`).
		Joins("LEFT JOIN rules ON rules.id = alerts.rule_id").
		Where("alerts.created_at >= ?", since).
		Group("alerts.rule_id, rules.name").
		Order("total DESC").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get rule alert stats: %w", err)
	}

	stats := make([]RuleAlertStats, len(results))
	for i, r := range results {
		stats[i] = RuleAlertStats{
			RuleID:    r.RuleID,
			RuleName:  r.RuleName,
			Total:     r.Total,
			Sent:      r.Sent,
			Failed:    r.Failed,
			LastAlert: r.LastAlert,
		}
	}

	return stats, nil
}

// GetRuleTimeSeriesStats returns time series alert statistics for top rules
func (s *Service) GetRuleTimeSeriesStats(duration time.Duration, intervalMinutes int) ([]RuleTimeSeriesStats, error) {
	since := time.Now().Add(-duration)
	now := time.Now()

	// Get top 5 rules by alert count
	var topRules []struct {
		RuleID   uint
		RuleName string
		Total    int64
	}

	err := database.DB.Model(&models.Alert{}).
		Select("alerts.rule_id, rules.name as rule_name, COUNT(*) as total").
		Joins("LEFT JOIN rules ON rules.id = alerts.rule_id").
		Where("alerts.created_at >= ?", since).
		Group("alerts.rule_id, rules.name").
		Order("total DESC").
		Limit(5).
		Scan(&topRules).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get top rules: %w", err)
	}

	if len(topRules) == 0 {
		return []RuleTimeSeriesStats{}, nil
	}

	// Generate time buckets
	numBuckets := int(duration.Minutes()) / intervalMinutes
	if numBuckets > 24 {
		numBuckets = 24 // Max 24 points for readability
	}
	if numBuckets < 6 {
		numBuckets = 6 // Min 6 points
	}

	result := make([]RuleTimeSeriesStats, len(topRules))

	for i, rule := range topRules {
		dataPoints := make([]TimeSeriesDataPoint, numBuckets)

		// Initialize all buckets
		for j := 0; j < numBuckets; j++ {
			bucketTime := since.Add(time.Duration(j) * time.Duration(intervalMinutes) * time.Minute)
			dataPoints[j] = TimeSeriesDataPoint{
				Time:  bucketTime.Format("15:04"),
				Value: 0,
			}
		}

		// Query alerts for this rule grouped by time bucket
		type BucketResult struct {
			BucketIndex int
			Count       int64
		}

		var bucketResults []BucketResult

		// PostgreSQL time bucketing using EXTRACT(EPOCH FROM timestamp)
		err := database.DB.Model(&models.Alert{}).
			Select(fmt.Sprintf(`
				CAST((EXTRACT(EPOCH FROM created_at)::bigint - %d) / %d AS INTEGER) as bucket_index,
				COUNT(*) as count
			`, since.Unix(), intervalMinutes*60)).
			Where("rule_id = ? AND created_at >= ? AND created_at <= ?", rule.RuleID, since, now).
			Group("bucket_index").
			Scan(&bucketResults).Error

		if err != nil {
			return nil, fmt.Errorf("failed to get time series data: %w", err)
		}

		// Fill in actual counts
		for _, br := range bucketResults {
			if br.BucketIndex >= 0 && br.BucketIndex < numBuckets {
				dataPoints[br.BucketIndex].Value = br.Count
			}
		}

		result[i] = RuleTimeSeriesStats{
			RuleID:     rule.RuleID,
			RuleName:   rule.RuleName,
			Total:      rule.Total,
			DataPoints: dataPoints,
		}
	}

	return result, nil
}
