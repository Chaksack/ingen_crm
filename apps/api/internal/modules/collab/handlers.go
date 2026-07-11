package collab

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"ingencore/api/internal/auth"
	"ingencore/api/internal/db"
	"ingencore/api/internal/push"
)

type Handler struct {
	pool *pgxpool.Pool
	hub  *Hub
	push *push.Sender
}

func NewHandler(pool *pgxpool.Pool, hub *Hub, pushSender *push.Sender) *Handler {
	return &Handler{pool: pool, hub: hub, push: pushSender}
}

// ListUsers returns the tenant's users so the caller can start a 1:1 chat.
func (h *Handler) ListUsers(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	selfID := auth.UserID(c)
	rows, err := db.Tx(c).Query(c.Context(),
		`SELECT id, display_name, email FROM users WHERE organization_id=$1 AND id<>$2 ORDER BY display_name`,
		orgID, selfID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list users")
	}
	defer rows.Close()
	out := []fiber.Map{}
	for rows.Next() {
		var id, displayName, email string
		if err := rows.Scan(&id, &displayName, &email); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not scan user")
		}
		out = append(out, fiber.Map{
			"id": id, "display_name": displayName, "email": email, "online": h.hub.IsOnline(id),
		})
	}
	return c.JSON(out)
}

// MessageHistory returns the 1:1 conversation between the caller and :userID.
func (h *Handler) MessageHistory(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	selfID := auth.UserID(c)
	otherID := c.Params("userID")

	rows, err := db.Tx(c).Query(c.Context(),
		`SELECT id, sender_user_id, recipient_user_id, body, created_at
		 FROM messages
		 WHERE organization_id=$1
		   AND ((sender_user_id=$2 AND recipient_user_id=$3) OR (sender_user_id=$3 AND recipient_user_id=$2))
		 ORDER BY created_at ASC
		 LIMIT 200`,
		orgID, selfID, otherID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not load messages")
	}
	defer rows.Close()
	out := []fiber.Map{}
	for rows.Next() {
		var id, sender, recipient, body string
		var createdAt time.Time
		if err := rows.Scan(&id, &sender, &recipient, &body, &createdAt); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not scan message")
		}
		out = append(out, fiber.Map{
			"id": id, "sender_user_id": sender, "recipient_user_id": recipient,
			"body": body, "created_at": createdAt.Format(time.RFC3339),
		})
	}
	return c.JSON(out)
}
