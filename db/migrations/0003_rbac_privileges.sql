-- Privilege-depth RBAC (PRD §5.1): per-entity privileges (create/read/write/
-- delete/assign) granted at a depth (none/user/bu/bu_children/org). A user's
-- effective depth for an entity+action is the most permissive depth across
-- all roles they hold.

CREATE TYPE privilege_action AS ENUM ('create', 'read', 'write', 'delete', 'assign');
CREATE TYPE privilege_depth AS ENUM ('none', 'user', 'bu', 'bu_children', 'org');

CREATE TABLE IF NOT EXISTS role_privileges (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  entity TEXT NOT NULL,
  action privilege_action NOT NULL,
  depth privilege_depth NOT NULL DEFAULT 'none',
  UNIQUE(role_id, entity, action)
);

CREATE INDEX IF NOT EXISTS role_privileges_role_idx ON role_privileges(role_id);

-- Default grants for the two built-in roles, across the entities that carry
-- an owner_user_id today. Admin gets org depth on everything. Member gets
-- org-depth create/read (can see and add records) but only user-depth
-- write/delete/assign (can only change what they own). This backfills
-- existing organizations and is also the shape new orgs are seeded with.
INSERT INTO role_privileges (role_id, entity, action, depth)
SELECT r.id, e.entity, a.action::privilege_action,
  CASE
    WHEN r.name = 'Admin' THEN 'org'::privilege_depth
    WHEN a.action IN ('create', 'read') THEN 'org'::privilege_depth
    ELSE 'user'::privilege_depth
  END
FROM roles r
CROSS JOIN (VALUES ('account'), ('contact'), ('lead'), ('case')) AS e(entity)
CROSS JOIN (VALUES ('create'), ('read'), ('write'), ('delete'), ('assign')) AS a(action)
WHERE r.name IN ('Admin', 'Member')
ON CONFLICT (role_id, entity, action) DO NOTHING;
