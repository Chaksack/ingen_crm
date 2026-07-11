package identity

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// privilegedEntities are the entities that carry an owner_user_id today and
// therefore participate in the privilege-depth model (PRD §5.1).
var privilegedEntities = []string{"account", "contact", "lead", "case"}

// seedRolePrivileges grants a role's default privilege set: Admins get org
// depth on every action; everyone else gets org-depth create/read (can see
// and add records) but only user-depth write/delete/assign (can only touch
// what they own). This mirrors db/migrations/0003_rbac_privileges.sql, which
// backfills the same defaults for organizations created before this existed.
func seedRolePrivileges(ctx context.Context, tx pgx.Tx, roleID string, isAdmin bool) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO role_privileges (role_id, entity, action, depth)
		SELECT $1, e.entity, a.action::privilege_action,
			CASE
				WHEN $2::boolean THEN 'org'::privilege_depth
				WHEN a.action IN ('create', 'read') THEN 'org'::privilege_depth
				ELSE 'user'::privilege_depth
			END
		FROM (SELECT unnest($3::text[])) AS e(entity)
		CROSS JOIN (VALUES ('create'), ('read'), ('write'), ('delete'), ('assign')) AS a(action)
		ON CONFLICT (role_id, entity, action) DO NOTHING`,
		roleID, isAdmin, privilegedEntities,
	)
	return err
}
