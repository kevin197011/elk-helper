-- 000002_add_updated_at_trigger.down.sql
-- 删除 updated_at 触发器

DROP TRIGGER IF EXISTS update_system_configs_updated_at ON system_configs;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_rules_updated_at ON rules;
DROP TRIGGER IF EXISTS update_lark_configs_updated_at ON lark_configs;
DROP TRIGGER IF EXISTS update_es_configs_updated_at ON es_configs;

DROP FUNCTION IF EXISTS update_updated_at_column();

