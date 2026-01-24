# System Configuration

## Purpose

Manage system-wide configuration settings including cleanup tasks and maintenance operations.

## Requirements

### Requirement: Cleanup Task Configuration

The system SHALL allow users to configure automatic cleanup of old alert records.

The system SHALL support the following configuration options:
- Enable/disable cleanup (boolean)
- Execution time (HH:MM format, 24-hour clock)
- Retention period (days)
- Last execution status (read-only, tracked automatically)
- Last execution time (read-only, tracked automatically)
- Last execution result (read-only, tracked automatically)

#### Scenario: Enable cleanup with retention
- **WHEN** a user enables cleanup with 30-day retention and execution time 03:00
- **THEN** system deletes alerts older than 30 days daily at 3:00 AM
- **AND** configuration is persisted
- **AND** execution status is tracked and displayed

#### Scenario: Disable cleanup
- **WHEN** a user disables cleanup
- **THEN** no automatic deletion occurs
- **AND** alerts are retained indefinitely

#### Scenario: Update cleanup schedule
- **WHEN** a user changes cleanup execution time
- **THEN** new schedule takes effect on next day
- **AND** previous cleanup tasks are cancelled

### Requirement: Cleanup Task Execution

The system SHALL execute cleanup tasks at configured time.

The system SHALL delete alerts older than retention period.

The system SHALL log cleanup operations (count of deleted records).

The system SHALL handle cleanup failures gracefully without affecting other operations.

The system SHALL track and update execution status after each cleanup execution.

#### Scenario: Execute cleanup
- **WHEN** cleanup time is reached and cleanup is enabled
- **THEN** system queries alerts older than retention period
- **AND** deletes matching records
- **AND** logs deletion count
- **AND** updates execution status to "success" with result description

#### Scenario: Cleanup with no records
- **WHEN** cleanup executes but no records match retention criteria
- **THEN** no deletion occurs
- **AND** cleanup completes successfully
- **AND** execution status is updated with "no data to clean" message

#### Scenario: Cleanup execution failure
- **WHEN** cleanup task fails due to database error
- **THEN** execution status is set to "failed"
- **AND** error message is recorded in execution result
- **AND** system continues normal operation

### Requirement: Cleanup Task Execution Status Tracking

The system SHALL track and display the last execution status of cleanup tasks.

The system SHALL record execution status after each cleanup execution (both automatic and manual).

The execution status SHALL include:
- Status: "success", "failed", or "never" (if not yet executed)
- Execution time: Timestamp of last execution
- Result description: Success message with deletion count, or error message on failure

The system SHALL update execution status immediately after cleanup completion.

#### Scenario: Track successful execution
- **WHEN** cleanup task executes successfully and deletes 150 alerts
- **THEN** execution status is set to "success"
- **AND** execution time is recorded
- **AND** result shows "成功删除 150 条告警数据"

#### Scenario: Track failed execution
- **WHEN** cleanup task fails due to database error
- **THEN** execution status is set to "failed"
- **AND** execution time is recorded
- **AND** result shows error message

#### Scenario: Display execution status
- **WHEN** user views cleanup configuration page
- **THEN** last execution status is displayed with appropriate visual indicator
- **AND** shows execution time and result description
- **AND** uses color coding (green for success, red for failure, gray for never executed)

#### Scenario: Status after manual cleanup
- **WHEN** user triggers manual cleanup
- **THEN** execution status is updated immediately after completion
- **AND** UI refreshes to show updated status

### Requirement: System Status Monitoring

The system SHALL provide status endpoint showing system health.

The status endpoint SHALL include:
- Total rule count
- Enabled rule count
- 24-hour alert statistics
- Elasticsearch data source status (success count, total count)

#### Scenario: Get system status
- **WHEN** a user requests system status
- **THEN** system returns comprehensive status information
- **AND** includes rule statistics and ES connectivity status

#### Scenario: ES connectivity check
- **WHEN** system status is requested
- **THEN** system checks connectivity to all configured ES data sources
- **AND** reports success/failure for each

### Requirement: Configuration Persistence

The system SHALL persist system configuration in database.

The system SHALL allow updating configuration through API.

Configuration changes SHALL take effect immediately.

#### Scenario: Save cleanup config
- **WHEN** a user updates cleanup configuration
- **THEN** configuration is saved to database
- **AND** changes are applied immediately

#### Scenario: Retrieve cleanup config
- **WHEN** a user requests cleanup configuration
- **THEN** system returns current configuration values
- **AND** includes all configured fields

