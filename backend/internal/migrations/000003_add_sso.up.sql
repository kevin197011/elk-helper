-- SSO providers and user auth source

CREATE TABLE IF NOT EXISTS sso_providers (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,

    type VARCHAR(32) NOT NULL DEFAULT 'oidc',
    name VARCHAR(255) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    config_json TEXT NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_sso_providers_enabled ON sso_providers(enabled) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_sso_providers_deleted_at ON sso_providers(deleted_at);

ALTER TABLE users ADD COLUMN IF NOT EXISTS auth_source VARCHAR(64) NOT NULL DEFAULT 'local';
ALTER TABLE users ALTER COLUMN password DROP NOT NULL;

CREATE INDEX IF NOT EXISTS idx_users_auth_source ON users(auth_source) WHERE deleted_at IS NULL;
