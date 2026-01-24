## ADDED Requirements

### Requirement: Graceful Shutdown

The backend service SHALL perform graceful shutdown on SIGTERM/SIGINT.

The backend service SHALL stop accepting new connections and allow in-flight requests to complete within a bounded timeout.

The backend service SHALL stop background workers (scheduler) during shutdown.

#### Scenario: SIGTERM during traffic
- **WHEN** the process receives SIGTERM while there are in-flight API requests
- **THEN** the service stops accepting new connections
- **AND** completes in-flight requests within the shutdown timeout
- **AND** exits with a clean shutdown result

#### Scenario: SIGTERM stops scheduler
- **WHEN** the process receives SIGTERM while the scheduler is running
- **THEN** the scheduler is stopped
- **AND** no new rule executions are started after shutdown begins

### Requirement: HTTP Server Timeouts

The backend service SHALL configure HTTP server timeouts to prevent resource exhaustion from slow connections.

Timeouts SHALL include at least read timeout, write timeout, and idle timeout.

#### Scenario: Slow client connection
- **WHEN** a client holds a connection open without completing a request
- **THEN** the server enforces read/idle timeouts
- **AND** the service remains responsive for other clients

