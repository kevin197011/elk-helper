-- 000001_init_schema.up.sql
-- 初始化数据库表结构

-- ES 数据源配置表
CREATE TABLE IF NOT EXISTS es_configs (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    
    name VARCHAR(255) NOT NULL,
    url VARCHAR(1024) NOT NULL,
    username VARCHAR(255),
    password TEXT,
    password_enc TEXT,
    use_ssl BOOLEAN NOT NULL DEFAULT FALSE,
    skip_verify BOOLEAN NOT NULL DEFAULT FALSE,
    ca_certificate TEXT,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    description TEXT,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    last_test_at TIMESTAMPTZ,
    test_status VARCHAR(50) NOT NULL DEFAULT 'unknown',
    test_error TEXT
);

-- Lark 配置表
CREATE TABLE IF NOT EXISTS lark_configs (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    
    name VARCHAR(255) NOT NULL,
    webhook_url TEXT NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    description TEXT,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    last_test_at TIMESTAMPTZ,
    test_status VARCHAR(50) NOT NULL DEFAULT 'unknown',
    test_error TEXT
);

-- 规则表
CREATE TABLE IF NOT EXISTS rules (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    
    name VARCHAR(255) NOT NULL,
    index_pattern VARCHAR(512) NOT NULL,
    queries TEXT,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    interval INTEGER NOT NULL DEFAULT 60,
    es_config_id BIGINT REFERENCES es_configs(id) ON DELETE SET NULL,
    lark_webhook TEXT,
    lark_config_id BIGINT REFERENCES lark_configs(id) ON DELETE SET NULL,
    description TEXT,
    
    -- 统计字段
    last_run_time TIMESTAMPTZ,
    run_count BIGINT NOT NULL DEFAULT 0,
    alert_count BIGINT NOT NULL DEFAULT 0
);

-- 告警记录表
CREATE TABLE IF NOT EXISTS alerts (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    
    rule_id BIGINT NOT NULL REFERENCES rules(id) ON DELETE CASCADE,
    index_name VARCHAR(512) NOT NULL,
    log_count INTEGER NOT NULL DEFAULT 0,
    logs TEXT,
    time_range VARCHAR(100),
    status VARCHAR(50) NOT NULL DEFAULT 'sent',
    error_msg TEXT
);

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    
    username VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at TIMESTAMPTZ
);

-- 系统配置表
CREATE TABLE IF NOT EXISTS system_configs (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    
    key VARCHAR(255) NOT NULL,
    value TEXT,
    description TEXT
);

-- 创建索引
CREATE UNIQUE INDEX IF NOT EXISTS idx_es_configs_name ON es_configs(name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_es_configs_deleted_at ON es_configs(deleted_at);

CREATE UNIQUE INDEX IF NOT EXISTS idx_lark_configs_name ON lark_configs(name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_lark_configs_deleted_at ON lark_configs(deleted_at);

CREATE UNIQUE INDEX IF NOT EXISTS idx_rules_name ON rules(name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_rules_deleted_at ON rules(deleted_at);
CREATE INDEX IF NOT EXISTS idx_rules_es_config_id ON rules(es_config_id);
CREATE INDEX IF NOT EXISTS idx_rules_lark_config_id ON rules(lark_config_id);
CREATE INDEX IF NOT EXISTS idx_rules_enabled ON rules(enabled);

CREATE INDEX IF NOT EXISTS idx_alerts_deleted_at ON alerts(deleted_at);
CREATE INDEX IF NOT EXISTS idx_alerts_rule_id ON alerts(rule_id);
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at);
CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);
CREATE INDEX IF NOT EXISTS idx_alerts_rule_created ON alerts(rule_id, created_at);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email) WHERE deleted_at IS NULL AND email IS NOT NULL AND email != '';
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

CREATE UNIQUE INDEX IF NOT EXISTS idx_system_configs_key ON system_configs(key) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_system_configs_deleted_at ON system_configs(deleted_at);

-- 注意：默认管理员用户由后端代码自动创建（InitDefaultAdmin），
-- 默认账号：admin，默认密码：admin123
-- 可通过环境变量 ADMIN_USERNAME、ADMIN_PASSWORD、ADMIN_EMAIL 自定义

-- 插入默认清理任务配置
INSERT INTO system_configs (key, value, description)
SELECT 'cleanup_config', '{"enabled":false,"hour":3,"minute":0,"retention_days":30,"last_execution_status":"never"}', '告警数据自动清理配置'
WHERE NOT EXISTS (SELECT 1 FROM system_configs WHERE key = 'cleanup_config' AND deleted_at IS NULL);

