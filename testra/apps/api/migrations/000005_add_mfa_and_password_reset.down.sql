DROP TABLE IF EXISTS password_reset_tokens;
ALTER TABLE users DROP COLUMN IF EXISTS mfa_secret;
ALTER TABLE users DROP COLUMN IF EXISTS mfa_enabled;
