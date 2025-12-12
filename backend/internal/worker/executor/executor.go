// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package executor

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/kk/elk-helper/backend/internal/models"
	"github.com/kk/elk-helper/backend/internal/service/alert"
	es_config "github.com/kk/elk-helper/backend/internal/service/esconfig"
	"github.com/kk/elk-helper/backend/internal/service/query"
	"github.com/kk/elk-helper/backend/internal/service/rule"
	"github.com/kk/elk-helper/backend/internal/worker/notifier"
)

// Executor executes rule queries and sends alerts
type Executor struct {
	defaultQueryService *query.Service // Fallback service using environment variables
	esConfigService     *es_config.Service
	ruleService         *rule.Service
	alertService        *alert.Service
	notifier            *notifier.LarkClient
	batchSize           int
	retryTimes          int
}

// NewExecutor creates a new executor
func NewExecutor(defaultQueryService *query.Service, esConfigService *es_config.Service, ruleService *rule.Service, alertService *alert.Service, retryTimes, batchSize int) *Executor {
	return &Executor{
		defaultQueryService: defaultQueryService,
		esConfigService:     esConfigService,
		ruleService:         ruleService,
		alertService:        alertService,
		batchSize:           batchSize,
		retryTimes:          retryTimes,
	}
}

// ExecuteRule executes a single rule
func (e *Executor) ExecuteRule(ctx context.Context, ruleModel *models.Rule) error {
	// Get last run time
	lastRun := time.Now().Add(-5 * time.Minute) // Default fallback
	if ruleModel.LastRunTime != nil {
		lastRun = *ruleModel.LastRunTime
		// Add a small overlap (2 seconds) to avoid missing data at boundaries
		// This ensures we don't miss logs that might have timestamps exactly at the boundary
		// or logs that arrived during the previous query execution
		lastRun = lastRun.Add(-2 * time.Second)
	}

	currentTime := time.Now()

	// Check if enough time has passed
	timeSinceLastRun := currentTime.Sub(lastRun)
	if timeSinceLastRun < time.Duration(ruleModel.Interval)*time.Second {
		return nil // Skip if not enough time has passed
	}

	// Get webhook URL from config or fallback to direct URL
	webhookURL := ruleModel.LarkWebhook
	if ruleModel.LarkConfigID != nil && ruleModel.LarkConfig != nil && ruleModel.LarkConfig.Enabled {
		webhookURL = ruleModel.LarkConfig.WebhookURL
	}
	if webhookURL == "" {
		return fmt.Errorf("no webhook URL configured for rule")
	}

	// Create notifier for this rule
	e.notifier = notifier.NewLarkClient(webhookURL)

	// Get query service based on rule's ES config
	queryService, err := e.getQueryService(ruleModel)
	if err != nil {
		return fmt.Errorf("failed to get query service: %w", err)
	}

	// Query logs using the adjusted lastRun time (with overlap to prevent data loss)
	// The overlap ensures we don't miss logs at the boundary
	logs, err := queryService.QueryLogs(ctx, ruleModel, lastRun, currentTime, e.batchSize)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	// Update last run time immediately after successful query (synchronous to prevent data loss)
	// This ensures the next query starts from the correct time
	now := currentTime
	if err := e.ruleService.UpdateLastRunTime(ruleModel.ID, &now); err != nil {
		slog.Warn("Failed to update last run time", "rule_id", ruleModel.ID, "error", err)
		// Continue even if update fails, but log the error
	}

	// Update run count asynchronously (non-critical)
	go func() {
		if err := e.ruleService.IncrementRunCount(ruleModel.ID); err != nil {
			slog.Warn("Failed to increment run count", "rule_id", ruleModel.ID, "error", err)
		}
	}()

	if len(logs) == 0 {
		return nil // No logs matched
	}

	// Send alert and create alert record in a separate goroutine
	timeRange := fmt.Sprintf("%s ~ %s", lastRun.Format("2006-01-02 15:04:05"), currentTime.Format("2006-01-02 15:04:05"))

	go e.sendAlertAsync(ruleModel, logs, lastRun, currentTime, timeRange)

	return nil
}

// sendAlertAsync sends alert asynchronously in a separate goroutine
func (e *Executor) sendAlertAsync(ruleModel *models.Rule, logs []map[string]interface{}, fromTime, toTime time.Time, timeRange string) {
	// Get webhook URL from config or fallback to direct URL
	webhookURL := ruleModel.LarkWebhook
	if ruleModel.LarkConfigID != nil && ruleModel.LarkConfig != nil && ruleModel.LarkConfig.Enabled {
		webhookURL = ruleModel.LarkConfig.WebhookURL
	}
	if webhookURL == "" {
		slog.Error("No webhook URL configured for rule", "rule_id", ruleModel.ID)
		return
	}

	// Create notifier for this alert
	notifier := notifier.NewLarkClient(webhookURL)

	// Send alert with timeout
	alertErrChan := make(chan error, 1)
	go func() {
		alertErrChan <- notifier.SendAlert(ruleModel.Name, ruleModel.IndexPattern, logs, fromTime, toTime, e.retryTimes)
	}()

	var err error
	select {
	case err = <-alertErrChan:
		// Alert sent
	case <-time.After(30 * time.Second):
		err = fmt.Errorf("alert send timeout after 30 seconds")
	}

	// Determine alert status
	alertStatus := models.AlertStatusSent
	errorMsg := ""
	if err != nil {
		alertStatus = models.AlertStatusFailed
		errorMsg = err.Error()
	}

	// Create alert record
	alertRecord := &models.Alert{
		RuleID:    ruleModel.ID,
		IndexName: ruleModel.IndexPattern,
		LogCount:  len(logs),
		Logs:      logs,
		TimeRange: timeRange,
		Status:    alertStatus,
		ErrorMsg:  errorMsg,
	}

	// Create alert record (async)
	if err := e.alertService.Create(alertRecord); err != nil {
		slog.Warn("Failed to create alert record", "rule_id", ruleModel.ID, "error", err)
	}

	// Update alert count if successful (async)
	if err == nil {
		go func() {
			if err := e.ruleService.IncrementAlertCount(ruleModel.ID, int64(len(logs))); err != nil {
				slog.Warn("Failed to increment alert count", "rule_id", ruleModel.ID, "error", err)
			}
		}()
	}
}

// getQueryService returns a query service based on rule's ES config
func (e *Executor) getQueryService(ruleModel *models.Rule) (*query.Service, error) {
	// If rule has ES config, use it
	if ruleModel.ESConfigID != nil && ruleModel.ESConfig != nil {
		if !ruleModel.ESConfig.Enabled {
			return nil, fmt.Errorf("ES config is disabled")
		}
		return query.NewServiceFromConfig(ruleModel.ESConfig)
	}

	// If rule has ES config ID but ESConfig is not loaded, fetch it
	if ruleModel.ESConfigID != nil {
		esConfig, err := e.esConfigService.GetByID(*ruleModel.ESConfigID)
		if err != nil {
			return nil, fmt.Errorf("failed to get ES config: %w", err)
		}
		if !esConfig.Enabled {
			return nil, fmt.Errorf("ES config is disabled")
		}
		return query.NewServiceFromConfig(esConfig)
	}

	// Fallback to default query service (using environment variables)
	if e.defaultQueryService == nil {
		return nil, fmt.Errorf("no query service available")
	}
	return e.defaultQueryService, nil
}
