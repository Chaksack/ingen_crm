package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

const (
	LocalUserID = "userID"
	LocalOrgID  = "orgID"
	LocalBUID   = "buID"
	LocalEmail  = "email"
)

// Middleware validates the Bearer JWT and stores tenant-scoping context in fiber.Locals.
func Middleware(issuer *TokenIssuer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		token := ""
		if strings.HasPrefix(header, "Bearer ") {
			token = strings.TrimPrefix(header, "Bearer ")
		} else {
			token = c.Query("token")
		}
		if token == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing bearer token")
		}
		claims, err := issuer.Parse(token)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired token")
		}
		c.Locals(LocalUserID, claims.UserID)
		c.Locals(LocalOrgID, claims.OrganizationID)
		c.Locals(LocalBUID, claims.BusinessUnitID)
		c.Locals(LocalEmail, claims.Email)
		return c.Next()
	}
}

func UserID(c *fiber.Ctx) string {
	v, _ := c.Locals(LocalUserID).(string)
	return v
}

func OrgID(c *fiber.Ctx) string {
	v, _ := c.Locals(LocalOrgID).(string)
	return v
}
