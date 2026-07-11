package identity

import (
	"github.com/gofiber/fiber/v2"

	"ingencore/api/internal/auth"
	"ingencore/api/internal/db"
)

// RequireAdmin gates team-management routes to users holding the org's
// "Admin" role. This is a coarse role-name check; it is superseded by the
// full privilege-depth model (user/BU/BU+children/org) planned next. Must be
// mounted under db.Middleware so db.Tx(c) resolves to the request's
// tenant-scoped transaction.
func RequireAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := auth.UserID(c)
		var isAdmin bool
		err := db.Tx(c).QueryRow(c.Context(),
			`SELECT EXISTS(
				SELECT 1 FROM user_roles ur
				JOIN roles r ON r.id = ur.role_id
				WHERE ur.user_id=$1 AND r.name='Admin'
			)`, userID,
		).Scan(&isAdmin)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not verify role")
		}
		if !isAdmin {
			return fiber.NewError(fiber.StatusForbidden, "admin role required")
		}
		return c.Next()
	}
}
