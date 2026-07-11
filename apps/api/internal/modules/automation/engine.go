// Package automation implements the workflow engine from PRD §5.1/E9: a
// trigger (record created/updated, or an SLA breach) evaluates conditions
// and, if they match, runs a sequence of actions (update a field, reassign
// the owner, create an in-app notification, or POST a webhook). Every run
// is logged to workflow_runs for auditability.
//
// Scoping notes (see README "Known gaps"): this runs in-process, triggered
// by direct calls from the CRUD handlers — there's no NATS/event-bus layer
// yet (PRD §7.2.6 describes that as the eventual architecture once the
// workflow engine needs to fan out to more than one consumer). Actions run
// with system/Admin-equivalent privilege, not through the privilege-depth
// model — a workflow is authored by an Admin and is trusted the same way a
// database trigger would be.
package automation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ingencore/api/internal/db"
	"ingencore/api/internal/push"
)

type Condition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"` // equals | not_equals | changed
	Value    string `json:"value"`
}

type Action struct {
	Type   string `json:"type"` // update_field | assign_owner | notify | webhook
	Field  string `json:"field,omitempty"`
	Value  string `json:"value,omitempty"`
	UserID string `json:"user_id,omitempty"` // notify/assign_owner; "owner" means "the record's current owner"
	Title  string `json:"title,omitempty"`
	Body   string `json:"body,omitempty"`
	URL    string `json:"url,omitempty"`
}

// fieldSetters allowlists which columns update_field may touch per entity —
// the same "validate against a fixed set before it ever reaches SQL"
// pattern used for privilege entities/actions elsewhere in this codebase.
var fieldSetters = map[string]map[string]bool{
	"account": {"name": true},
	"contact": {"first_name": true, "last_name": true, "email": true, "phone": true},
	"lead":    {"topic": true, "status": true},
	"case":    {"subject": true, "priority": true, "status": true},
}

var ownedEntities = map[string]bool{"account": true, "contact": true, "lead": true, "case": true}

var tableNames = map[string]string{
	"account": "accounts", "contact": "contacts", "lead": "leads", "case": "cases",
}

var listPaths = map[string]string{
	"account": "/sales/accounts", "contact": "/sales/contacts", "lead": "/sales/leads", "case": "/service/cases",
}

type Engine struct {
	pool *pgxpool.Pool
	push *push.Sender
}

func NewEngine(pool *pgxpool.Pool, pushSender *push.Sender) *Engine {
	return &Engine{pool: pool, push: pushSender}
}

// Run finds active workflows for org+entity+triggerEvent and executes any
// whose conditions match. oldFields is nil for a "created" trigger (there's
// nothing to compare against). Fields are compared as strings — v1 keeps
// conditions simple (equals/not_equals/changed), matching the "lite" scope.
//
// Intended to be called via `go engine.Run(...)` with a context derived from
// context.Background(), not a Fiber request context — see callers.
func (e *Engine) Run(ctx context.Context, orgID, entity, triggerEvent, recordID string, oldFields, newFields map[string]string) {
	err := db.WithOrgTx(ctx, e.pool, orgID, func(tx pgx.Tx) error {
		return e.run(ctx, tx, orgID, entity, triggerEvent, recordID, oldFields, newFields)
	})
	if err != nil {
		log.Printf("automation: run failed: %v", err)
	}
}

func (e *Engine) run(ctx context.Context, tx pgx.Tx, orgID, entity, triggerEvent, recordID string, oldFields, newFields map[string]string) error {
	rows, err := tx.Query(ctx, `
		SELECT id, conditions, actions FROM workflows
		WHERE organization_id=$1 AND entity=$2 AND trigger_event=$3 AND is_active`,
		orgID, entity, triggerEvent)
	if err != nil {
		return fmt.Errorf("could not load workflows: %w", err)
	}
	type row struct {
		id                  string
		conditions, actions []byte
	}
	var matched []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.conditions, &r.actions); err != nil {
			continue
		}
		matched = append(matched, r)
	}
	rows.Close()

	for _, r := range matched {
		var conditions []Condition
		_ = json.Unmarshal(r.conditions, &conditions)
		if !conditionsMatch(conditions, oldFields, newFields) {
			continue
		}
		var actions []Action
		_ = json.Unmarshal(r.actions, &actions)
		detail := e.executeActions(ctx, tx, orgID, entity, recordID, newFields, actions)
		status := "success"
		if strings.Contains(detail, "error:") {
			status = "error"
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO workflow_runs(workflow_id, organization_id, record_id, status, detail) VALUES ($1,$2,$3,$4,$5)`,
			r.id, orgID, recordID, status, detail,
		); err != nil {
			log.Printf("automation: could not log workflow run: %v", err)
		}
	}
	return nil
}

func conditionsMatch(conditions []Condition, oldFields, newFields map[string]string) bool {
	for _, c := range conditions {
		newVal := newFields[c.Field]
		switch c.Operator {
		case "equals":
			if newVal != c.Value {
				return false
			}
		case "not_equals":
			if newVal == c.Value {
				return false
			}
		case "changed":
			if oldFields == nil {
				continue // created trigger: nothing to have changed from
			}
			if oldFields[c.Field] == newVal {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func (e *Engine) executeActions(ctx context.Context, tx pgx.Tx, orgID, entity, recordID string, newFields map[string]string, actions []Action) string {
	var notes []string
	for _, a := range actions {
		switch a.Type {
		case "update_field":
			notes = append(notes, actionUpdateField(ctx, tx, orgID, entity, recordID, a))
		case "assign_owner":
			notes = append(notes, actionAssignOwner(ctx, tx, orgID, entity, recordID, a))
		case "notify":
			notes = append(notes, e.actionNotify(ctx, tx, orgID, entity, recordID, newFields, a))
		case "webhook":
			notes = append(notes, actionWebhook(entity, recordID, newFields, a))
		default:
			notes = append(notes, "error: unknown action type "+a.Type)
		}
	}
	return strings.Join(notes, "; ")
}

func actionUpdateField(ctx context.Context, tx pgx.Tx, orgID, entity, recordID string, a Action) string {
	if !fieldSetters[entity][a.Field] {
		return fmt.Sprintf("error: field %q not allowed for %s", a.Field, entity)
	}
	table := tableNames[entity]
	_, err := tx.Exec(ctx,
		fmt.Sprintf(`UPDATE %s SET %s=$1 WHERE id=$2 AND organization_id=$3`, table, a.Field),
		a.Value, recordID, orgID)
	if err != nil {
		return "error: update_field failed"
	}
	return fmt.Sprintf("set %s=%s", a.Field, a.Value)
}

func actionAssignOwner(ctx context.Context, tx pgx.Tx, orgID, entity, recordID string, a Action) string {
	if !ownedEntities[entity] {
		return "error: assign_owner not supported for " + entity
	}
	table := tableNames[entity]
	_, err := tx.Exec(ctx,
		fmt.Sprintf(`UPDATE %s SET owner_user_id=$1 WHERE id=$2 AND organization_id=$3`, table),
		a.UserID, recordID, orgID)
	if err != nil {
		return "error: assign_owner failed"
	}
	return "assigned to " + a.UserID
}

func (e *Engine) actionNotify(ctx context.Context, tx pgx.Tx, orgID, entity, recordID string, newFields map[string]string, a Action) string {
	userID := a.UserID
	if userID == "owner" {
		userID = newFields["owner_user_id"]
	}
	if userID == "" {
		return "error: notify has no resolvable user_id"
	}
	link := listPaths[entity]
	_, err := tx.Exec(ctx,
		`INSERT INTO notifications(organization_id, user_id, title, body, link) VALUES ($1,$2,$3,$4,$5)`,
		orgID, userID, a.Title, a.Body, link)
	if err != nil {
		return "error: notify failed"
	}
	_ = recordID
	e.push.SendToUser(ctx, tx, orgID, userID, a.Title, a.Body, link)
	return "notified " + userID
}

func actionWebhook(entity, recordID string, fields map[string]string, a Action) string {
	if a.URL == "" {
		return "error: webhook missing url"
	}
	payload, _ := json.Marshal(map[string]any{"entity": entity, "record_id": recordID, "fields": fields})
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodPost, a.URL, bytes.NewReader(payload))
	if err != nil {
		return "error: webhook request build failed"
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "error: webhook request failed"
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Sprintf("error: webhook returned status %d", resp.StatusCode)
	}
	return "webhook posted"
}
