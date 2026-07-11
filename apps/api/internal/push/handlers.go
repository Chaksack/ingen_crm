package push

import (
	"github.com/gofiber/fiber/v2"

	"ingencore/api/internal/auth"
	"ingencore/api/internal/db"
)

type Handler struct {
	Sender *Sender
}

func NewHandler(sender *Sender) *Handler {
	return &Handler{Sender: sender}
}

// VAPIDPublicKey is unauthenticated: the frontend needs it before it can
// even ask the browser to subscribe, which happens as part of getting the
// user logged in.
func (h *Handler) VAPIDPublicKey(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"public_key": h.Sender.PublicKey})
}

type subscribeRequest struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

func (h *Handler) Subscribe(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	userID := auth.UserID(c)
	var req subscribeRequest
	if err := c.BodyParser(&req); err != nil || req.Endpoint == "" || req.Keys.P256dh == "" || req.Keys.Auth == "" {
		return fiber.NewError(fiber.StatusBadRequest, "endpoint and keys.p256dh/keys.auth are required")
	}
	_, err := db.Tx(c).Exec(c.Context(), `
		INSERT INTO push_subscriptions(organization_id, user_id, endpoint, p256dh, auth)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (endpoint) DO UPDATE SET user_id=EXCLUDED.user_id, organization_id=EXCLUDED.organization_id,
			p256dh=EXCLUDED.p256dh, auth=EXCLUDED.auth`,
		orgID, userID, req.Endpoint, req.Keys.P256dh, req.Keys.Auth,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not save push subscription")
	}
	return c.SendStatus(fiber.StatusCreated)
}

type unsubscribeRequest struct {
	Endpoint string `json:"endpoint"`
}

func (h *Handler) Unsubscribe(c *fiber.Ctx) error {
	userID := auth.UserID(c)
	var req unsubscribeRequest
	if err := c.BodyParser(&req); err != nil || req.Endpoint == "" {
		return fiber.NewError(fiber.StatusBadRequest, "endpoint is required")
	}
	if _, err := db.Tx(c).Exec(c.Context(),
		`DELETE FROM push_subscriptions WHERE endpoint=$1 AND user_id=$2`, req.Endpoint, userID,
	); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not remove push subscription")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
