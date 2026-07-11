-- Postgres Row-Level Security (PRD §7.2.3, §6 tenancy NFR): a second,
-- database-enforced layer of tenant isolation on top of the existing
-- application-layer organization_id filters, so a forgotten WHERE clause
-- fails closed instead of leaking cross-tenant rows.
--
-- Important: RLS has no effect on a superuser or BYPASSRLS role, and the
-- 'user' role this project's docker-compose bootstraps IS a superuser (the
-- default behavior of the official postgres image). So the API's runtime
-- connection must run as a separate, unprivileged role — 'app_role' below —
-- while migrations continue to run as the superuser.

DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'app_role') THEN
    CREATE ROLE app_role LOGIN PASSWORD 'apppassword' NOSUPERUSER NOBYPASSRLS;
  END IF;
END
$$;

GRANT CONNECT ON DATABASE ingencore TO app_role;
GRANT USAGE ON SCHEMA public TO app_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_role;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_role;

-- Scope decision: organizations, users, sessions, and invites are
-- deliberately NOT RLS-protected. Each has a legitimate cross-tenant lookup
-- path that runs before an org context exists: login-by-email (organization
-- unknown until the row is found), invite-by-token and accept-invite (the
-- invitee has no session yet), and organizations itself (its own id IS the
-- tenant — there's no column to filter by). These remain protected by
-- existing application-layer organization_id checks only, exactly as today.

-- Tables with a direct organization_id column: a uniform policy.
DO $$
DECLARE
  t TEXT;
BEGIN
  FOREACH t IN ARRAY ARRAY[
    'business_units', 'roles', 'module_entitlements',
    'accounts', 'contacts', 'leads', 'queues', 'cases', 'messages',
    'sla_policies', 'kb_articles', 'workflows', 'workflow_runs', 'notifications'
  ]
  LOOP
    EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
    EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
    EXECUTE format('DROP POLICY IF EXISTS tenant_isolation ON %I', t);
    EXECUTE format(
      'CREATE POLICY tenant_isolation ON %I
         USING (organization_id = current_setting(''app.org_id'', true)::uuid)
         WITH CHECK (organization_id = current_setting(''app.org_id'', true)::uuid)',
      t
    );
  END LOOP;
END
$$;

-- Tables with no direct organization_id column: scope via their parent.
ALTER TABLE user_roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_roles FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON user_roles;
CREATE POLICY tenant_isolation ON user_roles
  USING (EXISTS (
    SELECT 1 FROM roles r WHERE r.id = user_roles.role_id
      AND r.organization_id = current_setting('app.org_id', true)::uuid
  ))
  WITH CHECK (EXISTS (
    SELECT 1 FROM roles r WHERE r.id = user_roles.role_id
      AND r.organization_id = current_setting('app.org_id', true)::uuid
  ));

ALTER TABLE role_privileges ENABLE ROW LEVEL SECURITY;
ALTER TABLE role_privileges FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON role_privileges;
CREATE POLICY tenant_isolation ON role_privileges
  USING (EXISTS (
    SELECT 1 FROM roles r WHERE r.id = role_privileges.role_id
      AND r.organization_id = current_setting('app.org_id', true)::uuid
  ))
  WITH CHECK (EXISTS (
    SELECT 1 FROM roles r WHERE r.id = role_privileges.role_id
      AND r.organization_id = current_setting('app.org_id', true)::uuid
  ));

ALTER TABLE queue_members ENABLE ROW LEVEL SECURITY;
ALTER TABLE queue_members FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON queue_members;
CREATE POLICY tenant_isolation ON queue_members
  USING (EXISTS (
    SELECT 1 FROM queues q WHERE q.id = queue_members.queue_id
      AND q.organization_id = current_setting('app.org_id', true)::uuid
  ))
  WITH CHECK (EXISTS (
    SELECT 1 FROM queues q WHERE q.id = queue_members.queue_id
      AND q.organization_id = current_setting('app.org_id', true)::uuid
  ));
