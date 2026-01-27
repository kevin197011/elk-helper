## MODIFIED Requirements

### Requirement: Chart Data Aggregation

The system SHALL aggregate alert execution counts by time buckets.

The system SHALL use COUNT(*) for aggregation (counting alert records, not log entries).

The system SHALL calculate time bucket interval dynamically for each rule based on its execution interval.

The system SHALL use bucket interval of approximately 5x the rule's execution interval, rounded to common intervals (5, 15, 30, or 60 minutes).

The system SHALL generate time buckets for a time range from current time minus 24 hours to current time.

**The time range SHALL be calculated as: [current_time - 24_hours, current_time], using actual current time (not aligned to bucket boundaries) for the query boundary.**

The system SHALL fill missing buckets with value 0.

#### Scenario: Aggregate data with dynamic intervals
- **WHEN** alerts are generated at various times
- **THEN** system groups alerts into time buckets with interval based on each rule's execution frequency
- **AND** each bucket shows alert execution count (COUNT of alert records)
- **AND** buckets with no alerts show 0
- **AND** rules with shorter execution intervals use finer-grained buckets (e.g., 5 minutes)
- **AND** rules with longer execution intervals use coarser-grained buckets (e.g., 60 minutes)

#### Scenario: Time range is exactly 24 hours from current time
- **WHEN** user views the chart at 14:30:45
- **THEN** chart displays data from 14:30:45 yesterday to 14:30:45 today
- **AND** query uses actual current time (14:30:45) as the end boundary, not aligned to bucket boundaries
- **AND** time buckets are generated for display purposes but do not affect the actual data query range

#### Scenario: Fill missing buckets
- **WHEN** a rule has no alerts in certain time periods
- **THEN** chart displays 0 for those time buckets
- **AND** chart remains continuous
- **AND** time range still represents exactly 24 hours from current time
