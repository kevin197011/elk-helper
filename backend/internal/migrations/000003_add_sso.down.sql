DROP INDEX IF EXISTS idx_users_auth_source;
ALTER TABLE users DROP COLUMN IF EXISTS auth_source;
ALTER TABLE users ALTER COLUMN password SET NOT NULL;

DROP INDEX IF EXISTS idx_sso_providers_deleted_at;
DROP INDEX IF EXISTS idx_sso_providers_enabled;
DROP TABLE IF EXISTS sso_providers;
