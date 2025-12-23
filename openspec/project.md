# Project Context

## Purpose

ELK Helper 是一个现代化的、高性能的日志监控告警系统，旨在帮助 DevOps 团队快速配置、管理和监控 Elasticsearch 日志数据源，实现智能化的告警通知。

**核心价值**：
- **零代码配置**：通过可视化界面配置告警规则，无需编写代码
- **高性能查询**：支持多 ES 节点负载均衡，优化查询性能
- **智能告警**：根据日志类型自动提取关键字段，发送简洁明了的告警消息
- **实时生效**：规则修改后自动生效，无需重启服务

## Tech Stack

### Backend
- **Go 1.23+**: 核心语言
- **Gin**: HTTP Web 框架
- **GORM**: ORM 数据库操作
- **PostgreSQL**: 关系型数据库
- **Elasticsearch Go Client v8**: ES 查询客户端
- **JWT (golang-jwt/jwt/v5)**: 身份认证

### Frontend
- **React 18**: UI 框架
- **TypeScript**: 类型安全
- **Vite**: 构建工具
- **TanStack Query**: 数据获取和缓存
- **React Router**: 路由管理
- **shadcn/ui**: UI 组件库
- **Tailwind CSS**: 样式框架
- **recharts**: 图表库

### Infrastructure
- **Docker & Docker Compose**: 容器化部署
- **Nginx**: 反向代理和静态文件服务
- **GitHub Actions**: CI/CD 自动化

## Development Workflow

### Analysis First, Implementation Second

**CRITICAL: Always analyze requirements before implementing code.**

When receiving a new feature request or change:
1. **Understand the requirement** - What exactly is being asked? What problem does it solve?
2. **Review existing codebase** - How are similar features implemented? What patterns are used?
3. **Check OpenSpec** - Are there existing specs or change proposals for related features?
4. **Identify impact** - What files/components will be affected? What are the dependencies?
5. **Plan approach** - Consider multiple implementation options and their trade-offs
6. **Create proposal** (if needed) - For new features, create OpenSpec change proposal first
7. **Then implement** - Start coding only after thorough analysis

This approach prevents:
- Rework due to misunderstandings
- Missing edge cases
- Breaking existing functionality
- Inconsistent implementation patterns

## Project Conventions

### Code Style

#### Go
- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 错误处理使用 `fmt.Errorf` 包装，包含上下文信息
- 函数注释使用 `// FunctionName does...` 格式
- 导出函数必须有注释

#### TypeScript/React
- 使用 ESLint 进行代码检查
- 组件使用函数式组件和 Hooks
- 类型定义优先使用 `interface`，避免使用 `type`（除非必要）
- 使用 `const` 而非 `let`，除非需要重新赋值
- 文件命名：组件使用 PascalCase，工具函数使用 camelCase

### Architecture Patterns

#### 后端架构
```
cmd/server/
  └── main.go              # 应用入口
internal/
  ├── api/                 # API 层
  │   ├── handlers/        # HTTP 处理器
  │   ├── middleware/      # 中间件（认证、CORS）
  │   └── routes/          # 路由定义
  ├── service/             # 业务逻辑层
  │   ├── rule/           # 规则管理
  │   ├── alert/          # 告警处理
  │   ├── query/          # 查询构建
  │   └── ...
  ├── models/             # 数据模型
  ├── repository/         # 数据访问层
  └── worker/             # 后台任务
      ├── scheduler/      # 规则调度
      ├── executor/       # 查询执行
      └── notifier/       # 通知发送
```

#### 前端架构
```
src/
  ├── pages/              # 页面组件
  ├── components/         # 可复用组件
  │   └── ui/            # shadcn/ui 组件
  ├── services/          # API 客户端
  ├── contexts/          # React Context
  ├── hooks/             # 自定义 Hooks
  └── lib/               # 工具函数
```

#### 设计原则
- **分层架构**：API → Service → Repository → Database
- **依赖注入**：通过构造函数注入依赖
- **单一职责**：每个 Service 只处理一个领域
- **错误处理**：统一错误格式，包含详细上下文

### Testing Strategy

#### 当前状态
- 项目目前**未包含自动化测试**
- 主要依赖手动测试和集成测试

#### 未来计划
- **单元测试**：核心业务逻辑（Service 层）
- **集成测试**：API 端点测试
- **端到端测试**：关键用户流程

### Git Workflow

#### 分支策略
- **main**: 生产环境代码，稳定版本
- **feature/***: 功能开发分支
- **fix/***: Bug 修复分支

#### Commit 规范
遵循 [Conventional Commits](https://www.conventionalcommits.org/)：

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type 类型**：
- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档变更
- `style`: 代码格式（不影响运行）
- `refactor`: 重构
- `perf`: 性能优化
- `test`: 测试相关
- `chore`: 构建/工具变更
- `ci`: CI 配置变更

**示例**：
```
feat(rule): add rule cloning feature

Implement rule cloning with inherited enabled status.
Includes backend API, frontend UI, and validation.

Closes #123
```

## Domain Context

### 核心概念

#### Rule (规则)
- 定义告警触发条件
- 包含查询条件、执行间隔、数据源配置、通知渠道
- 支持启用/禁用状态
- 自动统计执行次数和告警次数

#### Query Condition (查询条件)
- 灵活的 JSON 格式查询配置
- 支持多种操作符：`==`, `!=`, `>`, `>=`, `<`, `<=`, `contains`, `not_contains`
- 支持 AND/OR 逻辑组合
- 兼容 `operator` 和 `op` 两种字段名

#### Alert (告警)
- 规则触发后产生的告警记录
- 包含匹配的日志数据、时间范围、状态（已发送/失败）
- 支持查看详情、删除操作

#### ES Config (Elasticsearch 配置)
- 数据源配置，支持多个 ES 集群
- 支持多节点配置（分号分隔，自动轮询负载均衡）
- 支持 SSL/TLS 证书配置

#### Lark Config (Lark/飞书配置)
- 通知渠道配置
- 支持多个 Webhook URL
- 智能消息格式化（根据日志类型提取关键字段）

### 业务流程

#### 告警流程
1. **调度器启动**：Worker 服务启动，加载所有启用的规则
2. **定时执行**：根据规则配置的间隔时间，定时执行查询
3. **ES 查询**：构建查询条件，向 ES 发送查询请求
4. **结果处理**：如果匹配到日志，创建告警记录
5. **通知发送**：调用 Lark Webhook 发送告警消息
6. **状态更新**：更新规则统计信息（执行次数、告警次数）

#### 规则更新流程
- 用户通过前端修改规则配置
- 保存到数据库
- 调度器下次执行时自动加载最新配置
- **无需重启服务**

## Important Constraints

### 性能约束
- **列表查询不加载日志数据**：告警历史列表不返回 `logs` 字段，减少 99% 数据传输
- **详情查询限制**：告警详情只显示前 10 条日志
- **前端缓存**：30 秒缓存策略，减少重复请求
- **数据库索引**：关键字段添加索引，优化查询性能

### 数据约束
- **物理删除**：所有删除操作都是物理删除，无法恢复
- **级联删除**：删除规则时，关联的告警记录也会被删除
- **外键约束**：数据库外键约束确保数据一致性

### 安全约束
- **JWT 认证**：所有 API 请求需要 JWT Token（登录接口除外）
- **密码加密**：使用 bcrypt 加密存储密码
- **HTTPS 支持**：生产环境建议使用 HTTPS
- **环境变量**：敏感信息通过环境变量配置，不硬编码

### 部署约束
- **Docker 部署**：推荐使用 Docker Compose 部署
- **端口配置**：后端默认 8080，前端默认 3000
- **数据库初始化**：首次启动自动创建表结构
- **健康检查**：所有服务提供健康检查端点

## External Dependencies

### Elasticsearch
- **版本要求**：Elasticsearch 7.x 或 8.x
- **认证方式**：Basic Auth（用户名/密码）
- **SSL/TLS**：支持自签名证书
- **多节点**：支持分号分隔的多节点配置，自动轮询负载均衡

### PostgreSQL
- **版本要求**：PostgreSQL 12+
- **用途**：存储规则配置、告警历史、用户数据、系统配置
- **连接池**：GORM 自动管理连接池

### Lark/飞书
- **Webhook API**：通过 HTTP POST 发送告警消息
- **消息格式**：支持交互式卡片消息（Interactive Card）
- **重试机制**：失败后自动重试

## Key Design Decisions

### 为什么选择 Go？
- **高性能**：适合高并发场景，支持 1000+ 并发规则
- **简洁性**：代码简洁易维护
- **生态**：丰富的 ES 客户端和数据库驱动

### 为什么选择 React？
- **现代化**：组件化开发，可维护性强
- **生态**：丰富的 UI 组件库（shadcn/ui）
- **性能**：虚拟 DOM，高效的渲染性能

### 为什么使用 GORM？
- **易用性**：简洁的 API，快速开发
- **自动迁移**：自动创建和更新表结构
- **类型安全**：Go 的类型系统保证数据安全

### 为什么选择 PostgreSQL？
- **可靠性**：成熟稳定的关系型数据库
- **JSON 支持**：原生支持 JSON 数据类型
- **性能**：优秀的查询性能，支持复杂查询

## Deployment

### 开发环境
```bash
docker compose up -d
```

### 生产环境
```bash
docker compose -f docker-compose-prod.yml up -d
```

### 环境变量
关键环境变量（参考 `env.example`）：
- `BACKEND_PORT`: 后端服务端口
- `DB_PASSWORD`: 数据库密码
- `ES_URL`: Elasticsearch 地址（支持多节点，分号分隔）
- `JWT_SECRET`: JWT 密钥（生产环境必须修改）
- `ADMIN_USERNAME`: 管理员用户名
- `ADMIN_PASSWORD`: 管理员密码
- `WORKER_ENABLED`: 是否启用 Worker 服务
- `WORKER_CHECK_INTERVAL`: Worker 检查间隔（秒）

## Performance Targets

- **列表加载**：< 0.5 秒
- **详情查看**：< 0.3 秒
- **规则执行**：< 5 秒（取决于 ES 查询时间）
- **告警发送**：< 2 秒
- **并发支持**：1000+ 并发规则

## Security Considerations

- **密码安全**：bcrypt 加密，成本因子 10
- **JWT 安全**：Token 有效期 24 小时，密钥长度至少 32 字符
- **输入验证**：所有用户输入进行验证和清理
- **SQL 注入防护**：使用 GORM 参数化查询
- **XSS 防护**：前端使用 React 自动转义

## Maintenance

### 数据库备份
**强烈建议定期备份**（使用物理删除，无法恢复）：
```bash
docker exec elk-helper-postgres pg_dump -U postgres elk_helper | gzip > backup.sql.gz
```

### 日志清理
- 支持自动清理配置（系统配置页面）
- 默认保留 30 天数据
- 可配置清理时间和保留天数

### 监控指标
- 规则执行次数 (`run_count`)
- 告警触发次数 (`alert_count`)
- 最后执行时间 (`last_run_time`)
- 告警成功/失败状态
