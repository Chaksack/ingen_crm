package service

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"

	"ingencore/api/internal/auth"
	"ingencore/api/internal/authz"
	"ingencore/api/internal/db"
)

type kbArticleRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func (h *Handler) ListKBArticles(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	rows, err := db.Tx(c).Query(c.Context(),
		`SELECT id, title, status, created_at, updated_at FROM kb_articles WHERE organization_id=$1 ORDER BY updated_at DESC`, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list articles")
	}
	defer rows.Close()
	out := []fiber.Map{}
	for rows.Next() {
		var id, title, status string
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&id, &title, &status, &createdAt, &updatedAt); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not scan article")
		}
		out = append(out, fiber.Map{
			"id": id, "title": title, "status": status,
			"created_at": createdAt.Format(time.RFC3339), "updated_at": updatedAt.Format(time.RFC3339),
		})
	}
	return c.JSON(out)
}

func (h *Handler) CreateKBArticle(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	userID := auth.UserID(c)
	var req kbArticleRequest
	if err := c.BodyParser(&req); err != nil || req.Title == "" {
		return fiber.NewError(fiber.StatusBadRequest, "title is required")
	}
	var id string
	err := db.Tx(c).QueryRow(c.Context(),
		`INSERT INTO kb_articles(organization_id, title, body, created_by) VALUES ($1,$2,$3,$4) RETURNING id`,
		orgID, req.Title, req.Body, userID,
	).Scan(&id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create article")
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id, "title": req.Title, "status": "draft"})
}

func (h *Handler) GetKBArticle(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	var title, body, status string
	err := db.Tx(c).QueryRow(c.Context(),
		`SELECT title, body, status FROM kb_articles WHERE id=$1 AND organization_id=$2`, id, orgID,
	).Scan(&title, &body, &status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "article not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "could not load article")
	}
	return c.JSON(fiber.Map{"id": id, "title": title, "body": body, "status": status})
}

func (h *Handler) UpdateKBArticle(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	var req kbArticleRequest
	if err := c.BodyParser(&req); err != nil || req.Title == "" {
		return fiber.NewError(fiber.StatusBadRequest, "title is required")
	}
	cmd, err := db.Tx(c).Exec(c.Context(),
		`UPDATE kb_articles SET title=$1, body=$2, updated_at=now() WHERE id=$3 AND organization_id=$4`,
		req.Title, req.Body, id, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not update article")
	}
	if cmd.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "article not found")
	}
	return c.JSON(fiber.Map{"id": id, "title": req.Title})
}

func (h *Handler) setKBStatus(c *fiber.Ctx, status string) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	cmd, err := db.Tx(c).Exec(c.Context(),
		`UPDATE kb_articles SET status=$1, updated_at=now() WHERE id=$2 AND organization_id=$3`, status, id, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not update article status")
	}
	if cmd.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "article not found")
	}
	return c.JSON(fiber.Map{"id": id, "status": status})
}

func (h *Handler) PublishKBArticle(c *fiber.Ctx) error   { return h.setKBStatus(c, "published") }
func (h *Handler) UnpublishKBArticle(c *fiber.Ctx) error { return h.setKBStatus(c, "draft") }

func (h *Handler) DeleteKBArticle(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	id := c.Params("id")
	cmd, err := db.Tx(c).Exec(c.Context(), `DELETE FROM kb_articles WHERE id=$1 AND organization_id=$2`, id, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete article")
	}
	if cmd.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "article not found")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// SuggestKBArticles does a naive keyword match of a case's subject against
// published article titles/bodies — a placeholder for real search (Meilisearch
// is still on the backlog).
func (h *Handler) SuggestKBArticles(c *fiber.Ctx) error {
	orgID := auth.OrgID(c)
	caseID := c.Query("case_id")
	if caseID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "case_id query parameter is required")
	}

	scope, ok, err := authz.Guard(c.Context(), db.Tx(c), auth.UserID(c), "case", "read")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not check privileges")
	}
	if !ok {
		return fiber.NewError(fiber.StatusNotFound, "case not found")
	}
	where, wargs := authz.ScopedWhere(scope, 3)
	var subject string
	err = db.Tx(c).QueryRow(c.Context(),
		`SELECT subject FROM cases WHERE id=$1 AND organization_id=$2`+where,
		append([]any{caseID, orgID}, wargs...)...,
	).Scan(&subject)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "case not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "could not load case")
	}

	words := strings.Fields(subject)
	patterns := make([]string, 0, len(words))
	for _, w := range words {
		if len(w) < 3 {
			continue
		}
		patterns = append(patterns, "%"+w+"%")
	}
	if len(patterns) == 0 {
		return c.JSON([]fiber.Map{})
	}

	rows, err := db.Tx(c).Query(c.Context(), `
		SELECT id, title FROM kb_articles
		WHERE organization_id=$1 AND status='published'
		  AND (title ILIKE ANY($2) OR body ILIKE ANY($2))
		LIMIT 5`, orgID, patterns)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not search articles")
	}
	defer rows.Close()
	out := []fiber.Map{}
	for rows.Next() {
		var id, title string
		if err := rows.Scan(&id, &title); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not scan article")
		}
		out = append(out, fiber.Map{"id": id, "title": title})
	}
	return c.JSON(out)
}
