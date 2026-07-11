package identity

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"

	"ingencore/api/internal/auth"
	"ingencore/api/internal/db"
)

var allActions = []string{"create", "read", "write", "delete", "assign"}
var validDepths = map[string]bool{"none": true, "user": true, "bu": true, "bu_children": true, "org": true}

func isPrivilegedEntity(entity string) bool {
	for _, e := range privilegedEntities {
		if e == entity {
			return true
		}
	}
	return false
}

func isValidAction(action string) bool {
	for _, a := range allActions {
		if a == action {
			return true
		}
	}
	return false
}

func (h *Handler) ListRoles(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	rows, err := db.Tx(c).Query(c.Context(), `
		SELECT r.id, r.name, COUNT(ur.user_id)
		FROM roles r
		LEFT JOIN user_roles ur ON ur.role_id = r.id
		WHERE r.organization_id=$1
		GROUP BY r.id, r.name
		ORDER BY r.name`, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list roles")
	}
	defer rows.Close()
	out := []fiber.Map{}
	for rows.Next() {
		var id, name string
		var memberCount int
		if err := rows.Scan(&id, &name, &memberCount); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not scan role")
		}
		out = append(out, fiber.Map{
			"id": id, "name": name, "member_count": memberCount,
			"builtin": name == "Admin" || name == "Member",
		})
	}
	return c.JSON(out)
}

type roleRequest struct {
	Name string `json:"name"`
}

func (h *Handler) CreateRole(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	var req roleRequest
	if err := c.BodyParser(&req); err != nil || req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}

	ctx := c.Context()
	tx := db.Tx(c)

	var id string
	if err := tx.QueryRow(ctx,
		`INSERT INTO roles(organization_id, name) VALUES ($1,$2) RETURNING id`, orgID, req.Name,
	).Scan(&id); err != nil {
		return fiber.NewError(fiber.StatusConflict, "a role with that name already exists")
	}
	// New custom roles start with no access on anything; an Admin dials up
	// exactly what's needed from the privilege grid.
	if _, err := tx.Exec(ctx, `
		INSERT INTO role_privileges (role_id, entity, action, depth)
		SELECT $1, e.entity, a.action::privilege_action, 'none'
		FROM (SELECT unnest($2::text[])) AS e(entity)
		CROSS JOIN (VALUES ('create'), ('read'), ('write'), ('delete'), ('assign')) AS a(action)`,
		id, privilegedEntities,
	); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not seed role privileges")
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id, "name": req.Name, "member_count": 0, "builtin": false})
}

func (h *Handler) DeleteRole(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")

	var name string
	err := db.Tx(c).QueryRow(c.Context(), `SELECT name FROM roles WHERE id=$1 AND organization_id=$2`, id, orgID).Scan(&name)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "role not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "could not load role")
	}
	if name == "Admin" || name == "Member" {
		return fiber.NewError(fiber.StatusBadRequest, "built-in roles cannot be deleted")
	}

	var memberCount int
	if err := db.Tx(c).QueryRow(c.Context(), `SELECT COUNT(*) FROM user_roles WHERE role_id=$1`, id).Scan(&memberCount); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not check role membership")
	}
	if memberCount > 0 {
		return fiber.NewError(fiber.StatusBadRequest, "reassign members off this role before deleting it")
	}

	cmd, err := db.Tx(c).Exec(c.Context(), `DELETE FROM roles WHERE id=$1 AND organization_id=$2`, id, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete role")
	}
	if cmd.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "role not found")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) GetRolePrivileges(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	roleID := c.Params("id")

	var exists bool
	if err := db.Tx(c).QueryRow(c.Context(),
		`SELECT EXISTS(SELECT 1 FROM roles WHERE id=$1 AND organization_id=$2)`, roleID, orgID,
	).Scan(&exists); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not verify role")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "role not found")
	}

	rows, err := db.Tx(c).Query(c.Context(),
		`SELECT entity, action, depth FROM role_privileges WHERE role_id=$1`, roleID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not load privileges")
	}
	defer rows.Close()
	existing := map[string]string{}
	for rows.Next() {
		var entity, action, depth string
		if err := rows.Scan(&entity, &action, &depth); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not scan privilege")
		}
		existing[entity+":"+action] = depth
	}

	out := []fiber.Map{}
	for _, entity := range privilegedEntities {
		for _, action := range allActions {
			depth := existing[entity+":"+action]
			if depth == "" {
				depth = "none"
			}
			out = append(out, fiber.Map{"entity": entity, "action": action, "depth": depth})
		}
	}
	return c.JSON(out)
}

type privilegeEntry struct {
	Entity string `json:"entity"`
	Action string `json:"action"`
	Depth  string `json:"depth"`
}

func (h *Handler) UpdateRolePrivileges(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	roleID := c.Params("id")

	var exists bool
	if err := db.Tx(c).QueryRow(c.Context(),
		`SELECT EXISTS(SELECT 1 FROM roles WHERE id=$1 AND organization_id=$2)`, roleID, orgID,
	).Scan(&exists); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not verify role")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "role not found")
	}

	var entries []privilegeEntry
	if err := c.BodyParser(&entries); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	for _, e := range entries {
		if !isPrivilegedEntity(e.Entity) || !isValidAction(e.Action) || !validDepths[e.Depth] {
			return fiber.NewError(fiber.StatusBadRequest, "invalid privilege entry: "+e.Entity+"/"+e.Action+"/"+e.Depth)
		}
	}

	ctx := c.Context()
	tx := db.Tx(c)
	for _, e := range entries {
		if _, err := tx.Exec(ctx, `
			INSERT INTO role_privileges (role_id, entity, action, depth)
			VALUES ($1,$2,$3::privilege_action,$4::privilege_depth)
			ON CONFLICT (role_id, entity, action) DO UPDATE SET depth=EXCLUDED.depth`,
			roleID, e.Entity, e.Action, e.Depth,
		); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not update privilege")
		}
	}
	return c.SendStatus(fiber.StatusNoContent)
}
