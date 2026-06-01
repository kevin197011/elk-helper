// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package executor

import (
	"time"

	"github.com/kk/elk-helper/backend/internal/models"
)

const (
	// minLookbackWhenNoLastRun is used only when rule.Interval is invalid (<=0).
	minLookbackWhenNoLastRun = 5 * time.Minute
	// minRuleInterval matches scheduler minimum tick / interval floor.
	minRuleInterval = 10 * time.Second
	// esWindowOverlap prevents missing logs at bucket boundaries after a successful run.
	esWindowOverlap = 2 * time.Second
)

func ruleIntervalDuration(intervalSeconds int) time.Duration {
	d := time.Duration(intervalSeconds) * time.Second
	if d < minRuleInterval {
		return minRuleInterval
	}
	return d
}

// hasSuccessfulRun means at least one ES query completed and advanced the cursor.
func hasSuccessfulRun(rule *models.Rule) bool {
	return rule.RunCount > 0 && rule.LastRunTime != nil
}

// computeESWindowStart is the lower bound for @timestamp when querying Elasticsearch.
// Until the first successful run, always scan the last full interval window (never a sliding 5m stub).
func computeESWindowStart(rule *models.Rule, now time.Time) time.Time {
	if hasSuccessfulRun(rule) {
		return rule.LastRunTime.Add(-esWindowOverlap)
	}
	lookback := ruleIntervalDuration(rule.Interval)
	if lookback <= 0 {
		lookback = minLookbackWhenNoLastRun
	}
	return now.Add(-lookback)
}

// shouldRunNow applies the interval gate before hitting ES.
// - First tick (no last_run_time): always run.
// - Never succeeded (run_count=0) but last_run_time set: throttle retries using last_run_time as attempt marker only.
// - After success: throttle from last successful run time.
func shouldRunNow(rule *models.Rule, now time.Time, forceExecute bool) bool {
	if forceExecute {
		return true
	}
	required := ruleIntervalDuration(rule.Interval)
	if rule.LastRunTime == nil {
		return true
	}
	if !hasSuccessfulRun(rule) {
		return now.Sub(*rule.LastRunTime) >= required
	}
	anchor := rule.LastRunTime.Add(-esWindowOverlap)
	return now.Sub(anchor) >= required
}
