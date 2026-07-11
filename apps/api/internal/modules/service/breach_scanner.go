package service

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5"

	"ingencore/api/internal/db"
)

// breachRow is the subset of case columns needed to decide whether either
// SLA timer has breached and hasn't been notified yet.
type breachRow struct {
	id, subject, status                   string
	priority, owner                       *string
	firstDue, firstAt, resDue, resolvedAt *time.Time
	pausedAt                              *time.Time
	pausedSeconds                         int
	firstNotified, resNotified            *time.Time
}

// StartBreachScanner polls for cases whose first-response or resolution SLA
// has breached and fires the "sla_breach" workflow trigger exactly once per
// breach (dedup via the *_breach_notified_at columns), so that workflows can
// finally act on a breach (notify/reassign/escalate) instead of it only
// turning a badge red. Meant to be launched with `go serviceH.StartBreachScanner(ctx, interval)`.
func (h *Handler) StartBreachScanner(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		h.scanForBreaches(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// scanForBreaches is a background job with no request/tenant context of its
// own, so — unlike every HTTP handler — it must look across all orgs. cases
// is RLS-protected, so it scans org by org, opening a short tenant-scoped
// transaction per org (see db.WithOrgTx) rather than one global query that
// RLS would otherwise reduce to zero rows.
func (h *Handler) scanForBreaches(ctx context.Context) {
	orgRows, err := h.pool.Query(ctx, `SELECT id FROM organizations`)
	if err != nil {
		log.Printf("sla breach scan: could not list organizations: %v", err)
		return
	}
	var orgIDs []string
	for orgRows.Next() {
		var id string
		if err := orgRows.Scan(&id); err != nil {
			continue
		}
		orgIDs = append(orgIDs, id)
	}
	orgRows.Close()

	for _, orgID := range orgIDs {
		h.scanOrgForBreaches(ctx, orgID)
	}
}

func (h *Handler) scanOrgForBreaches(ctx context.Context, orgID string) {
	type breach struct {
		caseID, breachType string
		fields             map[string]string
	}
	var toFire []breach

	err := db.WithOrgTx(ctx, h.pool, orgID, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `
			SELECT id, subject, priority, status, owner_user_id, first_response_due_at,
			       first_response_at, resolution_due_at, resolved_at, paused_at, paused_seconds,
			       first_response_breach_notified_at, resolution_breach_notified_at
			FROM cases
			WHERE organization_id = $1
			  AND ((first_response_at IS NULL AND first_response_due_at IS NOT NULL AND first_response_breach_notified_at IS NULL)
			   OR (resolved_at IS NULL AND resolution_due_at IS NOT NULL AND resolution_breach_notified_at IS NULL))`,
			orgID)
		if err != nil {
			return err
		}
		var candidates []breachRow
		for rows.Next() {
			var r breachRow
			if err := rows.Scan(&r.id, &r.subject, &r.priority, &r.status, &r.owner, &r.firstDue,
				&r.firstAt, &r.resDue, &r.resolvedAt, &r.pausedAt, &r.pausedSeconds, &r.firstNotified, &r.resNotified); err != nil {
				continue
			}
			candidates = append(candidates, r)
		}
		rows.Close()

		now := time.Now()
		for _, r := range candidates {
			owner := ""
			if r.owner != nil {
				owner = *r.owner
			}
			priority := ""
			if r.priority != nil {
				priority = *r.priority
			}
			fields := func(breachType string) map[string]string {
				return map[string]string{
					"subject": r.subject, "priority": priority, "status": r.status,
					"owner_user_id": owner, "breach_type": breachType,
				}
			}

			if r.firstNotified == nil && r.firstDue != nil && r.firstAt == nil {
				due := effectiveDueTime(*r.firstDue, r.pausedSeconds, r.pausedAt)
				if now.After(due) && claimBreachNotification(ctx, tx, r.id, "first_response_breach_notified_at") {
					toFire = append(toFire, breach{caseID: r.id, breachType: "first_response", fields: fields("first_response")})
				}
			}
			if r.resNotified == nil && r.resDue != nil && r.resolvedAt == nil {
				due := effectiveDueTime(*r.resDue, r.pausedSeconds, r.pausedAt)
				if now.After(due) && claimBreachNotification(ctx, tx, r.id, "resolution_breach_notified_at") {
					toFire = append(toFire, breach{caseID: r.id, breachType: "resolution", fields: fields("resolution")})
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("sla breach scan: org %s: %v", orgID, err)
		return
	}

	for _, b := range toFire {
		h.fireWorkflow(orgID, "case", "sla_breach", b.caseID, nil, b.fields)
	}
}

// claimBreachNotification atomically marks a breach as notified, returning
// true only for the caller that wins the race — so two overlapping scan
// ticks (or a future multi-instance deployment) can never double-fire the
// same breach.
func claimBreachNotification(ctx context.Context, tx pgx.Tx, caseID, column string) bool {
	cmd, err := tx.Exec(ctx,
		`UPDATE cases SET `+column+` = now() WHERE id=$1 AND `+column+` IS NULL`, caseID)
	if err != nil {
		log.Printf("sla breach scan: could not claim notification for case %s: %v", caseID, err)
		return false
	}
	return cmd.RowsAffected() == 1
}
