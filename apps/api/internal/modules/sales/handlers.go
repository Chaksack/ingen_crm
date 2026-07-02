package sales

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

// ---- Accounts ----

type accountRequest struct {
	Name string `json:"name"`
}

func (h *Handler) ListAccounts(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	rows, err := h.pool.Query(c.Context(),
		`SELECT id, name, owner_user_id, created_at FROM accounts WHERE organization_id=$1 ORDER BY created_at DESC`, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list accounts")
	}
	defer rows.Close()
	var out []fiber.Map
	for rows.Next() {
		var id, name string
		var owner *string
		var createdAt interface{}
		if err := rows.Scan(&id, &name, &owner, &createdAt); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not scan account")
		}
		out = append(out, fiber.Map{"id": id, "name": name, "owner_user_id": owner, "created_at": createdAt})
	}
	return c.JSON(out)
}

func (h *Handler) CreateAccount(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	userID := auth.UserID(c)
	var req accountRequest
	if err := c.BodyParser(&req); err != nil || req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	var id string
	err := h.pool.QueryRow(c.Context(),
		`INSERT INTO accounts(organization_id, name, created_by, owner_user_id) VALUES ($1,$2,$3,$3) RETURNING id`,
		orgID, req.Name, userID,
	).Scan(&id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create account")
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id, "name": req.Name})
}

func (h *Handler) GetAccount(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	var name string
	var owner *string
	err := h.pool.QueryRow(c.Context(),
		`SELECT name, owner_user_id FROM accounts WHERE id=$1 AND organization_id=$2`, id, orgID,
	).Scan(&name, &owner)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "account not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "could not load account")
	}
	return c.JSON(fiber.Map{"id": id, "name": name, "owner_user_id": owner})
}

func (h *Handler) UpdateAccount(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	var req accountRequest
	if err := c.BodyParser(&req); err != nil || req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	cmd, err := h.pool.Exec(c.Context(),
		`UPDATE accounts SET name=$1 WHERE id=$2 AND organization_id=$3`, req.Name, id, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not update account")
	}
	if cmd.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "account not found")
	}
	return c.JSON(fiber.Map{"id": id, "name": req.Name})
}

func (h *Handler) DeleteAccount(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	cmd, err := h.pool.Exec(c.Context(), `DELETE FROM accounts WHERE id=$1 AND organization_id=$2`, id, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete account")
	}
	if cmd.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "account not found")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ---- Contacts ----

type contactRequest struct {
	AccountID *string `json:"account_id"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Email     string  `json:"email"`
	Phone     string  `json:"phone"`
}

func (h *Handler) ListContacts(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	rows, err := h.pool.Query(c.Context(),
		`SELECT id, account_id, first_name, last_name, email, phone, created_at FROM contacts WHERE organization_id=$1 ORDER BY created_at DESC`, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list contacts")
	}
	defer rows.Close()
	var out []fiber.Map
	for rows.Next() {
		var id string
		var accountID, firstName, lastName, email, phone *string
		var createdAt interface{}
		if err := rows.Scan(&id, &accountID, &firstName, &lastName, &email, &phone, &createdAt); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not scan contact")
		}
		out = append(out, fiber.Map{
			"id": id, "account_id": accountID, "first_name": firstName, "last_name": lastName,
			"email": email, "phone": phone, "created_at": createdAt,
		})
	}
	return c.JSON(out)
}

func (h *Handler) CreateContact(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	userID := auth.UserID(c)
	var req contactRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if req.FirstName == "" && req.LastName == "" {
		return fiber.NewError(fiber.StatusBadRequest, "first_name or last_name is required")
	}
	var id string
	err := h.pool.QueryRow(c.Context(),
		`INSERT INTO contacts(organization_id, account_id, first_name, last_name, email, phone, created_by, owner_user_id)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$7) RETURNING id`,
		orgID, req.AccountID, req.FirstName, req.LastName, req.Email, req.Phone, userID,
	).Scan(&id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create contact")
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) GetContact(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	var accountID, firstName, lastName, email, phone *string
	err := h.pool.QueryRow(c.Context(),
		`SELECT account_id, first_name, last_name, email, phone FROM contacts WHERE id=$1 AND organization_id=$2`, id, orgID,
	).Scan(&accountID, &firstName, &lastName, &email, &phone)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "contact not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "could not load contact")
	}
	return c.JSON(fiber.Map{
		"id": id, "account_id": accountID, "first_name": firstName, "last_name": lastName,
		"email": email, "phone": phone,
	})
}

func (h *Handler) UpdateContact(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	var req contactRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	cmd, err := h.pool.Exec(c.Context(),
		`UPDATE contacts SET account_id=$1, first_name=$2, last_name=$3, email=$4, phone=$5 WHERE id=$6 AND organization_id=$7`,
		req.AccountID, req.FirstName, req.LastName, req.Email, req.Phone, id, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not update contact")
	}
	if cmd.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "contact not found")
	}
	return c.JSON(fiber.Map{"id": id})
}

func (h *Handler) DeleteContact(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	cmd, err := h.pool.Exec(c.Context(), `DELETE FROM contacts WHERE id=$1 AND organization_id=$2`, id, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete contact")
	}
	if cmd.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "contact not found")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ---- Leads ----

type leadRequest struct {
	Topic     string  `json:"topic"`
	ContactID *string `json:"contact_id"`
	Status    string  `json:"status"`
}

func (h *Handler) ListLeads(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	rows, err := h.pool.Query(c.Context(),
		`SELECT id, topic, contact_id, status, created_at FROM leads WHERE organization_id=$1 ORDER BY created_at DESC`, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list leads")
	}
	defer rows.Close()
	var out []fiber.Map
	for rows.Next() {
		var id, topic, status string
		var contactID *string
		var createdAt interface{}
		if err := rows.Scan(&id, &topic, &contactID, &status, &createdAt); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not scan lead")
		}
		out = append(out, fiber.Map{"id": id, "topic": topic, "contact_id": contactID, "status": status, "created_at": createdAt})
	}
	return c.JSON(out)
}

func (h *Handler) CreateLead(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	userID := auth.UserID(c)
	var req leadRequest
	if err := c.BodyParser(&req); err != nil || req.Topic == "" {
		return fiber.NewError(fiber.StatusBadRequest, "topic is required")
	}
	if req.Status == "" {
		req.Status = "open"
	}
	var id string
	err := h.pool.QueryRow(c.Context(),
		`INSERT INTO leads(organization_id, topic, contact_id, status, created_by, owner_user_id)
		 VALUES ($1,$2,$3,$4,$5,$5) RETURNING id`,
		orgID, req.Topic, req.ContactID, req.Status, userID,
	).Scan(&id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create lead")
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id, "topic": req.Topic, "status": req.Status})
}

func (h *Handler) GetLead(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	var topic, status string
	var contactID *string
	err := h.pool.QueryRow(c.Context(),
		`SELECT topic, contact_id, status FROM leads WHERE id=$1 AND organization_id=$2`, id, orgID,
	).Scan(&topic, &contactID, &status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "lead not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "could not load lead")
	}
	return c.JSON(fiber.Map{"id": id, "topic": topic, "contact_id": contactID, "status": status})
}

func (h *Handler) UpdateLead(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	var req leadRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	cmd, err := h.pool.Exec(c.Context(),
		`UPDATE leads SET topic=$1, contact_id=$2, status=$3 WHERE id=$4 AND organization_id=$5`,
		req.Topic, req.ContactID, req.Status, id, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not update lead")
	}
	if cmd.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "lead not found")
	}
	return c.JSON(fiber.Map{"id": id})
}

func (h *Handler) DeleteLead(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	cmd, err := h.pool.Exec(c.Context(), `DELETE FROM leads WHERE id=$1 AND organization_id=$2`, id, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete lead")
	}
	if cmd.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "lead not found")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
