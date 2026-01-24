## ADDED Requirements

### Requirement: JWT Secret Management

The system SHALL require an explicit JWT signing secret configuration for issuing and validating tokens.

The system SHALL validate the JWT signing secret at startup.

The system SHALL reject weak or default JWT secrets in production deployments.

#### Scenario: Startup with strong secret
- **WHEN** the service starts with a configured JWT secret meeting minimum strength requirements
- **THEN** the service starts successfully
- **AND** issued tokens can be validated by the service

#### Scenario: Startup with missing or weak secret
- **WHEN** the service starts without a valid JWT secret (missing/weak/default)
- **THEN** the service fails fast with a clear error message

