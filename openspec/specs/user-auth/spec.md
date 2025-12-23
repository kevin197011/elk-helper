# User Authentication and Authorization

## Purpose

提供用户身份认证和授权功能，保护系统资源不被未授权访问。

## Requirements

### Requirement: User Login

The system SHALL provide a login endpoint that accepts username and password credentials.

The system SHALL validate credentials against the database.

The system SHALL return a JWT token upon successful authentication.

The system SHALL return an error message upon failed authentication.

#### Scenario: Successful login
- **WHEN** a user provides valid username and password
- **THEN** the system returns a JWT token
- **AND** the token expires after 24 hours

#### Scenario: Invalid credentials
- **WHEN** a user provides invalid username or password
- **THEN** the system returns HTTP 401 Unauthorized
- **AND** an error message is displayed

#### Scenario: Non-existent user
- **WHEN** a user provides a username that does not exist
- **THEN** the system returns HTTP 401 Unauthorized
- **AND** does not reveal whether the username exists

### Requirement: JWT Token Authentication

The system SHALL require JWT token in the Authorization header for all protected endpoints.

The system SHALL validate token signature and expiration.

The system SHALL reject requests with invalid or expired tokens.

#### Scenario: Valid token
- **WHEN** a request includes a valid JWT token in Authorization header
- **THEN** the request is processed normally

#### Scenario: Missing token
- **WHEN** a request does not include a JWT token
- **THEN** the system returns HTTP 401 Unauthorized

#### Scenario: Expired token
- **WHEN** a request includes an expired JWT token
- **THEN** the system returns HTTP 401 Unauthorized
- **AND** the user is redirected to login page

#### Scenario: Invalid token signature
- **WHEN** a request includes a token with invalid signature
- **THEN** the system returns HTTP 401 Unauthorized

### Requirement: Password Management

The system SHALL store passwords using bcrypt hashing with cost factor 10.

The system SHALL provide password change functionality for authenticated users.

The system SHALL validate password strength before accepting changes.

#### Scenario: Password change
- **WHEN** an authenticated user provides current password and new password
- **AND** current password is correct
- **AND** new password meets strength requirements
- **THEN** the password is updated
- **AND** the user must login again with new password

#### Scenario: Invalid current password
- **WHEN** a user attempts to change password with incorrect current password
- **THEN** the system returns an error message
- **AND** the password is not changed

### Requirement: Role-Based Access Control

The system SHALL support two roles: `admin` and `user`.

The system SHALL enforce role-based permissions on sensitive operations.

Admin users SHALL have full access to all features.

User role SHALL have limited access (implementation dependent).

#### Scenario: Admin access
- **WHEN** an admin user accesses a protected resource
- **THEN** access is granted

#### Scenario: User access
- **WHEN** a regular user accesses an admin-only resource
- **THEN** access is denied with HTTP 403 Forbidden

### Requirement: User Session Management

The system SHALL maintain user session state through JWT tokens.

The system SHALL not maintain server-side session storage.

The system SHALL allow multiple concurrent sessions for the same user.

#### Scenario: Multiple devices
- **WHEN** a user logs in from multiple devices
- **THEN** all devices receive valid tokens
- **AND** all sessions remain active independently

### Requirement: Default Admin Account

The system SHALL create a default admin account on first startup.

The default admin credentials SHALL be configurable via environment variables.

The system SHALL allow changing default admin password after first login.

#### Scenario: Initial setup
- **WHEN** the system starts for the first time
- **THEN** a default admin account is created
- **AND** credentials are read from environment variables (ADMIN_USERNAME, ADMIN_PASSWORD)

#### Scenario: Default credentials
- **WHEN** environment variables are not set
- **THEN** default credentials are used: username=`admin`, password=`admin123`

