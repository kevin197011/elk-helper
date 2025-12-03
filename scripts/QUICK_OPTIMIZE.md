# 快速性能优化

## ⚡ 最简单的方式（复制粘贴执行）

在服务器上执行以下命令：

```bash
cd /root/elk-helper

# 1. 添加索引（一行命令）
docker exec -i elk-helper-postgres psql -U postgres -d elk_helper -c "CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at DESC); CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status); CREATE INDEX IF NOT EXISTS idx_alerts_rule_created ON alerts(rule_id, created_at DESC); ANALYZE alerts;"

# 2. 验证索引
docker exec -i elk-helper-postgres psql -U postgres -d elk_helper -c "SELECT indexname FROM pg_indexes WHERE tablename = 'alerts';"

# 3. 更新服务
git pull origin main
docker compose -f docker-compose-prod.yml pull
docker compose -f docker-compose-prod.yml restart backend frontend

# 4. 等待重启
sleep 20

# 5. 检查状态
docker compose -f docker-compose-prod.yml ps

echo "✓ 优化完成！刷新浏览器测试速度"
```

## 🎯 或者分步执行

### 步骤 1: 添加索引
```bash
docker exec -i elk-helper-postgres psql -U postgres -d elk_helper -c "
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);
CREATE INDEX IF NOT EXISTS idx_alerts_rule_created ON alerts(rule_id, created_at DESC);
ANALYZE alerts;
SELECT '✓ 完成' as result;
"
```

### 步骤 2: 更新代码
```bash
cd /root/elk-helper
git pull origin main
```

### 步骤 3: 重启服务
```bash
docker compose -f docker-compose-prod.yml pull
docker compose -f docker-compose-prod.yml restart backend frontend
```

### 步骤 4: 验证
```bash
docker compose -f docker-compose-prod.yml ps
```

## 📊 预期效果

- ✅ 告警历史加载：**5-10秒 → <0.5秒**
- ✅ 详情查看：**2-3秒 → <0.3秒**
- ✅ 数据传输：减少 **99%**

## 🔍 验证索引已创建

```bash
docker exec -i elk-helper-postgres psql -U postgres -d elk_helper -c "\d alerts"
```

应该看到类似：
```
Indexes:
    "alerts_pkey" PRIMARY KEY, btree (id)
    "idx_alerts_created_at" btree (created_at DESC)
    "idx_alerts_status" btree (status)
    "idx_alerts_rule_created" btree (rule_id, created_at DESC)
```

完成！🎉

