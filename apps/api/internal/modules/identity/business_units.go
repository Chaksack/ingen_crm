package identity

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"ingencore/api/internal/auth"
	"ingencore/api/internal/db"
)

func (h *Handler) ListBusinessUnits(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	rows, err := db.Tx(c).Query(c.Context(),
		`SELECT id, name, parent_id FROM business_units WHERE organization_id=$1 ORDER BY name`, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list business units")
	}
	defer rows.Close()
	out := []fiber.Map{}
	for rows.Next() {
		var id, name string
		var parentID *string
		if err := rows.Scan(&id, &name, &parentID); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not scan business unit")
		}
		out = append(out, fiber.Map{"id": id, "name": name, "parent_id": parentID})
	}
	return c.JSON(out)
}

type businessUnitRequest struct {
	Name     string  `json:"name"`
	ParentID *string `json:"parent_id"`
}

func (h *Handler) CreateBusinessUnit(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	var req businessUnitRequest
	if err := c.BodyParser(&req); err != nil || req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	if req.ParentID != nil {
		ok, err := buExists(c.Context(), db.Tx(c), *req.ParentID, orgID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not verify parent business unit")
		}
		if !ok {
			return fiber.NewError(fiber.StatusBadRequest, "parent business unit not found")
		}
	}
	var id string
	err := db.Tx(c).QueryRow(c.Context(),
		`INSERT INTO business_units(organization_id, name, parent_id) VALUES ($1,$2,$3) RETURNING id`,
		orgID, req.Name, req.ParentID,
	).Scan(&id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create business unit")
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id, "name": req.Name, "parent_id": req.ParentID})
}

func (h *Handler) UpdateBusinessUnit(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	var req businessUnitRequest
	if err := c.BodyParser(&req); err != nil || req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	if req.ParentID != nil {
		if *req.ParentID == id {
			return fiber.NewError(fiber.StatusBadRequest, "a business unit cannot be its own parent")
		}
		ok, err := buExists(c.Context(), db.Tx(c), *req.ParentID, orgID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not verify parent business unit")
		}
		if !ok {
			return fiber.NewError(fiber.StatusBadRequest, "parent business unit not found")
		}
		isDescendant, err := buIsDescendantOrSelf(c.Context(), db.Tx(c), *req.ParentID, id)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not verify business unit hierarchy")
		}
		if isDescendant {
			return fiber.NewError(fiber.StatusBadRequest, "cannot move a business unit under its own descendant")
		}
	}
	cmd, err := db.Tx(c).Exec(c.Context(),
		`UPDATE business_units SET name=$1, parent_id=$2 WHERE id=$3 AND organization_id=$4`,
		req.Name, req.ParentID, id, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not update business unit")
	}
	if cmd.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "business unit not found")
	}
	return c.JSON(fiber.Map{"id": id, "name": req.Name, "parent_id": req.ParentID})
}

func (h *Handler) DeleteBusinessUnit(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")

	var parentID *string
	err := db.Tx(c).QueryRow(c.Context(),
		`SELECT parent_id FROM business_units WHERE id=$1 AND organization_id=$2`, id, orgID,
	).Scan(&parentID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "business unit not found")
	}
	if parentID == nil {
		return fiber.NewError(fiber.StatusBadRequest, "the root business unit cannot be deleted")
	}

	var userCount, childCount int
	if err := db.Tx(c).QueryRow(c.Context(), `SELECT COUNT(*) FROM users WHERE business_unit_id=$1`, id).Scan(&userCount); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not check business unit members")
	}
	if err := db.Tx(c).QueryRow(c.Context(), `SELECT COUNT(*) FROM business_units WHERE parent_id=$1`, id).Scan(&childCount); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not check child business units")
	}
	if userCount > 0 || childCount > 0 {
		return fiber.NewError(fiber.StatusBadRequest, "reassign members and child business units before deleting this one")
	}

	cmd, err := db.Tx(c).Exec(c.Context(), `DELETE FROM business_units WHERE id=$1 AND organization_id=$2`, id, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete business unit")
	}
	if cmd.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "business unit not found")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func buExists(ctx context.Context, q db.Queryer, id, orgID string) (bool, error) {
	var exists bool
	err := q.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM business_units WHERE id=$1 AND organization_id=$2)`, id, orgID,
	).Scan(&exists)
	return exists, err
}

// buIsDescendantOrSelf reports whether candidateID is ancestorID itself or a
// descendant of it — used to block a reparent that would create a cycle.
func buIsDescendantOrSelf(ctx context.Context, q db.Queryer, candidateID, ancestorID string) (bool, error) {
	var isDescendant bool
	err := q.QueryRow(ctx, `
		WITH RECURSIVE descendants AS (
			SELECT id FROM business_units WHERE id = $1
			UNION ALL
			SELECT bu.id FROM business_units bu JOIN descendants d ON bu.parent_id = d.id
		)
		SELECT EXISTS(SELECT 1 FROM descendants WHERE id = $2)`,
		ancestorID, candidateID,
	).Scan(&isDescendant)
	return isDescendant, err
}
