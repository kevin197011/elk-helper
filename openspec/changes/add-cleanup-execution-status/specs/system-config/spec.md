## MODIFIED Requirements

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

## ADDED Requirements

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

