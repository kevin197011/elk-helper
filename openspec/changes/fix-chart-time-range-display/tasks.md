## 1. Implementation
- [x] 1.1 Modify `backend/internal/service/alert/alert.go` to generate time buckets starting from actual current time minus 24 hours
- [x] 1.2 Ensure first bucket label shows the actual start time (24 hours ago) and last bucket shows current time
- [x] 1.3 Update `Requirement: Chart Visualization` in `specs/dashboard/spec.md` to specify rolling 24-hour window display
- [x] 1.4 Add scenario demonstrating correct time range display
