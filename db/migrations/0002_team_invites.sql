-- Team management: invite-based onboarding of additional users into an
-- existing organization, plus a default non-admin "Member" role.

CREATE TABLE IF NOT EXISTS invites (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  business_unit_id UUID REFERENCES business_units(id) ON DELETE SET NULL,
  email CITEXT NOT NULL,
  role_name TEXT NOT NULL DEFAULT 'Member',
  token TEXT UNIQUE NOT NULL,
  invited_by UUID REFERENCES users(id) ON DELETE SET NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  accepted_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS invites_org_pending_idx
  ON invites(organization_id)
  WHERE accepted_at IS NULL;

-- Backfill a default "Member" role for organizations created before this
-- migration, so existing orgs can invite non-admin teammates immediately.
INSERT INTO roles(organization_id, name)
SELECT id, 'Member' FROM organizations
ON CONFLICT (organization_id, name) DO NOTHING;
