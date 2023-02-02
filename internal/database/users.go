package database

import (
	"context"
	"finance/internal/models"

	"github.com/lib/pq"
	"github.com/pkg/errors"
)

// UserDB persist Users.
type UsersDB interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, userID models.UserID) (*models.User, error)
	GetUserByEmail(ctx context.Context, emial string) (*models.User, error)
	ListUsers(ctx context.Context) ([]*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, userID models.UserID) (bool, error)
}

var ErrUserExists = errors.New("user with that email exists")

const createUserQuery = `
	INSERT INTO users (
		email, password_hash
	)
	VALUES (
		:email, :password_hash
	)
	RETURNING user_id
`
func (d *database) CreateUser(ctx context.Context, user *models.User) error {
	rows, err := d.conn.NamedQueryContext(ctx, createUserQuery, user)
	if rows != nil {
		defer rows.Close()
	}

	if err != nil {
		if pqError, ok := err.(*pq.Error); ok {
			if pqError.Code.Name() == UniqueViolation {
				if pqError.Constraint == "user_email" {
					return ErrUserExists
				}
			}
		}
		return errors.Wrap(err, "could not create user")
	}

	rows.Next()

	if err := rows.Scan(&user.ID); err != nil {
		return errors.Wrap(err, "could not get created userID")
	}

	return nil
}

const getUserByIDQuery = `
	SELECT user_id, email, password_hash, created_at
	FROM users 
	WHERE user_id = $1 AND deleted_at IS NULL;
`
func (d *database) GetUserByID(ctx context.Context, userID models.UserID) (*models.User, error) {
	var user models.User
	if err := d.conn.GetContext(ctx, &user, getUserByIDQuery, userID); err != nil {
		return nil, err
	}

	return &user, nil
}

const getUserByEmailQuery = `
	SELECT user_id, email, password_hash, created_at
	FROM users 
	WHERE email = $1 AND deleted_at IS NULL;
`
func (d *database) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := d.conn.GetContext(ctx, &user, getUserByEmailQuery, email); err != nil {
		return nil, err
	}

	return &user, nil
}

const listUsersQuery = `
	SELECT user_id, email, password_hash, created_at
	FROM users
	WHERE deleted_at IS NULL;
`
func (d *database) ListUsers(ctx context.Context) ([]*models.User, error) {
	var users []*models.User
	if err := d.conn.SelectContext(ctx, &users, listUsersQuery); err != nil {
		return nil, errors.Wrap(err, "could not get users")
	}
	return users, nil
}

const updateUserQuery = `
	UPDATE users
	SET	password_hash = :password_hash
	WHERE user_id = :user_id;
`
func (d *database) UpdateUser(ctx context.Context, user *models.User) error {
	result, err := d.conn.NamedExecContext(ctx, updateUserQuery, user)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		return errors.New("User not found")
	}

	return nil
}

const DeleteUserQuery = `
	UPDATE users
	SET deleted_at = NOW(),
			email = CONCAT(email, '-DELETED-', uuid_generate_v4())
	WHERE user_id = $1 AND deleted_at IS NULL;
`
func (d *database) DeleteUser(ctx context.Context, userID models.UserID) (bool, error) {
	result, err := d.conn.ExecContext(ctx, DeleteUserQuery, userID)
	if err != nil {
		return false, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rows > 0, nil
}
