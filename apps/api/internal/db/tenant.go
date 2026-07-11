package db

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"ingencore/api/internal/auth"
)

// Queryer is satisfied by both *pgxpool.Pool and pgx.Tx, so authz/entitlement
// checks and handler queries can run against either a raw pool connection or
// a request-scoped, tenant-context transaction without caring which.
type Queryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

const txLocalsKey = "tenantTx"

// Middleware begins one transaction per request, sets the Postgres session
// variable the tenant_isolation RLS policies check (app.org_id) to the
// caller's org from the JWT, and commits (or rolls back on handler error)
// once the chain returns. Every RLS-protected table fails closed if this is
// never set for a request, so any authenticated route group MUST be mounted
// under this middleware — see main.go.
func Middleware(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		tx, err := pool.Begin(ctx)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not start transaction")
		}
		if err := SetOrgContext(ctx, tx, auth.OrgID(c)); err != nil {
			_ = tx.Rollback(ctx)
			return fiber.NewError(fiber.StatusInternalServerError, "could not set tenant context")
		}
		c.Locals(txLocalsKey, tx)

		if err := c.Next(); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not commit transaction")
		}
		return nil
	}
}

// Tx returns the current request's tenant-scoped transaction. Every route
// under Middleware has one; calling this outside that chain is a bug.
func Tx(c *fiber.Ctx) pgx.Tx {
	return c.Locals(txLocalsKey).(pgx.Tx)
}

// SetOrgContext sets the RLS session variable on an ad hoc transaction, for
// code that manages its own transaction outside the request middleware:
// tenant bootstrap during registration, invite acceptance, and background
// work (the workflow engine, the SLA breach scanner) that has no fiber.Ctx.
func SetOrgContext(ctx context.Context, tx pgx.Tx, orgID string) error {
	_, err := tx.Exec(ctx, `SELECT set_config('app.org_id', $1, true)`, orgID)
	return err
}

// WithOrgTx runs fn inside a fresh, tenant-scoped transaction, committing on
// success and rolling back on error.
func WithOrgTx(ctx context.Context, pool *pgxpool.Pool, orgID string, fn func(pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := SetOrgContext(ctx, tx, orgID); err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
