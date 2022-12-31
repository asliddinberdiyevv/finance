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
	GetUserByID(ctx context.Context, userID *models.UserID) (models.User, error)
}

var ErrUserExists = errors.New("user with that email exists")

var createUserQuery = `
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

var getUserByIDQuery = `
	SELECT user_id, email, password_hash, created_at, deleted_at
	FROM users 
	WHERE user_id = $1;
`

func (d *database) GetUserByID(ctx context.Context, userID *models.UserID) (models.User, error) {
	var user models.User
	if err := d.conn.GetContext(ctx, &user, getUserByIDQuery, userID); err != nil {
		return user, err
	}

	return user, nil
}
