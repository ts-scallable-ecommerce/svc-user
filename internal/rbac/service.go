package rbac

import (
	"context"
	"database/sql"
)

// Service encapsulates RBAC queries against the relational database.
type Service struct {
	db *sql.DB
}

// NewService constructs the RBAC service implementation.
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// AssignRole associates a role with the specified user.
func (s *Service) AssignRole(ctx context.Context, userID, role string) error {
	const query = `INSERT INTO user_roles (user_id, role_id)
SELECT $1, r.id FROM roles r WHERE r.name = $2
ON CONFLICT (user_id, role_id) DO NOTHING`
	_, err := s.db.ExecContext(ctx, query, userID, role)
	return err
}

// ListRoles returns the role names assigned to a user.
func (s *Service) ListRoles(ctx context.Context, userID string) ([]string, error) {
	const query = `SELECT r.name FROM roles r
JOIN user_roles ur ON ur.role_id = r.id
WHERE ur.user_id = $1`
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		roles = append(roles, name)
	}
	return roles, rows.Err()
}

// ResolvePermissions loads the permissions for a user identifier.
func (s *Service) ResolvePermissions(ctx context.Context, userID string) ([]string, error) {
	const query = `SELECT p.name FROM permissions p
JOIN role_permissions rp ON rp.perm_id = p.id
JOIN user_roles ur ON ur.role_id = rp.role_id
WHERE ur.user_id = $1`

	rows, err := s.db.QueryContext(ctx, query, userID)
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

// HasPermission checks whether a user has the given permission by name.
func (s *Service) HasPermission(ctx context.Context, userID, permission string) (bool, error) {
	const query = `SELECT EXISTS (
SELECT 1 FROM permissions p
JOIN role_permissions rp ON rp.perm_id = p.id
JOIN user_roles ur ON ur.role_id = rp.role_id
WHERE ur.user_id = $1 AND p.name = $2)`

	var exists bool
	if err := s.db.QueryRowContext(ctx, query, userID, permission).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}
