## ADDED Requirements

### Requirement: CORS Origin Allowlist

The backend service SHALL restrict cross-origin access using an explicit allowlist.

Allowed origins SHALL be configurable via environment (e.g., `CORS_ORIGINS`).

The service SHALL only return `Access-Control-Allow-Origin` for requests whose `Origin` is present in the allowlist.

#### Scenario: Allowed origin
- **WHEN** a browser request includes an `Origin` present in the allowlist
- **THEN** the response includes `Access-Control-Allow-Origin` matching the request origin
- **AND** the request is processed normally

#### Scenario: Disallowed origin
- **WHEN** a browser request includes an `Origin` not present in the allowlist
- **THEN** the response does not include `Access-Control-Allow-Origin`
- **AND** the browser blocks cross-origin access

