## ADDED Requirements

### Requirement: Encrypt Sensitive Configuration Values

The system SHALL support encrypting sensitive configuration values at rest (e.g., ES password, webhook URL).

Encryption-at-rest SHOULD be optional and controlled by deployment configuration.

The system SHALL remain backward compatible with existing plaintext values.

#### Scenario: Encryption enabled
- **WHEN** encryption is enabled via configuration
- **THEN** newly saved sensitive values are stored encrypted
- **AND** values are transparently decrypted when used by the backend

#### Scenario: Existing plaintext values
- **WHEN** encryption is enabled but existing values are stored in plaintext
- **THEN** the system continues to function using plaintext values

