package models

import (
	"time"

	"github.com/pkg/errors"
)

// MerchantID is identifier of Merchant
type MerchantID string

// NilMerchantID is an empty identifier of Merchant
var NilMerchantID MerchantID

type Merchant struct {
	ID        MerchantID `json:"id,omitempty" db:"merchant_id"`
	UserID    *UserID    `json:"user_id,omitempty" db:"user_id"`
	CreatedAt *time.Time `json:"-" db:"created_at"`
	DeletedAt *time.Time `json:"-" db:"deleted_at"`
	Name      *string    `json:"name,omitempty" db:"name"`
}

func (c *Merchant) Verify() error {
	if c.UserID == nil || len(*c.UserID) == 0 {
		return errors.New("user_id is required")
	}

	if c.Name == nil || len(*c.Name) == 0 {
		return errors.New("name is required")
	}

	return nil
}
