# Rule Management

## Purpose

Provide comprehensive rule management capabilities for defining, testing, cloning, and managing alert rules.

## Requirements

### Requirement: Rule Creation

The system SHALL allow users to create new alert rules.

The system SHALL require the following fields:
- Rule name (unique, required)
- Index pattern (required)
- Query conditions (JSON array, required)
- Execution interval in seconds (required, default: 60)
- ES data source configuration (optional)
- Lark notification configuration (optional)
- Description (optional)
- Enabled status (default: true)

#### Scenario: Create rule with all fields
- **WHEN** a user provides all required fields with valid values
- **THEN** the rule is created and saved to database
- **AND** a unique ID is assigned
- **AND** the rule is returned in the response

#### Scenario: Create rule with duplicate name
- **WHEN** a user attempts to create a rule with an existing name
- **THEN** the system returns HTTP 409 Conflict
- **AND** an error message indicates the name is already taken

#### Scenario: Create rule with invalid query
- **WHEN** a user provides invalid JSON for query conditions
- **THEN** the system returns HTTP 400 Bad Request
- **AND** validation error details are provided

### Requirement: Rule Query Conditions

The system SHALL support flexible query conditions in JSON format.

The system SHALL support the following operators:
- `==` or `equals`: Exact match
- `!=`: Not equal
- `>`: Greater than
- `>=`: Greater than or equal
- `<`: Less than
- `<=`: Less than or equal
- `contains`: String contains
- `not_contains`: String does not contain

The system SHALL support AND/OR logic combinations via `logic` field.

The system SHALL accept both `operator` and `op` field names for backward compatibility.

#### Scenario: Simple equality query
- **WHEN** a rule contains a query condition: `{"field": "response_code", "operator": "==", "value": 500}`
- **THEN** the system generates ES query: `{"term": {"response_code": 500}}`

#### Scenario: Range query
- **WHEN** a rule contains: `{"field": "responsetime", "operator": ">", "value": 3}`
- **THEN** the system generates ES query: `{"range": {"responsetime": {"gt": 3}}}`

#### Scenario: AND logic combination
- **WHEN** multiple conditions have `"logic": "and"` (except first)
- **THEN** all conditions are combined with AND logic in ES query

#### Scenario: OR logic combination
- **WHEN** multiple conditions have `"logic": "or"`
- **THEN** conditions are combined with OR logic in ES query

#### Scenario: Contains operator
- **WHEN** a rule contains: `{"field": "message", "operator": "contains", "value": "ERROR"}`
- **THEN** the system generates ES query: `{"wildcard": {"message": "*ERROR*"}}`

### Requirement: Rule Update

The system SHALL allow users to update existing rules.

The system SHALL validate all fields before saving.

The system SHALL preserve rule statistics (run_count, alert_count, last_run_time) during update.

Updated rules SHALL take effect on next scheduled execution without service restart.

#### Scenario: Update rule configuration
- **WHEN** a user updates rule fields (name excluded, if name change creates conflict)
- **THEN** the rule is updated in database
- **AND** changes take effect on next scheduler run

#### Scenario: Update rule name to existing name
- **WHEN** a user attempts to change rule name to an existing name
- **THEN** the system returns HTTP 409 Conflict
- **AND** the rule name is not changed

### Requirement: Rule Deletion

The system SHALL allow users to delete rules.

The system SHALL perform hard delete (permanently remove from database).

The system SHALL delete all associated alerts when a rule is deleted (cascade delete).

#### Scenario: Delete rule
- **WHEN** a user deletes a rule
- **THEN** the rule is permanently removed from database
- **AND** all associated alerts are also deleted
- **AND** HTTP 204 No Content is returned

#### Scenario: Delete non-existent rule
- **WHEN** a user attempts to delete a rule that does not exist
- **THEN** the system returns HTTP 404 Not Found

### Requirement: Rule Cloning

The system SHALL allow users to clone existing rules.

The cloned rule SHALL inherit all configuration from the original rule.

The cloned rule SHALL have a new name provided by the user.

The cloned rule SHALL inherit the enabled/disabled status from the original rule.

The cloned rule SHALL have reset statistics (run_count=0, alert_count=0, last_run_time=null).

#### Scenario: Clone rule
- **WHEN** a user clones a rule with a new name
- **THEN** a new rule is created with all fields copied
- **AND** the new rule has the provided name
- **AND** statistics are reset to zero
- **AND** enabled status matches the original rule

#### Scenario: Clone rule with duplicate name
- **WHEN** a user attempts to clone a rule with an existing name
- **THEN** the system returns HTTP 409 Conflict
- **AND** the clone operation is aborted

### Requirement: Rule Testing

The system SHALL provide a test endpoint to validate query conditions.

The test endpoint SHALL execute the query against the configured ES data source.

The test endpoint SHALL return matched log entries without creating an alert.

#### Scenario: Test valid query
- **WHEN** a user tests a rule with valid query conditions
- **THEN** the system executes the query against ES
- **AND** returns matched log entries (limited to first 10)
- **AND** no alert record is created

#### Scenario: Test query with no matches
- **WHEN** a user tests a query that matches no logs
- **THEN** the system returns empty results
- **AND** indicates zero matches

#### Scenario: Test query with ES connection error
- **WHEN** a user tests a query but ES is unreachable
- **THEN** the system returns an error message
- **AND** indicates connection failure

### Requirement: Rule Enable/Disable Toggle

The system SHALL allow users to enable or disable rules.

Disabled rules SHALL not be executed by the scheduler.

The system SHALL update the rule status immediately in the database.

#### Scenario: Disable rule
- **WHEN** a user disables an active rule
- **THEN** the rule status is updated to disabled
- **AND** the scheduler stops executing this rule on next check

#### Scenario: Enable rule
- **WHEN** a user enables a disabled rule
- **THEN** the rule status is updated to enabled
- **AND** the scheduler starts executing this rule on next check

### Requirement: Rule List and Search

The system SHALL provide a list endpoint to retrieve all rules.

The system SHALL support pagination (page, page_size).

The system SHALL include rule statistics in the list response.

#### Scenario: List all rules
- **WHEN** a user requests the rule list
- **THEN** the system returns all rules with pagination
- **AND** each rule includes statistics (run_count, alert_count, last_run_time)

#### Scenario: List rules with pagination
- **WHEN** a user requests page 2 with page_size 20
- **THEN** the system returns rules 21-40
- **AND** pagination metadata is included

### Requirement: Rule Statistics

The system SHALL track and update rule execution statistics.

Statistics SHALL include:
- `run_count`: Total number of rule executions
- `alert_count`: Total number of alerts triggered
- `last_run_time`: Timestamp of last execution

#### Scenario: Update statistics on execution
- **WHEN** a rule is executed by the scheduler
- **THEN** run_count is incremented
- **AND** last_run_time is updated

#### Scenario: Update statistics on alert
- **WHEN** a rule execution matches logs and creates an alert
- **THEN** alert_count is incremented
- **AND** run_count is incremented

### Requirement: Rule Real-Time Updates

The system SHALL support real-time rule configuration updates.

Modified rules SHALL be loaded by the scheduler on next execution cycle.

The system SHALL NOT require service restart for rule changes to take effect.

#### Scenario: Update rule interval
- **WHEN** a user changes rule execution interval
- **THEN** the new interval takes effect on next scheduler cycle
- **AND** no service restart is required

#### Scenario: Update rule query conditions
- **WHEN** a user modifies rule query conditions
- **THEN** new conditions are used in next execution
- **AND** no service restart is required

### Requirement: Quick Template Buttons

The frontend SHALL provide quick template buttons for common query patterns:
- Non-200 responses
- 4xx errors
- 5xx errors
- Slow queries (>3 seconds)
- 499 errors

Each template button SHALL populate the query editor with pre-configured JSON.

#### Scenario: Use 5xx error template
- **WHEN** a user clicks the "5xx错误" template button
- **THEN** the query editor is populated with: `[{"field": "response_code", "operator": ">=", "value": 500}]`

#### Scenario: Use slow query template
- **WHEN** a user clicks the "慢查询" template button
- **THEN** the query editor is populated with conditions for responsetime > 3 AND response_code == 200

