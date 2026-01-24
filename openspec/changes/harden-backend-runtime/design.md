## Context

本变更聚焦后端的“运行时护栏（guardrails）”：在容器滚动发布、高并发规则执行、浏览器访问与鉴权安全方面补齐底层保障，降低生产事故概率。

当前实现特征：
- HTTP 以 `gin.Engine.Run()` 启动，SIGTERM 时不会优雅 drain in-flight 请求。
- 规则执行存在多 goroutine 并发触发点，配置虽有 `WORKER_MAX_CONCURRENCY` 但缺少强制执行。
- CORS 策略需要可配置且默认不应过宽。
- JWT secret 存在默认值风险，需要启动期强校验。

## Goals / Non-Goals

### Goals
- 服务在 SIGTERM/SIGINT 时优雅停机，减少请求失败与数据不一致风险。
- worker 执行具备可配置并发上限，避免资源耗尽。
- CORS 按白名单精确放行，避免浏览器侧跨域策略过宽。
- JWT secret 启动期硬校验，避免弱密钥/默认值部署。

### Non-Goals
- 不在本变更中引入新的外部基础设施（如 Redis、消息队列）。
- 不在本变更中大规模重构 API 分层或引入复杂可观测性体系（仅在需要处增强日志字段）。
- 不在本变更中实现 P1/P2 项（分页、熔断、加密存储、rate limit 等），只记录为 follow-ups。

## Decisions

### Decision 1: 使用 `http.Server` 实现优雅停机
- **Why**：`gin.Engine.Run` 难以统一设置 timeout 与 shutdown；`http.Server` 支持 `Shutdown(ctx)`。
- **Details**：
  - 设置 `ReadTimeout`/`WriteTimeout`/`IdleTimeout`（保守值，避免慢连接拖垮）。
  - Shutdown 顺序：停止接收新连接 → 停止调度器 → 等待 in-flight 请求（或并行）→ 关闭 DB。

### Decision 2: Scheduler 执行使用 semaphore 实施 `MaxConcurrency`
- **Why**：改动小、易验证，符合“简单优先”。
- **Details**：
  - 新增 buffered channel 作为 semaphore，容量 = `WORKER_MAX_CONCURRENCY`。
  - 规则执行 goroutine 进入前先 acquire，执行完 release。
  - 如果 ctx 已取消，立即退出；如果 semaphore 满，按策略排队（阻塞等待）或快速失败（本 change 默认阻塞等待，避免丢执行）。

### Decision 3: CORS 使用 allowlist（精确 Origin 回显）
- **Why**：避免 `*` 造成浏览器侧安全风险；与 `CORS_ORIGINS` 配置一致。
- **Details**：
  - 仅当 `Origin` 在 allowlist 中才设置 `Access-Control-Allow-Origin` 为该 origin。
  - 允许 `Authorization` header 与常用 method；Preflight 正确返回。

### Decision 4: JWT secret 启动期强校验
- **Why**：防止默认值或弱密钥上线。
- **Details**：
  - `JWT_SECRET` 必须显式配置。
  - 推荐最小长度 32（或等效强度），否则启动失败。
  - 兼容开发模式：可通过 `GIN_MODE=debug` 放宽（但需明确记录并在日志提示）。

## Risks / Trade-offs

- **并发限制导致执行堆积**：当规则数量远大于并发上限，可能出现执行延迟；这是可接受的保护性退化，需要在 UI/监控侧可见（后续 work）。
- **更严格的 CORS/JWT 校验可能影响现有环境**：需要同步更新 `.env`/部署说明，避免升级后启动失败或前端跨域失败。

## Migration Plan

1. 实现 HTTP server + graceful shutdown + timeouts（最小可用）。
2. 加入 worker 并发限制（默认值不变，只强制生效）。
3. 收紧 CORS（优先兼容现有 `CORS_ORIGINS`）。
4. 加强 JWT secret 校验（同步更新 `env.example`）。
5. 在 docker-compose 环境下验证启动、登录、规则执行与停止流程。

