package models

import (
	"time"

	"github.com/pkg/errors"
)

// AccountID is an identifier for Account
type AccountID string

// NilAccountID is empty AccountID
var NilAccountID AccountID

// AccountType is type of account
type AccountType string

const (
	Cash   AccountType = "cash"
	Credit AccountType = "credit"
)

// Account is structure for Account
type Account struct {
	ID           AccountID    `json:"id,omitempty" db:"account_id"`
	UserID       *UserID      `json:"user_id,omitempty" db:"user_id"`
	Name         *string      `json:"name,omitempty" db:"account_name"`
	Type         *AccountType `json:"type,omitempty" db:"account_type"`
	StartBalance *int64       `json:"start_balance,omitempty" db:"start_balance"`
	Currency     *string      `json:"currency,omitempty" db:"currency"`
	CreatedAt    *time.Time   `json:"-" db:"created_at"`
	DeletedAt    *time.Time   `json:"-" db:"deleted_at"`
}

func (a *Account) Verify() error {
	if a.UserID == nil || len(*a.UserID) == 0 {
		return errors.New("user_id is required")
	}

	if a.Name == nil || len(*a.Name) == 0 {
		return errors.New("name is required")
	}

	if a.Type == nil || len(*a.Type) == 0 {
		return errors.New("type is required")
	}

	if a.StartBalance == nil {
		return errors.New("startBalance is required")
	}

	if a.Currency == nil || len(*a.Currency) == 0 {
		return errors.New("currency is required")
	}

	return nil
}
