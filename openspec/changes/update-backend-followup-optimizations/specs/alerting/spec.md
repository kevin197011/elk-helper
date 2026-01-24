## ADDED Requirements

### Requirement: Alert Logs Storage Guardrail

The system SHALL cap persisted alert logs payload size to prevent database bloat.

The system SHALL preserve the original matched log count separately from stored samples.

#### Scenario: Large match set
- **WHEN** a rule execution matches a large number of logs
- **THEN** the alert record stores only a capped sample of logs
- **AND** the alert record log_count reflects the original match count

