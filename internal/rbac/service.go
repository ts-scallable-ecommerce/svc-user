package rbac

import (
	"context"
	"database/sql"
)

// PermissionResolver resolves permissions for a user from persistent storage.
type PermissionResolver struct {
	db *sql.DB
}

// NewPermissionResolver constructs the resolver.
func NewPermissionResolver(db *sql.DB) *PermissionResolver {
	return &PermissionResolver{db: db}
}

// ResolvePermissions loads the permissions for a user identifier.
func (r *PermissionResolver) ResolvePermissions(ctx context.Context, userID string) ([]string, error) {
	const query = `
        SELECT p.name
        FROM permissions p
        JOIN role_permissions rp ON rp.perm_id = p.id
        JOIN user_roles ur ON ur.role_id = rp.role_id
        WHERE ur.user_id = $1`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		perms = append(perms, name)
	}
	return perms, rows.Err()
}
