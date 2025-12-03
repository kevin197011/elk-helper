-- 告警表性能优化 SQL

-- 1. 添加索引以加速查询
-- 创建时间索引（用于按时间排序和筛选）
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at DESC);

-- 状态索引（用于筛选已发送/失败）
CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);

-- 规则ID + 创建时间复合索引（用于按规则查询告警历史）
CREATE INDEX IF NOT EXISTS idx_alerts_rule_created ON alerts(rule_id, created_at DESC);

-- 2. 分析表以更新统计信息
ANALYZE alerts;

-- 3. 查看表大小和索引
SELECT
    pg_size_pretty(pg_total_relation_size('alerts')) as total_size,
    pg_size_pretty(pg_relation_size('alerts')) as table_size,
    pg_size_pretty(pg_total_relation_size('alerts') - pg_relation_size('alerts')) as indexes_size,
    (SELECT COUNT(*) FROM alerts) as row_count;

-- 4. 查看现有索引
SELECT
    indexname,
    indexdef
FROM pg_indexes
WHERE tablename = 'alerts'
ORDER BY indexname;

