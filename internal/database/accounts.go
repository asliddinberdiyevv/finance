package database

import (
	"context"
	"finance/internal/models"

	"github.com/pkg/errors"
)

type AccountDB interface {
	CreateAccount(ctx context.Context, account *models.Account) error
	UpdateAccount(ctx context.Context, account *models.Account) error
	GetAccountByID(ctx context.Context, accountID models.AccountID) (*models.Account, error)
	ListAccountByUserID(ctx context.Context, userID models.UserID) ([]*models.Account, error)
	DeleteAccount(ctx context.Context, accountID models.AccountID) (bool, error)
}

const createAccountQuery = `
	INSERT INTO accounts (user_id, start_balance, account_type, account_name, currency)
		VALUES (:user_id, :start_balance, :account_type, :account_name, :currency)
	RETURNING account_id;
`
func (d *database) CreateAccount(ctx context.Context, account *models.Account) error {
	rows, err := d.conn.NamedQueryContext(ctx, createAccountQuery, account)
	if err != nil {
		return err
	}

	defer rows.Close()
	rows.Next()
	if err := rows.Scan(&account.ID); err != nil {
		return err
	}

	return nil
}

const UpdateAccountQuery = `
	UPDATE accounts
		SET start_balance = :start_balance,
				account_type = :account_type,
				account_name = :account_name,
				currency = :currency
		WHERE account_id = :account_id;
`
func (d *database) UpdateAccount(ctx context.Context, account *models.Account) error {
	result, err := d.conn.NamedExecContext(ctx, UpdateAccountQuery, account)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		return errors.New("Account not found")
	}

	return nil
}

const getAccountByIDQuery = `
	SELECT account_id, user_id, start_balance, account_type, account_name, currency, created_at, deleted_at
	FROM accounts
	WHERE account_id = $1;
`
func (d *database) GetAccountByID(ctx context.Context, accountID models.AccountID) (*models.Account, error) {
	var account models.Account
	if err := d.conn.GetContext(ctx, &account, getAccountByIDQuery, accountID); err != nil {
		return nil, errors.Wrap(err, "could not get account")
	}

	return &account, nil
}

const listAccountByIDQuery = `
	SELECT account_id, user_id, start_balance, account_type, account_name, currency, created_at, deleted_at
	FROM accounts
	WHERE user_id = $1 AND deleted_at IS NULL;
`
func (d *database) ListAccountByUserID(ctx context.Context, userID models.UserID) ([]*models.Account, error) {
	var accounts []*models.Account
	if err := d.conn.SelectContext(ctx, &accounts, listAccountByIDQuery, userID); err != nil {
		return nil, errors.Wrap(err, "could not get user's accounts")
	}

	return accounts, nil
}

const DeleteAccountQuery = `
	UPDATE accounts
	SET deleted_at = NOW()
	WHERE account_id = $1 AND deleted_at IS NULL;
`
func (d *database) DeleteAccount(ctx context.Context, accountID models.AccountID) (bool, error) {
	result, err := d.conn.ExecContext(ctx, DeleteAccountQuery, accountID)
	if err != nil {
		return false, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rows > 0, nil
}
