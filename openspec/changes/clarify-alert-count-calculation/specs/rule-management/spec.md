## MODIFIED Requirements

### Requirement: Rule Statistics

The system SHALL track and update rule execution statistics.

Statistics SHALL include:
- `run_count`: Total number of rule executions
- `alert_count`: Total number of alert records created (execution count - one per successful alert notification)
- `last_run_time`: Timestamp of last execution

**Note**: `alert_count` represents the number of times alerts were successfully triggered, not the total number of matched log entries. Each successful alert notification increments the count by 1, regardless of how many log entries were matched in that execution.

#### Scenario: Update statistics on execution
- **WHEN** a rule is executed by the scheduler
- **THEN** run_count is incremented
- **AND** last_run_time is updated

#### Scenario: Update statistics on alert
- **WHEN** a rule execution matches logs and successfully creates an alert
- **THEN** alert_count is incremented by 1 (one per alert record)
- **AND** run_count is incremented

#### Scenario: Alert count represents execution count
- **WHEN** a rule execution finds 10 matching log entries and successfully sends an alert
- **THEN** alert_count increases by 1
- **AND** when the same rule later finds 5 matching log entries and sends another alert
- **THEN** alert_count increases by 1 (total becomes 2)
- **AND** the rules page displays alert_count as 2, representing 2 alert notifications triggered, regardless of the total number of matched log entries (15 in this example)
