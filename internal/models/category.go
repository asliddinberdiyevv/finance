package models

import (
	"time"

	"github.com/pkg/errors"
)

// CategoryID is identifier of Category
type CategoryID string

// NilCategoryID is identifier of Category
var NilCategoryID CategoryID

type Category struct {
	ID        CategoryID `json:"id" db:"category_id"`
	ParentID  CategoryID `json:"parent_id,omitempty" db:"parent_id"`
	UserID    *UserID    `json:"user_id,omitempty" db:"user_id"`
	CreatedAt *time.Time `json:"-" db:"created_at"`
	DeletedAt *time.Time `json:"-" db:"deleted_at"`
	Name      *string    `json:"name,omitempty" db:"name"`
}

func (c *Category) Verify() error {
	if c.UserID == nil || len(*c.UserID) == 0 {
		return errors.New("user_id is required")
	}

	if c.Name == nil || len(*c.Name) == 0 {
		return errors.New("name is required")
	}

	return nil
}
