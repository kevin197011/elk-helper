# Data Source Management

## Purpose

Manage Elasticsearch data source configurations, supporting multiple ES clusters and nodes with load balancing.

## Requirements

### Requirement: ES Config Creation

The system SHALL allow users to create ES data source configurations.

The system SHALL require the following fields:
- Name (unique, required)
- ES URL(s) - supports single or multiple URLs separated by semicolons (required)
- Username (optional, required if ES security enabled)
- Password (optional, required if ES security enabled)
- Skip certificate verification flag (default: false)
- Certificate content (optional, for custom CA certificates)

The system SHALL support multiple ES nodes in a single configuration (semicolon-separated).

#### Scenario: Create single node config
- **WHEN** a user creates config with single ES URL: `https://es.example.com:9200`
- **THEN** configuration is saved
- **AND** queries use this single endpoint

#### Scenario: Create multi-node config
- **WHEN** a user creates config with multiple URLs: `https://es1:9200;https://es2:9200;https://es3:9200`
- **THEN** configuration is saved
- **AND** system uses round-robin load balancing across nodes

#### Scenario: Create config with authentication
- **WHEN** a user creates config with username and password
- **THEN** credentials are stored securely
- **AND** used for all ES queries

#### Scenario: Create config with duplicate name
- **WHEN** a user attempts to create config with existing name
- **THEN** system returns HTTP 409 Conflict
- **AND** error indicates name already exists

### Requirement: ES Config Update

The system SHALL allow users to update existing ES configurations.

The system SHALL validate new configuration values before saving.

Updated configurations SHALL be used immediately for new queries.

#### Scenario: Update ES URL
- **WHEN** a user updates ES URL in configuration
- **THEN** new URL is used for subsequent queries
- **AND** existing queries complete with old URL

#### Scenario: Update credentials
- **WHEN** a user updates username or password
- **THEN** new credentials are used for subsequent queries
- **AND** password is stored securely

### Requirement: ES Config Deletion

The system SHALL allow users to delete ES configurations.

The system SHALL prevent deletion if configuration is referenced by active rules.

#### Scenario: Delete unused config
- **WHEN** a user deletes an ES config not used by any rules
- **THEN** configuration is deleted successfully
- **AND** HTTP 204 No Content is returned

#### Scenario: Delete config in use
- **WHEN** a user attempts to delete config used by active rules
- **THEN** system prevents deletion
- **AND** returns error indicating config is in use

### Requirement: ES Connection Testing

The system SHALL provide a test endpoint to verify ES connectivity.

The test endpoint SHALL attempt connection and basic query.

The test endpoint SHALL return success/failure with detailed error message if failed.

#### Scenario: Test successful connection
- **WHEN** a user tests ES config with valid credentials and reachable endpoint
- **THEN** system returns success status
- **AND** confirms connectivity

#### Scenario: Test connection failure
- **WHEN** a user tests ES config with unreachable endpoint
- **THEN** system returns error status
- **AND** provides detailed error message (connection timeout, DNS resolution, etc.)

#### Scenario: Test authentication failure
- **WHEN** a user tests ES config with invalid credentials
- **THEN** system returns error status
- **AND** indicates authentication failure

### Requirement: Multi-Node Load Balancing

The system SHALL support round-robin load balancing across multiple ES nodes.

When multiple URLs are configured (semicolon-separated), the system SHALL rotate through nodes for each query.

The system SHALL handle node failures gracefully by trying next available node.

#### Scenario: Round-robin across nodes
- **WHEN** a config has 3 ES nodes configured
- **AND** multiple queries are executed
- **THEN** queries are distributed evenly across all 3 nodes
- **AND** each query uses a different node (round-robin)

#### Scenario: Node failure handling
- **WHEN** one ES node becomes unavailable
- **THEN** system retries with next available node
- **AND** query succeeds if at least one node is available

#### Scenario: All nodes unavailable
- **WHEN** all configured ES nodes are unavailable
- **THEN** query fails with appropriate error
- **AND** alert generation is skipped for that execution

### Requirement: SSL/TLS Certificate Handling

The system SHALL support custom CA certificates for ES connections.

The system SHALL provide option to skip certificate verification (for self-signed certs in development).

The system SHALL validate certificate format when provided.

#### Scenario: Use custom certificate
- **WHEN** a user provides custom CA certificate content
- **THEN** system uses certificate for TLS verification
- **AND** validates ES certificate against provided CA

#### Scenario: Skip certificate verification
- **WHEN** a user enables "skip certificate verification"
- **THEN** system accepts any certificate (including self-signed)
- **AND** connection proceeds without TLS verification

### Requirement: ES Config List

The system SHALL provide an endpoint to list all ES configurations.

The list SHALL NOT include sensitive fields (password) in response.

The list SHALL include configuration status (testable, in use count).

#### Scenario: List all configs
- **WHEN** a user requests ES config list
- **THEN** system returns all configurations
- **AND** passwords are not included in response
- **AND** each config shows usage count (number of rules using it)

### Requirement: Default Data Source

The system SHALL support marking an ES config as default.

Rules without explicit ES config SHALL use the default configuration.

The system SHALL allow changing the default data source.

#### Scenario: Set default config
- **WHEN** a user sets an ES config as default
- **THEN** that config becomes the default
- **AND** previous default is unset

#### Scenario: Use default in rule
- **WHEN** a rule is created without ES config specified
- **THEN** rule uses the default ES configuration
- **AND** queries are executed against default ES cluster

