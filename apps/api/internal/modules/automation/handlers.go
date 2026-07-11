package automation

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"ingencore/api/internal/auth"
	"ingencore/api/internal/db"
	"ingencore/api/internal/push"
)

type Handler struct {
	pool   *pgxpool.Pool
	Engine *Engine
}

func NewHandler(pool *pgxpool.Pool, pushSender *push.Sender) *Handler {
	return &Handler{pool: pool, Engine: NewEngine(pool, pushSender)}
}

var validEntities = map[string]bool{"account": true, "contact": true, "lead": true, "case": true}
var validTriggers = map[string]bool{"created": true, "updated": true, "sla_breach": true}
var validOperators = map[string]bool{"equals": true, "not_equals": true, "changed": true}
var validActionTypes = map[string]bool{"update_field": true, "assign_owner": true, "notify": true, "webhook": true}

type workflowRequest struct {
	Name         string      `json:"name"`
	Entity       string      `json:"entity"`
	TriggerEvent string      `json:"trigger_event"`
	Conditions   []Condition `json:"conditions"`
	Actions      []Action    `json:"actions"`
	IsActive     *bool       `json:"is_active"`
}

func (req workflowRequest) validate() string {
	if req.Name == "" {
		return "name is required"
	}
	if !validEntities[req.Entity] {
		return "entity must be one of account, contact, lead, case"
	}
	if !validTriggers[req.TriggerEvent] {
		return "trigger_event must be one of created, updated, sla_breach"
	}
	for _, c := range req.Conditions {
		if !validOperators[c.Operator] {
			return "condition operator must be one of equals, not_equals, changed"
		}
	}
	for _, a := range req.Actions {
		if !validActionTypes[a.Type] {
			return "action type must be one of update_field, assign_owner, notify, webhook"
		}
	}
	return ""
}

func (h *Handler) ListWorkflows(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	rows, err := db.Tx(c).Query(c.Context(), `
		SELECT id, name, entity, trigger_event, conditions, actions, is_active, created_at
		FROM workflows WHERE organization_id=$1 ORDER BY created_at DESC`, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list workflows")
	}
	defer rows.Close()
	out := []fiber.Map{}
	for rows.Next() {
		var id, name, entity, triggerEvent string
		var conditions, actions []byte
		var isActive bool
		var createdAt time.Time
		if err := rows.Scan(&id, &name, &entity, &triggerEvent, &conditions, &actions, &isActive, &createdAt); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not scan workflow")
		}
		out = append(out, workflowJSON(id, name, entity, triggerEvent, conditions, actions, isActive, createdAt))
	}
	return c.JSON(out)
}

func workflowJSON(id, name, entity, triggerEvent string, conditions, actions []byte, isActive bool, createdAt time.Time) fiber.Map {
	var conds []Condition
	var acts []Action
	_ = json.Unmarshal(conditions, &conds)
	_ = json.Unmarshal(actions, &acts)
	return fiber.Map{
		"id": id, "name": name, "entity": entity, "trigger_event": triggerEvent,
		"conditions": conds, "actions": acts, "is_active": isActive,
		"created_at": createdAt.Format(time.RFC3339),
	}
}

func (h *Handler) CreateWorkflow(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	userID := auth.UserID(c)
	var req workflowRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if msg := req.validate(); msg != "" {
		return fiber.NewError(fiber.StatusBadRequest, msg)
	}
	conditionsJSON, _ := json.Marshal(req.Conditions)
	actionsJSON, _ := json.Marshal(req.Actions)

	var id string
	err := db.Tx(c).QueryRow(c.Context(), `
		INSERT INTO workflows(organization_id, name, entity, trigger_event, conditions, actions, created_by)
		VALUES ($1,$2,$3,$4,$5::jsonb,$6::jsonb,$7) RETURNING id`,
		orgID, req.Name, req.Entity, req.TriggerEvent, conditionsJSON, actionsJSON, userID,
	).Scan(&id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create workflow")
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) UpdateWorkflow(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	var req workflowRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if msg := req.validate(); msg != "" {
		return fiber.NewError(fiber.StatusBadRequest, msg)
	}
	conditionsJSON, _ := json.Marshal(req.Conditions)
	actionsJSON, _ := json.Marshal(req.Actions)
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	cmd, err := db.Tx(c).Exec(c.Context(), `
		UPDATE workflows SET name=$1, entity=$2, trigger_event=$3, conditions=$4::jsonb, actions=$5::jsonb,
		       is_active=$6, updated_at=now()
		WHERE id=$7 AND organization_id=$8`,
		req.Name, req.Entity, req.TriggerEvent, conditionsJSON, actionsJSON, isActive, id, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not update workflow")
	}
	if cmd.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "workflow not found")
	}
	return c.JSON(fiber.Map{"id": id})
}

func (h *Handler) DeleteWorkflow(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	cmd, err := db.Tx(c).Exec(c.Context(), `DELETE FROM workflows WHERE id=$1 AND organization_id=$2`, id, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete workflow")
	}
	if cmd.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "workflow not found")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ListWorkflowRuns(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	workflowID := c.Params("id")
	var exists bool
	if err := db.Tx(c).QueryRow(c.Context(),
		`SELECT EXISTS(SELECT 1 FROM workflows WHERE id=$1 AND organization_id=$2)`, workflowID, orgID,
	).Scan(&exists); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not verify workflow")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "workflow not found")
	}
	rows, err := db.Tx(c).Query(c.Context(),
		`SELECT id, record_id, status, detail, ran_at FROM workflow_runs WHERE workflow_id=$1 ORDER BY ran_at DESC LIMIT 50`,
		workflowID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list workflow runs")
	}
	defer rows.Close()
	out := []fiber.Map{}
	for rows.Next() {
		var id, recordID, status, detail string
		var ranAt time.Time
		if err := rows.Scan(&id, &recordID, &status, &detail, &ranAt); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not scan workflow run")
		}
		out = append(out, fiber.Map{
			"id": id, "record_id": recordID, "status": status, "detail": detail,
			"ran_at": ranAt.Format(time.RFC3339),
		})
	}
	return c.JSON(out)
}

// ---- Notifications ----

func (h *Handler) ListNotifications(c *fiber.Ctx) error {
	userID := auth.UserID(c)
	rows, err := db.Tx(c).Query(c.Context(),
		`SELECT id, title, body, link, is_read, created_at FROM notifications
		 WHERE user_id=$1 ORDER BY created_at DESC LIMIT 50`, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list notifications")
	}
	defer rows.Close()
	out := []fiber.Map{}
	for rows.Next() {
		var id, title, body string
		var link *string
		var isRead bool
		var createdAt time.Time
		if err := rows.Scan(&id, &title, &body, &link, &isRead, &createdAt); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not scan notification")
		}
		out = append(out, fiber.Map{
			"id": id, "title": title, "body": body, "link": link, "is_read": isRead,
			"created_at": createdAt.Format(time.RFC3339),
		})
	}
	return c.JSON(out)
}

func (h *Handler) UnreadNotificationCount(c *fiber.Ctx) error {
	userID := auth.UserID(c)
	var count int
	if err := db.Tx(c).QueryRow(c.Context(),
		`SELECT COUNT(*) FROM notifications WHERE user_id=$1 AND NOT is_read`, userID,
	).Scan(&count); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not count notifications")
	}
	return c.JSON(fiber.Map{"count": count})
}

func (h *Handler) MarkNotificationRead(c *fiber.Ctx) error {
	userID := auth.UserID(c)
	id := c.Params("id")
	cmd, err := db.Tx(c).Exec(c.Context(),
		`UPDATE notifications SET is_read=true WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not update notification")
	}
	if cmd.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "notification not found")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
