package audit

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"ingencore/api/internal/auth"
	"ingencore/api/internal/db"
)

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// List returns the org's audit trail, newest first, optionally filtered to
// one entity type and/or one record's history.
func (h *Handler) List(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	entity := c.Query("entity")
	entityID := c.Query("entity_id")

	query := `
		SELECT a.id, a.actor_user_id, u.display_name, a.entity, a.entity_id, a.action, a.before, a.after, a.created_at
		FROM audit_log a
		LEFT JOIN users u ON u.id = a.actor_user_id
		WHERE a.organization_id = $1`
	args := []any{orgID}
	if entity != "" {
		args = append(args, entity)
		query += fmt.Sprintf(" AND a.entity = $%d", len(args))
	}
	if entityID != "" {
		args = append(args, entityID)
		query += fmt.Sprintf(" AND a.entity_id = $%d", len(args))
	}
	query += " ORDER BY a.created_at DESC LIMIT 200"

	rows, err := db.Tx(c).Query(c.Context(), query, args...)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not load audit log")
	}
	defer rows.Close()

	out := []fiber.Map{}
	for rows.Next() {
		var id, ent, entID, action string
		var actorID *string
		var actorName *string
		var before, after []byte
		var createdAt time.Time
		if err := rows.Scan(&id, &actorID, &actorName, &ent, &entID, &action, &before, &after, &createdAt); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not scan audit entry")
		}
		var beforeVal, afterVal any
		_ = json.Unmarshal(before, &beforeVal)
		_ = json.Unmarshal(after, &afterVal)
		out = append(out, fiber.Map{
			"id": id, "actor_user_id": actorID, "actor_name": actorName,
			"entity": ent, "entity_id": entID, "action": action,
			"before": beforeVal, "after": afterVal,
			"created_at": createdAt.Format(time.RFC3339),
		})
	}
	return c.JSON(out)
}
