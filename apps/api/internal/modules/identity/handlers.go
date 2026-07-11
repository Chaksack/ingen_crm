package identity

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"ingencore/api/internal/auth"
	"ingencore/api/internal/db"
	"ingencore/api/internal/entitlement"
)

type Handler struct {
	pool   *pgxpool.Pool
	issuer *auth.TokenIssuer
}

func NewHandler(pool *pgxpool.Pool, issuer *auth.TokenIssuer) *Handler {
	return &Handler{pool: pool, issuer: issuer}
}

type registerRequest struct {
	OrganizationName string `json:"organization_name"`
	DisplayName      string `json:"display_name"`
	Email            string `json:"email"`
	Password         string `json:"password"`
}

// defaultEntitlements are granted to a newly self-registered tenant so a
// fresh signup has an immediately usable Phase 1 experience.
var defaultEntitlements = []string{"sales", "service", "collab"}

func (h *Handler) Register(c *fiber.Ctx) error {
	var req registerRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if req.OrganizationName == "" || req.Email == "" || req.Password == "" || req.DisplayName == "" {
		return fiber.NewError(fiber.StatusBadRequest, "organization_name, display_name, email and password are required")
	}
	if len(req.Password) < 8 {
		return fiber.NewError(fiber.StatusBadRequest, "password must be at least 8 characters")
	}

	ctx := c.Context()
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not hash password")
	}

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not start transaction")
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var orgID, buID, userID, roleID string
	if err := tx.QueryRow(ctx, `INSERT INTO organizations(name) VALUES ($1) RETURNING id`, req.OrganizationName).Scan(&orgID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create organization")
	}
	// Every table touched below is RLS-protected; set the session's tenant
	// context now that the org exists, before writing anything else to it.
	if err := db.SetOrgContext(ctx, tx, orgID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not set tenant context")
	}
	if err := tx.QueryRow(ctx, `INSERT INTO business_units(organization_id, name) VALUES ($1,'HQ') RETURNING id`, orgID).Scan(&buID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create business unit")
	}
	err = tx.QueryRow(ctx,
		`INSERT INTO users(organization_id, business_unit_id, email, password_hash, display_name)
		 VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		orgID, buID, req.Email, string(hash), req.DisplayName,
	).Scan(&userID)
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, "email already registered or invalid")
	}
	if err := tx.QueryRow(ctx, `INSERT INTO roles(organization_id, name) VALUES ($1,'Admin') RETURNING id`, orgID).Scan(&roleID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create admin role")
	}
	if _, err := tx.Exec(ctx, `INSERT INTO user_roles(user_id, role_id) VALUES ($1,$2)`, userID, roleID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not assign admin role")
	}
	if err := seedRolePrivileges(ctx, tx, roleID, true); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not seed admin privileges")
	}
	// Member is the default role for invited teammates; create it up front
	// so the first invite doesn't need special-case role provisioning.
	var memberRoleID string
	if err := tx.QueryRow(ctx, `INSERT INTO roles(organization_id, name) VALUES ($1,'Member') RETURNING id`, orgID).Scan(&memberRoleID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create member role")
	}
	if err := seedRolePrivileges(ctx, tx, memberRoleID, false); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not seed member privileges")
	}
	for _, m := range defaultEntitlements {
		if _, err := tx.Exec(ctx, `INSERT INTO module_entitlements(organization_id, module) VALUES ($1,$2)`, orgID, m); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not set entitlements")
		}
	}
	// Default SLA policy (60m first response / 480m resolution); Admins can
	// tune it from Service settings once entitled.
	if _, err := tx.Exec(ctx,
		`INSERT INTO sla_policies(organization_id, first_response_minutes, resolution_minutes) VALUES ($1,60,480)`, orgID,
	); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create default SLA policy")
	}
	if err := tx.Commit(ctx); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not commit registration")
	}

	return h.issueSession(c, userID, orgID, buID, req.Email)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	ctx := c.Context()
	var userID, orgID string
	var buID *string
	var hash string
	err := h.pool.QueryRow(ctx,
		`SELECT id, organization_id, business_unit_id, password_hash FROM users WHERE email=$1 AND is_active`,
		req.Email,
	).Scan(&userID, &orgID, &buID, &hash)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid email or password")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "login failed")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid email or password")
	}

	buVal := ""
	if buID != nil {
		buVal = *buID
	}
	return h.issueSession(c, userID, orgID, buVal, req.Email)
}

func (h *Handler) issueSession(c *fiber.Ctx, userID, orgID, buID, email string) error {
	ctx := c.Context()
	token, expiresAt, err := h.issuer.Issue(userID, orgID, buID, email)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not issue token")
	}
	if _, err := h.pool.Exec(ctx,
		`INSERT INTO sessions(user_id, expires_at) VALUES ($1,$2)`, userID, expiresAt,
	); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not record session")
	}
	return c.JSON(fiber.Map{
		"access_token": token,
		"expires_at":   expiresAt.Format(time.RFC3339),
		"user": fiber.Map{
			"id":               userID,
			"email":            email,
			"organization_id":  orgID,
			"business_unit_id": buID,
		},
	})
}

func (h *Handler) Me(c *fiber.Ctx) error {
	ctx := c.Context()
	userID := auth.UserID(c)

	var email, displayName, orgID string
	var buID *string
	err := db.Tx(c).QueryRow(ctx,
		`SELECT email, display_name, organization_id, business_unit_id FROM users WHERE id=$1`,
		userID,
	).Scan(&email, &displayName, &orgID, &buID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	}

	rows, err := db.Tx(c).Query(ctx,
		`SELECT r.name FROM roles r JOIN user_roles ur ON ur.role_id = r.id WHERE ur.user_id=$1`, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not load roles")
	}
	defer rows.Close()
	var roles []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not load roles")
		}
		roles = append(roles, name)
	}

	return c.JSON(fiber.Map{
		"id":               userID,
		"email":            email,
		"display_name":     displayName,
		"organization_id":  orgID,
		"business_unit_id": buID,
		"roles":            roles,
	})
}

type navItem struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Path  string `json:"path"`
	Icon  string `json:"icon"`
}

var moduleNav = map[string]navItem{
	"sales":     {Key: "sales", Label: "Sales", Path: "/sales", Icon: "lucide:handshake"},
	"service":   {Key: "service", Label: "Service", Path: "/service", Icon: "lucide:life-buoy"},
	"collab":    {Key: "collab", Label: "Chat", Path: "/chat", Icon: "lucide:message-circle"},
	"finance":   {Key: "finance", Label: "Finance", Path: "/finance", Icon: "lucide:banknote"},
	"scm":       {Key: "scm", Label: "Supply Chain", Path: "/scm", Icon: "lucide:package"},
	"projects":  {Key: "projects", Label: "Projects", Path: "/projects", Icon: "lucide:kanban"},
	"hr":        {Key: "hr", Label: "HR", Path: "/hr", Icon: "lucide:users"},
	"marketing": {Key: "marketing", Label: "Marketing", Path: "/marketing", Icon: "lucide:megaphone"},
}

// Manifest returns the effective set of modules, navigation, and dashboard
// widgets for the caller's tenant, per PRD §7.4 (module entitlement architecture).
func (h *Handler) Manifest(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	mods, err := entitlement.List(c.Context(), db.Tx(c), orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not load manifest")
	}
	nav := make([]navItem, 0, len(mods))
	for _, m := range mods {
		if item, ok := moduleNav[m]; ok {
			nav = append(nav, item)
		}
	}
	return c.JSON(fiber.Map{
		"modules": mods,
		"nav":     nav,
	})
}
