package service

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"ingencore/api/internal/auth"
	"ingencore/api/internal/authz"
	"ingencore/api/internal/db"
)

// pausingStatuses freeze both SLA timers while a case sits in them (e.g.
// waiting on the customer to respond isn't the agent's clock running).
var pausingStatuses = map[string]bool{"waiting": true}

// resolvedStatuses stop the resolution timer and record resolved_at.
var resolvedStatuses = map[string]bool{"resolved": true, "closed": true}

type slaPolicyRequest struct {
	FirstResponseMinutes int `json:"first_response_minutes"`
	ResolutionMinutes    int `json:"resolution_minutes"`
}

func (h *Handler) GetSLAPolicy(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	var first, res int
	err := db.Tx(c).QueryRow(c.Context(),
		`SELECT first_response_minutes, resolution_minutes FROM sla_policies WHERE organization_id=$1`, orgID,
	).Scan(&first, &res)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not load SLA policy")
	}
	return c.JSON(fiber.Map{"first_response_minutes": first, "resolution_minutes": res})
}

func (h *Handler) UpdateSLAPolicy(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	var req slaPolicyRequest
	if err := c.BodyParser(&req); err != nil || req.FirstResponseMinutes <= 0 || req.ResolutionMinutes <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "first_response_minutes and resolution_minutes must be positive")
	}
	_, err := db.Tx(c).Exec(c.Context(), `
		INSERT INTO sla_policies(organization_id, first_response_minutes, resolution_minutes)
		VALUES ($1,$2,$3)
		ON CONFLICT (organization_id) DO UPDATE SET first_response_minutes=$2, resolution_minutes=$3`,
		orgID, req.FirstResponseMinutes, req.ResolutionMinutes,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not update SLA policy")
	}
	return c.JSON(fiber.Map{"first_response_minutes": req.FirstResponseMinutes, "resolution_minutes": req.ResolutionMinutes})
}

// slaDueDates computes due timestamps for a case created right now, from the
// org's policy (falling back to sensible defaults if the policy row is
// somehow missing).
func (h *Handler) slaDueDates(ctx context.Context, q db.Queryer, orgID string) (*time.Time, *time.Time) {
	firstMin, resMin := 60, 480
	_ = q.QueryRow(ctx,
		`SELECT first_response_minutes, resolution_minutes FROM sla_policies WHERE organization_id=$1`, orgID,
	).Scan(&firstMin, &resMin)
	now := time.Now()
	first := now.Add(time.Duration(firstMin) * time.Minute)
	res := now.Add(time.Duration(resMin) * time.Minute)
	return &first, &res
}

// effectiveDueTime extends a due date by however long the case has spent (or
// is currently spending) in a paused status, so pausing genuinely stops the
// clock rather than just hiding that it kept running.
func effectiveDueTime(dueAt time.Time, pausedSeconds int, pausedAt *time.Time) time.Time {
	total := pausedSeconds
	if pausedAt != nil {
		total += int(time.Since(*pausedAt).Seconds())
	}
	return dueAt.Add(time.Duration(total) * time.Second)
}

// computeSLA renders one timer (first-response or resolution) as a status
// the UI can badge directly: ok / warning (last 20% of the window) /
// breached / met / met_late.
func computeSLA(createdAt time.Time, dueAt *time.Time, achievedAt *time.Time, pausedSeconds int, pausedAt *time.Time) fiber.Map {
	if dueAt == nil {
		return fiber.Map{"status": "n/a"}
	}
	due := effectiveDueTime(*dueAt, pausedSeconds, pausedAt)
	if achievedAt != nil {
		status := "met"
		if achievedAt.After(due) {
			status = "met_late"
		}
		return fiber.Map{"status": status, "due_at": due.Format(time.RFC3339), "achieved_at": achievedAt.Format(time.RFC3339)}
	}
	now := time.Now()
	if now.After(due) {
		return fiber.Map{"status": "breached", "due_at": due.Format(time.RFC3339)}
	}
	total := due.Sub(createdAt)
	warningStart := due.Add(-total / 5) // last 20% of the window
	status := "ok"
	if now.After(warningStart) {
		status = "warning"
	}
	return fiber.Map{"status": status, "due_at": due.Format(time.RFC3339)}
}

// RespondToCase marks first-response as achieved. Idempotent: calling it
// again after the first time is a no-op, not an error.
func (h *Handler) RespondToCase(c *fiber.Ctx) error {
	return h.touchCase(c, "first_response_at = COALESCE(first_response_at, now())")
}

func (h *Handler) touchCase(c *fiber.Ctx, setClause string) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	scope, ok, err := authz.Guard(c.Context(), db.Tx(c), auth.UserID(c), "case", "write")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not check privileges")
	}
	if !ok {
		return fiber.NewError(fiber.StatusNotFound, "case not found")
	}
	where, wargs := authz.ScopedWhere(scope, 3)
	cmd, err := db.Tx(c).Exec(c.Context(),
		`UPDATE cases SET `+setClause+` WHERE id=$1 AND organization_id=$2`+where,
		append([]any{id, orgID}, wargs...)...)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not update case")
	}
	if cmd.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "case not found")
	}
	return c.JSON(fiber.Map{"id": id})
}
