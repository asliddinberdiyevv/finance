package database

import (
	"context"
	"finance/internal/models"
	"fmt"

	"github.com/pkg/errors"
)

type MerchantDB interface {
	CreateMerchant(ctx context.Context, merchant *models.Merchant) error
	UpdateMerchant(ctx context.Context, merchant *models.Merchant) error
	GetMerchantByID(ctx context.Context, merchantID models.MerchantID) (*models.Merchant, error)
	ListMerchantByUserID(ctx context.Context, userID models.UserID) ([]*models.Merchant, error)
	DeleteMerchant(ctx context.Context, merchantID models.MerchantID) (bool, error)
}

var createMerchantQuery = `
	INSERT INTO merchants (user_id, name)
	VALUES (:user_id, :name)
	RETURNING merchant_id;
`

func (d *database) CreateMerchant(ctx context.Context, merchant *models.Merchant) error {
	rows, err := d.conn.NamedQueryContext(ctx, createMerchantQuery, merchant)
	if err != nil {
		return err
	}

	defer rows.Close()
	rows.Next()
	if err := rows.Scan(&merchant.ID); err != nil {
		return err
	}

	return nil
}

var updateMerchantQuery = `
	UPDATE merchants
	SET name = :name
	WHERE merchant_id = :merchant_id;
`

func (d *database) UpdateMerchant(ctx context.Context, merchant *models.Merchant) error {
	result, err := d.conn.NamedExecContext(ctx, updateMerchantQuery, merchant)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		return errors.New("Merchant not found")
	}

	return nil
}

var getMerchantByIDQuery = `
	SELECT merchant_id, user_id, name, created_at, deleted_at
	FROM merchants
	WHERE merchant_id = $1;
`

func (d *database) GetMerchantByID(ctx context.Context, merchantID models.MerchantID) (*models.Merchant, error) {
	var merchant models.Merchant
	if err := d.conn.GetContext(ctx, &merchant, getMerchantByIDQuery, merchantID); err != nil {
		fmt.Println(err)
		fmt.Println(err)
		fmt.Println(err)
		return nil, errors.Wrap(err, "could not get merchant")
	}
	return &merchant, nil
}

var listMerchantByIDQuery = `
	SELECT merchant_id, user_id, name, created_at, deleted_at
	FROM merchants
	WHERE user_id = $1 AND deleted_at IS NULL;
`

func (d *database) ListMerchantByUserID(ctx context.Context, userID models.UserID) ([]*models.Merchant, error) {
	var merchants []*models.Merchant
	if err := d.conn.SelectContext(ctx, &merchants, listMerchantByIDQuery, userID); err != nil {
		return nil, errors.Wrap(err, "could not get user's merchants")
	}
	return merchants, nil
}

var DeleteMerchantQuery = `
	UPDATE merchants
	SET deleted_at = NOW()
	WHERE merchant_id = $1 AND deleted_at IS NULL;
`

func (d *database) DeleteMerchant(ctx context.Context, merchantID models.MerchantID) (bool, error) {
	result, err := d.conn.ExecContext(ctx, DeleteMerchantQuery, merchantID)
	if err != nil {
		return false, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rows > 0, nil
}
