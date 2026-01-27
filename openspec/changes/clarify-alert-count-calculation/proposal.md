# Change: Change Alert Count to Execution Count

## Why

The current implementation increments `alert_count` by the number of matched log entries (`originalLogCount`) each time a rule execution finds matches and successfully sends an alert. However, users expect "告警次数" (Alert Count) to represent the number of times alerts were triggered (execution count), not the total number of matched log entries.

This change aligns the implementation with user expectations: each successful alert notification should increment the count by 1, regardless of how many log entries were matched.

## What Changes

- **MODIFIED**: Update `Requirement: Rule Statistics` in `specs/rule-management/spec.md` to state that `alert_count` represents the number of alert records created (execution count).
- **MODIFIED**: Change `backend/internal/worker/executor/executor.go` to increment `alert_count` by 1 instead of `originalLogCount`.
- Add scenarios that demonstrate this behavior with concrete examples.

## Impact

- Affected specs: `rule-management`
- Affected code: `backend/internal/worker/executor/executor.go`
- User-facing: "告警次数" now represents the number of times alerts were triggered, making it easier to understand rule activity frequency
