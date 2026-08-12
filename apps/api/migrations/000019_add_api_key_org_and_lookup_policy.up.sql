-- Add denormalized organization_id to api_keys so API-key authentication can
-- resolve the tenant without first knowing the tenant (chicken-and-egg with RLS).
-- Also add a permissive lookup policy that allows a connection to locate an
-- API key by its hash when the app.tenant_id has not yet been established.

ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS organization_id UUID;

UPDATE api_keys k
SET organization_id = w.organization_id
FROM workspaces w
WHERE k.workspace_id = w.id AND k.organization_id IS NULL;

ALTER TABLE api_keys ALTER COLUMN organization_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_api_keys_organization ON api_keys(organization_id);

-- Enable the API-key lookup path before the tenant id is known.  The
-- application sets app.lookup_key_hash to the SHA-256 hash it wants to fetch
-- and app.tenant_id to the nil UUID while validating the key; once validated,
-- app.tenant_id is set to the key's organization_id.
CREATE POLICY api_keys_lookup_by_hash ON api_keys
    USING (
        key_hash = COALESCE(current_setting('app.lookup_key_hash', true), '')
    );

-- Ensure new keys carry their organization_id automatically.
CREATE OR REPLACE FUNCTION api_keys_set_organization_id()
RETURNS TRIGGER AS $$
BEGIN
    SELECT organization_id INTO NEW.organization_id
    FROM workspaces WHERE id = NEW.workspace_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS api_keys_set_organization_id_trigger ON api_keys;
CREATE TRIGGER api_keys_set_organization_id_trigger
    BEFORE INSERT OR UPDATE OF workspace_id ON api_keys
    FOR EACH ROW
    EXECUTE FUNCTION api_keys_set_organization_id();
