package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

type ids struct {
	orgID       string
	buID        string
	userID      string
	adminRoleID string
}

func mustEnv() string {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		log.Fatal("DATABASE_URL is required in environment")
	}
	return url
}

func connect(url string) *pgxpool.Pool {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		log.Fatalf("db connect error: %v", err)
	}
	ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctx2); err != nil {
		log.Fatalf("db ping error: %v", err)
	}
	return pool
}

func seed(pool *pgxpool.Pool) (*ids, error) {
	ctx := context.Background()
	var out ids

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// 1) Organization
	const orgName = "Dev Org"
	err = tx.QueryRow(ctx, `SELECT id FROM organizations WHERE name=$1`, orgName).Scan(&out.orgID)
	if err != nil {
		if err == pgx.ErrNoRows {
			err = tx.QueryRow(ctx, `INSERT INTO organizations(name) VALUES ($1) RETURNING id`, orgName).Scan(&out.orgID)
			if err != nil {
				return nil, fmt.Errorf("insert org: %w", err)
			}
		} else {
			return nil, fmt.Errorf("select org: %w", err)
		}
	}

	// 2) Business Unit
	const buName = "HQ"
	err = tx.QueryRow(ctx, `SELECT id FROM business_units WHERE organization_id=$1 AND name=$2`, out.orgID, buName).Scan(&out.buID)
	if err != nil {
		if err == pgx.ErrNoRows {
			err = tx.QueryRow(ctx, `INSERT INTO business_units(organization_id, name) VALUES ($1,$2) RETURNING id`, out.orgID, buName).Scan(&out.buID)
			if err != nil {
				return nil, fmt.Errorf("insert bu: %w", err)
			}
		} else {
			return nil, fmt.Errorf("select bu: %w", err)
		}
	}

	// 3) User (dev admin)
	const email = "dev@ingen.local"
	const displayName = "Dev Admin"
	const plainPassword = "password"

	err = tx.QueryRow(ctx, `SELECT id FROM users WHERE email=$1`, email).Scan(&out.userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			hash, herr := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
			if herr != nil {
				return nil, fmt.Errorf("hash pw: %w", herr)
			}
			err = tx.QueryRow(ctx, `INSERT INTO users(organization_id, business_unit_id, email, password_hash, display_name)
                                     VALUES ($1,$2,$3,$4,$5) RETURNING id`,
				out.orgID, out.buID, email, string(hash), displayName,
			).Scan(&out.userID)
			if err != nil {
				return nil, fmt.Errorf("insert user: %w", err)
			}
		} else {
			return nil, fmt.Errorf("select user: %w", err)
		}
	}

	// 4) Role: Admin and assignment
	const roleName = "Admin"
	err = tx.QueryRow(ctx, `SELECT id FROM roles WHERE organization_id=$1 AND name=$2`, out.orgID, roleName).Scan(&out.adminRoleID)
	if err != nil {
		if err == pgx.ErrNoRows {
			err = tx.QueryRow(ctx, `INSERT INTO roles(organization_id, name) VALUES ($1,$2) RETURNING id`, out.orgID, roleName).Scan(&out.adminRoleID)
			if err != nil {
				return nil, fmt.Errorf("insert role: %w", err)
			}
		} else {
			return nil, fmt.Errorf("select role: %w", err)
		}
	}
	// Assign role if not yet
	_, err = tx.Exec(ctx, `INSERT INTO user_roles(user_id, role_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, out.userID, out.adminRoleID)
	if err != nil {
		return nil, fmt.Errorf("assign role: %w", err)
	}

	// 5) Module entitlements for sales, service, collab
	_, err = tx.Exec(ctx, `INSERT INTO module_entitlements(organization_id, module) VALUES ($1,'sales') ON CONFLICT (organization_id, module) DO NOTHING`, out.orgID)
	if err != nil {
		return nil, fmt.Errorf("entitle sales: %w", err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO module_entitlements(organization_id, module) VALUES ($1,'service') ON CONFLICT (organization_id, module) DO NOTHING`, out.orgID)
	if err != nil {
		return nil, fmt.Errorf("entitle service: %w", err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO module_entitlements(organization_id, module) VALUES ($1,'collab') ON CONFLICT (organization_id, module) DO NOTHING`, out.orgID)
	if err != nil {
		return nil, fmt.Errorf("entitle collab: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	tx = nil
	return &out, nil
}

func main() {
	url := mustEnv()
	pool := connect(url)
	defer pool.Close()
	got, err := seed(pool)
	if err != nil {
		log.Fatalf("seed error: %v", err)
	}
	fmt.Println("Seed complete:")
	fmt.Printf("  Organization ID: %s\n", got.orgID)
	fmt.Printf("  Business Unit ID: %s\n", got.buID)
	fmt.Printf("  Admin User ID: %s\n", got.userID)
	fmt.Println("  Credentials: dev@ingen.local / password")
}
