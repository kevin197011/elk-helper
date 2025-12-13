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
}

// NewScheduler creates a new scheduler
func NewScheduler(ruleService *rule.Service, queryService *query.Service, esConfigService *es_config.Service, alertService *alert.Service, systemConfigService *system_config.Service, checkInterval time.Duration, retryTimes, batchSize int) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	return &Scheduler{
		ruleService:         ruleService,
		queryService:        queryService,
		alertService:        alertService,
		systemConfigService: systemConfigService,
		executor:            executor.NewExecutor(queryService, esConfigService, ruleService, alertService, retryTimes, batchSize),
		ctx:                 ctx,
		cancel:              cancel,
		runningRules:        make(map[uint]context.CancelFunc),
		checkInterval:       checkInterval,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	slog.Info("Scheduler started", "check_interval", s.checkInterval)

	// Start rule monitor goroutine
	s.wg.Add(1)
	go s.monitorRules()

	// Start cleanup task goroutine (runs daily at 3 AM)
	s.wg.Add(1)
	go s.startCleanupTask()

	return nil
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

	// Create map of enabled rule IDs
	enabledRuleIDs := make(map[uint]bool)
	for _, r := range rules {
		enabledRuleIDs[r.ID] = true
	}

	// Stop goroutines for disabled rules
	for ruleID, cancel := range s.runningRules {
		if !enabledRuleIDs[ruleID] {
			slog.Info("Stopping rule", "rule_id", ruleID, "reason", "disabled")
			cancel()
			delete(s.runningRules, ruleID)
		}
	}

	// Start goroutines for new enabled rules
	runningRuleIDs := make(map[uint]bool)
	for ruleID := range s.runningRules {
		runningRuleIDs[ruleID] = true
	}

	for _, r := range rules {
		if !runningRuleIDs[r.ID] {
			slog.Info("Starting rule", "rule_id", r.ID, "rule_name", r.Name)
			ruleCtx, ruleCancel := context.WithCancel(s.ctx)
			s.runningRules[r.ID] = ruleCancel

			// Start goroutine for this rule
			s.wg.Add(1)
			go s.runRule(ruleCtx, r)
		}
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
	// Reload rule from database to get latest configuration
	rule, err := s.ruleService.GetByID(ruleID)
	if err != nil {
		slog.Error("Failed to reload rule", "rule_id", ruleID, "error", err)
		return
	}

	s.executeRule(ctx, rule)
}

// executeRule executes a single rule execution in a separate goroutine
func (s *Scheduler) executeRule(ctx context.Context, ruleModel *models.Rule) {
	// Execute in a separate goroutine to avoid blocking
	go func() {
		if err := s.executor.ExecuteRule(ctx, ruleModel); err != nil {
			slog.Error("Failed to execute rule", "rule_id", ruleModel.ID, "rule_name", ruleModel.Name, "error", err)
		}
	}()
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
	if err == nil && config != nil && config.Enabled {
		nextRun = s.nextRunTime(config.Hour, config.Minute)
		retentionDays = config.RetentionDays
		slog.Info("Cleanup task enabled", "scheduled_time", nextRun.Format("2006-01-02 15:04:05"), "retention_days", retentionDays)
	} else {
		slog.Info("Cleanup task disabled")
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
					slog.Debug("Cleanup task disabled, clearing next run")
					nextRun = nil
				}
				continue
			}

			// If configuration changed, reschedule
			newNextRun := s.nextRunTime(config.Hour, config.Minute)
			if nextRun == nil || !nextRun.Equal(*newNextRun) || retentionDays != config.RetentionDays {
				nextRun = newNextRun
				retentionDays = config.RetentionDays
				slog.Info("Cleanup task rescheduled", "scheduled_time", nextRun.Format("2006-01-02 15:04:05"), "retention_days", retentionDays)
			}

			// Check if it's time to run
			now := time.Now()
			if nextRun != nil {
				// Use truncate to minute for comparison to allow execution within the same minute
				nowTruncated := now.Truncate(time.Minute)
				nextRunTruncated := nextRun.Truncate(time.Minute)

				if nowTruncated.After(nextRunTruncated) || nowTruncated.Equal(nextRunTruncated) {
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
						"current_time", time.Now().Format("2006-01-02 15:04:05"))
				} else {
					// Log periodically for debugging (every 10 minutes)
					if now.Minute()%10 == 0 && now.Second() < 10 {
						slog.Debug("Cleanup task waiting",
							"current_time", now.Format("2006-01-02 15:04:05"),
							"next_run", nextRun.Format("2006-01-02 15:04:05"),
							"time_until_run", nextRun.Sub(now).String())
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

	// If today's scheduled time has passed, schedule for tomorrow
	if now.After(todayRun) || now.Equal(todayRun) {
		todayRun = todayRun.Add(24 * time.Hour)
	}

	return &todayRun
}
