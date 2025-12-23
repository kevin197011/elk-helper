-- 000001_init_schema.down.sql
-- 回滚初始化表结构（谨慎使用！）

-- 删除索引
DROP INDEX IF EXISTS idx_system_configs_deleted_at;
DROP INDEX IF EXISTS idx_system_configs_key;
DROP INDEX IF EXISTS idx_users_deleted_at;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_alerts_rule_created;
DROP INDEX IF EXISTS idx_alerts_status;
DROP INDEX IF EXISTS idx_alerts_created_at;
DROP INDEX IF EXISTS idx_alerts_rule_id;
DROP INDEX IF EXISTS idx_alerts_deleted_at;
DROP INDEX IF EXISTS idx_rules_enabled;
DROP INDEX IF EXISTS idx_rules_lark_config_id;
DROP INDEX IF EXISTS idx_rules_es_config_id;
DROP INDEX IF EXISTS idx_rules_deleted_at;
DROP INDEX IF EXISTS idx_rules_name;
DROP INDEX IF EXISTS idx_lark_configs_deleted_at;
DROP INDEX IF EXISTS idx_lark_configs_name;
DROP INDEX IF EXISTS idx_es_configs_deleted_at;
DROP INDEX IF EXISTS idx_es_configs_name;

-- 删除表（按依赖顺序）
DROP TABLE IF EXISTS system_configs;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS alerts;
DROP TABLE IF EXISTS rules;
DROP TABLE IF EXISTS lark_configs;
DROP TABLE IF EXISTS es_configs;

