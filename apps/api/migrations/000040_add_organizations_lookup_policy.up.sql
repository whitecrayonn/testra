-- Add lookup policy for organizations so the GET /organizations endpoint
-- can list all orgs the authenticated user is a member of.
-- The existing org_tenant_isolation policy only allows seeing the single
-- org set in app.tenant_id, but the list endpoint needs to see all of the
-- user's orgs via app.lookup_user_id.

DROP POLICY IF EXISTS organizations_lookup_user ON organizations;
CREATE POLICY organizations_lookup_user ON organizations
    USING (id IN (
        SELECT organization_id FROM organization_members
        WHERE user_id = NULLIF(current_setting('app.lookup_user_id', true), '')::uuid
    ));
