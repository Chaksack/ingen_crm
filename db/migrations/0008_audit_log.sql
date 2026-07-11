-- Audit log (PRD §5.1/E8, and named twice in the §6 NFRs: "audit trail
-- immutable (append-only)"): entity-level auditing of writes with actor,
-- timestamp, and before/after values.
--
-- Scope decision: wired into the same four entities the privilege-depth
-- RBAC model already treats as "the" business records (account/contact/
-- lead/case — see identity.privilegedEntities). Admin-config tables (roles,
-- business units, workflows, queues, KB) are not audited yet; a reasonable
-- follow-up, not done here to keep this pass bounded.

CREATE TABLE IF NOT EXISTS audit_log (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  entity TEXT NOT NULL,
  entity_id UUID NOT NULL,
  action TEXT NOT NULL, -- 'create' | 'update' | 'delete'
  before JSONB,
  after JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS audit_log_org_created_idx ON audit_log(organization_id, created_at DESC);
CREATE INDEX IF NOT EXISTS audit_log_entity_idx ON audit_log(organization_id, entity, entity_id, created_at DESC);

-- Append-only: no UPDATE/DELETE grants, not even for app_role. Only INSERT
-- and SELECT are needed by the application; revoking the rest means a bug
-- (or a compromised app_role credential) still can't rewrite history.
ALTER TABLE audit_log ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_log FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON audit_log;
CREATE POLICY tenant_isolation ON audit_log
  USING (organization_id = current_setting('app.org_id', true)::uuid)
  WITH CHECK (organization_id = current_setting('app.org_id', true)::uuid);

REVOKE UPDATE, DELETE ON audit_log FROM app_role;
GRANT SELECT, INSERT ON audit_log TO app_role;
