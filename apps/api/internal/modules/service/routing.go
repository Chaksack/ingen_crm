package service

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"ingencore/api/internal/auth"
	"ingencore/api/internal/db"
)

var validRoutingStrategies = map[string]bool{"manual": true, "round_robin": true, "capacity": true}

type queueRoutingRequest struct {
	RoutingStrategy string `json:"routing_strategy"`
}

func (h *Handler) UpdateQueueRouting(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	var req queueRoutingRequest
	if err := c.BodyParser(&req); err != nil || !validRoutingStrategies[req.RoutingStrategy] {
		return fiber.NewError(fiber.StatusBadRequest, "routing_strategy must be manual, round_robin, or capacity")
	}
	cmd, err := db.Tx(c).Exec(c.Context(),
		`UPDATE queues SET routing_strategy=$1 WHERE id=$2 AND organization_id=$3`,
		req.RoutingStrategy, id, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not update queue routing")
	}
	if cmd.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "queue not found")
	}
	return c.JSON(fiber.Map{"id": id, "routing_strategy": req.RoutingStrategy})
}

func (h *Handler) ListQueueMembers(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	queueID := c.Params("id")
	rows, err := db.Tx(c).Query(c.Context(), `
		SELECT u.id, u.display_name, u.email
		FROM queue_members qm
		JOIN users u ON u.id = qm.user_id
		JOIN queues q ON q.id = qm.queue_id
		WHERE qm.queue_id=$1 AND q.organization_id=$2
		ORDER BY u.display_name`, queueID, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list queue members")
	}
	defer rows.Close()
	out := []fiber.Map{}
	for rows.Next() {
		var id, displayName, email string
		if err := rows.Scan(&id, &displayName, &email); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not scan queue member")
		}
		out = append(out, fiber.Map{"id": id, "display_name": displayName, "email": email})
	}
	return c.JSON(out)
}

type queueMemberRequest struct {
	UserID string `json:"user_id"`
}

func (h *Handler) AddQueueMember(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	queueID := c.Params("id")
	var req queueMemberRequest
	if err := c.BodyParser(&req); err != nil || req.UserID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "user_id is required")
	}

	var queueExists, userExists bool
	if err := db.Tx(c).QueryRow(c.Context(),
		`SELECT EXISTS(SELECT 1 FROM queues WHERE id=$1 AND organization_id=$2)`, queueID, orgID,
	).Scan(&queueExists); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not verify queue")
	}
	if err := db.Tx(c).QueryRow(c.Context(),
		`SELECT EXISTS(SELECT 1 FROM users WHERE id=$1 AND organization_id=$2)`, req.UserID, orgID,
	).Scan(&userExists); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not verify user")
	}
	if !queueExists || !userExists {
		return fiber.NewError(fiber.StatusBadRequest, "queue or user not found")
	}

	if _, err := db.Tx(c).Exec(c.Context(),
		`INSERT INTO queue_members(queue_id, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, queueID, req.UserID,
	); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not add queue member")
	}
	return c.SendStatus(fiber.StatusCreated)
}

func (h *Handler) RemoveQueueMember(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	queueID := c.Params("id")
	userID := c.Params("userId")
	cmd, err := db.Tx(c).Exec(c.Context(), `
		DELETE FROM queue_members
		WHERE queue_id=$1 AND user_id=$2
		  AND queue_id IN (SELECT id FROM queues WHERE organization_id=$3)`,
		queueID, userID, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not remove queue member")
	}
	if cmd.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "queue member not found")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// autoAssign picks an owner for a new case dropped into queueID, per the
// queue's routing strategy. Returns "" (no error) when the strategy is
// "manual" or the queue has no members, leaving the caller to fall back to
// its own default (the case creator).
func (h *Handler) autoAssign(ctx context.Context, q db.Queryer, queueID, orgID string) (string, error) {
	var strategy string
	var lastAssigned *string
	err := q.QueryRow(ctx,
		`SELECT routing_strategy, last_assigned_user_id FROM queues WHERE id=$1 AND organization_id=$2`,
		queueID, orgID,
	).Scan(&strategy, &lastAssigned)
	if err != nil || strategy == "manual" {
		return "", err
	}

	switch strategy {
	case "round_robin":
		return h.assignRoundRobin(ctx, q, queueID, lastAssigned)
	case "capacity":
		return h.assignByCapacity(ctx, q, queueID, orgID)
	default:
		return "", nil
	}
}

func (h *Handler) assignRoundRobin(ctx context.Context, q db.Queryer, queueID string, lastAssigned *string) (string, error) {
	rows, err := q.Query(ctx, `SELECT user_id FROM queue_members WHERE queue_id=$1 ORDER BY user_id`, queueID)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var members []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return "", err
		}
		members = append(members, id)
	}
	if len(members) == 0 {
		return "", nil
	}

	next := members[0]
	if lastAssigned != nil {
		for i, m := range members {
			if m == *lastAssigned {
				next = members[(i+1)%len(members)]
				break
			}
		}
	}
	if _, err := q.Exec(ctx, `UPDATE queues SET last_assigned_user_id=$1 WHERE id=$2`, next, queueID); err != nil {
		return "", err
	}
	return next, nil
}

func (h *Handler) assignByCapacity(ctx context.Context, q db.Queryer, queueID, orgID string) (string, error) {
	var best *string
	err := q.QueryRow(ctx, `
		SELECT qm.user_id
		FROM queue_members qm
		LEFT JOIN cases c ON c.owner_user_id = qm.user_id
			AND c.organization_id = $2
			AND c.status NOT IN ('resolved', 'closed')
		WHERE qm.queue_id = $1
		GROUP BY qm.user_id
		ORDER BY COUNT(c.id) ASC, qm.user_id ASC
		LIMIT 1`, queueID, orgID,
	).Scan(&best)
	if err != nil || best == nil {
		return "", err
	}
	return *best, nil
}
