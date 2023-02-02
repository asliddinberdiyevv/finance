package database

import (
	"context"
	"finance/internal/models"

	"github.com/pkg/errors"
)

type UserRoleDB interface {
	GrantRole(ctx context.Context, userID models.UserID, role models.Role) error
	RevokeRole(ctx context.Context, userID models.UserID, role models.Role) error
	GetRolesByUser(ctx context.Context, userID models.UserID) ([]*models.UserRole, error)
}

const grantUserRoleQuery = `
	INSERT INTO user_roles (user_id, role)
		VALUES ($1, $2);
`
func (d *database) GrantRole(ctx context.Context, userID models.UserID, role models.Role) error {
	if _, err := d.conn.ExecContext(ctx, grantUserRoleQuery, userID, role); err != nil {
		return errors.Wrap(err, "could not grant user role")
	}
	return nil
}

const revokeUserRoleQuery = `
	DELETE FROM user_roles 
	WHERE user_id = $1 AND role = $2;
`
func (d *database) RevokeRole(ctx context.Context, userID models.UserID, role models.Role) error {
	if _, err := d.conn.ExecContext(ctx, revokeUserRoleQuery, userID, role); err != nil {
		return errors.Wrap(err, "could not revoke user role")
	}
	return nil
}

const getRolesByUserIDQuery = `
	SELECT role
	FROM user_roles
	WHERE user_id = $1;
`
func (d *database) GetRolesByUser(ctx context.Context, userID models.UserID) ([]*models.UserRole, error) {
	var roles []*models.UserRole
	if err := d.conn.SelectContext(ctx, &roles, getRolesByUserIDQuery, userID); err != nil {
		return nil, errors.Wrap(err, "could not get user roles")
	}
	return roles, nil
}
