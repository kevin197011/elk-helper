# Change: Update Backend Follow-up Optimizations

## Why

上一轮运行时护栏（优雅停机/并发限制/CORS/JWT）已落地，但仍有一批后端“可预见会在规模与安全上踩坑”的优化项尚未实现（分页、DB/ES 超时、通知重试策略、告警日志存储上限、敏感信息加密存储、登录限流）。

## What Changes

- **Rule list pagination**：对齐 `rule-management` spec，`/api/v1/rules` 支持分页与分页元信息（兼容旧响应）。
- **DB query timeout guardrail**：为关键 DB 操作增加超时（避免慢 SQL 拖死 API/worker）。
- **ES query timeout**：为 ES 查询增加 timeout（避免规则执行卡死占用并发 slot）。
- **Webhook retry/backoff improvements**：优化告警通知重试（退避/上限/更合理的超时控制）。
- **Alert log storage cap**：限制单条告警持久化 logs 的体积/条数，避免 DB 膨胀与响应体过大。
- **Encrypt secrets at rest (best-effort)**：在不做迁移的前提下，为 ES 密码、Webhook URL 等敏感字段提供可选加密存储（兼容旧明文数据）。
- **Login rate limiting**：为 `/api/v1/auth/login` 增加基于 IP 的限流，降低爆破/误操作风险。

## Impact

- Affected specs:
  - `specs/rule-management/spec.md` (pagination)
  - `specs/user-auth/spec.md` (login rate limiting)
  - `specs/notification/spec.md` (retry/backoff behavior refinement)
  - `specs/alerting/spec.md` (logs storage guardrail)
  - `specs/data-source/spec.md` (secret storage)

- Affected code (expected):
  - `backend/internal/api/handlers/rule_handler.go`
  - `backend/internal/service/rule/rule.go`
  - `backend/internal/repository/database/database.go`
  - `backend/internal/worker/executor/executor.go`
  - `backend/internal/worker/notifier/lark.go`
  - `backend/internal/service/{esconfig,larkconfig}/*`
  - `backend/internal/config/config.go`
  - `backend/internal/api/routes/routes.go` + middleware
  - `env.example`, `README.md`

