# Change: Harden Backend Runtime Guardrails

## Why

当前后端在容器化部署与高并发规则执行场景下，存在稳定性与安全性隐患（例如：HTTP 非优雅停机、规则执行并发不受控、CORS 过宽、JWT 密钥可能误用默认值）。这些问题会在滚动发布、规则数量增长或外部依赖抖动时放大，导致请求失败、资源耗尽或安全风险。

## What Changes

### In scope (P0)

- **Graceful shutdown**：后端在收到 SIGTERM/SIGINT 时，支持优雅停机并在超时内完成收尾（停止调度器、停止接收新连接、等待 in-flight 请求结束）。
- **HTTP server timeouts**：为 HTTP 服务补齐合理的 Read/Write/Idle 超时，降低慢连接/资源占用风险。
- **Worker concurrency guardrail**：规则执行增加全局并发上限（使用 `WORKER_MAX_CONCURRENCY`），避免在规则数量大时出现 goroutine/内存失控。
- **CORS allowlist**：CORS 中间件按 `CORS_ORIGINS` 白名单精确放行 `Origin`，避免 `*` 过宽策略。
- **JWT secret hardening**：启动时校验 `JWT_SECRET`（例如最小长度/禁止默认值），避免生产误用弱密钥。

### Out of scope (P1/P2, tracked as follow-ups)

- rules 列表分页、DB 查询 context timeout、ES/Webhook 熔断/退避、敏感配置加密存储、API rate limit 等（在本 change 中仅记录，不实施）。

## Impact

- Affected specs:
  - `specs/user-auth/spec.md`
  - `specs/rule-management/spec.md`
  - **NEW**: `specs/service-runtime/spec.md`（在本 change 的 delta 中新增能力）
  - **NEW**: `specs/api-security/spec.md`（在本 change 的 delta 中新增能力）

- Affected code (expected):
  - `backend/cmd/server/main.go`
  - `backend/internal/worker/scheduler/scheduler.go`
  - `backend/internal/api/middleware/cors.go`
  - `backend/internal/service/auth/auth.go`
  - `backend/internal/config/config.go`
  - `docker-compose*.yml`, `env.example`（如需补充/调整环境变量说明）

