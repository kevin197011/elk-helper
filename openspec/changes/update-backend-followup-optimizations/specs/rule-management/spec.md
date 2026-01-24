## MODIFIED Requirements

### Requirement: Rule List and Search

The system SHALL provide a list endpoint to retrieve all rules.

The system SHALL support pagination (page, page_size).

The system SHALL include rule statistics in the list response.

The system SHALL remain backward compatible when pagination parameters are not provided.

#### Scenario: List all rules
- **WHEN** a user requests the rule list without pagination parameters
- **THEN** the system returns all rules
- **AND** existing response shape remains compatible

#### Scenario: List rules with pagination
- **WHEN** a user requests page 2 with page_size 20
- **THEN** the system returns rules 21-40
- **AND** pagination metadata is included

