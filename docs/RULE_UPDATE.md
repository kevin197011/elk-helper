# 规则更新机制说明

## 🐛 问题描述

### 修改前的问题

当你更新了规则配置后，Worker 仍然使用旧的规则配置执行查询。

**原因**：
- Worker 在启动时为每个规则创建一个 goroutine
- Goroutine 中保存的是**启动时的规则配置副本**
- 即使数据库中规则已更新，运行中的 goroutine 仍使用旧配置
- 需要重启 backend 服务才能应用新配置

### 用户体验

```
1. 创建规则 A（查询条件：response_code != 200）
2. Worker 启动，开始执行规则 A
3. 修改规则 A（查询条件：response_code >= 500）
4. ❌ Worker 仍然使用旧配置（!= 200）执行查询
5. 💡 需要重启 backend 才能应用新配置
```

## ✅ 修复方案

### 修改后的机制

每次执行规则时，从数据库**重新加载最新配置**，无需重启服务。

**工作流程**：
1. Worker 启动时为每个规则创建 goroutine
2. Goroutine 按照时间间隔定期执行
3. **每次执行前，从数据库重新加载该规则的最新配置** ✅
4. 使用最新配置执行查询
5. 如果间隔时间也改了，自动调整执行频率

### 新的用户体验

```
1. 创建规则 A（查询条件：response_code != 200）
2. Worker 启动，开始执行规则 A
3. 修改规则 A（查询条件：response_code >= 500）
4. ✅ 下次执行时自动使用新配置（>= 500）
5. 🚀 无需重启，立即生效
```

## 📋 支持的实时更新

以下配置修改后，在下次执行时自动生效：

### ✅ 查询配置
- **索引模式** (`index_pattern`)
- **查询条件** (`queries`)
- **执行间隔** (`interval`) - 动态调整定时器

### ✅ 数据源配置
- **ES 配置** (`es_config_id`)
- 切换到不同的 ES 数据源

### ✅ 告警配置
- **Lark Webhook** (`lark_webhook`)
- **Lark 配置** (`lark_config_id`)
- 切换到不同的告警渠道

### ✅ 规则元数据
- **规则名称** (`name`)
- **规则描述** (`description`)

### ✅ 规则状态
- **启用/禁用** (`enabled`)
  - 禁用：立即停止执行
  - 启用：在下个检查周期启动（默认 30 秒）

## 🕐 更新生效时间

| 操作 | 生效时间 | 说明 |
|------|---------|------|
| 修改规则配置 | 下次执行时 | 等待当前执行周期结束 |
| 修改执行间隔 | 下次执行后 | 新间隔在下次执行后生效 |
| 禁用规则 | 30 秒内 | Scheduler 每 30 秒检查一次 |
| 启用规则 | 30 秒内 | Scheduler 每 30 秒检查一次 |
| 删除规则 | 30 秒内 | Scheduler 会停止该规则 |

**示例时间线**：

```
16:00:00 - 规则 A 执行（使用旧配置）
16:00:30 - 你修改了规则 A 的查询条件
16:01:00 - 规则 A 执行（使用新配置）✓ 已生效
16:02:00 - 规则 A 执行（使用新配置）
...
```

## 🔄 Scheduler 检查机制

### 检查周期
- **默认**: 30 秒
- **配置**: `WORKER_CHECK_INTERVAL` 环境变量

### 检查内容
1. **新增启用的规则** → 启动新 goroutine
2. **禁用的规则** → 停止对应 goroutine
3. **已删除的规则** → 停止对应 goroutine

### 规则配置更新
- 每次执行前从数据库重新加载
- 包括：查询条件、ES 配置、Lark 配置等
- **无需重启服务**

## 💡 最佳实践

### 1. 测试规则后再保存

```
1. 在规则编辑页面填写配置
2. 点击"测试规则"按钮
3. 查看测试结果，确认查询条件正确
4. 点击"保存"
5. ✓ 新配置会在下次执行时生效
```

### 2. 修改关键规则时

```
1. 临时禁用规则
2. 修改配置并测试
3. 确认无误后重新启用
4. 等待最多 30 秒开始执行
```

### 3. 快速应用更新

如果不想等待下次执行周期，可以：

**方式 1**: 禁用后重新启用（推荐）
```
1. 点击规则的"禁用"开关
2. 等待 30 秒（规则停止）
3. 点击"启用"开关
4. 等待 30 秒（规则启动，使用新配置）
```

**方式 2**: 重启 Backend 服务
```bash
docker compose -f docker-compose-prod.yml restart backend
```

## 🔍 查看规则执行情况

### 查看后端日志

```bash
# 查看规则启动/停止日志
docker compose -f docker-compose-prod.yml logs backend | grep "Starting rule\|Stopping rule"

# 查看规则执行日志
docker compose -f docker-compose-prod.yml logs backend | grep "Rule.*executed"

# 查看规则更新日志
docker compose -f docker-compose-prod.yml logs backend | grep "interval updated"

# 实时查看
docker compose -f docker-compose-prod.yml logs -f backend
```

### 日志示例

```
[INFO] Starting rule 1 (Nginx 500 错误告警)
[INFO] Rule 1 (Nginx 500 错误告警) interval updated to 1m0s
[INFO] Stopping rule 2 (disabled)
```

## 📊 配置更新流程图

```
修改规则
    ↓
保存到数据库 ✓
    ↓
等待下次执行周期
    ↓
Worker 执行前 → 从数据库重新加载规则 ✓
    ↓
使用最新配置执行查询 ✓
    ↓
发送告警（使用新配置）✓
```

## 🛠️ 故障排查

### 问题 1: 规则更新后还是用旧配置

**检查**：
```bash
# 1. 确认数据库中规则已更新
docker exec -it elk-helper-postgres psql -U postgres -d elk_helper -c \
"SELECT id, name, queries FROM rules WHERE id = 1;"

# 2. 查看后端日志，确认规则正在运行
docker compose -f docker-compose-prod.yml logs backend | grep "rule 1"

# 3. 等待下次执行周期（查看 interval 配置）
```

**解决**：
- 等待当前执行周期结束
- 或禁用后重新启用规则
- 或重启 backend 服务

### 问题 2: 规则没有执行

**检查**：
```bash
# 1. 确认规则已启用
docker exec -it elk-helper-postgres psql -U postgres -d elk_helper -c \
"SELECT id, name, enabled FROM rules WHERE id = 1;"

# 2. 查看 Worker 状态
docker compose -f docker-compose-prod.yml logs backend | grep "Scheduler started"

# 3. 查看是否有错误日志
docker compose -f docker-compose-prod.yml logs backend | grep ERROR
```

**解决**：
- 确保规则已启用
- 检查 ES 数据源配置是否正确
- 检查 Lark Webhook 配置是否正确

### 问题 3: 间隔时间没有更新

**原因**：间隔时间更新在下次执行后才会生效

**时间线**：
```
16:00:00 - 执行（旧间隔 5 分钟）
16:00:30 - 修改间隔为 1 分钟
16:05:00 - 执行（仍然是 5 分钟后）
16:06:00 - 执行（新间隔 1 分钟生效）✓
16:07:00 - 执行（1 分钟间隔）
```

**快速生效**：禁用后重新启用规则

## 🚀 部署更新

```bash
# 1. 提交代码
git add .
git commit -m "fix(worker): 规则执行前重新加载配置，支持实时更新"
git push origin main

# 2. 等待 GitHub Actions 构建

# 3. 服务器更新
docker compose -f docker-compose-prod.yml pull
docker compose -f docker-compose-prod.yml restart backend

# 4. 验证
docker compose -f docker-compose-prod.yml logs -f backend | grep "Starting rule"
```

## 📝 版本对比

| 特性 | 旧版本 | 新版本 |
|------|--------|--------|
| 规则配置更新 | 需要重启服务 | 自动加载，无需重启 ✓ |
| 间隔时间更新 | 需要重启服务 | 动态调整 ✓ |
| 查询条件更新 | 需要重启服务 | 实时生效 ✓ |
| ES 数据源切换 | 需要重启服务 | 实时切换 ✓ |
| Webhook 更新 | 需要重启服务 | 实时更新 ✓ |

现在修改规则后，下次执行时就会自动使用新配置了！🎉

