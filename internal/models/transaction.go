package models

import (
	"time"

	"github.com/pkg/errors"
)

// TransactionID is identifier of Transaction
type TransactionID string

// NilTransactionID is an empty identifier of Transaction
var NilTransactionID TransactionID

// TransactionType is string representation of Transction type
type TransactionType string

const (
	Income  TransactionType = "income"
	Expense TransactionType = "expense"
)

type Transaction struct {
	ID         TransactionID `json:"id,omitempty" db:"transaction_id"`
	UserID     *UserID       `json:"user_id,omitempty" db:"user_id"`
	AccountID  *AccountID    `json:"account_id,omitempty" db:"account_id"`
	CategoryID *CategoryID   `json:"category_id,omitempty" db:"category_id"`

	CreatedAt *time.Time `json:"-" db:"created_at"`
	DeletedAt *time.Time `json:"-" db:"deleted_at"`

	Date   *time.Time       `json:"date" db:"transaction_date"`
	Type   *TransactionType `json:"type" db:"transaction_type"`
	Amount *int64           `json:"amount" db:"amount"`
	Notes  string           `json:"notes,omitempty" db:"notes"`
}

func (c *Transaction) Verify() error {
	if c.UserID == nil || len(*c.UserID) == 0 {
		return errors.New("user_id is required")
	}

	if c.AccountID == nil || len(*c.AccountID) == 0 {
		return errors.New("account_id is required")
	}

	if c.CategoryID == nil || len(*c.CategoryID) == 0 {
		return errors.New("category_id is required")
	}

	if c.Date == nil {
		return errors.New("date is required")
	}

	if c.Type == nil || len(*c.Type) == 0 {
		return errors.New("type is required")
	}

	if c.Amount == nil {
		return errors.New("amount is required")
	}

	return nil
}
