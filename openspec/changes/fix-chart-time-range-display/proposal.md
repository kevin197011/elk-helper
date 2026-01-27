# Change: Fix Chart Time Range Display to Show Current Time Minus 24 Hours

## Why

The current implementation generates time buckets starting from an aligned time boundary (aligned to bucket intervals), which causes the chart to display time labels that don't accurately represent "current time minus 24 hours". Users see time ranges like "00:00 to 23:00" instead of the actual rolling 24-hour window from the current time.

The chart should display time labels that reflect the actual data range: from 24 hours ago to the current time, not aligned to day boundaries.

## What Changes

- **MODIFIED**: Update `backend/internal/service/alert/alert.go` to generate time buckets starting from actual current time minus 24 hours (not aligned to bucket boundaries), while still maintaining bucket alignment for consistent indexing.
- **MODIFIED**: Update time label generation to ensure the first and last buckets represent the actual time range boundaries.
- **MODIFIED**: Update `Requirement: Chart Visualization` in `specs/dashboard/spec.md` to specify that X-axis labels must reflect the actual 24-hour rolling window from current time.
- Add scenarios demonstrating correct time range display.

## Impact

- Affected specs: `dashboard`
- Affected code: `backend/internal/service/alert/alert.go`
- User-facing: Chart X-axis will now show the actual rolling 24-hour window (e.g., "14:30 yesterday to 14:30 today") instead of day-aligned times (e.g., "00:00 to 23:00")
