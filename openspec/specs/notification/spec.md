# Notification Management

## Purpose

Manage notification channel configurations for delivering alerts to external systems (Lark/Feishu).

## Requirements

### Requirement: Lark Config Creation

The system SHALL allow users to create Lark/Feishu Webhook configurations.

The system SHALL require the following fields:
- Name (unique, required)
- Webhook URL (required, must be valid HTTP/HTTPS URL)

The system SHALL validate webhook URL format before saving.

#### Scenario: Create Lark config
- **WHEN** a user creates config with valid webhook URL
- **THEN** configuration is saved
- **AND** can be used in rule configurations

#### Scenario: Create config with invalid URL
- **WHEN** a user provides invalid webhook URL format
- **THEN** system returns validation error
- **AND** configuration is not saved

#### Scenario: Create config with duplicate name
- **WHEN** a user attempts to create config with existing name
- **THEN** system returns HTTP 409 Conflict
- **AND** error indicates name already exists

### Requirement: Lark Config Update

The system SHALL allow users to update existing Lark configurations.

The system SHALL validate new webhook URL before saving.

Updated configurations SHALL be used immediately for new notifications.

#### Scenario: Update webhook URL
- **WHEN** a user updates webhook URL in configuration
- **THEN** new URL is used for subsequent notifications
- **AND** change takes effect immediately

### Requirement: Lark Config Deletion

The system SHALL allow users to delete Lark configurations.

The system SHALL prevent deletion if configuration is referenced by active rules.

#### Scenario: Delete unused config
- **WHEN** a user deletes a Lark config not used by any rules
- **THEN** configuration is deleted successfully

#### Scenario: Delete config in use
- **WHEN** a user attempts to delete config used by active rules
- **THEN** system prevents deletion
- **AND** returns error indicating config is in use

### Requirement: Notification Testing

The system SHALL provide a test endpoint to send test notifications.

The test endpoint SHALL send a sample message to verify webhook connectivity.

The test endpoint SHALL return success/failure status.

#### Scenario: Test successful notification
- **WHEN** a user tests Lark config with valid webhook
- **THEN** test message is delivered to Lark/Feishu
- **AND** system returns success status

#### Scenario: Test notification failure
- **WHEN** a user tests Lark config with invalid webhook URL
- **THEN** system returns error status
- **AND** provides detailed error message

### Requirement: Lark Message Format

The system SHALL format alert messages as Lark Interactive Cards.

The message SHALL include:
- Alert title with emoji indicator
- Rule name
- Time range and alert count
- Index name
- Log summary (first 3 entries) in structured format

The system SHALL extract key fields based on log type automatically.

#### Scenario: Format nginx alert message
- **WHEN** alert contains nginx logs
- **THEN** message includes: @timestamp, request (URL path), cf_ray, response_code
- **AND** response_code is highlighted in red
- **AND** cf_ray shows "-" if missing

#### Scenario: Format java alert message
- **WHEN** alert contains java application logs
- **THEN** message includes: module, node_ip, message, @timestamp
- **AND** fields are displayed in card format

### Requirement: Notification Retry

The system SHALL retry failed notifications according to configured policy.

The system SHALL track retry attempts and final status.

The system SHALL update alert status based on notification result.

#### Scenario: Retry on failure
- **WHEN** initial notification fails
- **THEN** system retries up to configured retry times (default: 3)
- **AND** alert status is updated to "sent" if any retry succeeds
- **AND** alert status remains "failed" if all retries fail

#### Scenario: Successful retry
- **WHEN** first attempt fails but second attempt succeeds
- **THEN** alert status is updated to "sent"
- **AND** no error message is stored

### Requirement: Lark Config List

The system SHALL provide an endpoint to list all Lark configurations.

The list SHALL NOT include webhook URLs in response for security.

The list SHALL include usage count (number of rules using each config).

#### Scenario: List all configs
- **WHEN** a user requests Lark config list
- **THEN** system returns all configurations
- **AND** webhook URLs are not included in response
- **AND** each config shows usage count

### Requirement: Multiple Notification Channels

The system SHALL support configuring multiple Lark webhooks per rule.

Rules SHALL be able to use different Lark configurations.

The system SHALL send notifications to all configured channels when alert is generated.

#### Scenario: Rule with multiple channels
- **WHEN** a rule is configured with multiple Lark configs
- **THEN** notifications are sent to all configured webhooks
- **AND** each notification is tracked independently

### Requirement: Notification Rate Limiting

The system SHALL respect rate limits when sending notifications.

The system SHALL handle rate limit errors gracefully.

The system SHALL not exceed reasonable notification frequency per webhook.

#### Scenario: Handle rate limit
- **WHEN** notification exceeds webhook rate limit
- **THEN** system receives rate limit error
- **AND** retries after appropriate delay
- **AND** alert status is updated accordingly

