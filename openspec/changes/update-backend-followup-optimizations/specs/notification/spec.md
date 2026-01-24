## MODIFIED Requirements

### Requirement: Notification Retry Mechanism

The system SHALL retry failed notification deliveries.

The retry mechanism SHALL use exponential backoff with an upper bound to avoid retry storms.

The system SHALL enforce a request timeout for each notification attempt.

#### Scenario: Notification retry with backoff
- **WHEN** the notification delivery fails due to transient network error
- **THEN** the system retries with exponential backoff
- **AND** retries stop after configured attempts

