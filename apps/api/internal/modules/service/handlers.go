package service

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ingencore/api/internal/audit"
	"ingencore/api/internal/auth"
	"ingencore/api/internal/authz"
	"ingencore/api/internal/db"
	"ingencore/api/internal/modules/automation"
)

type Handler struct {
	pool   *pgxpool.Pool
	engine *automation.Engine
}

func NewHandler(pool *pgxpool.Pool, engine *automation.Engine) *Handler {
	return &Handler{pool: pool, engine: engine}
}

// fireWorkflow runs matching workflows in the background, decoupled from the
// request's own context (which Fiber/fasthttp may recycle once the handler
// returns) so a slow webhook action never adds latency to the API response.
func (h *Handler) fireWorkflow(orgID, entity, event, recordID string, oldFields, newFields map[string]string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		h.engine.Run(ctx, orgID, entity, event, recordID, oldFields, newFields)
	}()
}

// ---- Queues ----

type queueRequest struct {
	Name string `json:"name"`
}

func (h *Handler) ListQueues(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	rows, err := db.Tx(c).Query(c.Context(),
		`SELECT id, name, routing_strategy, created_at FROM queues WHERE organization_id=$1 ORDER BY created_at DESC`, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list queues")
	}
	defer rows.Close()
	out := []fiber.Map{}
	for rows.Next() {
		var id, name, routingStrategy string
		var createdAt interface{}
		if err := rows.Scan(&id, &name, &routingStrategy, &createdAt); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not scan queue")
		}
		out = append(out, fiber.Map{"id": id, "name": name, "routing_strategy": routingStrategy, "created_at": createdAt})
	}
	return c.JSON(out)
}

func (h *Handler) CreateQueue(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	var req queueRequest
	if err := c.BodyParser(&req); err != nil || req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	var id string
	err := db.Tx(c).QueryRow(c.Context(),
		`INSERT INTO queues(organization_id, name) VALUES ($1,$2) RETURNING id`, orgID, req.Name,
	).Scan(&id)
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, "could not create queue (name may already exist)")
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id, "name": req.Name})
}

func (h *Handler) DeleteQueue(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	cmd, err := db.Tx(c).Exec(c.Context(), `DELETE FROM queues WHERE id=$1 AND organization_id=$2`, id, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete queue")
	}
	if cmd.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "queue not found")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ---- Cases ----

type caseRequest struct {
	Subject   string  `json:"subject"`
	Priority  string  `json:"priority"`
	Status    string  `json:"status"`
	QueueID   *string `json:"queue_id"`
	ContactID *string `json:"contact_id"`
}

func (h *Handler) ListCases(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	scope, ok, err := authz.Guard(c.Context(), db.Tx(c), auth.UserID(c), "case", "read")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not check privileges")
	}
	if !ok {
		return c.JSON([]fiber.Map{})
	}
	where, wargs := authz.ScopedWhere(scope, 2)
	rows, err := db.Tx(c).Query(c.Context(),
		`SELECT id, subject, priority, status, queue_id, contact_id, owner_user_id, created_at,
		        first_response_due_at, first_response_at, resolution_due_at, resolved_at, paused_at, paused_seconds
		 FROM cases WHERE organization_id=$1`+where+` ORDER BY created_at DESC`,
		append([]any{orgID}, wargs...)...)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list cases")
	}
	defer rows.Close()
	out := []fiber.Map{}
	for rows.Next() {
		var row caseRow
		if err := rows.Scan(&row.id, &row.subject, &row.priority, &row.status, &row.queueID, &row.contactID, &row.owner, &row.createdAt,
			&row.firstDue, &row.firstAt, &row.resDue, &row.resolvedAt, &row.pausedAt, &row.pausedSeconds); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not scan case")
		}
		out = append(out, row.toJSON())
	}
	return c.JSON(out)
}

// caseRow is the shared shape scanned by both ListCases and GetCase so the
// SLA computation (which needs several extra columns) isn't duplicated.
type caseRow struct {
	id, subject, status                             string
	priority, queueID, contactID, owner             *string
	createdAt                                       time.Time
	firstDue, firstAt, resDue, resolvedAt, pausedAt *time.Time
	pausedSeconds                                   int
}

func (r caseRow) toJSON() fiber.Map {
	return fiber.Map{
		"id": r.id, "subject": r.subject, "priority": r.priority, "status": r.status,
		"queue_id": r.queueID, "contact_id": r.contactID, "owner_user_id": r.owner,
		"created_at":         r.createdAt.Format(time.RFC3339),
		"first_response_sla": computeSLA(r.createdAt, r.firstDue, r.firstAt, r.pausedSeconds, r.pausedAt),
		"resolution_sla":     computeSLA(r.createdAt, r.resDue, r.resolvedAt, r.pausedSeconds, r.pausedAt),
	}
}

func (h *Handler) CreateCase(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	userID := auth.UserID(c)
	_, ok, err := authz.Guard(c.Context(), db.Tx(c), userID, "case", "create")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not check privileges")
	}
	if !ok {
		return fiber.NewError(fiber.StatusForbidden, "not permitted to create cases")
	}
	var req caseRequest
	if err := c.BodyParser(&req); err != nil || req.Subject == "" {
		return fiber.NewError(fiber.StatusBadRequest, "subject is required")
	}
	if req.Status == "" {
		req.Status = "new"
	}

	ownerID := userID
	if req.QueueID != nil {
		if assignee, aerr := h.autoAssign(c.Context(), db.Tx(c), *req.QueueID, orgID); aerr == nil && assignee != "" {
			ownerID = assignee
		}
	}
	firstDue, resDue := h.slaDueDates(c.Context(), db.Tx(c), orgID)

	var id string
	err = db.Tx(c).QueryRow(c.Context(),
		`INSERT INTO cases(organization_id, subject, priority, status, queue_id, contact_id, created_by, owner_user_id, first_response_due_at, resolution_due_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`,
		orgID, req.Subject, req.Priority, req.Status, req.QueueID, req.ContactID, userID, ownerID, firstDue, resDue,
	).Scan(&id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create case")
	}
	if err := audit.Log(c.Context(), db.Tx(c), orgID, userID, "case", id, "create", nil,
		fiber.Map{"subject": req.Subject, "priority": req.Priority, "status": req.Status}); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not record audit log")
	}
	h.fireWorkflow(orgID, "case", "created", id, nil, map[string]string{
		"subject": req.Subject, "priority": req.Priority, "status": req.Status, "owner_user_id": ownerID,
	})
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id, "subject": req.Subject, "status": req.Status, "owner_user_id": ownerID})
}

func (h *Handler) GetCase(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	scope, ok, err := authz.Guard(c.Context(), db.Tx(c), auth.UserID(c), "case", "read")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not check privileges")
	}
	if !ok {
		return fiber.NewError(fiber.StatusNotFound, "case not found")
	}
	where, wargs := authz.ScopedWhere(scope, 3)
	row := caseRow{id: id}
	err = db.Tx(c).QueryRow(c.Context(),
		`SELECT subject, priority, status, queue_id, contact_id, owner_user_id, created_at,
		        first_response_due_at, first_response_at, resolution_due_at, resolved_at, paused_at, paused_seconds
		 FROM cases WHERE id=$1 AND organization_id=$2`+where,
		append([]any{id, orgID}, wargs...)...,
	).Scan(&row.subject, &row.priority, &row.status, &row.queueID, &row.contactID, &row.owner, &row.createdAt,
		&row.firstDue, &row.firstAt, &row.resDue, &row.resolvedAt, &row.pausedAt, &row.pausedSeconds)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "case not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "could not load case")
	}
	return c.JSON(row.toJSON())
}

func (h *Handler) UpdateCase(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	scope, ok, err := authz.Guard(c.Context(), db.Tx(c), auth.UserID(c), "case", "write")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not check privileges")
	}
	if !ok {
		return fiber.NewError(fiber.StatusNotFound, "case not found")
	}

	where0, wargs0 := authz.ScopedWhere(scope, 3)
	var oldSubject, oldStatus string
	var oldPriority, oldOwner *string
	var oldPausedAt, oldResolvedAt *time.Time
	err = db.Tx(c).QueryRow(c.Context(),
		`SELECT subject, priority, status, owner_user_id, paused_at, resolved_at FROM cases WHERE id=$1 AND organization_id=$2`+where0,
		append([]any{id, orgID}, wargs0...)...,
	).Scan(&oldSubject, &oldPriority, &oldStatus, &oldOwner, &oldPausedAt, &oldResolvedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "case not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "could not load case")
	}

	var req caseRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	where, wargs := authz.ScopedWhere(scope, 8)
	cmd, err := db.Tx(c).Exec(c.Context(),
		`UPDATE cases SET subject=$1, priority=$2, status=$3, queue_id=$4, contact_id=$5 WHERE id=$6 AND organization_id=$7`+where,
		append([]any{req.Subject, req.Priority, req.Status, req.QueueID, req.ContactID, id, orgID}, wargs...)...)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not update case")
	}
	if cmd.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "case not found")
	}

	now := time.Now()
	switch {
	case pausingStatuses[req.Status] && !pausingStatuses[oldStatus]:
		_, _ = db.Tx(c).Exec(c.Context(), `UPDATE cases SET paused_at=$1 WHERE id=$2`, now, id)
	case !pausingStatuses[req.Status] && pausingStatuses[oldStatus] && oldPausedAt != nil:
		elapsed := int(now.Sub(*oldPausedAt).Seconds())
		_, _ = db.Tx(c).Exec(c.Context(), `UPDATE cases SET paused_seconds = paused_seconds + $1, paused_at = NULL WHERE id=$2`, elapsed, id)
	}
	switch {
	case resolvedStatuses[req.Status] && oldResolvedAt == nil:
		_, _ = db.Tx(c).Exec(c.Context(), `UPDATE cases SET resolved_at=$1 WHERE id=$2`, now, id)
	case !resolvedStatuses[req.Status] && oldResolvedAt != nil:
		_, _ = db.Tx(c).Exec(c.Context(), `UPDATE cases SET resolved_at=NULL WHERE id=$1`, id)
	}

	oldPriorityStr := ""
	if oldPriority != nil {
		oldPriorityStr = *oldPriority
	}
	oldOwnerID := ""
	if oldOwner != nil {
		oldOwnerID = *oldOwner
	}
	if err := audit.Log(c.Context(), db.Tx(c), orgID, auth.UserID(c), "case", id, "update",
		fiber.Map{"subject": oldSubject, "priority": oldPriorityStr, "status": oldStatus},
		fiber.Map{"subject": req.Subject, "priority": req.Priority, "status": req.Status}); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not record audit log")
	}
	h.fireWorkflow(orgID, "case", "updated", id,
		map[string]string{"subject": oldSubject, "priority": oldPriorityStr, "status": oldStatus, "owner_user_id": oldOwnerID},
		map[string]string{"subject": req.Subject, "priority": req.Priority, "status": req.Status, "owner_user_id": oldOwnerID})

	return c.JSON(fiber.Map{"id": id})
}

func (h *Handler) DeleteCase(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	scope, ok, err := authz.Guard(c.Context(), db.Tx(c), auth.UserID(c), "case", "delete")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not check privileges")
	}
	if !ok {
		return fiber.NewError(fiber.StatusNotFound, "case not found")
	}
	where, wargs := authz.ScopedWhere(scope, 3)
	var deletedSubject string
	err = db.Tx(c).QueryRow(c.Context(),
		`DELETE FROM cases WHERE id=$1 AND organization_id=$2`+where+` RETURNING subject`,
		append([]any{id, orgID}, wargs...)...,
	).Scan(&deletedSubject)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "case not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete case")
	}
	if err := audit.Log(c.Context(), db.Tx(c), orgID, auth.UserID(c), "case", id, "delete",
		fiber.Map{"subject": deletedSubject}, nil); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not record audit log")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
