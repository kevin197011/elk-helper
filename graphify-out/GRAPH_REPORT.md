# Graph Report - kk-elk-helper  (2026-06-01)

## Corpus Check
- 122 files · ~64,497 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1139 nodes · 1368 edges · 103 communities (83 shown, 20 thin omitted)
- Extraction: 93% EXTRACTED · 7% INFERRED · 0% AMBIGUOUS · INFERRED: 91 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `c68bf41a`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- [[_COMMUNITY_Community 0|Community 0]]
- [[_COMMUNITY_Community 1|Community 1]]
- [[_COMMUNITY_Community 2|Community 2]]
- [[_COMMUNITY_Community 3|Community 3]]
- [[_COMMUNITY_Community 4|Community 4]]
- [[_COMMUNITY_Community 5|Community 5]]
- [[_COMMUNITY_Community 6|Community 6]]
- [[_COMMUNITY_Community 7|Community 7]]
- [[_COMMUNITY_Community 8|Community 8]]
- [[_COMMUNITY_Community 9|Community 9]]
- [[_COMMUNITY_Community 10|Community 10]]
- [[_COMMUNITY_Community 11|Community 11]]
- [[_COMMUNITY_Community 12|Community 12]]
- [[_COMMUNITY_Community 13|Community 13]]
- [[_COMMUNITY_Community 14|Community 14]]
- [[_COMMUNITY_Community 15|Community 15]]
- [[_COMMUNITY_Community 16|Community 16]]
- [[_COMMUNITY_Community 17|Community 17]]
- [[_COMMUNITY_Community 18|Community 18]]
- [[_COMMUNITY_Community 19|Community 19]]
- [[_COMMUNITY_Community 20|Community 20]]
- [[_COMMUNITY_Community 21|Community 21]]
- [[_COMMUNITY_Community 22|Community 22]]
- [[_COMMUNITY_Community 23|Community 23]]
- [[_COMMUNITY_Community 24|Community 24]]
- [[_COMMUNITY_Community 25|Community 25]]
- [[_COMMUNITY_Community 26|Community 26]]
- [[_COMMUNITY_Community 27|Community 27]]
- [[_COMMUNITY_Community 28|Community 28]]
- [[_COMMUNITY_Community 29|Community 29]]
- [[_COMMUNITY_Community 30|Community 30]]
- [[_COMMUNITY_Community 31|Community 31]]
- [[_COMMUNITY_Community 32|Community 32]]
- [[_COMMUNITY_Community 33|Community 33]]
- [[_COMMUNITY_Community 34|Community 34]]
- [[_COMMUNITY_Community 35|Community 35]]
- [[_COMMUNITY_Community 36|Community 36]]
- [[_COMMUNITY_Community 37|Community 37]]
- [[_COMMUNITY_Community 38|Community 38]]
- [[_COMMUNITY_Community 39|Community 39]]
- [[_COMMUNITY_Community 40|Community 40]]
- [[_COMMUNITY_Community 41|Community 41]]
- [[_COMMUNITY_Community 42|Community 42]]
- [[_COMMUNITY_Community 43|Community 43]]
- [[_COMMUNITY_Community 44|Community 44]]
- [[_COMMUNITY_Community 45|Community 45]]
- [[_COMMUNITY_Community 46|Community 46]]
- [[_COMMUNITY_Community 47|Community 47]]
- [[_COMMUNITY_Community 48|Community 48]]
- [[_COMMUNITY_Community 49|Community 49]]
- [[_COMMUNITY_Community 50|Community 50]]
- [[_COMMUNITY_Community 51|Community 51]]
- [[_COMMUNITY_Community 52|Community 52]]
- [[_COMMUNITY_Community 53|Community 53]]
- [[_COMMUNITY_Community 54|Community 54]]
- [[_COMMUNITY_Community 55|Community 55]]
- [[_COMMUNITY_Community 56|Community 56]]
- [[_COMMUNITY_Community 57|Community 57]]
- [[_COMMUNITY_Community 58|Community 58]]
- [[_COMMUNITY_Community 59|Community 59]]
- [[_COMMUNITY_Community 60|Community 60]]
- [[_COMMUNITY_Community 61|Community 61]]
- [[_COMMUNITY_Community 62|Community 62]]
- [[_COMMUNITY_Community 63|Community 63]]
- [[_COMMUNITY_Community 64|Community 64]]
- [[_COMMUNITY_Community 65|Community 65]]
- [[_COMMUNITY_Community 66|Community 66]]
- [[_COMMUNITY_Community 67|Community 67]]
- [[_COMMUNITY_Community 68|Community 68]]
- [[_COMMUNITY_Community 69|Community 69]]
- [[_COMMUNITY_Community 70|Community 70]]
- [[_COMMUNITY_Community 71|Community 71]]
- [[_COMMUNITY_Community 72|Community 72]]
- [[_COMMUNITY_Community 73|Community 73]]
- [[_COMMUNITY_Community 74|Community 74]]
- [[_COMMUNITY_Community 75|Community 75]]
- [[_COMMUNITY_Community 76|Community 76]]
- [[_COMMUNITY_Community 77|Community 77]]
- [[_COMMUNITY_Community 78|Community 78]]
- [[_COMMUNITY_Community 79|Community 79]]
- [[_COMMUNITY_Community 80|Community 80]]
- [[_COMMUNITY_Community 81|Community 81]]
- [[_COMMUNITY_Community 82|Community 82]]
- [[_COMMUNITY_Community 83|Community 83]]
- [[_COMMUNITY_Community 84|Community 84]]
- [[_COMMUNITY_Community 85|Community 85]]
- [[_COMMUNITY_Community 86|Community 86]]
- [[_COMMUNITY_Community 87|Community 87]]
- [[_COMMUNITY_Community 88|Community 88]]
- [[_COMMUNITY_Community 89|Community 89]]
- [[_COMMUNITY_Community 90|Community 90]]
- [[_COMMUNITY_Community 91|Community 91]]
- [[_COMMUNITY_Community 92|Community 92]]
- [[_COMMUNITY_Community 93|Community 93]]
- [[_COMMUNITY_Community 100|Community 100]]
- [[_COMMUNITY_Community 101|Community 101]]
- [[_COMMUNITY_Community 102|Community 102]]

## God Nodes (most connected - your core abstractions)
1. `WithTimeout()` - 42 edges
2. `ELK Helper - 智能日志告警系统` - 21 edges
3. `compilerOptions` - 19 edges
4. `SetupRoutes()` - 17 edges
5. `Service` - 16 edges
6. `ESTestLogGenerator` - 16 edges
7. `Service` - 15 edges
8. `OpenSpec Instructions` - 15 edges
9. `Scheduler` - 14 edges
10. `useAuth()` - 13 edges

## Surprising Connections (you probably didn't know these)
- `SetupRoutes()` --calls--> `NewIPRateLimiter()`  [INFERRED]
  backend/internal/api/routes/routes.go → backend/internal/api/middleware/rate_limit.go
- `SetupRoutes()` --calls--> `NewAuthHandler()`  [INFERRED]
  backend/internal/api/routes/routes.go → backend/internal/api/handlers/auth_handler.go
- `SetupRoutes()` --calls--> `NewSSOHandler()`  [INFERRED]
  backend/internal/api/routes/routes.go → backend/internal/api/handlers/sso_handler.go
- `SetupRoutes()` --calls--> `NewRuleHandler()`  [INFERRED]
  backend/internal/api/routes/routes.go → backend/internal/api/handlers/rule_handler.go
- `SetupRoutes()` --calls--> `NewAlertHandler()`  [INFERRED]
  backend/internal/api/routes/routes.go → backend/internal/api/handlers/alert_handler.go

## Communities (103 total, 20 thin omitted)

### Community 0 - "Community 0"
Cohesion: 0.04
Nodes (45): Alert (告警), Analysis First, Implementation Second, Backend, code:bash (docker compose up -d), code:bash (docker compose -f docker-compose-prod.yml up -d), code:bash (docker exec elk-helper-postgres pg_dump -U postgres elk_help), Deployment, Development Workflow (+37 more)

### Community 1 - "Community 1"
Cohesion: 0.04
Nodes (43): MODIFIED Requirements, Purpose, Requirement: Quick Template Buttons, Requirement: Rule Cloning, Requirement: Rule Creation, Requirement: Rule Deletion, Requirement: Rule Enable/Disable Toggle, Requirement: Rule List and Search (+35 more)

### Community 2 - "Community 2"
Cohesion: 0.06
Nodes (12): calculateBucketInterval(), RuleAlertStats, RuleTimeSeriesStats, Service, TimeSeriesDataPoint, WithTimeout(), Service, Service (+4 more)

### Community 3 - "Community 3"
Cohesion: 0.10
Nodes (18): Executor, computeESWindowStart(), hasSuccessfulRun(), ruleIntervalDuration(), shouldRunNow(), TestComputeESWindowStart_neverSucceededUsesFullInterval(), TestHasSuccessfulRun_andESWindowAfterSuccess(), TestShouldRunNow_afterSuccessRespectsInterval() (+10 more)

### Community 4 - "Community 4"
Cohesion: 0.06
Nodes (31): Alerting, Purpose, Requirement: Alert Cleanup, Requirement: Alert Deletion, Requirement: Alert Generation, Requirement: Alert History Retrieval, Requirement: Alert History Storage, Requirement: Alert Message Formatting (+23 more)

### Community 5 - "Community 5"
Cohesion: 0.06
Nodes (30): Data Source Management, Purpose, Requirement: Default Data Source, Requirement: ES Config Creation, Requirement: ES Config Deletion, Requirement: ES Config List, Requirement: ES Config Update, Requirement: ES Connection Testing (+22 more)

### Community 6 - "Community 6"
Cohesion: 0.07
Nodes (29): dependencies, @ant-design/icons, antd, axios, dayjs, react, react-dom, react-router-dom (+21 more)

### Community 7 - "Community 7"
Cohesion: 0.07
Nodes (26): Dashboard, MODIFIED Requirements, Purpose, Requirement: Chart Data Aggregation, Requirement: Chart Visualization, Requirement: Data Loading and Caching, Requirement: Responsive Design, Requirement: Rule Alert Time Series Chart (+18 more)

### Community 8 - "Community 8"
Cohesion: 0.07
Nodes (27): Notification Management, Purpose, Requirement: Lark Config Creation, Requirement: Lark Config Deletion, Requirement: Lark Config List, Requirement: Lark Config Update, Requirement: Lark Message Format, Requirement: Multiple Notification Channels (+19 more)

### Community 9 - "Community 9"
Cohesion: 0.08
Nodes (24): ADDED Requirements, MODIFIED Requirements, Purpose, Requirement: Cleanup Task Configuration, Requirement: Cleanup Task Execution, Requirement: Cleanup Task Execution Status Tracking, Requirement: Configuration Persistence, Requirement: System Status Monitoring (+16 more)

### Community 10 - "Community 10"
Cohesion: 0.22
Nodes (3): NewExecutor(), Scheduler, NewScheduler()

### Community 11 - "Community 11"
Cohesion: 0.11
Nodes (6): NewESConfigHandler(), ESConfigHandler, escapeWildcardLiteral(), NewServiceFromConfig(), parseESAddresses(), Service

### Community 12 - "Community 12"
Cohesion: 0.08
Nodes (23): Purpose, Requirement: Default Admin Account, Requirement: JWT Token Authentication, Requirement: Password Management, Requirement: Role-Based Access Control, Requirement: User Login, Requirement: User Session Management, Requirements (+15 more)

### Community 13 - "Community 13"
Cohesion: 0.09
Nodes (22): compilerOptions, allowSyntheticDefaultImports, baseUrl, esModuleInterop, isolatedModules, jsx, lib, module (+14 more)

### Community 14 - "Community 14"
Cohesion: 0.11
Nodes (16): ChangePasswordDialogProps, api, authApi, CleanupConfig, CreateUserPayload, getRequestBearerToken(), PaginatedResponse, QueryCondition (+8 more)

### Community 15 - "Community 15"
Cohesion: 0.14
Nodes (16): AdminRoute(), AdminRouteProps, Layout(), LayoutProps, ProtectedRoute(), ProtectedRouteProps, AuthContext, AuthContextType (+8 more)

### Community 16 - "Community 16"
Cohesion: 0.09
Nodes (20): NewSSOAdminHandler(), parseAdminSSOID(), SSOAdminHandler, SSOProviderRequest, BM25, detect_domain(), _load_csv(), Lowercase, split, remove punctuation, filter short words (+12 more)

### Community 17 - "Community 17"
Cohesion: 0.11
Nodes (19): 1. 多 ES 节点支持, 2. 智能告警格式, 3. 查询条件配置, 4. 规则实时更新, 5. 规则导入导出, 6. 清理任务配置, 7. 性能优化, code:json ([) (+11 more)

### Community 18 - "Community 18"
Cohesion: 0.11
Nodes (6): getEnv(), Claims, Service, isLocalAuthSource(), randomPassword(), UpdateUserInput

### Community 19 - "Community 19"
Cohesion: 0.14
Nodes (7): ImportRulesRequest, ImportRulesResponse, compareOptionalUint(), compareQueryConditions(), NewRuleHandler(), RuleHandler, GetGlobalScheduler()

### Community 20 - "Community 20"
Cohesion: 0.11
Nodes (18): Architecture Patterns, Code Style, code:block1 (cmd/server/), code:block2 (src/), code:block3 (<type>(<scope>): <subject>), code:block4 (feat(rule): add rule cloning feature), Commit 规范, Git Workflow (+10 more)

### Community 21 - "Community 21"
Cohesion: 0.10
Nodes (13): NewSSOHandler(), oidcStateCookieName(), parseSSOProviderID(), SSOHandler, AdminProviderItem, OIDCConfig, oidcRuntime, ProviderInfo (+5 more)

### Community 23 - "Community 23"
Cohesion: 0.13
Nodes (20): AuthConfig, Config, getEnv(), getEnvSlice(), Load(), parseBoolWithDefault(), parseIntWithDefault(), resolveSSOFrontendBaseURL() (+12 more)

### Community 24 - "Community 24"
Cohesion: 0.16
Nodes (3): RuleEditDialogProps, Rule, rulesApi

### Community 25 - "Community 25"
Cohesion: 0.27
Nodes (8): createUserRequest, resetPasswordRequest, updateUserRequest, mapUserServiceError(), NewUserHandler(), parseUintParam(), sanitizeUser(), UserHandler

### Community 26 - "Community 26"
Cohesion: 0.14
Nodes (14): 1. 配置 ES 数据源, 2. 创建告警规则, 3. 配置告警通知（Webhook）, 4. 查看告警历史, code:block15 (ES 地址: https://es.example.com:9200), code:block16 (ES 地址: https://es-node1.example.com:9200;https://es-node2.ex), code:block17 (规则名称: Nginx 5xx 错误告警), code:block18 (规则名称: Java ERROR 级别日志) (+6 more)

### Community 27 - "Community 27"
Cohesion: 0.18
Nodes (4): PageHeaderProps, Alert, alertsApi, statusApi

### Community 28 - "Community 28"
Cohesion: 0.17
Nodes (11): Context, Decision 1: 使用 `http.Server` 实现优雅停机, Decision 2: Scheduler 执行使用 semaphore 实施 `MaxConcurrency`, Decision 3: CORS 使用 allowlist（精确 Origin 回显）, Decision 4: JWT secret 启动期强校验, Decisions, Goals, Goals / Non-Goals (+3 more)

### Community 29 - "Community 29"
Cohesion: 0.20
Nodes (9): code:block1 (┌─────────────────────────────────────────────┐), code:block38 (feat: 新功能), Commit 规范, ELK Helper - 智能日志告警系统, 📚 参考文档, 🏗️ 技术架构, 👥 支持, 📄 许可证 (+1 more)

### Community 31 - "Community 31"
Cohesion: 0.20
Nodes (5): NewAuthHandler(), AuthHandler, LoginRequest, LoginResponse, UpdatePasswordRequest

### Community 33 - "Community 33"
Cohesion: 0.20
Nodes (9): Before Any Task, code:bash (# 1) Explore current state), code:block2 (openspec/), Directory Structure, Happy Path Script, OpenSpec Instructions, Search Guidance, TL;DR Quick Checklist (+1 more)

### Community 34 - "Community 34"
Cohesion: 0.22
Nodes (9): code:bash (# 查看所有服务), code:bash (# 1. 拉取最新镜像), code:bash (# 手动备份), code:bash (# 创建 devops 管理员账户), 备份数据（重要！）, 手动创建用户, 更新服务, 查看服务状态 (+1 more)

### Community 35 - "Community 35"
Cohesion: 0.22
Nodes (9): 1. Web 服务器监控, 2. 应用错误监控, 3. 慢请求监控, code:block31 (规则名称: Nginx 5xx 错误告警), code:block32 (规则名称: 应用 ERROR 日志), code:block33 (规则名称: API 响应超时), 推荐的告警规则, 📊 监控与告警 (+1 more)

### Community 36 - "Community 36"
Cohesion: 0.22
Nodes (9): 告警推送, 告警规则, 安全与权限, 性能优化, 数据库管理, 数据源管理, 架构与部署, 🎯 核心特性 (+1 more)

### Community 37 - "Community 37"
Cohesion: 0.25
Nodes (7): compilerOptions, allowSyntheticDefaultImports, composite, module, moduleResolution, skipLibCheck, include

### Community 38 - "Community 38"
Cohesion: 0.25
Nodes (8): code:bash (# 1. JWT 密钥（至少 32 字符）), code:bash (# 获取免费 SSL 证书), code:bash (cd ssl), Let's Encrypt（推荐）, SSL/TLS 配置, 🔐 安全配置, 生产环境必改项, 自签名证书（测试环境）

### Community 39 - "Community 39"
Cohesion: 0.25
Nodes (8): code:block3 (New request?), code:markdown (# Change: [Brief description of change]), code:markdown (## ADDED Requirements), code:markdown (## 1. Implementation), code:markdown (## Context), Creating Change Proposals, Decision Tree, Proposal Structure

### Community 40 - "Community 40"
Cohesion: 0.25
Nodes (8): code:markdown (## RENAMED Requirements), code:markdown (#### Scenario: User login success), code:markdown (- **Scenario: User login**  ❌), Critical: Scenario Formatting, Delta Operations, Requirement Wording, Spec File Format, When to use ADDED vs MODIFIED

### Community 41 - "Community 41"
Cohesion: 0.25
Nodes (6): ADDED Requirements, Requirement: JWT Secret Management, Requirement: Login Rate Limiting, Scenario: Exceed login rate limit, Scenario: Startup with missing or weak secret, Scenario: Startup with strong secret

### Community 43 - "Community 43"
Cohesion: 0.29
Nodes (3): ErrorBoundary, Props, State

### Community 44 - "Community 44"
Cohesion: 0.38
Nodes (3): ESConfigEditDialogProps, ESConfig, esConfigApi

### Community 45 - "Community 45"
Cohesion: 0.38
Nodes (3): UserEditDialogProps, User, usersApi

### Community 46 - "Community 46"
Cohesion: 0.38
Nodes (3): LarkConfigEditDialogProps, LarkConfig, larkConfigApi

### Community 47 - "Community 47"
Cohesion: 0.29
Nodes (6): Change: Harden Backend Runtime Guardrails, Impact, In scope (P0), Out of scope (P1/P2, tracked as follow-ups), What Changes, Why

### Community 48 - "Community 48"
Cohesion: 0.32
Nodes (8): code:bash (# 查看迁移状态), code:bash (# 执行索引优化脚本), code:block23 (启用清理: ✓), 定期清理数据, ⚡ 性能优化, ⚡ 性能优化, 数据库索引优化, 数据库迁移管理

### Community 49 - "Community 49"
Cohesion: 0.33
Nodes (4): AuthMiddleware(), RequireAdmin(), CORSMiddleware(), SetupRoutes()

### Community 50 - "Community 50"
Cohesion: 0.38
Nodes (4): ipLimiter, IPRateLimiter, NewIPRateLimiter(), RateLimitMiddleware()

### Community 51 - "Community 51"
Cohesion: 0.29
Nodes (3): QueryCondition, QueryConditions, Rule

### Community 52 - "Community 52"
Cohesion: 0.29
Nodes (6): ADDED Requirements, Requirement: Graceful Shutdown, Requirement: HTTP Server Timeouts, Scenario: SIGTERM during traffic, Scenario: SIGTERM stops scheduler, Scenario: Slow client connection

### Community 54 - "Community 54"
Cohesion: 0.33
Nodes (6): 502 Bad Gateway, code:bash (# 1. 确保代码最新), code:bash (# 执行性能优化), 告警未触发, 🐛 故障排查, 页面加载慢

### Community 56 - "Community 56"
Cohesion: 0.33
Nodes (6): Best Practices, Capability Naming, Change ID Naming, Clear References, Complexity Triggers, Simplicity First

### Community 57 - "Community 57"
Cohesion: 0.40
Nodes (4): Change: Add Cleanup Task Execution Status, Impact, What Changes, Why

### Community 58 - "Community 58"
Cohesion: 0.40
Nodes (4): Change: Add Cleanup Task Execution Status, Impact, What Changes, Why

### Community 59 - "Community 59"
Cohesion: 0.40
Nodes (4): ADDED Requirements, Requirement: CORS Origin Allowlist, Scenario: Allowed origin, Scenario: Disallowed origin

### Community 60 - "Community 60"
Cohesion: 0.40
Nodes (4): Change: Change Alert Count to Execution Count, Impact, What Changes, Why

### Community 61 - "Community 61"
Cohesion: 0.40
Nodes (4): ADDED Requirements, Requirement: Encrypt Sensitive Configuration Values, Scenario: Encryption enabled, Scenario: Existing plaintext values

### Community 62 - "Community 62"
Cohesion: 0.40
Nodes (4): Change: Ensure Chart Shows Current Time Minus 24 Hours, Impact, What Changes, Why

### Community 63 - "Community 63"
Cohesion: 0.40
Nodes (4): Change: Fix Chart Time Range Display to Show Current Time Minus 24 Hours, Impact, What Changes, Why

### Community 64 - "Community 64"
Cohesion: 0.40
Nodes (5): code:yaml (services:), 📦 Docker 镜像, GitHub Container Registry, 构建优化, 镜像标签

### Community 65 - "Community 65"
Cohesion: 0.40
Nodes (5): code:bash (# 1. 克隆项目), code:bash (# 1. 克隆项目), 🚀 快速开始, 方式 1: 使用预构建镜像（推荐生产环境）, 方式 2: 本地构建（开发环境）

### Community 66 - "Community 66"
Cohesion: 0.40
Nodes (5): code:bash (# 使用默认配置（https://localhost:9200, elastic/changeme）), code:bash (# 生成密码哈希), 密码管理, 🛠️ 开发工具, 测试日志生成

### Community 67 - "Community 67"
Cohesion: 0.40
Nodes (5): code:bash (# 后端端口（默认 8080）), code:bash (# 1. 复制配置文件), Nginx 反向代理配置, 基本配置（.env 文件）, 📋 环境配置

### Community 68 - "Community 68"
Cohesion: 0.40
Nodes (5): code:yaml (# docker-compose-prod.yml), code:nginx (# nginx/simple.conf), Nginx 反向代理, 生产环境部署, 🔧 部署配置

### Community 70 - "Community 70"
Cohesion: 0.40
Nodes (5): CLI Essentials, code:bash (openspec list              # What's in progress?), File Purposes, Quick Reference, Stage Indicators

### Community 71 - "Community 71"
Cohesion: 0.40
Nodes (4): Change: Update Backend Follow-up Optimizations, Impact, What Changes, Why

### Community 72 - "Community 72"
Cohesion: 0.50
Nodes (3): ADDED Requirements, Requirement: Alert Logs Storage Guardrail, Scenario: Large match set

### Community 74 - "Community 74"
Cohesion: 0.50
Nodes (4): v1.0.0 (初始版本), v1.1.0 (2025-12-03), v1.2.0 (2025-12-23), 🔄 更新日志

### Community 75 - "Community 75"
Cohesion: 0.50
Nodes (3): MODIFIED Requirements, Requirement: Notification Retry Mechanism, Scenario: Notification retry with backoff

### Community 76 - "Community 76"
Cohesion: 0.50
Nodes (4): CLI Commands, code:bash (# Essential commands), Command Flags, Quick Start

### Community 77 - "Community 77"
Cohesion: 0.50
Nodes (4): code:bash (# Always use strict mode for comprehensive checks), Common Errors, Troubleshooting, Validation Tips

### Community 78 - "Community 78"
Cohesion: 0.50
Nodes (4): Change Conflicts, Error Recovery, Missing Context, Validation Failures

### Community 79 - "Community 79"
Cohesion: 0.50
Nodes (4): code:block13 (openspec/changes/add-2fa-notify/), code:markdown (## ADDED Requirements), code:markdown (## ADDED Requirements), Multi-Capability Example

### Community 80 - "Community 80"
Cohesion: 0.50
Nodes (4): Stage 1: Creating Changes, Stage 2: Implementing Changes, Stage 3: Archiving Changes, Three-Stage Workflow

### Community 81 - "Community 81"
Cohesion: 0.50
Nodes (3): 1. Implementation (P0 Guardrails), 2. Validation, 3. Follow-ups (not in this change)

### Community 82 - "Community 82"
Cohesion: 0.50
Nodes (3): ADDED Requirements, Requirement: Rule Execution Concurrency Limit, Scenario: Concurrency under load

### Community 100 - "Community 100"
Cohesion: 0.25
Nodes (4): DEFAULT_OIDC_CONFIG, FormValues, SSOAdminProvider, SSOProviderConfig

### Community 102 - "Community 102"
Cohesion: 0.33
Nodes (5): Admin setup, API summary, Environment, Flow, OIDC SSO

## Knowledge Gaps
- **438 isolated node(s):** `composite`, `skipLibCheck`, `module`, `moduleResolution`, `allowSyntheticDefaultImports` (+433 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **20 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `SetupRoutes()` connect `Community 49` to `Community 32`, `Community 69`, `Community 11`, `Community 16`, `Community 50`, `Community 19`, `Community 21`, `Community 55`, `Community 23`, `Community 25`, `Community 30`, `Community 31`?**
  _High betweenness centrality (0.062) - this node is a cross-community bridge._
- **Why does `WithTimeout()` connect `Community 2` to `Community 32`, `Community 3`, `Community 11`, `Community 23`?**
  _High betweenness centrality (0.057) - this node is a cross-community bridge._
- **Why does `main()` connect `Community 23` to `Community 49`, `Community 10`, `Community 2`?**
  _High betweenness centrality (0.051) - this node is a cross-community bridge._
- **Are the 41 inferred relationships involving `WithTimeout()` (e.g. with `main()` and `.TestLarkConfig()`) actually correct?**
  _`WithTimeout()` has 41 INFERRED edges - model-reasoned connections that need verification._
- **Are the 16 inferred relationships involving `SetupRoutes()` (e.g. with `main()` and `CORSMiddleware()`) actually correct?**
  _`SetupRoutes()` has 16 INFERRED edges - model-reasoned connections that need verification._
- **What connects `composite`, `skipLibCheck`, `module` to the rest of the system?**
  _448 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Community 0` be split into smaller, more focused modules?**
  _Cohesion score 0.043478260869565216 - nodes in this community are weakly interconnected._