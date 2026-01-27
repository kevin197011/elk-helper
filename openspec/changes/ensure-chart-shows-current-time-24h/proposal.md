# Change: Ensure Chart Shows Current Time Minus 24 Hours

## Why

The current implementation of the rule alert trend chart uses time bucket alignment which may cause the displayed time range to not precisely represent "current time minus 24 hours". The chart should always show data from exactly 24 hours ago to the current time, without being affected by time bucket boundary alignment.

This ensures users see the most recent 24 hours of data accurately, making the chart more useful for real-time monitoring.

## What Changes

- **MODIFIED**: Update `Requirement: Chart Data Aggregation` in `specs/dashboard/spec.md` to explicitly state that the time range must be from current time minus 24 hours to current time.
- **MODIFIED**: Update `backend/internal/service/alert/alert.go` to ensure time range calculation uses actual current time (not aligned time) for the end boundary, while still using aligned time for bucket generation.
- Add scenarios demonstrating the precise time range behavior.

## Impact

- Affected specs: `dashboard`
- Affected code: `backend/internal/service/alert/alert.go`
- User-facing: Chart will always show exactly the last 24 hours from current time, making it more accurate for real-time monitoring
