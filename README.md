# ELK Helper - 智能日志告警系统

一个现代化的、高性能的日志监控告警系统，支持多数据源、多日志类型、灵活的查询配置和实时告警推送。

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go)
![React](https://img.shields.io/badge/React-18-61DAFB?logo=react)
![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)

## 🎯 核心特性

### 架构与部署
- ✅ **前后端分离架构**：Go + React，现代化技术栈
- ✅ **Docker 部署**：支持 Docker Compose 一键部署
- ✅ **GitHub Actions CI/CD**：自动构建和发布 Docker 镜像到 GHCR
- ✅ **生产级配置**：资源限制、健康检查、优雅重启

### 用户界面
- ✅ **现代化 UI**：Ant Design + React 18，响应式设计
- ✅ **简洁优雅风格**：浅色系统一设计
- ✅ **直观的操作**：可视化配置，无需编写代码
- ✅ **实时数据刷新**：仪表盘每 5 分钟自动更新

### 数据源管理
- ✅ **多 ES 节点支持**：支持配置多个 ES 节点（分号分隔），自动轮询负载均衡
- ✅ **多数据源配置**：可配置多个 Elasticsearch 集群
- ✅ **SSL/TLS 支持**：完整的证书配置，支持自签名证书
- ✅ **连接测试**：一键测试数据源连通性
- ✅ **默认数据源**：灵活切换不同环境

### 告警规则
- ✅ **灵活的查询配置**：
  - 支持多种操作符（==, !=, >, >=, <, <=, contains, not_contains）
  - 支持复合查询（AND/OR 逻辑组合）
  - 支持 `operator` 和 `op` 两种字段名
- ✅ **规则实时更新**：修改规则后自动生效，无需重启服务
- ✅ **规则自动执行**：新创建/更新/启用的规则立即触发检测
- ✅ **规则导入导出**：支持 JSON 格式，自动去重和更新
- ✅ **智能编辑器**：
  - 支持 Tab 键缩进
  - 一键格式化 JSON
  - 快速示例模板（非200响应、4xx/5xx错误、慢查询等）
- ✅ **规则测试**：保存前可测试查询条件
- ✅ **弹框式新建/编辑**：规则新建与修改在列表页弹框内完成（更高效）
- ✅ **规则名称唯一约束**：数据库级别保证规则名称唯一性

### 告警推送
- ✅ **智能消息格式**：
  - 自动识别日志类型（nginx, java, go, c++, python, nodejs）
  - 智能提取关键字段
  - 显示前 5 条日志摘要
  - 消息大小减少 95-99%
- ✅ **多通知渠道**：支持配置多个告警 Webhook（飞书/Lark 等）
- ✅ **告警重试**：失败自动重试，确保送达
- ✅ **告警历史**：完整记录，支持查询和筛选

### 性能优化
- ✅ **高性能查询**：
  - 列表查询不加载 logs 字段（减少 99% 数据量）
  - 详情查询限制前 10 条日志
  - 数据库索引优化
  - 前端查询缓存（30 秒）
- ✅ **高并发支持**：
  - 每个规则独立 goroutine
  - 支持 1000+ 并发规则
  - ES 连接池优化
  - PostgreSQL 连接优化

### 安全与权限
- ✅ **JWT 认证**：Token 有效期 24 小时
- ✅ **角色权限**：Admin/User 角色管理
- ✅ **密码加密**：bcrypt 加密存储
- ✅ **HTTPS 支持**：完整的 SSL/TLS 配置
- ✅ **ES 证书验证**：支持跳过证书验证（测试环境）

### 数据库管理
- ✅ **自动迁移**：使用 golang-migrate 自动管理数据库 schema
- ✅ **版本控制**：SQL 迁移文件版本化管理
- ✅ **零停机更新**：服务重启时自动应用迁移

## 🏗️ 技术架构

```
┌─────────────────────────────────────────────┐
│  前端 (React 18 + TypeScript + Ant Design)  │
│  - 规则管理、告警历史、数据源配置            │
│  - 实时编辑、格式化、快速示例               │
│  - 仪表盘自动刷新、清理任务配置              │
└────────────────┬────────────────────────────┘
                 │ REST API (JWT)
┌────────────────┴────────────────────────────┐
│  后端 (Go 1.23 + Gin)                       │
│  ├─ API 服务：CRUD、认证、配置管理          │
│  └─ Worker 服务：规则调度、查询执行         │
└─────┬──────────────────────┬────────────────┘
      │                      │
      ▼                      ▼
┌─────────────┐      ┌──────────────────┐
│ PostgreSQL  │      │ Elasticsearch    │
│ - 规则配置  │      │ - 多节点轮询     │
│ - 告警历史  │      │ - SSL/TLS        │
│ - 用户数据  │      │ - 高并发查询     │
└─────────────┘      └──────────────────┘
      │
      ▼
┌─────────────┐
│ Lark/飞书   │
│ - 智能摘要  │
│ - 多渠道    │
└─────────────┘
```

## 🚀 快速开始

### 方式 1: 使用预构建镜像（推荐生产环境）

```bash
# 1. 克隆项目
git clone https://github.com/kevin197011/elk-helper.git
cd elk-helper

# 2. 配置环境变量
cp env.example .env
vim .env  # 修改配置

# 3. 启动服务（使用 GitHub Container Registry 镜像）
docker compose -f docker-compose-prod.yml up -d

# 4. 查看服务状态
docker compose -f docker-compose-prod.yml ps

# 5. 查看日志
docker compose -f docker-compose-prod.yml logs -f
```

**访问地址**：
- 前端界面：http://localhost:3000
- 后端 API：http://localhost:8080

**默认账户**：
- 用户名：`admin`
- 密码：`admin123`

> 💡 **提示**：首次登录后建议立即修改默认密码

### 方式 2: 本地构建（开发环境）

```bash
# 1. 克隆项目
git clone <repository-url>
cd elk-helper

# 2. 启动所有服务（包括 Elasticsearch）
docker compose up -d

# 3. 查看服务状态
docker compose ps
```

## 📋 环境配置

### 基本配置（.env 文件）

```bash
# 后端端口（默认 8080）
BACKEND_PORT=8080

# 数据库密码
DB_PASSWORD=postgres

# 单次 DB 查询超时（秒，默认: 5）
DB_QUERY_TIMEOUT_SECONDS=5

# Elasticsearch 配置（支持多节点，分号分隔）
ES_URL=https://elasticsearch:9200
# 或多节点：ES_URL=https://es-node1:9200;https://es-node2:9200;https://es-node3:9200
ES_USERNAME=elastic
ES_PASSWORD=changeme

# 是否使用 SSL/TLS（默认: URL 为 https 时自动启用）
ES_USE_SSL=true
# 是否跳过证书验证（开发环境可用，生产环境不建议）
ES_SKIP_VERIFY=true
# 可选：CA 证书内容（PEM），用于自签证书校验
ES_CA_CERTIFICATE=

# 单次 ES 查询超时（秒，默认: 30）
ES_QUERY_TIMEOUT_SECONDS=30

# 管理员账户
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin123
ADMIN_EMAIL=admin@example.com

# JWT 密钥（生产环境必须修改！release 模式会校验最小长度 >= 32）
JWT_SECRET=elk-helper-dev-compose-secret-please-change-32chars

# 跨域允许列表（多个用逗号分隔）
CORS_ORIGINS=http://localhost:3000

# Worker 配置
WORKER_ENABLED=true
WORKER_CHECK_INTERVAL=30
WORKER_BATCH_SIZE=500
WORKER_MAX_CONCURRENCY=1000

# 单次告警发送最大耗时（秒，默认: 20）
ALERT_SEND_TIMEOUT_SECONDS=20

# 可选：敏感信息加密（base64 编码的 32 字节 key；用于 ES 密码、Webhook 等）
APP_ENCRYPTION_KEY=

# 登录接口限流（默认: 60/min，burst 20）
LOGIN_RATE_LIMIT_PER_MINUTE=60
LOGIN_RATE_LIMIT_BURST=20
```

### Nginx 反向代理配置

支持通过宿主机 Nginx 反向代理访问：

```bash
# 1. 复制配置文件
sudo cp nginx/simple.conf /etc/nginx/sites-available/elk-helper.conf

# 2. 修改域名
sudo vim /etc/nginx/sites-available/elk-helper.conf

# 3. 配置 SSL 证书
sudo certbot --nginx -d your-domain.com

# 4. 启用配置
sudo ln -s /etc/nginx/sites-available/elk-helper.conf /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

详细配置请参考 `nginx/README.md`

## 🎨 功能特性

### 1. 多 ES 节点支持

支持配置单个或多个 Elasticsearch 节点，实现负载均衡和高可用：

**单节点配置**：
```
https://10.170.1.54:9200
```

**多节点配置**（分号分隔，自动轮询）：
```
https://10.170.1.54:9200;https://10.170.1.55:9200;https://10.170.1.56:9200
```

**优势**：
- 查询性能提升 2-3 倍
- 高可用，单节点故障不影响服务
- 自动故障转移和重试

### 2. 智能告警格式

根据日志类型自动提取关键字段，发送简洁明了的告警消息：

#### Nginx 日志告警
```
🚨 ELK 告警
📋 规则名称: Nginx 5xx 错误告警
⏰ 时间范围: 2025-12-03 16:30 ~ 16:35 | 🔔 告警数量: 156 条
📊 索引名称: prod-nginx-access-*

📝 日志摘要
#1 `response_code`: 500 | `ip`: 192.168.1.100 | `request`: /api/users | `method`: POST
#2 `response_code`: 502 | `ip`: 192.168.1.101 | `request`: /api/orders | `method`: GET
#3 `response_code`: 500 | `ip`: 192.168.1.102 | `request`: /api/login | `method`: POST
...
```

#### Java 应用日志告警
```
🚨 ELK 告警
📋 规则名称: Java ERROR 级别日志
⏰ 时间范围: 2025-12-03 18:45 ~ 18:50 | 🔔 告警数量: 23 条
📊 索引名称: prod-app-logs-*

📝 日志摘要
#1 `module`: user-service | `node_ip`: 10.0.1.21 | `message`: [ERROR] NullPointerException...
#2 `module`: user-service | `node_ip`: 10.0.1.22 | `message`: [ERROR] SQLException...
...
```

**支持的日志类型**：nginx, java, go, c++, python, nodejs, application

### 3. 查询条件配置

#### 使用快速示例按钮
- 非200响应
- 4xx/5xx错误
- 5xx错误
- 慢查询

#### 或手动编辑 JSON

**单条件**：
```json
[
  {
    "field": "response_code",
    "operator": "!=",
    "value": 200
  }
]
```

**复合查询（AND + OR）**：
```json
[
  {
    "field": "request_method",
    "operator": "==",
    "value": "POST",
    "logic": "and"
  },
  {
    "field": "response_code",
    "operator": "==",
    "value": 500,
    "logic": "or"
  },
  {
    "field": "response_code",
    "operator": "==",
    "value": 502,
    "logic": "or"
  }
]
```

**支持的操作符**：`==`, `!=`, `>`, `>=`, `<`, `<=`, `contains`, `not_contains`, `exists`

### 4. 规则实时更新

修改规则配置后，下次执行时自动生效，无需重启服务：

- ✅ 查询条件更新 → 实时生效
- ✅ 执行间隔修改 → 动态调整
- ✅ 数据源切换 → 即时切换
- ✅ 告警渠道更改 → 立即应用
- ✅ 启用/禁用 → 30 秒内生效

### 5. 规则导入导出

支持规则的批量导入和导出，方便规则迁移和备份：

**导出规则**：
- 路径：规则管理 → 导出
- 格式：JSON 文件
- 包含：规则配置、ES 配置、告警配置

**导入规则**：
- 路径：规则管理 → 导入
- 自动去重：基于规则名称
- 智能更新：检测变更并更新现有规则
- 跳过未变更：相同规则自动跳过
- 详细反馈：显示创建/更新/跳过/错误数量

### 6. 清理任务配置

**路径**：系统配置 → 清理任务配置

- ✅ 定时清理：每天指定时间自动清理历史告警
- ✅ 立即清理：手动触发清理任务
- ✅ 执行状态：显示上次执行时间和结果
- ✅ 状态持久化：配置更新时保留执行状态

### 7. 性能优化

#### 告警历史加速
- 列表查询不加载日志数据（减少 99% 传输量）
- 详情查询限制前 10 条日志
- 数据库索引优化（查询加速 10-100 倍）
- 前端缓存策略（减少 70% 请求）

**性能对比**：
| 操作 | 优化前 | 优化后 | 提升 |
|------|--------|--------|------|
| 列表加载 | 5-10 秒 | <0.5 秒 | **90-95%** ⬆️ |
| 详情查看 | 2-3 秒 | <0.3 秒 | **90%** ⬆️ |
| 数据传输 | 5-10 MB | 20-50 KB | **99%** ⬇️ |

## 📦 Docker 镜像

### GitHub Container Registry

项目提供自动构建的 Docker 镜像：

```yaml
services:
  backend:
    image: ghcr.io/kevin197011/elk-helper/backend:latest

  frontend:
    image: ghcr.io/kevin197011/elk-helper/frontend:latest
```

### 镜像标签

- `latest` - 最新的 main 分支构建
- `v1.0.0` - 版本号标签
- `main-sha-abc1234` - 特定提交

### 构建优化

- ✅ 只构建 linux/amd64（加速 60-70%）
- ✅ Go 模块缓存（二次构建提速 80-90%）
- ✅ npm 缓存（二次构建提速 90%）
- ✅ 编译优化（减小镜像体积 20-30%）

## 🔧 部署配置

### 生产环境部署

使用预构建镜像 + 外部 Elasticsearch：

```yaml
# docker-compose-prod.yml
services:
  backend:
    image: ghcr.io/kevin197011/elk-helper/backend:latest
    ports:
      - "8081:8080"  # 宿主机:容器
    environment:
      - ES_URL=https://your-es-cluster:9200  # 外部 ES 集群

  frontend:
    image: ghcr.io/kevin197011/elk-helper/frontend:latest
    ports:
      - "3000:80"
```

### Nginx 反向代理

提供简单的 Nginx 配置模板：

```nginx
# nginx/simple.conf
server {
    listen 443 ssl http2;
    server_name elk-helper.yourdomain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 📖 使用指南

### 1. 配置 ES 数据源

**路径**: 数据源配置 → 新建配置

**单节点**：
```
ES 地址: https://es.example.com:9200
用户名: elastic
密码: ******
跳过证书验证: ✓
```

**多节点**（负载均衡）：
```
ES 地址: https://es-node1.example.com:9200;https://es-node2.example.com:9200;https://es-node3.example.com:9200
用户名: elastic
密码: ******
跳过证书验证: ✓
```

点击"测试连接"验证配置。

### 2. 创建告警规则

**路径**: 规则管理 → 新建规则（弹框）

#### 示例 1: 监控 Nginx 5xx 错误

```
规则名称: Nginx 5xx 错误告警
索引模式: prod-nginx-access-*
查询条件: 点击 "5xx错误" 快捷按钮
执行间隔: 60 秒
ES 数据源: 选择已配置的数据源
告警配置: 选择告警群
```

#### 示例 2: 监控 Java ERROR 日志

```
规则名称: Java ERROR 级别日志
索引模式: app-logs-*
查询条件:
[
  {
    "field": "log_type",
    "operator": "==",
    "value": "java",
    "logic": "and"
  },
  {
    "field": "message",
    "operator": "contains",
    "value": "ERROR"
  }
]
执行间隔: 300 秒
```

#### 示例 3: 监控慢请求

点击 "慢查询" 快捷按钮，自动生成：
```json
[
  {
    "field": "responsetime",
    "operator": ">",
    "value": 3,
    "logic": "and"
  },
  {
    "field": "response_code",
    "operator": "==",
    "value": 200,
    "logic": "and"
  }
]
```

### 3. 配置告警通知（Webhook）

**路径**: 告警配置 → 新建告警配置

```
配置名称: 生产告警群
Webhook URL: https://open.feishu.cn/open-apis/bot/v2/hook/...
```

点击"测试发送"验证配置。

### 4. 查看告警历史

**路径**: 告警历史

- 搜索规则名称、索引
- 筛选状态（已发送/失败）
- 查看告警详情（前 10 条日志）
- 复制 JSON 到剪贴板

**⚠️ 删除说明**：
- 所有删除操作均为**物理删除**（立即从数据库移除）
- 删除后**无法恢复**，请谨慎操作
- 建议配置自动清理任务管理历史数据
- **强烈建议定期备份数据库**

## ⚡ 性能优化

### 数据库迁移管理

项目使用 `golang-migrate` 自动管理数据库 schema：

```bash
# 查看迁移状态
rake db:status

# 手动执行迁移（通常不需要，服务启动时自动执行）
rake db:migrate

# 回滚迁移
rake db:rollback

# 查看所有迁移
rake db:list
```

**迁移文件位置**：`backend/internal/migrations/`

**自动迁移**：服务启动时自动应用所有未执行的迁移

### 数据库索引优化

```bash
# 执行索引优化脚本
bash scripts/add-indexes.sh

# 或手动执行
docker exec -i elk-helper-postgres psql -U postgres -d elk_helper -c "
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);
CREATE INDEX IF NOT EXISTS idx_alerts_rule_created ON alerts(rule_id, created_at DESC);
ANALYZE alerts;
"
```

**效果**：告警历史加载速度提升 **90-95%**

### 定期清理数据

**路径**: 系统配置 → 清理任务配置

```
启用清理: ✓
执行时间: 03:00
保留天数: 30
```

系统会每天自动清理 30 天前的告警记录。

**执行状态**：
- 显示上次执行时间
- 显示执行结果（成功/失败）
- 显示删除的记录数量
- 配置更新时状态自动保留

## 🔐 安全配置

### 生产环境必改项

```bash
# 1. JWT 密钥（至少 32 字符）
JWT_SECRET=$(openssl rand -base64 32)

# 2. 管理员密码
ADMIN_PASSWORD=$(openssl rand -base64 16)

# 3. 数据库密码
DB_PASSWORD=$(openssl rand -base64 16)

# 4. ES 密码
ES_PASSWORD=your-secure-password
```

### SSL/TLS 配置

#### Let's Encrypt（推荐）

```bash
# 获取免费 SSL 证书
sudo certbot --nginx -d elk-helper.yourdomain.com
```

#### 自签名证书（测试环境）

```bash
cd ssl
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout key.pem -out cert.pem \
  -subj "/CN=elk-helper.local"
```

## 🛠️ 运维管理

### 查看服务状态

```bash
# 查看所有服务
docker compose -f docker-compose-prod.yml ps

# 查看日志
docker compose -f docker-compose-prod.yml logs -f backend
docker compose -f docker-compose-prod.yml logs -f frontend

# 查看资源使用
docker stats
```

### 更新服务

```bash
# 1. 拉取最新镜像
docker compose -f docker-compose-prod.yml pull

# 2. 重启服务
docker compose -f docker-compose-prod.yml up -d

# 3. 清理旧镜像
docker image prune -f
```

### 备份数据（重要！）

⚠️ **由于使用物理删除，强烈建议定期备份数据**

```bash
# 手动备份
docker exec elk-helper-postgres pg_dump -U postgres elk_helper | gzip > elk_helper_$(date +%Y%m%d).sql.gz

# 恢复数据
gunzip < elk_helper_20251203.sql.gz | docker exec -i elk-helper-postgres psql -U postgres -d elk_helper

# 配置自动备份（添加到 crontab）
echo "0 2 * * * docker exec elk-helper-postgres pg_dump -U postgres elk_helper | gzip > /backup/elk_helper_\$(date +\%Y\%m\%d).sql.gz" | crontab -
```

**备份建议**：
- 每日自动备份
- 保留最近 30 天备份
- 重要操作前手动备份

### 手动创建用户

```bash
# 创建 devops 管理员账户
docker exec -i elk-helper-postgres psql -U postgres -d elk_helper -c "
INSERT INTO users (username, password, email, role, enabled, created_at, updated_at)
VALUES (
    'devops',
    '\$2a\$10\$rOZvN8g8WQYxP1nH9xhQNOqKJYmXZvDxH3jPKp0Y7jzGqN5xVWE4e',
    'devops@example.com',
    'admin',
    true,
    NOW(),
    NOW()
);
"
# 用户名: devops, 密码: devops123
```

## 📊 监控与告警

### 系统监控指标

- 规则执行次数 (`run_count`)
- 告警触发次数 (`alert_count`)
- 最后执行时间 (`last_run_time`)
- 告警成功/失败状态

### 推荐的告警规则

#### 1. Web 服务器监控

```
规则名称: Nginx 5xx 错误告警
查询条件: response_code >= 500
间隔: 60 秒
阈值: 10 条/5分钟
```

#### 2. 应用错误监控

```
规则名称: 应用 ERROR 日志
查询条件: level == "ERROR" OR message contains "ERROR"
间隔: 180 秒
阈值: 5 条/5分钟
```

#### 3. 慢请求监控

```
规则名称: API 响应超时
查询条件: responsetime > 3 AND response_code == 200
间隔: 300 秒
阈值: 20 条/5分钟
```

## 🐛 故障排查

### 502 Bad Gateway

**症状**：Frontend 无法连接 Backend

**解决**：
```bash
# 1. 确保代码最新
git pull origin main

# 2. 设置正确的端口
echo "BACKEND_PORT=8080" > .env

# 3. 完全重启
docker compose -f docker-compose-prod.yml down
docker compose -f docker-compose-prod.yml up -d

# 4. 等待启动
sleep 60

# 5. 检查状态
docker compose -f docker-compose-prod.yml ps
```

### 告警未触发

**检查清单**：
- [ ] 规则已启用
- [ ] Worker 服务运行中（查看日志）
- [ ] ES 数据源连接正常
- [ ] 查询条件正确（使用测试功能）
- [ ] Lark Webhook 配置正确

### 页面加载慢

**优化方案**：
```bash
# 执行性能优化
bash scripts/add-indexes.sh

# 更新到最新版本
git pull origin main
docker compose -f docker-compose-prod.yml pull
docker compose -f docker-compose-prod.yml restart backend frontend
```

## 📚 参考文档

- `nginx/README.md` - Nginx 反向代理配置指南
- `env.example` - 环境变量配置示例
- `scripts/generate-test-logs.rb` - 测试日志生成脚本（Ruby）
- `scripts/reset-admin-password.sh` - 重置管理员密码脚本
- `Rakefile` - 数据库迁移管理任务

## 🛠️ 开发工具

### 测试日志生成

使用 Ruby 脚本生成测试日志到 Elasticsearch：

```bash
# 使用默认配置（https://localhost:9200, elastic/changeme）
ruby scripts/generate-test-logs.rb --type nginx --count 1000

# 自定义配置
export ES_URL="https://es.example.com:9200"
export ES_USERNAME="elastic"
export ES_PASSWORD="your-password"
ruby scripts/generate-test-logs.rb --type java --count 500 --prefix app-logs
```

### 密码管理

```bash
# 生成密码哈希
./scripts/generate-password-hash.sh

# 重置管理员密码
./scripts/reset-admin-password.sh
```

## 🔄 更新日志

### v1.2.0 (2025-12-23)

**新功能**：
- ✅ 前端框架迁移到 Ant Design
- ✅ 规则导入导出功能（自动去重和更新）
- ✅ 清理任务配置页面（定时清理、立即清理、执行状态）
- ✅ 数据库自动迁移（golang-migrate）
- ✅ 规则创建/更新/启用后立即执行
- ✅ 仪表盘自动刷新（每 5 分钟）
- ✅ 简洁优雅的统一 UI 风格（浅色系、轻量网格动效登录页）
- ✅ Logo / Icon 重新设计（更高对比、更易识别）
- ✅ 规则名称唯一约束

**改进**：
- ✅ 清理任务执行状态持久化
- ✅ ES 数据源密码保存修复
- ✅ 告警趋势图表时间对齐修复
- ✅ 规则执行日志增强
- ✅ 删除操作提示优化

**Bug 修复**：
- ✅ 修复 ES 密码首次创建不保存的问题
- ✅ 修复清理任务状态更新后消失的问题
- ✅ 修复告警趋势图表滞后 1 小时的问题

### v1.1.0 (2025-12-03)

**新功能**：
- ✅ GitHub Actions 自动构建和发布 Docker 镜像
- ✅ 多 ES 节点支持（分号分隔，轮询负载均衡）
- ✅ 告警消息智能格式化（根据日志类型提取关键字段）
- ✅ 规则实时更新（修改后自动生效，无需重启）
- ✅ 查询条件编辑器改进（Tab 缩进、格式化、快速示例）
- ✅ 告警历史性能优化（列表不加载 logs，数据库索引）
- ✅ 动态端口配置支持
- ✅ 支持 `operator` 和 `op` 两种字段名

**性能优化**：
- ✅ 镜像构建速度提升 60-90%
- ✅ 告警列表加载速度提升 90-95%
- ✅ 告警消息大小减少 95-99%
- ✅ 数据库查询加速 10-100 倍

**Bug 修复**：
- ✅ 修复规则更新后仍使用旧配置的问题
- ✅ 修复 operator 字段保存丢失的问题
- ✅ 修复测试失败后状态不更新的问题
- ✅ 修复编辑规则时下拉框显示错误的问题
- ✅ 修复 PostgreSQL 初始化参数错误

### v1.0.0 (初始版本)

- 基础功能实现

## 🤝 贡献指南

1. Fork 本项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'feat: Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

### Commit 规范

遵循 [Conventional Commits](https://www.conventionalcommits.org/)：

```
feat: 新功能
fix: Bug 修复
docs: 文档更新
style: 代码格式
refactor: 重构
perf: 性能优化
test: 测试
chore: 构建/工具变更
```

## 📄 许可证

MIT License - Copyright (c) 2025 kk

## 👥 支持

- 🐛 Issues: https://github.com/your-org/elk-helper/issues
- 📖 文档: 查看项目 `docs/` 目录
- 💬 讨论: GitHub Discussions

---

**如有问题或建议，欢迎提交 Issue 或 PR！**

🌟 如果这个项目对你有帮助，请给个 Star！
