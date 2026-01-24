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

	"github.com/kk/elk-helper/backend/internal/config"
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

// ExecuteRule executes a single rule (with time interval check)
func (e *Executor) ExecuteRule(ctx context.Context, ruleModel *models.Rule) error {
	return e.ExecuteRuleWithOptions(ctx, ruleModel, false)
}

// ExecuteRuleForce executes a single rule immediately (skip time interval check)
func (e *Executor) ExecuteRuleForce(ctx context.Context, ruleModel *models.Rule) error {
	return e.ExecuteRuleWithOptions(ctx, ruleModel, true)
}

// ExecuteRuleWithOptions executes a single rule with options
// forceExecute: if true, skip the time interval check and execute immediately
func (e *Executor) ExecuteRuleWithOptions(ctx context.Context, ruleModel *models.Rule, forceExecute bool) error {
	slog.Info("ExecuteRule called", "rule_id", ruleModel.ID, "rule_name", ruleModel.Name, "interval", ruleModel.Interval, "force_execute", forceExecute)

	// Get last run time
	lastRun := time.Now().Add(-5 * time.Minute) // Default fallback
	if ruleModel.LastRunTime != nil {
		lastRun = *ruleModel.LastRunTime
		// Add a small overlap (2 seconds) to avoid missing data at boundaries
		// This ensures we don't miss logs that might have timestamps exactly at the boundary
		// or logs that arrived during the previous query execution
		lastRun = lastRun.Add(-2 * time.Second)
		slog.Info("Using last run time", "rule_id", ruleModel.ID, "last_run", ruleModel.LastRunTime.Format("2006-01-02 15:04:05"), "adjusted_last_run", lastRun.Format("2006-01-02 15:04:05"))
	} else {
		slog.Info("No last run time, using default", "rule_id", ruleModel.ID, "default_last_run", lastRun.Format("2006-01-02 15:04:05"))
	}

	currentTime := time.Now()

	// Check if enough time has passed (skip if forceExecute is true)
	timeSinceLastRun := currentTime.Sub(lastRun)
	requiredInterval := time.Duration(ruleModel.Interval) * time.Second

	if !forceExecute && timeSinceLastRun < requiredInterval {
		slog.Info("Skipping execution - not enough time passed", "rule_id", ruleModel.ID, "time_since_last_run", timeSinceLastRun, "required_interval", requiredInterval)
		return nil // Skip if not enough time has passed
	}

	if forceExecute {
		slog.Info("Force executing rule (skip time check)", "rule_id", ruleModel.ID, "time_since_last_run", timeSinceLastRun, "required_interval", requiredInterval)
	} else {
		slog.Info("Proceeding with execution", "rule_id", ruleModel.ID, "time_since_last_run", timeSinceLastRun, "required_interval", requiredInterval)
	}

	// Get webhook URL from config or fallback to direct URL
	webhookURL := ruleModel.LarkWebhook
	if ruleModel.LarkConfigID != nil && ruleModel.LarkConfig != nil && ruleModel.LarkConfig.Enabled {
		webhookURL = ruleModel.LarkConfig.WebhookURL
		slog.Info("Using Lark config webhook", "rule_id", ruleModel.ID, "lark_config_id", *ruleModel.LarkConfigID, "webhook_url", webhookURL)
	} else if ruleModel.LarkWebhook != "" {
		slog.Info("Using direct webhook", "rule_id", ruleModel.ID, "webhook_url", ruleModel.LarkWebhook)
	}

	if webhookURL == "" {
		errMsg := fmt.Sprintf("no webhook URL configured for rule: lark_webhook=%s, lark_config_id=%v, lark_config_loaded=%v, lark_config_enabled=%v",
			ruleModel.LarkWebhook,
			ruleModel.LarkConfigID,
			ruleModel.LarkConfig != nil,
			ruleModel.LarkConfig != nil && ruleModel.LarkConfig.Enabled)
		slog.Error("No webhook URL configured", "rule_id", ruleModel.ID, "rule_name", ruleModel.Name, "details", errMsg)
		return fmt.Errorf("%s", errMsg)
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
	slog.Info("Querying logs", "rule_id", ruleModel.ID, "index_pattern", ruleModel.IndexPattern, "from_time", lastRun.Format("2006-01-02 15:04:05"), "to_time", currentTime.Format("2006-01-02 15:04:05"))
	logs, err := queryService.QueryLogs(ctx, ruleModel, lastRun, currentTime, e.batchSize)
	if err != nil {
		slog.Error("Query failed", "rule_id", ruleModel.ID, "error", err)
		return fmt.Errorf("query failed: %w", err)
	}

	slog.Info("Query completed", "rule_id", ruleModel.ID, "logs_found", len(logs))

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
		slog.Info("No logs matched, skipping alert", "rule_id", ruleModel.ID, "rule_name", ruleModel.Name)
		return nil // No logs matched
	}

	// Send alert and create alert record in a separate goroutine
	timeRange := fmt.Sprintf("%s ~ %s", lastRun.Format("2006-01-02 15:04:05"), currentTime.Format("2006-01-02 15:04:05"))
	slog.Info("Found logs, triggering alert", "rule_id", ruleModel.ID, "rule_name", ruleModel.Name, "log_count", len(logs), "time_range", timeRange)

	go e.sendAlertAsync(ruleModel, logs, lastRun, currentTime, timeRange)

	return nil
}

// sendAlertAsync sends alert asynchronously in a separate goroutine
func (e *Executor) sendAlertAsync(ruleModel *models.Rule, logs []map[string]interface{}, fromTime, toTime time.Time, timeRange string) {
	slog.Info("sendAlertAsync started", "rule_id", ruleModel.ID, "rule_name", ruleModel.Name, "log_count", len(logs))

	originalLogCount := len(logs)
	logsForNotify := logs
	if len(logsForNotify) > 10 {
		// 告警通知只需要少量样本，避免 payload 过大
		logsForNotify = logsForNotify[:10]
	}

	// Get webhook URL from config or fallback to direct URL
	webhookURL := ruleModel.LarkWebhook
	if ruleModel.LarkConfigID != nil && ruleModel.LarkConfig != nil && ruleModel.LarkConfig.Enabled {
		webhookURL = ruleModel.LarkConfig.WebhookURL
		slog.Info("Using Lark config webhook in sendAlertAsync", "rule_id", ruleModel.ID, "lark_config_id", *ruleModel.LarkConfigID, "webhook_url", webhookURL)
	} else if ruleModel.LarkWebhook != "" {
		slog.Info("Using direct webhook in sendAlertAsync", "rule_id", ruleModel.ID, "webhook_url", ruleModel.LarkWebhook)
	}

	if webhookURL == "" {
		slog.Error("No webhook URL configured for rule in sendAlertAsync", "rule_id", ruleModel.ID, "rule_name", ruleModel.Name)
		return
	}

	// Create notifier for this alert
	notifier := notifier.NewLarkClient(webhookURL)
	slog.Info("Sending alert notification", "rule_id", ruleModel.ID, "rule_name", ruleModel.Name, "webhook_url", webhookURL, "retry_times", e.retryTimes)

	sendTimeout := 20 * time.Second
	if config.AppConfig != nil && config.AppConfig.Worker.AlertSendTimeoutSeconds > 0 {
		sendTimeout = time.Duration(config.AppConfig.Worker.AlertSendTimeoutSeconds) * time.Second
	}

	sendCtx, cancel := context.WithTimeout(context.Background(), sendTimeout)
	defer cancel()

	err := func() error {
		// notifier 内部 http client 有 timeout；这里再用 context 做整体兜底
		type result struct{ err error }
		ch := make(chan result, 1)
		go func() {
			ch <- result{err: notifier.SendAlert(ruleModel.Name, ruleModel.IndexPattern, logsForNotify, originalLogCount, fromTime, toTime, e.retryTimes)}
		}()

		select {
		case r := <-ch:
			return r.err
		case <-sendCtx.Done():
			return fmt.Errorf("alert send timeout after %s", sendTimeout)
		}
	}()

	if err != nil {
		slog.Error("Alert send failed", "rule_id", ruleModel.ID, "rule_name", ruleModel.Name, "error", err)
	} else {
		slog.Info("Alert sent successfully", "rule_id", ruleModel.ID, "rule_name", ruleModel.Name)
	}

	// Determine alert status
	alertStatus := models.AlertStatusSent
	errorMsg := ""
	if err != nil {
		alertStatus = models.AlertStatusFailed
		errorMsg = err.Error()
	}

	// Persist only a capped sample of logs to prevent DB bloat.
	logsForStorage := logs
	if len(logsForStorage) > 50 {
		logsForStorage = logsForStorage[:50]
	}

	// Create alert record
	alertRecord := &models.Alert{
		RuleID:    ruleModel.ID,
		IndexName: ruleModel.IndexPattern,
		LogCount:  originalLogCount,
		Logs:      models.LogData(logsForStorage),
		TimeRange: timeRange,
		Status:    alertStatus,
		ErrorMsg:  errorMsg,
	}

	// Create alert record (async)
	if err := e.alertService.Create(alertRecord); err != nil {
		slog.Error("Failed to create alert record", "rule_id", ruleModel.ID, "error", err)
	} else {
		slog.Info("Alert record created", "rule_id", ruleModel.ID, "alert_status", alertStatus, "log_count", len(logs))
	}

	// Update alert count if successful (async)
	if err == nil {
		go func() {
			if err := e.ruleService.IncrementAlertCount(ruleModel.ID, int64(originalLogCount)); err != nil {
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
