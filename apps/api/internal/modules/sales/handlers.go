package sales

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

// ---- Accounts ----

type accountRequest struct {
	Name string `json:"name"`
}

func (h *Handler) ListAccounts(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	scope, ok, err := authz.Guard(c.Context(), db.Tx(c), auth.UserID(c), "account", "read")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not check privileges")
	}
	if !ok {
		return c.JSON([]fiber.Map{})
	}
	where, wargs := authz.ScopedWhere(scope, 2)
	rows, err := db.Tx(c).Query(c.Context(),
		`SELECT id, name, owner_user_id, created_at FROM accounts WHERE organization_id=$1`+where+` ORDER BY created_at DESC`,
		append([]any{orgID}, wargs...)...)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list accounts")
	}
	defer rows.Close()
	out := []fiber.Map{}
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
	_, ok, err := authz.Guard(c.Context(), db.Tx(c), userID, "account", "create")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not check privileges")
	}
	if !ok {
		return fiber.NewError(fiber.StatusForbidden, "not permitted to create accounts")
	}
	var req accountRequest
	if err := c.BodyParser(&req); err != nil || req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	var id string
	err = db.Tx(c).QueryRow(c.Context(),
		`INSERT INTO accounts(organization_id, name, created_by, owner_user_id) VALUES ($1,$2,$3,$3) RETURNING id`,
		orgID, req.Name, userID,
	).Scan(&id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create account")
	}
	if err := audit.Log(c.Context(), db.Tx(c), orgID, userID, "account", id, "create", nil, fiber.Map{"name": req.Name}); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not record audit log")
	}
	h.fireWorkflow(orgID, "account", "created", id, nil, map[string]string{"name": req.Name, "owner_user_id": userID})
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id, "name": req.Name})
}

func (h *Handler) GetAccount(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	scope, ok, err := authz.Guard(c.Context(), db.Tx(c), auth.UserID(c), "account", "read")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not check privileges")
	}
	if !ok {
		return fiber.NewError(fiber.StatusNotFound, "account not found")
	}
	where, wargs := authz.ScopedWhere(scope, 3)
	var name string
	var owner *string
	err = db.Tx(c).QueryRow(c.Context(),
		`SELECT name, owner_user_id FROM accounts WHERE id=$1 AND organization_id=$2`+where,
		append([]any{id, orgID}, wargs...)...,
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
	scope, ok, err := authz.Guard(c.Context(), db.Tx(c), auth.UserID(c), "account", "write")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not check privileges")
	}
	if !ok {
		return fiber.NewError(fiber.StatusNotFound, "account not found")
	}
	var req accountRequest
	if err := c.BodyParser(&req); err != nil || req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	readWhere, readArgs := authz.ScopedWhere(scope, 3)
	var oldName string
	var ownerID *string
	_ = db.Tx(c).QueryRow(c.Context(),
		`SELECT name, owner_user_id FROM accounts WHERE id=$1 AND organization_id=$2`+readWhere,
		append([]any{id, orgID}, readArgs...)...,
	).Scan(&oldName, &ownerID)

	where, wargs := authz.ScopedWhere(scope, 4)
	cmd, err := db.Tx(c).Exec(c.Context(),
		`UPDATE accounts SET name=$1 WHERE id=$2 AND organization_id=$3`+where,
		append([]any{req.Name, id, orgID}, wargs...)...)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not update account")
	}
	if cmd.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "account not found")
	}
	owner := ""
	if ownerID != nil {
		owner = *ownerID
	}
	if err := audit.Log(c.Context(), db.Tx(c), orgID, auth.UserID(c), "account", id, "update",
		fiber.Map{"name": oldName}, fiber.Map{"name": req.Name}); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not record audit log")
	}
	h.fireWorkflow(orgID, "account", "updated", id,
		map[string]string{"name": oldName, "owner_user_id": owner},
		map[string]string{"name": req.Name, "owner_user_id": owner})
	return c.JSON(fiber.Map{"id": id, "name": req.Name})
}

func (h *Handler) DeleteAccount(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	scope, ok, err := authz.Guard(c.Context(), db.Tx(c), auth.UserID(c), "account", "delete")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not check privileges")
	}
	if !ok {
		return fiber.NewError(fiber.StatusNotFound, "account not found")
	}
	where, wargs := authz.ScopedWhere(scope, 3)
	var deletedName string
	err = db.Tx(c).QueryRow(c.Context(),
		`DELETE FROM accounts WHERE id=$1 AND organization_id=$2`+where+` RETURNING name`,
		append([]any{id, orgID}, wargs...)...,
	).Scan(&deletedName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "account not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete account")
	}
	if err := audit.Log(c.Context(), db.Tx(c), orgID, auth.UserID(c), "account", id, "delete",
		fiber.Map{"name": deletedName}, nil); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not record audit log")
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
	scope, ok, err := authz.Guard(c.Context(), db.Tx(c), auth.UserID(c), "contact", "read")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not check privileges")
	}
	if !ok {
		return c.JSON([]fiber.Map{})
	}
	where, wargs := authz.ScopedWhere(scope, 2)
	rows, err := db.Tx(c).Query(c.Context(),
		`SELECT id, account_id, first_name, last_name, email, phone, created_at FROM contacts WHERE organization_id=$1`+where+` ORDER BY created_at DESC`,
		append([]any{orgID}, wargs...)...)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list contacts")
	}
	defer rows.Close()
	out := []fiber.Map{}
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
	_, ok, err := authz.Guard(c.Context(), db.Tx(c), userID, "contact", "create")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not check privileges")
	}
	if !ok {
		return fiber.NewError(fiber.StatusForbidden, "not permitted to create contacts")
	}
	var req contactRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if req.FirstName == "" && req.LastName == "" {
		return fiber.NewError(fiber.StatusBadRequest, "first_name or last_name is required")
	}
	var id string
	err = db.Tx(c).QueryRow(c.Context(),
		`INSERT INTO contacts(organization_id, account_id, first_name, last_name, email, phone, created_by, owner_user_id)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$7) RETURNING id`,
		orgID, req.AccountID, req.FirstName, req.LastName, req.Email, req.Phone, userID,
	).Scan(&id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create contact")
	}
	if err := audit.Log(c.Context(), db.Tx(c), orgID, userID, "contact", id, "create", nil, fiber.Map{
		"first_name": req.FirstName, "last_name": req.LastName, "email": req.Email, "phone": req.Phone,
	}); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not record audit log")
	}
	h.fireWorkflow(orgID, "contact", "created", id, nil, map[string]string{
		"first_name": req.FirstName, "last_name": req.LastName, "email": req.Email, "phone": req.Phone, "owner_user_id": userID,
	})
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) GetContact(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	scope, ok, err := authz.Guard(c.Context(), db.Tx(c), auth.UserID(c), "contact", "read")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not check privileges")
	}
	if !ok {
		return fiber.NewError(fiber.StatusNotFound, "contact not found")
	}
	where, wargs := authz.ScopedWhere(scope, 3)
	var accountID, firstName, lastName, email, phone *string
	err = db.Tx(c).QueryRow(c.Context(),
		`SELECT account_id, first_name, last_name, email, phone FROM contacts WHERE id=$1 AND organization_id=$2`+where,
		append([]any{id, orgID}, wargs...)...,
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
	scope, ok, err := authz.Guard(c.Context(), db.Tx(c), auth.UserID(c), "contact", "write")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not check privileges")
	}
	if !ok {
		return fiber.NewError(fiber.StatusNotFound, "contact not found")
	}
	var req contactRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	readWhere, readArgs := authz.ScopedWhere(scope, 3)
	var oldFirstName, oldLastName, oldEmail, oldPhone string
	_ = db.Tx(c).QueryRow(c.Context(),
		`SELECT first_name, last_name, email, phone FROM contacts WHERE id=$1 AND organization_id=$2`+readWhere,
		append([]any{id, orgID}, readArgs...)...,
	).Scan(&oldFirstName, &oldLastName, &oldEmail, &oldPhone)

	where, wargs := authz.ScopedWhere(scope, 8)
	cmd, err := db.Tx(c).Exec(c.Context(),
		`UPDATE contacts SET account_id=$1, first_name=$2, last_name=$3, email=$4, phone=$5 WHERE id=$6 AND organization_id=$7`+where,
		append([]any{req.AccountID, req.FirstName, req.LastName, req.Email, req.Phone, id, orgID}, wargs...)...)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not update contact")
	}
	if cmd.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "contact not found")
	}
	if err := audit.Log(c.Context(), db.Tx(c), orgID, auth.UserID(c), "contact", id, "update",
		fiber.Map{"first_name": oldFirstName, "last_name": oldLastName, "email": oldEmail, "phone": oldPhone},
		fiber.Map{"first_name": req.FirstName, "last_name": req.LastName, "email": req.Email, "phone": req.Phone},
	); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not record audit log")
	}
	h.fireWorkflow(orgID, "contact", "updated", id,
		map[string]string{"first_name": oldFirstName, "last_name": oldLastName, "email": oldEmail, "phone": oldPhone},
		map[string]string{"first_name": req.FirstName, "last_name": req.LastName, "email": req.Email, "phone": req.Phone})
	return c.JSON(fiber.Map{"id": id})
}

func (h *Handler) DeleteContact(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	scope, ok, err := authz.Guard(c.Context(), db.Tx(c), auth.UserID(c), "contact", "delete")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not check privileges")
	}
	if !ok {
		return fiber.NewError(fiber.StatusNotFound, "contact not found")
	}
	where, wargs := authz.ScopedWhere(scope, 3)
	var delFirstName, delLastName string
	err = db.Tx(c).QueryRow(c.Context(),
		`DELETE FROM contacts WHERE id=$1 AND organization_id=$2`+where+` RETURNING first_name, last_name`,
		append([]any{id, orgID}, wargs...)...,
	).Scan(&delFirstName, &delLastName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "contact not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete contact")
	}
	if err := audit.Log(c.Context(), db.Tx(c), orgID, auth.UserID(c), "contact", id, "delete",
		fiber.Map{"first_name": delFirstName, "last_name": delLastName}, nil); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not record audit log")
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
	scope, ok, err := authz.Guard(c.Context(), db.Tx(c), auth.UserID(c), "lead", "read")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not check privileges")
	}
	if !ok {
		return c.JSON([]fiber.Map{})
	}
	where, wargs := authz.ScopedWhere(scope, 2)
	rows, err := db.Tx(c).Query(c.Context(),
		`SELECT id, topic, contact_id, status, created_at FROM leads WHERE organization_id=$1`+where+` ORDER BY created_at DESC`,
		append([]any{orgID}, wargs...)...)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list leads")
	}
	defer rows.Close()
	out := []fiber.Map{}
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
	_, ok, err := authz.Guard(c.Context(), db.Tx(c), userID, "lead", "create")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not check privileges")
	}
	if !ok {
		return fiber.NewError(fiber.StatusForbidden, "not permitted to create leads")
	}
	var req leadRequest
	if err := c.BodyParser(&req); err != nil || req.Topic == "" {
		return fiber.NewError(fiber.StatusBadRequest, "topic is required")
	}
	if req.Status == "" {
		req.Status = "open"
	}
	var id string
	err = db.Tx(c).QueryRow(c.Context(),
		`INSERT INTO leads(organization_id, topic, contact_id, status, created_by, owner_user_id)
		 VALUES ($1,$2,$3,$4,$5,$5) RETURNING id`,
		orgID, req.Topic, req.ContactID, req.Status, userID,
	).Scan(&id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create lead")
	}
	if err := audit.Log(c.Context(), db.Tx(c), orgID, userID, "lead", id, "create", nil,
		fiber.Map{"topic": req.Topic, "status": req.Status}); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not record audit log")
	}
	h.fireWorkflow(orgID, "lead", "created", id, nil, map[string]string{
		"topic": req.Topic, "status": req.Status, "owner_user_id": userID,
	})
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id, "topic": req.Topic, "status": req.Status})
}

func (h *Handler) GetLead(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	scope, ok, err := authz.Guard(c.Context(), db.Tx(c), auth.UserID(c), "lead", "read")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not check privileges")
	}
	if !ok {
		return fiber.NewError(fiber.StatusNotFound, "lead not found")
	}
	where, wargs := authz.ScopedWhere(scope, 3)
	var topic, status string
	var contactID *string
	err = db.Tx(c).QueryRow(c.Context(),
		`SELECT topic, contact_id, status FROM leads WHERE id=$1 AND organization_id=$2`+where,
		append([]any{id, orgID}, wargs...)...,
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
	scope, ok, err := authz.Guard(c.Context(), db.Tx(c), auth.UserID(c), "lead", "write")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not check privileges")
	}
	if !ok {
		return fiber.NewError(fiber.StatusNotFound, "lead not found")
	}
	var req leadRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	readWhere, readArgs := authz.ScopedWhere(scope, 3)
	var oldTopic, oldStatus string
	_ = db.Tx(c).QueryRow(c.Context(),
		`SELECT topic, status FROM leads WHERE id=$1 AND organization_id=$2`+readWhere,
		append([]any{id, orgID}, readArgs...)...,
	).Scan(&oldTopic, &oldStatus)

	where, wargs := authz.ScopedWhere(scope, 6)
	cmd, err := db.Tx(c).Exec(c.Context(),
		`UPDATE leads SET topic=$1, contact_id=$2, status=$3 WHERE id=$4 AND organization_id=$5`+where,
		append([]any{req.Topic, req.ContactID, req.Status, id, orgID}, wargs...)...)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not update lead")
	}
	if cmd.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "lead not found")
	}
	if err := audit.Log(c.Context(), db.Tx(c), orgID, auth.UserID(c), "lead", id, "update",
		fiber.Map{"topic": oldTopic, "status": oldStatus},
		fiber.Map{"topic": req.Topic, "status": req.Status}); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not record audit log")
	}
	h.fireWorkflow(orgID, "lead", "updated", id,
		map[string]string{"topic": oldTopic, "status": oldStatus},
		map[string]string{"topic": req.Topic, "status": req.Status})
	return c.JSON(fiber.Map{"id": id})
}

func (h *Handler) DeleteLead(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	scope, ok, err := authz.Guard(c.Context(), db.Tx(c), auth.UserID(c), "lead", "delete")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not check privileges")
	}
	if !ok {
		return fiber.NewError(fiber.StatusNotFound, "lead not found")
	}
	where, wargs := authz.ScopedWhere(scope, 3)
	var deletedTopic string
	err = db.Tx(c).QueryRow(c.Context(),
		`DELETE FROM leads WHERE id=$1 AND organization_id=$2`+where+` RETURNING topic`,
		append([]any{id, orgID}, wargs...)...,
	).Scan(&deletedTopic)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "lead not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete lead")
	}
	if err := audit.Log(c.Context(), db.Tx(c), orgID, auth.UserID(c), "lead", id, "delete",
		fiber.Map{"topic": deletedTopic}, nil); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not record audit log")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
