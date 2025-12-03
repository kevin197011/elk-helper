docker exec -i elk-helper-postgres psql -U postgres -d elk_helper << 'EOF'
-- 创建索引
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);
CREATE INDEX IF NOT EXISTS idx_alerts_rule_created ON alerts(rule_id, created_at DESC);

-- 分析表
ANALYZE alerts;

-- 显示结果
\echo '✓ 索引创建完成'
\echo ''
\echo '当前索引:'
SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'alerts';

\echo ''
\echo '表大小统计:'
SELECT
    pg_size_pretty(pg_total_relation_size('alerts')) as total_size,
    pg_size_pretty(pg_relation_size('alerts')) as table_size,
    pg_size_pretty(pg_total_relation_size('alerts') - pg_relation_size('alerts')) as indexes_size,
    (SELECT COUNT(*) FROM alerts) as row_count;
EOF