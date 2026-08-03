DROP TRIGGER IF EXISTS api_keys_set_organization_id_trigger ON api_keys;
DROP FUNCTION IF EXISTS api_keys_set_organization_id();
DROP POLICY IF EXISTS api_keys_lookup_by_hash ON api_keys;
DROP INDEX IF EXISTS idx_api_keys_organization;
ALTER TABLE api_keys DROP COLUMN IF EXISTS organization_id;
