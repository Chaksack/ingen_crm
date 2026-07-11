package identity

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"ingencore/api/internal/auth"
	"ingencore/api/internal/db"
)

const inviteTTL = 7 * 24 * time.Hour

var assignableRoles = map[string]bool{"Admin": true, "Member": true}

func generateToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// ---- Invites ----

type createInviteRequest struct {
	Email          string  `json:"email"`
	Role           string  `json:"role"`
	BusinessUnitID *string `json:"business_unit_id"`
}

func (h *Handler) CreateInvite(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	inviterID := auth.UserID(c)

	var req createInviteRequest
	if err := c.BodyParser(&req); err != nil || req.Email == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email is required")
	}
	if req.Role == "" {
		req.Role = "Member"
	}
	if !assignableRoles[req.Role] {
		return fiber.NewError(fiber.StatusBadRequest, "role must be Admin or Member")
	}

	ctx := c.Context()

	if req.BusinessUnitID != nil {
		ok, err := buExists(ctx, db.Tx(c), *req.BusinessUnitID, orgID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not verify business unit")
		}
		if !ok {
			return fiber.NewError(fiber.StatusBadRequest, "business unit not found")
		}
	}

	var existingUser bool
	if err := db.Tx(c).QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE organization_id=$1 AND email=$2)`, orgID, req.Email,
	).Scan(&existingUser); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not check existing users")
	}
	if existingUser {
		return fiber.NewError(fiber.StatusConflict, "that email is already a member of this organization")
	}

	token, err := generateToken()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not generate invite token")
	}

	// Replace any prior pending invite for this email so re-inviting issues a fresh link.
	if _, err := db.Tx(c).Exec(ctx,
		`DELETE FROM invites WHERE organization_id=$1 AND email=$2 AND accepted_at IS NULL`, orgID, req.Email,
	); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not clear prior invite")
	}

	var id string
	expiresAt := time.Now().Add(inviteTTL)
	err = db.Tx(c).QueryRow(ctx,
		`INSERT INTO invites(organization_id, business_unit_id, email, role_name, token, invited_by, expires_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		orgID, req.BusinessUnitID, req.Email, req.Role, token, inviterID, expiresAt,
	).Scan(&id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create invite")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":         id,
		"email":      req.Email,
		"role":       req.Role,
		"token":      token,
		"expires_at": expiresAt.Format(time.RFC3339),
	})
}

func (h *Handler) ListInvites(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	rows, err := db.Tx(c).Query(c.Context(),
		`SELECT id, email, role_name, expires_at, created_at
		 FROM invites WHERE organization_id=$1 AND accepted_at IS NULL ORDER BY created_at DESC`, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list invites")
	}
	defer rows.Close()
	out := []fiber.Map{}
	for rows.Next() {
		var id, email, role string
		var expiresAt, createdAt time.Time
		if err := rows.Scan(&id, &email, &role, &expiresAt, &createdAt); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not scan invite")
		}
		out = append(out, fiber.Map{
			"id": id, "email": email, "role": role,
			"expires_at": expiresAt.Format(time.RFC3339),
			"created_at": createdAt.Format(time.RFC3339),
			"expired":    expiresAt.Before(time.Now()),
		})
	}
	return c.JSON(out)
}

func (h *Handler) RevokeInvite(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	cmd, err := db.Tx(c).Exec(c.Context(),
		`DELETE FROM invites WHERE id=$1 AND organization_id=$2 AND accepted_at IS NULL`, id, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not revoke invite")
	}
	if cmd.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "invite not found")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// GetInviteByToken is unauthenticated: it's how the accept-invite page
// greets someone who isn't a user yet.
func (h *Handler) GetInviteByToken(c *fiber.Ctx) error {
	token := c.Params("token")
	var email, orgName string
	var expiresAt time.Time
	var acceptedAt *time.Time
	err := h.pool.QueryRow(c.Context(),
		`SELECT i.email, o.name, i.expires_at, i.accepted_at
		 FROM invites i JOIN organizations o ON o.id = i.organization_id
		 WHERE i.token=$1`, token,
	).Scan(&email, &orgName, &expiresAt, &acceptedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "invite not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "could not load invite")
	}
	if acceptedAt != nil || expiresAt.Before(time.Now()) {
		return fiber.NewError(fiber.StatusNotFound, "invite not found")
	}
	return c.JSON(fiber.Map{"email": email, "organization_name": orgName})
}

type acceptInviteRequest struct {
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

func (h *Handler) AcceptInvite(c *fiber.Ctx) error {
	token := c.Params("token")
	var req acceptInviteRequest
	if err := c.BodyParser(&req); err != nil || req.DisplayName == "" || req.Password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "display_name and password are required")
	}
	if len(req.Password) < 8 {
		return fiber.NewError(fiber.StatusBadRequest, "password must be at least 8 characters")
	}

	ctx := c.Context()
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not start transaction")
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var inviteID, orgID, email, roleName string
	var buID *string
	var expiresAt time.Time
	var acceptedAt *time.Time
	err = tx.QueryRow(ctx,
		`SELECT id, organization_id, business_unit_id, email, role_name, expires_at, accepted_at
		 FROM invites WHERE token=$1 FOR UPDATE`, token,
	).Scan(&inviteID, &orgID, &buID, &email, &roleName, &expiresAt, &acceptedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "invite not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "could not load invite")
	}
	if acceptedAt != nil || expiresAt.Before(time.Now()) {
		return fiber.NewError(fiber.StatusNotFound, "invite not found")
	}
	// roles/user_roles are RLS-protected; now that the invite tells us which
	// org this is, set the session's tenant context before touching them.
	if err := db.SetOrgContext(ctx, tx, orgID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not set tenant context")
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not hash password")
	}

	var userID string
	err = tx.QueryRow(ctx,
		`INSERT INTO users(organization_id, business_unit_id, email, password_hash, display_name)
		 VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		orgID, buID, email, string(hashBytes), req.DisplayName,
	).Scan(&userID)
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, "email already registered or invalid")
	}

	var roleID string
	if err := tx.QueryRow(ctx,
		`SELECT id FROM roles WHERE organization_id=$1 AND name=$2`, orgID, roleName,
	).Scan(&roleID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "invite role no longer exists")
	}
	if _, err := tx.Exec(ctx, `INSERT INTO user_roles(user_id, role_id) VALUES ($1,$2)`, userID, roleID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not assign role")
	}
	if _, err := tx.Exec(ctx, `UPDATE invites SET accepted_at=now() WHERE id=$1`, inviteID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not mark invite accepted")
	}

	if err := tx.Commit(ctx); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not commit invite acceptance")
	}

	buVal := ""
	if buID != nil {
		buVal = *buID
	}
	return h.issueSession(c, userID, orgID, buVal, email)
}

// ---- Members ----

func (h *Handler) ListMembers(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	rows, err := db.Tx(c).Query(c.Context(),
		`SELECT u.id, u.email, u.display_name, u.is_active, u.created_at, u.business_unit_id, bu.name,
		        COALESCE((SELECT r.name FROM roles r JOIN user_roles ur ON ur.role_id=r.id WHERE ur.user_id=u.id AND r.organization_id=u.organization_id LIMIT 1), 'Member') AS role
		 FROM users u LEFT JOIN business_units bu ON bu.id = u.business_unit_id
		 WHERE u.organization_id=$1 ORDER BY u.created_at`, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list members")
	}
	defer rows.Close()
	out := []fiber.Map{}
	for rows.Next() {
		var id, email, displayName, role string
		var isActive bool
		var createdAt time.Time
		var businessUnitID, businessUnitName *string
		if err := rows.Scan(&id, &email, &displayName, &isActive, &createdAt, &businessUnitID, &businessUnitName, &role); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not scan member")
		}
		out = append(out, fiber.Map{
			"id": id, "email": email, "display_name": displayName,
			"is_active": isActive, "role": role, "created_at": createdAt.Format(time.RFC3339),
			"business_unit_id": businessUnitID, "business_unit_name": businessUnitName,
		})
	}
	return c.JSON(out)
}

type updateMemberRequest struct {
	Role           *string `json:"role"`
	IsActive       *bool   `json:"is_active"`
	BusinessUnitID *string `json:"business_unit_id"`
}

func (h *Handler) UpdateMember(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	targetID := c.Params("id")

	var req updateMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	ctx := c.Context()
	tx := db.Tx(c)

	var exists bool
	if err := tx.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE id=$1 AND organization_id=$2)`, targetID, orgID,
	).Scan(&exists); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not load member")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "member not found")
	}

	if req.IsActive != nil {
		if _, err := tx.Exec(ctx, `UPDATE users SET is_active=$1 WHERE id=$2`, *req.IsActive, targetID); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not update member status")
		}
	}
	if req.Role != nil {
		var roleID string
		if err := tx.QueryRow(ctx,
			`SELECT id FROM roles WHERE organization_id=$1 AND name=$2`, orgID, *req.Role,
		).Scan(&roleID); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "no role with that name exists in this organization")
		}
		if _, err := tx.Exec(ctx,
			`DELETE FROM user_roles WHERE user_id=$1 AND role_id IN (SELECT id FROM roles WHERE organization_id=$2)`,
			targetID, orgID,
		); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not clear existing role")
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO user_roles(user_id, role_id) VALUES ($1,$2)`, targetID, roleID,
		); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not assign role")
		}
	}
	if req.BusinessUnitID != nil {
		var buParam any
		if *req.BusinessUnitID != "" {
			ok, err := buExists(ctx, tx, *req.BusinessUnitID, orgID)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "could not verify business unit")
			}
			if !ok {
				return fiber.NewError(fiber.StatusBadRequest, "business unit not found")
			}
			buParam = *req.BusinessUnitID
		}
		if _, err := tx.Exec(ctx, `UPDATE users SET business_unit_id=$1 WHERE id=$2`, buParam, targetID); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not update business unit")
		}
	}

	// Guard rail: never leave an org with zero active Admins.
	var activeAdmins int
	if err := tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM users u
		 JOIN user_roles ur ON ur.user_id = u.id
		 JOIN roles r ON r.id = ur.role_id
		 WHERE u.organization_id=$1 AND u.is_active AND r.name='Admin'`, orgID,
	).Scan(&activeAdmins); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not verify admin count")
	}
	if activeAdmins == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "organization must retain at least one active Admin")
	}

	return c.JSON(fiber.Map{"id": targetID})
}
