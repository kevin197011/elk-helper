## ADDED Requirements

### Requirement: Login Rate Limiting

The system SHALL apply rate limiting to the login endpoint to mitigate brute-force attempts.

Rate limiting SHALL be applied per client identifier (e.g., IP address).

When the rate limit is exceeded, the system SHALL return HTTP 429 Too Many Requests.

#### Scenario: Exceed login rate limit
- **WHEN** a client repeatedly calls the login endpoint beyond the allowed rate
- **THEN** the system returns HTTP 429
- **AND** the response indicates rate limit exceeded

