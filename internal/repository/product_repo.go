package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/ifortepay/ApiBackendTest/internal/database"
	"github.com/ifortepay/ApiBackendTest/internal/model"
)

type ProductRepository interface {
	GetProducts(ctx context.Context, executor database.DBExecutor, filter model.ProductFilter) ([]model.Product, int, error)
	GetProductBySKU(ctx context.Context, executor database.DBExecutor, sku string) (*model.Product, error)
	UpdateProductStock(ctx context.Context, executor database.DBExecutor, sku string, quantity int) error
	GetProductsBySKUs(ctx context.Context, executor database.DBExecutor, skus []string) ([]model.Product, error)
}

type productRepository struct{}

func NewProductRepository() ProductRepository {
	return &productRepository{}
}

func (r *productRepository) GetProducts(ctx context.Context, executor database.DBExecutor, filter model.ProductFilter) ([]model.Product, int, error) {
	var products []model.Product
	var total int

	baseQuery := "FROM products WHERE 1=1"
	countQuery := "SELECT COUNT(*) " + baseQuery
	selectQuery := "SELECT sku, name, price, stock_quantity, created_at, updated_at " + baseQuery

	args := []interface{}{}
	argCount := 1

	if filter.Search != "" {
		selectQuery += fmt.Sprintf(" AND (name ILIKE $%d OR sku ILIKE $%d)", argCount, argCount)
		countQuery += fmt.Sprintf(" AND (name ILIKE $%d OR sku ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+filter.Search+"%")
		argCount++
	}

	if filter.SKU != "" {
		selectQuery += fmt.Sprintf(" AND sku = $%d", argCount)
		countQuery += fmt.Sprintf(" AND sku = $%d", argCount)
		args = append(args, filter.SKU)
		argCount++
	}

	if filter.MinPrice > 0 {
		selectQuery += fmt.Sprintf(" AND price >= $%d", argCount)
		countQuery += fmt.Sprintf(" AND price >= $%d", argCount)
		args = append(args, filter.MinPrice)
		argCount++
	}

	if filter.MaxPrice > 0 {
		selectQuery += fmt.Sprintf(" AND price <= $%d", argCount)
		countQuery += fmt.Sprintf(" AND price <= $%d", argCount)
		args = append(args, filter.MaxPrice)
		argCount++
	}

	if err := executor.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	offset := (filter.Page - 1) * filter.Limit
	selectQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, filter.Limit, offset)

	if err := executor.SelectContext(ctx, &products, selectQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to get products: %w", err)
	}

	return products, total, nil
}

func (r *productRepository) GetProductBySKU(ctx context.Context, executor database.DBExecutor, sku string) (*model.Product, error) {
	var product model.Product
	query := "SELECT sku, name, price, stock_quantity, created_at, updated_at FROM products WHERE sku = $1"

	if err := executor.GetContext(ctx, &product, query, sku); err != nil {
		return nil, fmt.Errorf("failed to get product by SKU: %w", err)
	}

	return &product, nil
}

func (r *productRepository) UpdateProductStock(ctx context.Context, executor database.DBExecutor, sku string, quantity int) error {
	query := "UPDATE products SET stock_quantity = stock_quantity - $1, updated_at = NOW() WHERE sku = $2 AND stock_quantity >= $1"

	result, err := executor.ExecContext(ctx, query, quantity, sku)
	if err != nil {
		return fmt.Errorf("failed to update product stock: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("insufficient stock for product %s", sku)
	}

	return nil
}

func (r *productRepository) GetProductsBySKUs(ctx context.Context, executor database.DBExecutor, skus []string) ([]model.Product, error) {
	if len(skus) == 0 {
		return []model.Product{}, nil
	}

	placeholders := make([]string, len(skus))
	args := make([]interface{}, len(skus))
	for i, sku := range skus {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = sku
	}

	query := fmt.Sprintf("SELECT sku, name, price, stock_quantity, created_at, updated_at FROM products WHERE sku IN (%s)", strings.Join(placeholders, ","))

	var products []model.Product
	if err := executor.SelectContext(ctx, &products, query, args...); err != nil {
		return nil, fmt.Errorf("failed to get products by SKUs: %w", err)
	}

	return products, nil
}
