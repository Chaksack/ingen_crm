// Package authz implements the privilege-depth RBAC model from PRD §5.1:
// per-entity privileges (create/read/write/delete/assign), each granted at a
// depth (none/user/bu/bu_children/org). A user's effective depth for an
// entity+action is the most permissive depth across all roles they hold.
package authz

import (
	"context"
	"fmt"

	"ingencore/api/internal/db"
)

type Depth string

const (
	DepthNone       Depth = "none"
	DepthUser       Depth = "user"
	DepthBU         Depth = "bu"
	DepthBUChildren Depth = "bu_children"
	DepthOrg        Depth = "org"
)

var depthRank = map[Depth]int{
	DepthNone: 0, DepthUser: 1, DepthBU: 2, DepthBUChildren: 3, DepthOrg: 4,
}

func maxDepth(a, b Depth) Depth {
	if depthRank[a] >= depthRank[b] {
		return a
	}
	return b
}

// EffectiveDepth returns the most permissive depth across all roles held by
// userID for entity+action. A user with no matching privilege gets DepthNone.
func EffectiveDepth(ctx context.Context, q db.Queryer, userID, entity, action string) (Depth, error) {
	rows, err := q.Query(ctx, `
		SELECT rp.depth
		FROM user_roles ur
		JOIN role_privileges rp ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1 AND rp.entity = $2 AND rp.action = $3`,
		userID, entity, action,
	)
	if err != nil {
		return DepthNone, err
	}
	defer rows.Close()

	best := DepthNone
	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			return DepthNone, err
		}
		best = maxDepth(best, Depth(d))
	}
	return best, rows.Err()
}

// Scope is a depth resolved down to a concrete set of owner_user_ids a
// caller may act on, so handlers can filter with a plain
// `owner_user_id = ANY($n)` instead of building dynamic SQL per depth.
type Scope struct {
	// Restricted is false at org depth: no owner filter is needed at all.
	Restricted bool
	// OwnerIDs is the set of user ids in scope. An empty slice with
	// Restricted=true means no records are visible/actionable.
	OwnerIDs []string
}

// ScopedWhere returns an " AND owner_user_id = ANY($n)" SQL fragment (and
// its arg) when scope is restricted, or "" at org depth (no restriction
// needed). placeholder is the next free positional parameter index.
func ScopedWhere(scope Scope, placeholder int) (string, []any) {
	if !scope.Restricted {
		return "", nil
	}
	return fmt.Sprintf(" AND owner_user_id = ANY($%d)", placeholder), []any{scope.OwnerIDs}
}

// Guard resolves the caller's scope for entity+action in one call. ok=false
// means the caller holds no privilege at all (depth none) — callers should
// treat that as "nothing visible" for lists, or 404 for a single record,
// consistent with how disabled modules are hidden rather than 403'd.
func Guard(ctx context.Context, q db.Queryer, userID, entity, action string) (Scope, bool, error) {
	depth, err := EffectiveDepth(ctx, q, userID, entity, action)
	if err != nil {
		return Scope{}, false, err
	}
	if depth == DepthNone {
		return Scope{}, false, nil
	}
	scope, err := resolveScope(ctx, q, userID, depth)
	return scope, true, err
}

func resolveScope(ctx context.Context, q db.Queryer, userID string, depth Depth) (Scope, error) {
	switch depth {
	case DepthOrg:
		return Scope{Restricted: false}, nil
	case DepthUser:
		return Scope{Restricted: true, OwnerIDs: []string{userID}}, nil
	case DepthBU, DepthBUChildren:
		ids, err := usersInBUScope(ctx, q, userID, depth == DepthBUChildren)
		if err != nil {
			return Scope{}, err
		}
		return Scope{Restricted: true, OwnerIDs: ids}, nil
	default:
		return Scope{Restricted: true, OwnerIDs: []string{}}, nil
	}
}

func usersInBUScope(ctx context.Context, q db.Queryer, userID string, includeChildren bool) ([]string, error) {
	var buID *string
	if err := q.QueryRow(ctx, `SELECT business_unit_id FROM users WHERE id=$1`, userID).Scan(&buID); err != nil {
		return nil, err
	}
	if buID == nil {
		return []string{userID}, nil
	}

	buIDs := []string{*buID}
	if includeChildren {
		rows, err := q.Query(ctx, `
			WITH RECURSIVE descendants AS (
				SELECT id FROM business_units WHERE id = $1
				UNION ALL
				SELECT bu.id FROM business_units bu
				JOIN descendants d ON bu.parent_id = d.id
			)
			SELECT id FROM descendants`, *buID,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		buIDs = buIDs[:0]
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return nil, err
			}
			buIDs = append(buIDs, id)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	rows, err := q.Query(ctx, `SELECT id FROM users WHERE business_unit_id = ANY($1)`, buIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	userIDs := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, id)
	}
	return userIDs, rows.Err()
}
