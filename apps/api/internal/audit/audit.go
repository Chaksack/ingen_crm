// Package audit implements entity-level write auditing (PRD §5.1/E8, named
// twice in the §6 NFRs: actor, timestamp, before/after values, an immutable
// append-only trail). Log runs inside the caller's own transaction — if the
// audit insert fails, the whole request rolls back, so a mutation can never
// commit without also being recorded.
package audit

import (
	"context"
	"encoding/json"

	"ingencore/api/internal/db"
)

// Log records one write. before is nil for a create; after is nil for a
// delete. actorID may be empty for system-initiated writes.
func Log(ctx context.Context, q db.Queryer, orgID, actorID, entity, entityID, action string, before, after any) error {
	beforeJSON, err := json.Marshal(before)
	if err != nil {
		return err
	}
	afterJSON, err := json.Marshal(after)
	if err != nil {
		return err
	}
	var actor any
	if actorID != "" {
		actor = actorID
	}
	_, err = q.Exec(ctx,
		`INSERT INTO audit_log(organization_id, actor_user_id, entity, entity_id, action, before, after)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		orgID, actor, entity, entityID, action, beforeJSON, afterJSON,
	)
	return err
}
