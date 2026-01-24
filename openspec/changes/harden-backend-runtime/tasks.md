## 1. Implementation (P0 Guardrails)

- [x] 1.1 Implement HTTP graceful shutdown with `http.Server` (timeouts + `Shutdown(ctx)`)
- [x] 1.2 Enforce worker execution concurrency limit using `WORKER_MAX_CONCURRENCY`
- [x] 1.3 Tighten CORS to an explicit allowlist from `CORS_ORIGINS`
- [x] 1.4 Harden JWT secret handling (startup validation; fail fast on weak/default secret)
- [x] 1.5 Update `env.example` and README deployment notes for the new guardrails

## 2. Validation

- [x] 2.1 `docker compose up -d` starts cleanly and `/health` passes
- [x] 2.2 Login works; protected endpoints return 401 when missing token
- [x] 2.3 Create/enable rule triggers execution without goroutine explosion under load (basic sanity)
- [x] 2.4 SIGTERM behavior: container stop/restart drains without dropping in-flight requests (best-effort manual check)

## 3. Follow-ups (not in this change)

- 3.1 Add pagination to rules list endpoint (align with spec requirement)
- 3.2 Add context timeouts to DB queries in service layer
- 3.3 Add ES/Lark webhook retry backoff + (optional) circuit breaker
- 3.4 Reduce stored alert log payload size / move large payloads out of DB
- 3.5 Encrypt sensitive stored secrets (ES password, webhook secrets) at rest
- 3.6 Add API rate limiting (at least `/auth/login`)

