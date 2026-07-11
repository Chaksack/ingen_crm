// Package push implements Web Push delivery (PRD §5.11/§7.5): browsers
// register a subscription (endpoint + encryption keys) while the PWA is
// open; the server can then push a notification to it even when the app is
// closed, via VAPID-authenticated requests to the browser's push service.
package push

import (
	"context"
	"encoding/json"
	"log"

	webpush "github.com/SherClockHolmes/webpush-go"

	"ingencore/api/internal/db"
)

// Sender holds the VAPID keypair used to authenticate outgoing push
// requests. Configured from .env (see internal/config); if the public or
// private key is empty, SendToUser silently no-ops so local dev without
// push configured still works.
type Sender struct {
	PublicKey  string
	PrivateKey string
	Contact    string
}

func NewSender(publicKey, privateKey, contact string) *Sender {
	return &Sender{PublicKey: publicKey, PrivateKey: privateKey, Contact: contact}
}

func (s *Sender) configured() bool {
	return s.PublicKey != "" && s.PrivateKey != ""
}

type payload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url,omitempty"`
}

// SendToUser pushes a notification to every device userID has subscribed
// from. Subscriptions that the push service reports as gone (404/410) are
// pruned, per PRD §6 ("push subscriptions pruned on failure"). Best-effort:
// failures are logged, never returned, since a missed push should never
// fail the caller's own request/workflow run.
func (s *Sender) SendToUser(ctx context.Context, q db.Queryer, orgID, userID, title, body, url string) {
	if !s.configured() {
		return
	}
	rows, err := q.Query(ctx,
		`SELECT id, endpoint, p256dh, auth FROM push_subscriptions WHERE organization_id=$1 AND user_id=$2`,
		orgID, userID)
	if err != nil {
		log.Printf("push: could not load subscriptions: %v", err)
		return
	}
	type sub struct{ id, endpoint, p256dh, auth string }
	var subs []sub
	for rows.Next() {
		var r sub
		if err := rows.Scan(&r.id, &r.endpoint, &r.p256dh, &r.auth); err != nil {
			continue
		}
		subs = append(subs, r)
	}
	rows.Close()

	msg, err := json.Marshal(payload{Title: title, Body: body, URL: url})
	if err != nil {
		return
	}

	for _, sub := range subs {
		resp, err := webpush.SendNotificationWithContext(ctx, msg, &webpush.Subscription{
			Endpoint: sub.endpoint,
			Keys:     webpush.Keys{P256dh: sub.p256dh, Auth: sub.auth},
		}, &webpush.Options{
			Subscriber:      s.Contact,
			VAPIDPublicKey:  s.PublicKey,
			VAPIDPrivateKey: s.PrivateKey,
			TTL:             30,
		})
		if err != nil {
			log.Printf("push: send failed for subscription %s: %v", sub.id, err)
			continue
		}
		resp.Body.Close()
		switch {
		case resp.StatusCode == 404 || resp.StatusCode == 410:
			if _, err := q.Exec(ctx, `DELETE FROM push_subscriptions WHERE id=$1`, sub.id); err != nil {
				log.Printf("push: could not prune dead subscription %s: %v", sub.id, err)
			}
		case resp.StatusCode >= 300:
			log.Printf("push: subscription %s: push service returned status %d", sub.id, resp.StatusCode)
		}
	}
}
