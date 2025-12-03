#!/bin/bash
# 为告警表添加索引以提升性能

echo "=== 添加数据库索引 ==="

# 创建临时 SQL 文件
cat > /tmp/optimize-alerts.sql << 'EOFMARKER'
-- 创建索引
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);
CREATE INDEX IF NOT EXISTS idx_alerts_rule_created ON alerts(rule_id, created_at DESC);

-- 分析表
ANALYZE alerts;

-- 查看结果
SELECT 'idx_alerts_created_at' as index_name UNION ALL
SELECT 'idx_alerts_status' UNION ALL
SELECT 'idx_alerts_rule_created';

-- 查看表大小
SELECT
    pg_size_pretty(pg_total_relation_size('alerts')) as total_size,
    (SELECT COUNT(*) FROM alerts) as row_count;
EOFMARKER

# 复制到容器并执行
docker cp /tmp/optimize-alerts.sql elk-helper-postgres:/tmp/
docker exec -i elk-helper-postgres psql -U postgres -d elk_helper -f /tmp/optimize-alerts.sql

# 清理
rm -f /tmp/optimize-alerts.sql
docker exec -i elk-helper-postgres rm -f /tmp/optimize-alerts.sql

echo ""
echo "✓ 索引优化完成！"
echo ""
echo "现在更新服务以应用代码优化:"
echo "  git pull origin main"
echo "  docker compose -f docker-compose-prod.yml pull"
echo "  docker compose -f docker-compose-prod.yml restart backend frontend"
