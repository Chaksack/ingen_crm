package entitlement

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"ingen.one/api/internal/auth"
)

var AllModules = []string{"sales", "service", "collab", "finance", "scm", "projects", "hr", "marketing"}

func Enabled(ctx context.Context, pool *pgxpool.Pool, orgID, module string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM module_entitlements WHERE organization_id=$1 AND module=$2)`,
		orgID, module,
	).Scan(&exists)
	return exists, err
}

func List(ctx context.Context, pool *pgxpool.Pool, orgID string) ([]string, error) {
	rows, err := pool.Query(ctx, `SELECT module FROM module_entitlements WHERE organization_id=$1`, orgID)
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
// per the platform rule that disabled modules are invisible.
func Require(pool *pgxpool.Pool, module string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		orgID := auth.OrgID(c)
		ok, err := Enabled(c.Context(), pool, orgID, module)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "entitlement check failed")
		}
		if !ok {
			return fiber.NewError(fiber.StatusNotFound, "not found")
		}
		return c.Next()
	}
}
