// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/kk/elk-helper/backend/internal/models"
	"github.com/kk/elk-helper/backend/internal/service/alert"
	es_config "github.com/kk/elk-helper/backend/internal/service/esconfig"
	"github.com/kk/elk-helper/backend/internal/service/query"
	"github.com/kk/elk-helper/backend/internal/service/rule"
	system_config "github.com/kk/elk-helper/backend/internal/service/systemconfig"
	"github.com/kk/elk-helper/backend/internal/worker/executor"
)

// Scheduler manages rule execution schedule
type Scheduler struct {
	ruleService         *rule.Service
	queryService        *query.Service
	alertService        *alert.Service
	systemConfigService *system_config.Service
	executor            *executor.Executor
	ctx                 context.Context
	cancel              context.CancelFunc
	wg                  sync.WaitGroup
	mu                  sync.RWMutex
	runningRules        map[uint]context.CancelFunc // Track running rule goroutines
	checkInterval       time.Duration
	triggerChan         chan uint // Channel to trigger immediate rule sync/execution
	maxConcurrency      int
	execSem             chan struct{}
}

// Global scheduler instance for triggering from handlers
var globalScheduler *Scheduler

// GetGlobalScheduler returns the global scheduler instance
func GetGlobalScheduler() *Scheduler {
	return globalScheduler
}

// NewScheduler creates a new scheduler
func NewScheduler(ruleService *rule.Service, queryService *query.Service, esConfigService *es_config.Service, alertService *alert.Service, systemConfigService *system_config.Service, checkInterval time.Duration, retryTimes, batchSize, maxConcurrency int) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	if maxConcurrency <= 0 {
		maxConcurrency = 1
	}

	s := &Scheduler{
		ruleService:         ruleService,
		queryService:        queryService,
		alertService:        alertService,
		systemConfigService: systemConfigService,
		executor:            executor.NewExecutor(queryService, esConfigService, ruleService, alertService, retryTimes, batchSize),
		ctx:                 ctx,
		cancel:              cancel,
		runningRules:        make(map[uint]context.CancelFunc),
		checkInterval:       checkInterval,
		triggerChan:         make(chan uint, 100), // Buffer for rule triggers
		maxConcurrency:      maxConcurrency,
		execSem:             make(chan struct{}, maxConcurrency),
	}

	// Set global instance for access from handlers
	globalScheduler = s

	return s
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	slog.Info("Scheduler started", "check_interval", s.checkInterval, "max_concurrency", s.maxConcurrency)

	// Start rule monitor goroutine
	s.wg.Add(1)
	go s.monitorRules()

	// Start cleanup task goroutine (runs daily at 3 AM)
	s.wg.Add(1)
	go s.startCleanupTask()

	return nil
}

// TriggerRule notifies the scheduler to immediately check and execute a rule
// This is called when a rule is created, updated, or enabled
func (s *Scheduler) TriggerRule(ruleID uint) {
	select {
	case s.triggerChan <- ruleID:
		slog.Info("Rule trigger sent", "rule_id", ruleID)
	default:
		// Channel is full, sync will happen on next check interval
		slog.Warn("Rule trigger channel full, will sync on next interval", "rule_id", ruleID)
	}
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.cancel()

	// Cancel all running rule goroutines
	s.mu.Lock()
	for ruleID, cancel := range s.runningRules {
		slog.Info("Stopping rule", "rule_id", ruleID)
		cancel()
	}
	s.mu.Unlock()

	s.wg.Wait()
	slog.Info("Scheduler stopped")
}

// monitorRules periodically checks for rule changes and starts/stops rule goroutines
func (s *Scheduler) monitorRules() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	// Initial check
	s.syncRules()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.syncRules()
		case ruleID := <-s.triggerChan:
			// Triggered by rule creation/update/enable - sync and execute immediately
			slog.Info("Rule trigger received, syncing immediately", "rule_id", ruleID)
			s.syncRulesAndExecute(ruleID)
		}
	}
}

// syncRules synchronizes running rule goroutines with enabled rules
func (s *Scheduler) syncRules() {
	rules, err := s.ruleService.GetEnabled()
	if err != nil {
		slog.Error("Failed to get enabled rules", "error", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Log enabled rules found
	enabledRuleIDs := make([]uint, 0, len(rules))
	for _, r := range rules {
		enabledRuleIDs = append(enabledRuleIDs, r.ID)
	}

	// Always log sync activity for debugging
	runningRuleIDList := make([]uint, 0, len(s.runningRules))
	for ruleID := range s.runningRules {
		runningRuleIDList = append(runningRuleIDList, ruleID)
	}
	slog.Info("Syncing rules", "enabled_count", len(rules), "enabled_rule_ids", enabledRuleIDs, "running_count", len(s.runningRules), "running_rule_ids", runningRuleIDList)

	// Create map of enabled rule IDs
	enabledRuleIDMap := make(map[uint]bool)
	for _, r := range rules {
		enabledRuleIDMap[r.ID] = true
	}

	// Stop goroutines for disabled rules
	stoppedCount := 0
	for ruleID, cancel := range s.runningRules {
		if !enabledRuleIDMap[ruleID] {
			slog.Info("Stopping rule", "rule_id", ruleID, "reason", "disabled")
			cancel()
			delete(s.runningRules, ruleID)
			stoppedCount++
		}
	}

	// Start goroutines for new enabled rules
	runningRuleIDs := make(map[uint]bool)
	for ruleID := range s.runningRules {
		runningRuleIDs[ruleID] = true
	}

	startedCount := 0
	for _, r := range rules {
		if !runningRuleIDs[r.ID] {
			slog.Info("Starting rule", "rule_id", r.ID, "rule_name", r.Name, "interval", r.Interval, "index_pattern", r.IndexPattern)
			ruleCtx, ruleCancel := context.WithCancel(s.ctx)
			s.runningRules[r.ID] = ruleCancel

			// Start goroutine for this rule
			s.wg.Add(1)
			go s.runRule(ruleCtx, r)
			startedCount++
		}
	}

	// Always log sync result for debugging
	if stoppedCount > 0 || startedCount > 0 {
		slog.Info("Rule sync completed", "started", startedCount, "stopped", stoppedCount, "total_running", len(s.runningRules))
	} else {
		slog.Info("Rule sync completed - no changes", "enabled_count", len(rules), "running_count", len(s.runningRules))
	}
}

// syncRulesAndExecute syncs rules and immediately executes the specified rule
// This is called when a rule is created/updated/enabled to execute it immediately
func (s *Scheduler) syncRulesAndExecute(ruleID uint) {
	// First sync to ensure the rule goroutine is started
	s.syncRules()

	// Check if this rule is already running (meaning it was just started by syncRules)
	s.mu.RLock()
	_, isRunning := s.runningRules[ruleID]
	s.mu.RUnlock()

	if isRunning {
		// Rule goroutine was just started, it will execute immediately on its own
		slog.Info("Rule already started by sync, will execute in its goroutine", "rule_id", ruleID)
		return
	}

	// If rule is not running (might be disabled), try to execute it once anyway
	// This handles the case where rule was updated but not enabled
	rule, err := s.ruleService.GetByID(ruleID)
	if err != nil {
		slog.Error("Failed to get rule for immediate execution", "rule_id", ruleID, "error", err)
		return
	}

	if rule.Enabled {
		// Rule is enabled but not running yet, execute it directly
		slog.Info("Executing rule immediately after trigger", "rule_id", ruleID, "rule_name", rule.Name)
		s.executeRuleForce(s.ctx, rule)
	} else {
		slog.Info("Rule is disabled, skipping immediate execution", "rule_id", ruleID, "rule_name", rule.Name)
	}
}

// runRule runs a single rule in its own goroutine
func (s *Scheduler) runRule(ctx context.Context, ruleModel models.Rule) {
	defer s.wg.Done()

	ruleID := ruleModel.ID
	ruleName := ruleModel.Name

	// Use initial interval, will be updated on each execution
	interval := time.Duration(ruleModel.Interval) * time.Second
	if interval < 10*time.Second {
		interval = 10 * time.Second // Minimum 10 seconds
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Execute immediately on start
	slog.Info("Rule goroutine started, executing immediately", "rule_id", ruleID, "rule_name", ruleName, "interval", interval)
	s.executeRuleWithReload(ctx, ruleID)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Rule stopped", "rule_id", ruleID, "rule_name", ruleName)
			return
		case <-ticker.C:
			// Reload rule from database to get latest configuration
			rule, err := s.ruleService.GetByID(ruleID)
			if err != nil {
				slog.Error("Failed to reload rule", "rule_id", ruleID, "error", err)
				continue
			}

			// Update ticker interval if changed
			newInterval := time.Duration(rule.Interval) * time.Second
			if newInterval < 10*time.Second {
				newInterval = 10 * time.Second
			}
			if newInterval != interval {
				interval = newInterval
				ticker.Reset(interval)
				slog.Info("Rule interval updated", "rule_id", ruleID, "rule_name", rule.Name, "interval", interval)
			}

			// Execute with latest configuration
			s.executeRule(ctx, rule)
		}
	}
}

// executeRuleWithReload loads the latest rule configuration and executes it
func (s *Scheduler) executeRuleWithReload(ctx context.Context, ruleID uint) {
	slog.Info("executeRuleWithReload called", "rule_id", ruleID)

	// Reload rule from database to get latest configuration
	rule, err := s.ruleService.GetByID(ruleID)
	if err != nil {
		slog.Error("Failed to reload rule", "rule_id", ruleID, "error", err)
		return
	}

	slog.Info("Rule reloaded", "rule_id", ruleID, "rule_name", rule.Name,
		"lark_config_id", rule.LarkConfigID,
		"lark_config_loaded", rule.LarkConfig != nil,
		"es_config_id", rule.ESConfigID,
		"es_config_loaded", rule.ESConfig != nil)

	// Validate rule configuration before executing
	if rule.LarkConfigID != nil && (rule.LarkConfig == nil || !rule.LarkConfig.Enabled) {
		slog.Warn("Rule has LarkConfigID but config is not loaded or disabled",
			"rule_id", ruleID,
			"rule_name", rule.Name,
			"lark_config_id", *rule.LarkConfigID,
			"lark_config_loaded", rule.LarkConfig != nil,
			"lark_config_enabled", rule.LarkConfig != nil && rule.LarkConfig.Enabled)
	}

	webhookURL := rule.LarkWebhook
	if rule.LarkConfigID != nil && rule.LarkConfig != nil && rule.LarkConfig.Enabled {
		webhookURL = rule.LarkConfig.WebhookURL
	}

	larkConfigWebhookURL := ""
	if rule.LarkConfig != nil {
		larkConfigWebhookURL = rule.LarkConfig.WebhookURL
	}

	if webhookURL == "" {
		slog.Error("Rule has no webhook URL configured, skipping execution",
			"rule_id", ruleID,
			"rule_name", rule.Name,
			"lark_webhook", rule.LarkWebhook,
			"lark_config_id", rule.LarkConfigID,
			"lark_config_loaded", rule.LarkConfig != nil,
			"lark_config_enabled", rule.LarkConfig != nil && rule.LarkConfig.Enabled,
			"lark_config_webhook_url", larkConfigWebhookURL)
		return
	}

	slog.Info("Rule configuration validated, force executing on startup", "rule_id", ruleID, "rule_name", rule.Name, "webhook_url", webhookURL)
	s.executeRuleForce(ctx, rule)
}

// executeRule executes a single rule execution in a separate goroutine (with time interval check)
func (s *Scheduler) executeRule(ctx context.Context, ruleModel *models.Rule) {
	s.executeRuleWithOptions(ctx, ruleModel, false)
}

// executeRuleForce executes a single rule immediately in a separate goroutine (skip time interval check)
func (s *Scheduler) executeRuleForce(ctx context.Context, ruleModel *models.Rule) {
	s.executeRuleWithOptions(ctx, ruleModel, true)
}

// executeRuleWithOptions executes a single rule execution in a separate goroutine
func (s *Scheduler) executeRuleWithOptions(ctx context.Context, ruleModel *models.Rule, forceExecute bool) {
	// Acquire global concurrency slot to avoid unbounded goroutine growth
	select {
	case s.execSem <- struct{}{}:
		// acquired
	case <-ctx.Done():
		slog.Info("Rule execution skipped due to cancellation", "rule_id", ruleModel.ID, "rule_name", ruleModel.Name, "force_execute", forceExecute)
		return
	}
	defer func() { <-s.execSem }()

	var err error
	if forceExecute {
		err = s.executor.ExecuteRuleForce(ctx, ruleModel)
	} else {
		err = s.executor.ExecuteRule(ctx, ruleModel)
	}

	if err != nil {
		slog.Error("Failed to execute rule",
			"rule_id", ruleModel.ID,
			"rule_name", ruleModel.Name,
			"error", err,
			"force_execute", forceExecute,
			"lark_config_id", ruleModel.LarkConfigID,
			"lark_config_loaded", ruleModel.LarkConfig != nil,
			"es_config_id", ruleModel.ESConfigID,
			"es_config_loaded", ruleModel.ESConfig != nil)
	} else {
		slog.Info("Rule executed successfully", "rule_id", ruleModel.ID, "rule_name", ruleModel.Name, "force_execute", forceExecute)
	}
}

// startCleanupTask runs a daily cleanup task based on configuration
func (s *Scheduler) startCleanupTask() {
	defer s.wg.Done()

	// Check configuration every minute to handle updates
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	var nextRun *time.Time
	var retentionDays int

	// Initial configuration load
	config, err := s.systemConfigService.GetCleanupConfig()
	if err != nil {
		slog.Error("Failed to load cleanup config on startup", "error", err)
	} else if config == nil {
		slog.Warn("Cleanup config not found, using defaults", "default_enabled", true, "default_hour", 3, "default_minute", 0, "default_retention_days", 90)
		// Use default config
		nextRun = s.nextRunTime(3, 0)
		retentionDays = 90
		slog.Info("Cleanup task enabled with defaults", "scheduled_time", nextRun.Format("2006-01-02 15:04:05"), "retention_days", retentionDays)
	} else if !config.Enabled {
		slog.Info("Cleanup task disabled in configuration", "hour", config.Hour, "minute", config.Minute, "retention_days", config.RetentionDays)
	} else {
		nextRun = s.nextRunTime(config.Hour, config.Minute)
		retentionDays = config.RetentionDays
		slog.Info("Cleanup task enabled", "scheduled_time", nextRun.Format("2006-01-02 15:04:05"), "retention_days", retentionDays, "hour", config.Hour, "minute", config.Minute)
	}

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			// Reload configuration
			config, err := s.systemConfigService.GetCleanupConfig()
			if err != nil {
				slog.Error("Failed to load cleanup config", "error", err)
				continue
			}

			if !config.Enabled {
				if nextRun != nil {
					slog.Info("Cleanup task disabled, clearing next run", "was_scheduled_for", nextRun.Format("2006-01-02 15:04:05"))
					nextRun = nil
				}
				continue
			}

			// If configuration changed, reschedule
			newNextRun := s.nextRunTime(config.Hour, config.Minute)
			configChanged := nextRun == nil || !nextRun.Equal(*newNextRun) || retentionDays != config.RetentionDays
			if configChanged {
				nextRun = newNextRun
				retentionDays = config.RetentionDays
				slog.Info("Cleanup task rescheduled", "scheduled_time", nextRun.Format("2006-01-02 15:04:05"), "retention_days", retentionDays, "hour", config.Hour, "minute", config.Minute)
			}

			// Check if it's time to run
			now := time.Now()
			if nextRun == nil {
				// This should not happen if config.Enabled is true, but log it for debugging
				slog.Warn("Cleanup task enabled but nextRun is nil, rescheduling", "hour", config.Hour, "minute", config.Minute)
				nextRun = s.nextRunTime(config.Hour, config.Minute)
				retentionDays = config.RetentionDays
				slog.Info("Cleanup task rescheduled", "scheduled_time", nextRun.Format("2006-01-02 15:04:05"), "retention_days", retentionDays)
			} else {
				// Use truncate to minute for comparison to allow execution within the same minute
				nowTruncated := now.Truncate(time.Minute)
				nextRunTruncated := nextRun.Truncate(time.Minute)

				// Log debug info every minute to help diagnose issues
				if now.Second() < 5 {
					slog.Info("Cleanup task check",
						"current_time", now.Format("2006-01-02 15:04:05"),
						"next_run", nextRun.Format("2006-01-02 15:04:05"),
						"now_truncated", nowTruncated.Format("2006-01-02 15:04:05"),
						"next_run_truncated", nextRunTruncated.Format("2006-01-02 15:04:05"),
						"should_run", nowTruncated.After(nextRunTruncated) || nowTruncated.Equal(nextRunTruncated),
						"retention_days", retentionDays,
						"enabled", config.Enabled)
				}

				// Check if we should execute now
				shouldExecute := nowTruncated.After(nextRunTruncated) || nowTruncated.Equal(nextRunTruncated)
				if shouldExecute {
					slog.Info("Cleanup task triggered",
						"triggered_at", now.Format("2006-01-02 15:04:05"),
						"scheduled_for", nextRun.Format("2006-01-02 15:04:05"),
						"now_truncated", nowTruncated.Format("2006-01-02 15:04:05"),
						"next_run_truncated", nextRunTruncated.Format("2006-01-02 15:04:05"))

					// Execute cleanup
					retentionDuration := time.Duration(retentionDays) * 24 * time.Hour
					rowsAffected, err := s.alertService.CleanupOldData(retentionDuration)
					if err != nil {
						slog.Error("Failed to cleanup old alerts", "error", err)
						// Update execution status to failed
						statusErr := s.systemConfigService.UpdateCleanupExecutionStatus("failed", fmt.Sprintf("清理失败: %v", err))
						if statusErr != nil {
							slog.Error("Failed to update cleanup execution status", "error", statusErr)
						} else {
							slog.Info("Cleanup execution status updated", "status", "failed")
						}
					} else {
						slog.Info("Cleanup task completed", "rows_affected", rowsAffected, "retention_days", retentionDays)
						// Update execution status to success
						resultMsg := fmt.Sprintf("成功删除 %d 条告警数据", rowsAffected)
						if rowsAffected == 0 {
							resultMsg = "没有需要清理的数据"
						}
						statusErr := s.systemConfigService.UpdateCleanupExecutionStatus("success", resultMsg)
						if statusErr != nil {
							slog.Error("Failed to update cleanup execution status", "error", statusErr)
						} else {
							slog.Info("Cleanup execution status updated", "status", "success", "message", resultMsg)
						}
					}

					// Schedule next run immediately to prevent multiple executions in the same minute
					nextRun = s.nextRunTime(config.Hour, config.Minute)
					slog.Info("Next cleanup task scheduled",
						"scheduled_time", nextRun.Format("2006-01-02 15:04:05"),
						"current_time", time.Now().Format("2006-01-02 15:04:05"),
						"retention_days", retentionDays)
				} else {
					// Log periodically for debugging (every 10 minutes)
					if now.Minute()%10 == 0 && now.Second() < 10 {
						slog.Info("Cleanup task waiting",
							"current_time", now.Format("2006-01-02 15:04:05"),
							"next_run", nextRun.Format("2006-01-02 15:04:05"),
							"time_until_run", nextRun.Sub(now).String(),
							"retention_days", retentionDays)
					}
				}
			}
		}
	}
}

// nextRunTime calculates the next run time based on hour and minute
func (s *Scheduler) nextRunTime(hour, minute int) *time.Time {
	now := time.Now()
	// Get today's scheduled time
	todayRun := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())

	// Truncate current time to minute for comparison
	nowTruncated := now.Truncate(time.Minute)
	todayRunTruncated := todayRun.Truncate(time.Minute)

	// If today's scheduled time has passed (more than 1 minute), schedule for tomorrow
	// Allow execution within the same minute (nowTruncated.Equal(todayRunTruncated) means we're in the execution window)
	if nowTruncated.After(todayRunTruncated) {
		todayRun = todayRun.Add(24 * time.Hour)
		slog.Debug("Cleanup task scheduled for tomorrow",
			"today_run", todayRunTruncated.Format("2006-01-02 15:04:05"),
			"now", nowTruncated.Format("2006-01-02 15:04:05"),
			"next_run", todayRun.Format("2006-01-02 15:04:05"))
	} else {
		slog.Debug("Cleanup task scheduled for today",
			"scheduled_time", todayRun.Format("2006-01-02 15:04:05"),
			"now", nowTruncated.Format("2006-01-02 15:04:05"))
	}

	return &todayRun
}
