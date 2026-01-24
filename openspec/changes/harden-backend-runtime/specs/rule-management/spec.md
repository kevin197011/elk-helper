## ADDED Requirements

### Requirement: Rule Execution Concurrency Limit

The system SHALL enforce a global maximum concurrency for rule executions.

The maximum concurrency SHALL be configurable via environment (e.g., `WORKER_MAX_CONCURRENCY`).

When the maximum concurrency is reached, additional rule executions SHALL wait until capacity is available.

#### Scenario: Concurrency under load
- **WHEN** multiple enabled rules are scheduled to execute concurrently
- **AND** the number of in-flight executions reaches `WORKER_MAX_CONCURRENCY`
- **THEN** additional executions are queued (wait) until a slot is available
- **AND** the service remains stable without unbounded goroutine growth

