package collab

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"

	"ingen.one/api/internal/auth"
)

type inboundMessage struct {
	To   string `json:"to"`
	Body string `json:"body"`
}

type outboundMessage struct {
	Type            string `json:"type"`
	ID              string `json:"id,omitempty"`
	SenderUserID    string `json:"sender_user_id"`
	RecipientUserID string `json:"recipient_user_id"`
	Body            string `json:"body"`
	CreatedAt       string `json:"created_at"`
}

// Serve handles one WebSocket connection: register/unregister with the hub,
// persist inbound messages, and relay them to the recipient if online.
func (h *Handler) Serve(c *websocket.Conn) {
	userID, _ := c.Locals(auth.LocalUserID).(string)
	orgID, _ := c.Locals(auth.LocalOrgID).(string)
	if userID == "" || orgID == "" {
		_ = c.Close()
		return
	}

	h.hub.Register(userID, c)
	defer h.hub.Unregister(userID)

	for {
		var in inboundMessage
		if err := c.ReadJSON(&in); err != nil {
			break
		}
		if in.To == "" || in.Body == "" {
			continue
		}

		var id string
		createdAt := time.Now().UTC()
		err := h.pool.QueryRow(context.Background(),
			`INSERT INTO messages(organization_id, sender_user_id, recipient_user_id, body)
			 VALUES ($1,$2,$3,$4) RETURNING id, created_at`,
			orgID, userID, in.To, in.Body,
		).Scan(&id, &createdAt)
		if err != nil {
			log.Printf("collab: could not persist message: %v", err)
			continue
		}

		out := outboundMessage{
			Type: "message", ID: id, SenderUserID: userID, RecipientUserID: in.To,
			Body: in.Body, CreatedAt: createdAt.Format(time.RFC3339),
		}
		_ = c.WriteJSON(out) // echo to sender for optimistic UI reconciliation
		h.hub.Send(in.To, out)
	}
}
