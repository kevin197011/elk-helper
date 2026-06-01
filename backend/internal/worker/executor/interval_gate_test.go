// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package executor

import (
	"testing"
	"time"

	"github.com/kk/elk-helper/backend/internal/models"
)

func TestComputeESWindowStart_neverSucceededUsesFullInterval(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	attempt := now.Add(-5 * time.Minute)
	rule := &models.Rule{Interval: 600, LastRunTime: &attempt, RunCount: 0}

	got := computeESWindowStart(rule, now)
	want := now.Add(-10 * time.Minute)
	if !got.Equal(want) {
		t.Fatalf("ES window start = %v, want %v (must not use sliding 5m attempt marker)", got, want)
	}
}

func TestShouldRunNow_neverSucceededThrottlesOnAttemptMarker(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	attempt := now.Add(-5 * time.Minute)
	rule := &models.Rule{Interval: 600, LastRunTime: &attempt, RunCount: 0}

	if shouldRunNow(rule, now, false) {
		t.Fatal("should wait for interval after failed attempt")
	}
	attempt = now.Add(-11 * time.Minute)
	rule.LastRunTime = &attempt
	if !shouldRunNow(rule, now, false) {
		t.Fatal("should run after interval elapsed since attempt")
	}
}

func TestShouldRunNow_noLastRunTimeAlwaysRuns(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	rule := &models.Rule{Interval: 600}
	if !shouldRunNow(rule, now, false) {
		t.Fatal("first tick must run")
	}
}

func TestShouldRunNow_oldSlidingFiveMinuteDeadlock(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	// Legacy bug: anchor always ~5m ago while interval 10m — with new logic, nil last_run runs;
	// if we wrongly used 5m sliding without attempt marker, first run still passes when LastRunTime nil.
	rule := &models.Rule{Interval: 600}
	if !shouldRunNow(rule, now, false) {
		t.Fatal("nil last_run_time must not deadlock")
	}
}

func TestHasSuccessfulRun_andESWindowAfterSuccess(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	last := now.Add(-15 * time.Minute)
	rule := &models.Rule{Interval: 600, LastRunTime: &last, RunCount: 3}

	if !hasSuccessfulRun(rule) {
		t.Fatal("expected successful run")
	}
	got := computeESWindowStart(rule, now)
	want := last.Add(-esWindowOverlap)
	if !got.Equal(want) {
		t.Fatalf("ES from success cursor: got %v want %v", got, want)
	}
	if !shouldRunNow(rule, now, false) {
		t.Fatal("should run when interval elapsed since last success")
	}
}

func TestShouldRunNow_afterSuccessRespectsInterval(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	last := now.Add(-3 * time.Minute)
	rule := &models.Rule{Interval: 600, LastRunTime: &last, RunCount: 1}
	if shouldRunNow(rule, now, false) {
		t.Fatal("should skip when interval not elapsed since last success")
	}
	last = now.Add(-11 * time.Minute)
	rule.LastRunTime = &last
	if !shouldRunNow(rule, now, false) {
		t.Fatal("should run when interval elapsed")
	}
}
