package database

import (
	"context"
	"finance/internal/models"
	"time"

	"github.com/pkg/errors"
)

type TransactionDB interface {
	CreateTransaction(ctx context.Context, transaction *models.Transaction) error
	UpdateTransaction(ctx context.Context, transaction *models.Transaction) error
	GetTransactionByID(ctx context.Context, transactionID models.TransactionID) (*models.Transaction, error)
	ListTransactionByUserID(ctx context.Context, userID models.UserID, from, to time.Time) ([]*models.Transaction, error) //we will filter by selected time frame (current month, last month etc)
	ListTransactionByAccountID(ctx context.Context, accountID models.AccountID, from, to time.Time) ([]*models.Transaction, error)
	ListTransactionByCategoryID(ctx context.Context, categoryID models.CategoryID, from, to time.Time) ([]*models.Transaction, error)
	DeleteTransaction(ctx context.Context, transactionID models.TransactionID) (bool, error)
}

const createTransactionQuery = `
	INSERT INTO transactions (user_id, account_id, category_id, transaction_date, transaction_type, amount, notes)
	VALUES (:user_id, :account_id, :category_id, :transaction_date, :transaction_type, :amount, :notes)
	RETURNING transaction_id;
`

func (d *database) CreateTransaction(ctx context.Context, transaction *models.Transaction) error {
	rows, err := d.conn.NamedQueryContext(ctx, createTransactionQuery, transaction)
	if err != nil {
		return err
	}

	defer rows.Close()
	rows.Next()
	if err := rows.Scan(&transaction.ID); err != nil {
		return err
	}

	return nil
}

const updateTransactionQuery = `
	UPDATE transactions
	SET account_id = :account_id, category_id = :category_id, transaction_date = :transaction_date, transaction_type = :transaction_type, amount = :amount, notes = :notes
	WHERE transaction_id = :transaction_id;
`

func (d *database) UpdateTransaction(ctx context.Context, transaction *models.Transaction) error {
	result, err := d.conn.NamedExecContext(ctx, updateTransactionQuery, transaction)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		return errors.New("Transaction not found")
	}

	return nil
}

const getTransactionByIDQuery = `
	SELECT transaction_id, user_id, account_id, category_id, created_at, deleted_at, transaction_date, transaction_type, amount, notes
	FROM transactions
	WHERE transaction_id = $1;
`

func (d *database) GetTransactionByID(ctx context.Context, transactionID models.TransactionID) (*models.Transaction, error) {
	var transaction models.Transaction
	if err := d.conn.GetContext(ctx, &transaction, getTransactionByIDQuery, transactionID); err != nil {
		return nil, errors.Wrap(err, "could not get transaction")
	}
	return &transaction, nil
}

const listTransactioByUserIDQuery = `
	SELECT transaction_id, user_id, account_id, category_id, created_at, deleted_at, transaction_date, transaction_type, amount, notes
	FROM transactions
	WHERE user_id = $1 
				AND deleted_at IS NULL
				AND date > $2
				AND date < $3;
`
func (d *database) ListTransactionByUserID(ctx context.Context, userID models.UserID, from, to time.Time) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	if err := d.conn.SelectContext(ctx, &transactions, listTransactioByUserIDQuery, userID, from, to); err != nil {
		return nil, errors.Wrap(err, "could not get user's transactions")
	}
	return transactions, nil
}


const listTransactioByAccountIDQuery = `
	SELECT transaction_id, user_id, account_id, category_id, created_at, deleted_at, transaction_date, transaction_type, amount, notes
	FROM transactions
	WHERE account_id = $1 
				AND deleted_at IS NULL
				AND date > $2
				AND date < $3;
`
func (d *database) ListTransactionByAccountID(ctx context.Context, accountID models.AccountID, from, to time.Time) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	if err := d.conn.SelectContext(ctx, &transactions, listTransactioByAccountIDQuery, accountID, from, to); err != nil {
		return nil, errors.Wrap(err, "could not get account's transactions")
	}
	return transactions, nil
}


const listTransactioByACategoryQuery = `
	SELECT transaction_id, user_id, account_id, category_id, created_at, deleted_at, transaction_date, transaction_type, amount, notes
	FROM transactions
	WHERE category_id = $1 
				AND deleted_at IS NULL
				AND date > $2
				AND date < $3;
`
func (d *database) ListTransactionByCategoryID(ctx context.Context, categoryID models.CategoryID, from, to time.Time) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	if err := d.conn.SelectContext(ctx, &transactions, listTransactioByACategoryQuery, categoryID, from, to); err != nil {
		return nil, errors.Wrap(err, "could not get category's transactions")
	}
	return transactions, nil
}

const DeleteTransactionQuery = `
	UPDATE transactions
	SET deleted_at = NOW()
	WHERE transaction_id = $1 AND deleted_at IS NULL;
`
func (d *database) DeleteTransaction(ctx context.Context, transactionID models.TransactionID) (bool, error) {
	result, err := d.conn.ExecContext(ctx, DeleteTransactionQuery, transactionID)
	if err != nil {
		return false, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rows > 0, nil
}
