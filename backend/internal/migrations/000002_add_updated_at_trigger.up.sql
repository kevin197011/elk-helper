-- 000002_add_updated_at_trigger.up.sql
-- 添加 updated_at 自动更新触发器

-- 创建通用的更新时间戳函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为各表创建触发器（如果不存在）
DO $$
BEGIN
    -- es_configs
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_es_configs_updated_at') THEN
        CREATE TRIGGER update_es_configs_updated_at
            BEFORE UPDATE ON es_configs
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
    END IF;
    
    -- lark_configs
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_lark_configs_updated_at') THEN
        CREATE TRIGGER update_lark_configs_updated_at
            BEFORE UPDATE ON lark_configs
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
    END IF;
    
    -- rules
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_rules_updated_at') THEN
        CREATE TRIGGER update_rules_updated_at
            BEFORE UPDATE ON rules
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
    END IF;
    
    -- users
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_users_updated_at') THEN
        CREATE TRIGGER update_users_updated_at
            BEFORE UPDATE ON users
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
    END IF;
    
    -- system_configs
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_system_configs_updated_at') THEN
        CREATE TRIGGER update_system_configs_updated_at
            BEFORE UPDATE ON system_configs
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
    END IF;
END
$$;

