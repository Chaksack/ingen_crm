package service

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ingen.one/api/internal/auth"
)

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// ---- Queues ----

type queueRequest struct {
	Name string `json:"name"`
}

func (h *Handler) ListQueues(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	rows, err := h.pool.Query(c.Context(),
		`SELECT id, name, created_at FROM queues WHERE organization_id=$1 ORDER BY created_at DESC`, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list queues")
	}
	defer rows.Close()
	var out []fiber.Map
	for rows.Next() {
		var id, name string
		var createdAt interface{}
		if err := rows.Scan(&id, &name, &createdAt); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not scan queue")
		}
		out = append(out, fiber.Map{"id": id, "name": name, "created_at": createdAt})
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
	err := h.pool.QueryRow(c.Context(),
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
	cmd, err := h.pool.Exec(c.Context(), `DELETE FROM queues WHERE id=$1 AND organization_id=$2`, id, orgID)
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
	rows, err := h.pool.Query(c.Context(),
		`SELECT id, subject, priority, status, queue_id, contact_id, owner_user_id, created_at
		 FROM cases WHERE organization_id=$1 ORDER BY created_at DESC`, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list cases")
	}
	defer rows.Close()
	var out []fiber.Map
	for rows.Next() {
		var id, subject, status string
		var priority, queueID, contactID, owner *string
		var createdAt interface{}
		if err := rows.Scan(&id, &subject, &priority, &status, &queueID, &contactID, &owner, &createdAt); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not scan case")
		}
		out = append(out, fiber.Map{
			"id": id, "subject": subject, "priority": priority, "status": status,
			"queue_id": queueID, "contact_id": contactID, "owner_user_id": owner, "created_at": createdAt,
		})
	}
	return c.JSON(out)
}

func (h *Handler) CreateCase(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	userID := auth.UserID(c)
	var req caseRequest
	if err := c.BodyParser(&req); err != nil || req.Subject == "" {
		return fiber.NewError(fiber.StatusBadRequest, "subject is required")
	}
	if req.Status == "" {
		req.Status = "new"
	}
	var id string
	err := h.pool.QueryRow(c.Context(),
		`INSERT INTO cases(organization_id, subject, priority, status, queue_id, contact_id, created_by, owner_user_id)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$7) RETURNING id`,
		orgID, req.Subject, req.Priority, req.Status, req.QueueID, req.ContactID, userID,
	).Scan(&id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create case")
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id, "subject": req.Subject, "status": req.Status})
}

func (h *Handler) GetCase(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	var subject, status string
	var priority, queueID, contactID, owner *string
	err := h.pool.QueryRow(c.Context(),
		`SELECT subject, priority, status, queue_id, contact_id, owner_user_id
		 FROM cases WHERE id=$1 AND organization_id=$2`, id, orgID,
	).Scan(&subject, &priority, &status, &queueID, &contactID, &owner)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "case not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "could not load case")
	}
	return c.JSON(fiber.Map{
		"id": id, "subject": subject, "priority": priority, "status": status,
		"queue_id": queueID, "contact_id": contactID, "owner_user_id": owner,
	})
}

func (h *Handler) UpdateCase(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	var req caseRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	cmd, err := h.pool.Exec(c.Context(),
		`UPDATE cases SET subject=$1, priority=$2, status=$3, queue_id=$4, contact_id=$5 WHERE id=$6 AND organization_id=$7`,
		req.Subject, req.Priority, req.Status, req.QueueID, req.ContactID, id, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not update case")
	}
	if cmd.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "case not found")
	}
	return c.JSON(fiber.Map{"id": id})
}

func (h *Handler) DeleteCase(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	cmd, err := h.pool.Exec(c.Context(), `DELETE FROM cases WHERE id=$1 AND organization_id=$2`, id, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete case")
	}
	if cmd.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "case not found")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
