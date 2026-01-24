## 1. Implementation

- [x] 1.1 Add pagination support for `/api/v1/rules` (page/page_size + pagination metadata; backward compatible)
- [x] 1.2 Add DB query timeout guardrail for key operations
- [x] 1.3 Add ES query timeout guardrail for rule executions and rule test endpoint
- [x] 1.4 Improve webhook retry/backoff behavior (cap/jitter + request timeout)
- [x] 1.5 Cap persisted alert logs payload (store samples; keep log_count accurate)
- [x] 1.6 Add optional encryption-at-rest for sensitive fields (backward compatible `enc:` format)
- [x] 1.7 Add rate limiting for `/api/v1/auth/login` (per-IP token bucket)
- [x] 1.8 Update `env.example` and `README.md` for new settings

## 2. Validation

- [x] 2.1 `go test ./...` passes
- [x] 2.2 `docker compose up -d --build` healthy and login works
- [x] 2.3 `/api/v1/rules` returns paginated response when `page/page_size` provided
- [x] 2.4 Alert record creation stores capped logs but retains original `log_count`
- [x] 2.5 Secrets encryption is backward compatible (plaintext still works; `enc:` values decrypt)
- [x] 2.6 Login endpoint rate limiting returns HTTP 429 when exceeded

