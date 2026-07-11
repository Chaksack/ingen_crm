package entitlement

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"ingencore/api/internal/auth"
	"ingencore/api/internal/db"
)

var AllModules = []string{"sales", "service", "collab", "finance", "scm", "projects", "hr", "marketing"}

func Enabled(ctx context.Context, q db.Queryer, orgID, module string) (bool, error) {
	var exists bool
	err := q.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM module_entitlements WHERE organization_id=$1 AND module=$2)`,
		orgID, module,
	).Scan(&exists)
	return exists, err
}

func List(ctx context.Context, q db.Queryer, orgID string) ([]string, error) {
	rows, err := q.Query(ctx, `SELECT module FROM module_entitlements WHERE organization_id=$1`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var mods []string
	for rows.Next() {
		var m string
		if err := rows.Scan(&m); err != nil {
			return nil, err
		}
		mods = append(mods, m)
	}
	return mods, rows.Err()
}

// Require returns 404 (not 403) when the module is disabled for the tenant,
// per the platform rule that disabled modules are invisible. Must be mounted
// under db.Middleware so db.Tx(c) resolves to the request's tenant-scoped
// transaction.
func Require(module string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		orgID := auth.OrgID(c)
		ok, err := Enabled(c.Context(), db.Tx(c), orgID, module)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "entitlement check failed")
		}
		if !ok {
			return fiber.NewError(fiber.StatusNotFound, "not found")
		}
		return c.Next()
	}
}
