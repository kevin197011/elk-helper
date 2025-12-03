// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kk/elk-helper/backend/internal/models"
	"github.com/kk/elk-helper/backend/internal/service/alert"
	"github.com/kk/elk-helper/backend/internal/service/esconfig"
	"github.com/kk/elk-helper/backend/internal/service/query"
	"github.com/kk/elk-helper/backend/internal/service/rule"
	"github.com/kk/elk-helper/backend/internal/service/systemconfig"
	"github.com/kk/elk-helper/backend/internal/worker/executor"
)

// Scheduler manages rule execution schedule
type Scheduler struct {
	ruleService       *rule.Service
	queryService      *query.Service
	alertService      *alert.Service
	systemConfigService *system_config.Service
	executor          *executor.Executor
	ctx               context.Context
	cancel            context.CancelFunc
	wg                sync.WaitGroup
	mu                sync.RWMutex
	runningRules      map[uint]context.CancelFunc // Track running rule goroutines
	checkInterval     time.Duration
}

// NewScheduler creates a new scheduler
func NewScheduler(ruleService *rule.Service, queryService *query.Service, esConfigService *es_config.Service, alertService *alert.Service, systemConfigService *system_config.Service, checkInterval time.Duration, retryTimes, batchSize int) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	return &Scheduler{
		ruleService:        ruleService,
		queryService:       queryService,
		alertService:       alertService,
		systemConfigService: systemConfigService,
		executor:           executor.NewExecutor(queryService, esConfigService, ruleService, alertService, retryTimes, batchSize),
		ctx:                ctx,
		cancel:             cancel,
		runningRules:       make(map[uint]context.CancelFunc),
		checkInterval:      checkInterval,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	fmt.Printf("[INFO] Scheduler started, checking for rule changes every %v\n", s.checkInterval)

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
		fmt.Printf("[INFO] Stopping rule %d\n", ruleID)
		cancel()
	}
	s.mu.Unlock()

	s.wg.Wait()
	fmt.Println("[INFO] Scheduler stopped")
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
		fmt.Printf("[ERROR] Failed to get enabled rules: %v\n", err)
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
			fmt.Printf("[INFO] Stopping rule %d (disabled)\n", ruleID)
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
			fmt.Printf("[INFO] Starting rule %d (%s)\n", r.ID, r.Name)
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
			fmt.Printf("[INFO] Rule %d (%s) stopped\n", ruleID, ruleName)
			return
		case <-ticker.C:
			// Reload rule from database to get latest configuration
			rule, err := s.ruleService.GetByID(ruleID)
			if err != nil {
				fmt.Printf("[ERROR] Failed to reload rule %d: %v\n", ruleID, err)
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
				fmt.Printf("[INFO] Rule %d (%s) interval updated to %v\n", ruleID, rule.Name, interval)
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
		fmt.Printf("[ERROR] Failed to reload rule %d: %v\n", ruleID, err)
		return
	}

	s.executeRule(ctx, rule)
}

// executeRule executes a single rule execution in a separate goroutine
func (s *Scheduler) executeRule(ctx context.Context, ruleModel *models.Rule) {
	// Execute in a separate goroutine to avoid blocking
	go func() {
		if err := s.executor.ExecuteRule(ctx, ruleModel); err != nil {
			fmt.Printf("[ERROR] Failed to execute rule %d (%s): %v\n", ruleModel.ID, ruleModel.Name, err)
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
		fmt.Printf("[INFO] Cleanup task enabled: scheduled for %s (deletes alerts older than %d days)\n",
			nextRun.Format("2006-01-02 15:04:05"), retentionDays)
	} else {
		fmt.Printf("[INFO] Cleanup task disabled\n")
	}

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			// Reload configuration
			config, err := s.systemConfigService.GetCleanupConfig()
			if err != nil {
				fmt.Printf("[ERROR] Failed to load cleanup config: %v\n", err)
				continue
			}

			if !config.Enabled {
				nextRun = nil
				continue
			}

			// If configuration changed, reschedule
			newNextRun := s.nextRunTime(config.Hour, config.Minute)
			if nextRun == nil || !nextRun.Equal(*newNextRun) || retentionDays != config.RetentionDays {
				nextRun = newNextRun
				retentionDays = config.RetentionDays
				fmt.Printf("[INFO] Cleanup task rescheduled for %s (deletes alerts older than %d days)\n",
					nextRun.Format("2006-01-02 15:04:05"), retentionDays)
			}

			// Check if it's time to run
			if nextRun != nil && time.Now().After(*nextRun) {
				// Execute cleanup
				retentionDuration := time.Duration(retentionDays) * 24 * time.Hour
				rowsAffected, err := s.alertService.CleanupOldData(retentionDuration)
				if err != nil {
					fmt.Printf("[ERROR] Failed to cleanup old alerts: %v\n", err)
				} else {
					fmt.Printf("[INFO] Cleanup task completed: deleted %d alerts older than %d days\n", rowsAffected, retentionDays)
				}

				// Schedule next run
				nextRun = s.nextRunTime(config.Hour, config.Minute)
				fmt.Printf("[INFO] Next cleanup task scheduled for %s\n", nextRun.Format("2006-01-02 15:04:05"))
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
