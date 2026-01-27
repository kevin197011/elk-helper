// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package alert

import (
	"context"
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
	db, cancel := database.WithTimeout(context.Background())
	defer cancel()

	if err := db.Create(alert).Error; err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}
	return nil
}

// GetAll returns all alerts with pagination (without logs for performance)
func (s *Service) GetAll(page, pageSize int) ([]models.Alert, int64, error) {
	var alerts []models.Alert
	var total int64

	offset := (page - 1) * pageSize

	db, cancel := database.WithTimeout(context.Background())
	defer cancel()

	if err := db.Model(&models.Alert{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count alerts: %w", err)
	}

	// Optimize: Don't load logs field in list view for performance
	// Logs can be hundreds of KB or even MBs, causing slow page loads
	// Only load logs when viewing individual alert details
	if err := db.Preload("Rule").
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
	db, cancel := database.WithTimeout(context.Background())
	defer cancel()

	if err := db.Preload("Rule").First(&alert, id).Error; err != nil {
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

// calculateBucketInterval calculates appropriate time bucket interval based on rule execution interval
// Returns interval in minutes
func calculateBucketInterval(ruleIntervalSeconds int) int {
	// Convert rule interval to minutes
	ruleIntervalMinutes := ruleIntervalSeconds / 60
	if ruleIntervalMinutes < 1 {
		ruleIntervalMinutes = 1
	}

	// Calculate bucket interval: use 5-10x the rule interval, but with reasonable bounds
	// This ensures we capture multiple executions per bucket while keeping reasonable granularity
	bucketMinutes := ruleIntervalMinutes * 5

	// Apply bounds for readability and performance
	if bucketMinutes < 1 {
		bucketMinutes = 1 // Minimum 1 minute
	} else if bucketMinutes > 60 {
		bucketMinutes = 60 // Maximum 60 minutes (1 hour)
	}

	// Round to common intervals for cleaner display
	if bucketMinutes <= 5 {
		bucketMinutes = 5
	} else if bucketMinutes <= 15 {
		bucketMinutes = 15
	} else if bucketMinutes <= 30 {
		bucketMinutes = 30
	} else {
		bucketMinutes = 60
	}

	return bucketMinutes
}

// GetRuleTimeSeriesStats returns time series alert statistics for all enabled rules
// Each rule uses its own execution interval to calculate appropriate time bucket size
func (s *Service) GetRuleTimeSeriesStats(duration time.Duration, _ int) ([]RuleTimeSeriesStats, error) {
	// Explicitly get UTC time to ensure consistency
	// Note: time.Now() respects TZ env var in Docker, but we explicitly convert to UTC for database queries
	now := time.Now()
	utcNow := now.UTC()
	since := utcNow.Add(-duration)

	// Get all enabled rules
	var allRules []models.Rule
	err := database.DB.Where("enabled = ?", true).Find(&allRules).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get rules: %w", err)
	}

	if len(allRules) == 0 {
		return []RuleTimeSeriesStats{}, nil
	}

	// Get alert counts for each rule in the time period
	type RuleAlertCount struct {
		RuleID uint
		Total  int64
	}

	var alertCounts []RuleAlertCount
	err = database.DB.Model(&models.Alert{}).
		Select("rule_id, COUNT(*) as total").
		Where("created_at >= ?", since).
		Group("rule_id").
		Scan(&alertCounts).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get alert counts: %w", err)
	}

	// Create map for quick lookup
	alertCountMap := make(map[uint]int64)
	for _, ac := range alertCounts {
		alertCountMap[ac.RuleID] = ac.Total
	}

	// Load timezone for display (explicitly use Asia/Hong_Kong to match Docker TZ setting)
	localTZ, err := time.LoadLocation("Asia/Hong_Kong")
	if err != nil {
		// Fallback to UTC if timezone loading fails
		localTZ = time.UTC
	}

	result := make([]RuleTimeSeriesStats, len(allRules))

	for i, rule := range allRules {
		// Calculate bucket interval based on rule's execution interval
		bucketIntervalMinutes := calculateBucketInterval(rule.Interval)

		// Calculate number of buckets for this rule
		numBuckets := int(duration.Minutes()) / bucketIntervalMinutes
		if numBuckets > 48 {
			numBuckets = 48 // Max 48 points for readability
		}
		if numBuckets < 6 {
			numBuckets = 6 // Min 6 points
		}

		dataPoints := make([]TimeSeriesDataPoint, numBuckets)

		// Use actual current time (not aligned) for query boundary to ensure exact 24-hour range
		// Time range: [current_time - 24h, current_time]
		// Generate time buckets starting from actual start time (since) to show rolling 24-hour window
		intervalSeconds := bucketIntervalMinutes * 60
		sinceUnix := since.Unix()

		// Align start time down to bucket boundary for consistent bucket indexing in database query
		// This ensures buckets align nicely for data aggregation
		alignedSinceUnix := (sinceUnix / int64(intervalSeconds)) * int64(intervalSeconds)
		alignedSince := time.Unix(alignedSinceUnix, 0).UTC()

		// Calculate bucket duration
		bucketDuration := time.Duration(bucketIntervalMinutes) * time.Minute

		// Initialize all buckets - from oldest (since) to newest (utcNow)
		// Time labels start from actual start time (since) to show rolling 24-hour window
		// But we need to map these to the correct bucket indices for data aggregation
		for j := 0; j < numBuckets; j++ {
			// Calculate bucket label time: from actual since time, not aligned time
			// This ensures the first bucket label shows the actual start time (24 hours ago)
			// and the last bucket shows time near current time
			bucketLabelTime := since.Add(time.Duration(j) * bucketDuration)
			
			// Ensure last bucket doesn't exceed current time
			if bucketLabelTime.After(utcNow) {
				bucketLabelTime = utcNow
			}
			
			// Convert to Hong Kong time for display
			bucketTimeHK := bucketLabelTime.In(localTZ)
			dataPoints[j] = TimeSeriesDataPoint{
				Time:  bucketTimeHK.Format("15:04"),
				Value: 0,
			}
		}

		// Query alerts for this rule grouped by time bucket
		// Use COUNT(*) to count alert records (execution count), not log entries
		type BucketResult struct {
			BucketIndex int
			Count       int64
		}

		var bucketResults []BucketResult

		// PostgreSQL time bucketing using EXTRACT(EPOCH FROM timestamp)
		// Use COUNT(*) to count alert records (execution count)
		// Query boundary uses actual current time (since, utcNow) for exact 24-hour range
		// Bucket indexing uses aligned start time for consistent bucket assignment
		err := database.DB.Model(&models.Alert{}).
			Select(fmt.Sprintf(`
				CAST((EXTRACT(EPOCH FROM created_at)::bigint - %d) / %d AS INTEGER) as bucket_index,
				COUNT(*) as count
			`, alignedSince.Unix(), intervalSeconds)).
			Where("rule_id = ? AND created_at >= ? AND created_at <= ?", rule.ID, since, utcNow).
			Group("bucket_index").
			Scan(&bucketResults).Error

		if err != nil {
			return nil, fmt.Errorf("failed to get time series data for rule %d: %w", rule.ID, err)
		}

		// Fill in actual counts
		// Map database bucket indices to display bucket indices
		// Database uses alignedSince for indexing, but we display from actual since time
		for _, br := range bucketResults {
			// Calculate the actual time for this database bucket
			dbBucketTime := alignedSince.Add(time.Duration(br.BucketIndex) * bucketDuration)
			
			// Find the corresponding display bucket index
			// Display buckets start from 'since', so we calculate the offset
			displayBucketIndex := int(dbBucketTime.Sub(since) / bucketDuration)
			
			// Ensure the index is within bounds
			if displayBucketIndex >= 0 && displayBucketIndex < numBuckets {
				dataPoints[displayBucketIndex].Value = br.Count
			}
		}

		result[i] = RuleTimeSeriesStats{
			RuleID:     rule.ID,
			RuleName:   rule.Name,
			Total:      alertCountMap[rule.ID], // Will be 0 if not in map
			DataPoints: dataPoints,
		}
	}

	return result, nil
}
