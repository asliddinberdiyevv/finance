package database

import (
	"context"
	"finance/internal/models"
	"finance/internal/utils"

	"github.com/pkg/errors"
)

type CategoryDB interface {
	CreateCategory(ctx context.Context, category *models.Category) error
	UpdateCategory(ctx context.Context, category *models.Category) error
	GetCategoryByID(ctx context.Context, categoryID models.CategoryID) (*models.Category, error)
	ListCategoryByUserID(ctx context.Context, userID models.UserID) ([]*models.Category, error)
	DeleteCategory(ctx context.Context, categoryID models.CategoryID) (bool, error)
}

var createCategoryQuery = `
	INSERT INTO categories (parent_id, user_id, name)
		VALUES (:parent_id, :user_id, :name)
	RETURNING category_id;
`

func (d *database) CreateCategory(ctx context.Context, category *models.Category) error {
	rows, err := d.conn.NamedQueryContext(ctx, createCategoryQuery, category)
	if err != nil {
		return err
	}

	defer rows.Close()
	rows.Next()
	if err := rows.Scan(&category.ID); err != nil {
		return err
	}

	return nil
}

var updateCategoryQuery = `
	UPDATE categories
		SET parent_id = :parent_id,
				name = :name
		WHERE category_id = :category_id;
`

func (d *database) UpdateCategory(ctx context.Context, category *models.Category) error {
	result, err := d.conn.NamedExecContext(ctx, updateCategoryQuery, category)
	utils.CheckError(err)

	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		return errors.New("Category not found")
	}

	return nil
}

var getCategoryByIDQuery = `
	SELECT category_id, parent_id, user_id, name, created_at, deleted_at
	FROM categories
	WHERE category_id = $1;
`

func (d *database) GetCategoryByID(ctx context.Context, categoryID models.CategoryID) (*models.Category, error) {
	var category models.Category
	if err := d.conn.GetContext(ctx, &category, getCategoryByIDQuery, categoryID); err != nil {
		return nil, errors.Wrap(err, "could not get category")
	}
	return &category, nil
}

var listCategoryByIDQuery = `
	SELECT category_id, parent_id, user_id, name, created_at, deleted_at
	FROM categories
	WHERE user_id = $1 AND deleted_at IS NULL;
`

func (d *database) ListCategoryByUserID(ctx context.Context, userID models.UserID) ([]*models.Category, error) {
	var categories []*models.Category
	if err := d.conn.SelectContext(ctx, &categories, listCategoryByIDQuery, userID); err != nil {
		return nil, errors.Wrap(err, "could not get user's categories")
	}
	return categories, nil
}

var DeleteCategoryQuery = `
	UPDATE categories
	SET deleted_at = NOW()
	WHERE category_id = $1 AND deleted_at IS NULL;
`

func (d *database) DeleteCategory(ctx context.Context, categoryID models.CategoryID) (bool, error) {
	result, err := d.conn.ExecContext(ctx, DeleteCategoryQuery, categoryID)
	if err != nil {
		return false, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rows > 0, nil
}
