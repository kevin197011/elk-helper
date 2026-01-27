# Alerting

## Purpose

Handle alert generation, storage, notification delivery, and historical tracking when rule conditions are met.

## Requirements

### Requirement: Alert Generation

The system SHALL automatically generate alerts when rule query conditions match log entries.

The system SHALL create an alert record for each rule execution that finds matching logs.

The alert record SHALL include:
- Rule ID reference
- Index name that was queried
- Log count (number of matched entries)
- Log data (actual log entries, limited to configured batch size)
- Time range of the query window
- Status (sent/failed)
- Error message (if notification failed)

#### Scenario: Generate alert on match
- **WHEN** a rule execution finds matching log entries
- **THEN** an alert record is created with all matched logs
- **AND** log_count reflects the number of entries
- **AND** time_range indicates the query window

#### Scenario: No alert on no match
- **WHEN** a rule execution finds no matching log entries
- **THEN** no alert record is created
- **AND** rule statistics are still updated (run_count incremented)

### Requirement: Alert Notification Delivery

The system SHALL send notifications via Lark Webhook when alerts are generated.

The system SHALL use the configured Lark Webhook URL from rule configuration.

The system SHALL format alert messages according to log type (nginx, java, go, etc.).

The system SHALL include only the first 3 log entries in the notification for brevity.

The system SHALL retry failed notification attempts according to configured retry policy.

#### Scenario: Send notification successfully
- **WHEN** an alert is generated and notification is sent successfully
- **THEN** alert status is set to "sent"
- **AND** notification is delivered to Lark/Feishu group

#### Scenario: Notification failure
- **WHEN** notification delivery fails (network error, invalid webhook, etc.)
- **THEN** alert status is set to "failed"
- **AND** error message is stored in error_msg field
- **AND** retry is attempted according to policy

#### Scenario: Retry on failure
- **WHEN** initial notification fails
- **THEN** system retries up to configured retry times
- **AND** status is updated to "sent" if retry succeeds
- **AND** status remains "failed" if all retries fail

### Requirement: Alert Message Formatting

The system SHALL automatically detect log type from log entries.

The system SHALL extract key fields based on log type:
- Nginx: response_code, request, method, ip, cf_ray, @timestamp
- Java: module, node_ip, message, level, @timestamp
- Go/Application: message, level, module, @timestamp

The system SHALL format timestamps as YYYY-MM-DD HH:MM:SS.

The system SHALL truncate request URLs to 50 characters if longer.

The system SHALL highlight response_code in red color for visual emphasis.

The system SHALL always display cf_ray field, showing "-" if value is missing.

#### Scenario: Format nginx log alert
- **WHEN** alert contains nginx log entries
- **THEN** notification includes: @timestamp, request (URL path only), cf_ray, response_code
- **AND** response_code is highlighted
- **AND** only first 3 entries are shown

#### Scenario: Format java log alert
- **WHEN** alert contains java application logs
- **THEN** notification includes: module, node_ip, message, @timestamp
- **AND** message is truncated if too long

### Requirement: Alert History Storage

The system SHALL store all alert records in the database.

The system SHALL maintain alert records even after rule deletion (via cascade delete).

Alert records SHALL be queryable by:
- Rule ID
- Status (sent/failed)
- Time range
- Index name

#### Scenario: Store alert record
- **WHEN** an alert is generated
- **THEN** record is saved to alerts table
- **AND** rule_id foreign key references the triggering rule

#### Scenario: Query alerts by rule
- **WHEN** a user queries alerts for a specific rule
- **THEN** system returns all alerts associated with that rule
- **AND** results are paginated

### Requirement: Alert History Retrieval

The system SHALL provide an endpoint to list alert history.

The system SHALL support pagination with configurable page size (default: 20).

The system SHALL NOT load logs field in list queries to improve performance (99% data reduction).

The system SHALL provide an endpoint to retrieve full alert details including logs.

The detail endpoint SHALL limit logs to first 10 entries for performance.

#### Scenario: List alerts without logs
- **WHEN** a user requests alert list
- **THEN** system returns alert metadata without logs field
- **AND** response is fast (<0.5 seconds)
- **AND** data size is minimal

#### Scenario: Get alert details
- **WHEN** a user requests alert details by ID
- **THEN** system returns full alert record including logs
- **AND** logs are limited to first 10 entries

#### Scenario: Paginate alert list
- **WHEN** a user requests page 2 with 20 items per page
- **THEN** system returns alerts 21-40
- **AND** pagination metadata includes total count and total pages

### Requirement: Alert Search and Filtering

The system SHALL support searching alerts by:
- Rule name
- Index name
- Time range

The system SHALL support filtering by status (all, sent, failed).

#### Scenario: Search by rule name
- **WHEN** a user searches for alerts with rule name containing "nginx"
- **THEN** system returns matching alerts
- **AND** search is case-insensitive

#### Scenario: Filter by status
- **WHEN** a user filters alerts by status "failed"
- **THEN** system returns only failed alerts
- **AND** pagination works correctly with filtered results

### Requirement: Alert Deletion

The system SHALL allow users to delete alert records.

The system SHALL perform hard delete (permanently remove from database).

Deleted alerts SHALL not be recoverable.

#### Scenario: Delete alert
- **WHEN** a user deletes an alert
- **THEN** alert is permanently removed from database
- **AND** HTTP 204 No Content is returned

### Requirement: Alert Statistics

The system SHALL provide time-series statistics for rule alerts.

The system SHALL aggregate alert execution counts by time buckets with intervals dynamically calculated based on each rule's execution frequency.

The system SHALL calculate bucket interval as approximately 5x the rule's execution interval, rounded to common intervals (5, 15, 30, or 60 minutes).

The system SHALL include all enabled rules, showing 0 for rules with no alerts.

The system SHALL return data for the last 24 hours by default.

#### Scenario: Get time-series stats
- **WHEN** a user requests rule time-series statistics
- **THEN** system returns aggregated data per rule per time bucket
- **AND** rules with no alerts show value 0 for all buckets
- **AND** alert execution count is aggregated using COUNT(*) (counting alert records)
- **AND** each rule uses time bucket interval appropriate for its execution frequency

#### Scenario: Time bucket aggregation with dynamic intervals
- **WHEN** alerts are generated at various times
- **THEN** system groups alerts into time buckets with interval based on each rule's execution interval
- **AND** each bucket shows alert execution count for that period
- **AND** rules with 60-second interval use 5-minute buckets
- **AND** rules with 3600-second interval use 60-minute buckets

### Requirement: Alert Cleanup

The system SHALL support automatic cleanup of old alert records.

The system SHALL allow configuration of retention period (days).

The system SHALL execute cleanup at configured time (e.g., 03:00 daily).

#### Scenario: Auto cleanup enabled
- **WHEN** cleanup is enabled with 30-day retention
- **THEN** system deletes alerts older than 30 days daily at configured time

#### Scenario: Cleanup disabled
- **WHEN** cleanup is disabled
- **THEN** no automatic deletion occurs
- **AND** alerts are retained indefinitely

