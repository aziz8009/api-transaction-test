package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ifortepay/ApiBackendTest/internal/database"
	"github.com/ifortepay/ApiBackendTest/internal/model"
	"github.com/lib/pq"
)

type CartRepository interface {
	CreateCart(ctx context.Context, executor database.DBExecutor) (*model.Cart, error)
	GetCart(ctx context.Context, executor database.DBExecutor, cartID uuid.UUID) (*model.CartWithItems, error)
	GetActiveCart(ctx context.Context, executor database.DBExecutor) (*model.CartWithItems, error)
	AddItemToCart(ctx context.Context, executor database.DBExecutor, cartID uuid.UUID, item model.CartItem) error
	UpdateCartItem(ctx context.Context, executor database.DBExecutor, itemID uuid.UUID, quantity int) error
	RemoveCartItem(ctx context.Context, executor database.DBExecutor, itemID uuid.UUID) error
	UpdateCartTotals(ctx context.Context, executor database.DBExecutor, cartID uuid.UUID, totalAmount, discountAmount, finalAmount float64) error
	ClearCart(ctx context.Context, executor database.DBExecutor, cartID uuid.UUID) error
	GetCartItemByProduct(ctx context.Context, executor database.DBExecutor, cartID uuid.UUID, productSKU string) (*model.CartItem, error)
	GetCartItems(ctx context.Context, executor database.DBExecutor, cartID *uuid.UUID) ([]model.CartItem, error)
	GetCartItemByID(ctx context.Context, executor database.DBExecutor, itemID uuid.UUID) (*model.CartItem, error)
}

type cartRepository struct{}

func NewCartRepository() CartRepository {
	return &cartRepository{}
}

func (r *cartRepository) CreateCart(ctx context.Context, executor database.DBExecutor) (*model.Cart, error) {
	var cart model.Cart
	query := "INSERT INTO carts (total_amount, discount_amount, final_amount, status) VALUES (0, 0, 0, 'active') RETURNING id, total_amount, discount_amount, final_amount, status, created_at, updated_at"

	if err := executor.GetContext(ctx, &cart, query); err != nil {
		return nil, fmt.Errorf("failed to create cart: %w", err)
	}

	return &cart, nil
}

func (r *cartRepository) GetCart(ctx context.Context, executor database.DBExecutor, cartID uuid.UUID) (*model.CartWithItems, error) {
	var cart model.Cart
	query := "SELECT id, total_amount, discount_amount, final_amount, status, created_at, updated_at FROM carts WHERE id = $1"

	if err := executor.GetContext(ctx, &cart, query, cartID); err != nil {
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}

	items, err := r.GetCartItems(ctx, executor, &cartID)
	if err != nil {
		return nil, err
	}

	enrichedItems, err := r.enrichCartItems(ctx, executor, items)
	if err != nil {
		return nil, err
	}

	return &model.CartWithItems{
		Cart:  cart,
		Items: enrichedItems,
	}, nil
}

func (r *cartRepository) GetActiveCart(ctx context.Context, executor database.DBExecutor) (*model.CartWithItems, error) {
	var cart model.Cart
	query := "SELECT id, total_amount, discount_amount, final_amount, status, created_at, updated_at FROM carts WHERE status = 'active' ORDER BY created_at DESC LIMIT 1"

	err := executor.GetContext(ctx, &cart, query)
	if err != nil {
		// If no active cart, create one
		return r.createCartWithItems(ctx, executor)
	}

	items, err := r.GetCartItems(ctx, executor, &cart.ID)
	if err != nil {
		return nil, err
	}

	enrichedItems, err := r.enrichCartItems(ctx, executor, items)
	if err != nil {
		return nil, err
	}

	return &model.CartWithItems{
		Cart:  cart,
		Items: enrichedItems,
	}, nil
}

func (r *cartRepository) createCartWithItems(ctx context.Context, executor database.DBExecutor) (*model.CartWithItems, error) {
	cart, err := r.CreateCart(ctx, executor)
	if err != nil {
		return nil, err
	}

	return &model.CartWithItems{
		Cart:  *cart,
		Items: []model.CartItemDetail{},
	}, nil
}

func (r *cartRepository) AddItemToCart(ctx context.Context, executor database.DBExecutor, cartID uuid.UUID, item model.CartItem) error {
	query := `
		INSERT INTO cart_items (cart_id, product_sku, quantity, unit_price, total_price, discount_amount)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (cart_id, product_sku) 
		DO UPDATE SET 
			quantity = cart_items.quantity + $3,
			total_price = (cart_items.quantity + $3) * cart_items.unit_price,
			updated_at = NOW()
	`

	_, err := executor.ExecContext(ctx, query, cartID, item.ProductSKU, item.Quantity, item.UnitPrice, item.TotalPrice, item.DiscountAmount)
	if err != nil {
		return fmt.Errorf("failed to add item to cart: %w", err)
	}

	return nil
}

func (r *cartRepository) UpdateCartItem(ctx context.Context, executor database.DBExecutor, itemID uuid.UUID, quantity int) error {
	query := "UPDATE cart_items SET quantity = $1, total_price = quantity * unit_price, updated_at = NOW() WHERE id = $2"

	result, err := executor.ExecContext(ctx, query, quantity, itemID)
	if err != nil {
		return fmt.Errorf("failed to update cart item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("cart item not found")
	}

	return nil
}

func (r *cartRepository) RemoveCartItem(ctx context.Context, executor database.DBExecutor, itemID uuid.UUID) error {
	query := "DELETE FROM cart_items WHERE id = $1"

	result, err := executor.ExecContext(ctx, query, itemID)
	if err != nil {
		return fmt.Errorf("failed to remove cart item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("cart item not found")
	}

	return nil
}

func (r *cartRepository) UpdateCartTotals(ctx context.Context, executor database.DBExecutor, cartID uuid.UUID, totalAmount, discountAmount, finalAmount float64) error {
	query := "UPDATE carts SET total_amount = $1, discount_amount = $2, final_amount = $3, updated_at = NOW() WHERE id = $4"

	_, err := executor.ExecContext(ctx, query, totalAmount, discountAmount, finalAmount, cartID)
	if err != nil {
		return fmt.Errorf("failed to update cart totals: %w", err)
	}

	return nil
}

func (r *cartRepository) ClearCart(ctx context.Context, executor database.DBExecutor, cartID uuid.UUID) error {
	query := "DELETE FROM cart_items WHERE cart_id = $1"

	_, err := executor.ExecContext(ctx, query, cartID)
	if err != nil {
		return fmt.Errorf("failed to clear cart: %w", err)
	}

	return nil
}

func (r *cartRepository) GetCartItemByProduct(ctx context.Context, executor database.DBExecutor, cartID uuid.UUID, productSKU string) (*model.CartItem, error) {
	var item model.CartItem
	query := "SELECT id, cart_id, product_sku, quantity, unit_price, total_price, discount_amount, created_at, updated_at FROM cart_items WHERE cart_id = $1 AND product_sku = $2"

	if err := executor.GetContext(ctx, &item, query, cartID, productSKU); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("cart item not found")
		}

		return nil, fmt.Errorf("failed to get cart item")
	}

	return &item, nil
}

func (r *cartRepository) GetCartItems(ctx context.Context, executor database.DBExecutor, cartID *uuid.UUID) ([]model.CartItem, error) {
	var items []model.CartItem

	query := `
		SELECT id,cart_id,product_sku,quantity,unit_price,total_price,discount_amount,created_at,updated_at FROM cart_items
	`

	args := []interface{}{}

	if cartID != nil {
		query += " WHERE cart_id = $1"
		args = append(args, *cartID)
	}

	if err := executor.SelectContext(ctx, &items, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("cart item not found")
		}

		return nil, fmt.Errorf("failed to get cart item")
	}

	return items, nil
}

func (r *cartRepository) GetCartItemByID(
	ctx context.Context,
	executor database.DBExecutor,
	itemID uuid.UUID,
) (*model.CartItem, error) {
	var item model.CartItem

	query := `
		SELECT id, cart_id, product_sku, quantity, unit_price, total_price, discount_amount, created_at, updated_at
		FROM cart_items
		WHERE id = $1
	`

	if err := executor.GetContext(ctx, &item, query, itemID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("cart item not found")
		}

		return nil, fmt.Errorf("failed to get cart item")
	}

	return &item, nil
}

func (r *cartRepository) enrichCartItems(ctx context.Context, executor database.DBExecutor, items []model.CartItem) ([]model.CartItemDetail, error) {
	if len(items) == 0 {
		return []model.CartItemDetail{}, nil
	}

	skus := make([]string, len(items))
	for i, item := range items {
		skus[i] = item.ProductSKU
	}

	products, err := r.getProductsBySKUs(ctx, executor, skus)
	if err != nil {
		return nil, err
	}

	productMap := make(map[string]model.Product)
	for _, product := range products {
		productMap[product.SKU] = product
	}

	var enriched []model.CartItemDetail
	for _, item := range items {
		detail := model.CartItemDetail{
			ID:             item.ID,
			CartID:         item.CartID,
			ProductSKU:     item.ProductSKU,
			Quantity:       item.Quantity,
			UnitPrice:      item.UnitPrice,
			TotalPrice:     item.TotalPrice,
			DiscountAmount: item.DiscountAmount,
			FinalPrice:     item.TotalPrice - item.DiscountAmount,
		}
		if product, exists := productMap[item.ProductSKU]; exists {
			detail.ProductName = product.Name
		}
		enriched = append(enriched, detail)
	}

	return enriched, nil
}

func (r *cartRepository) getProductsBySKUs(ctx context.Context, executor database.DBExecutor, skus []string) ([]model.Product, error) {
	if len(skus) == 0 {
		return nil, nil // Mengembalikan nil slice lebih idiomatis di Go jika data kosong
	}

	query := `
        SELECT sku, name, price, stock_quantity, created_at, updated_at 
        FROM products 
        WHERE sku = ANY($1)
    `

	var products []model.Product

	if err := executor.SelectContext(ctx, &products, query, pq.Array(skus)); err != nil {
		return nil, fmt.Errorf("failed to get products by skus: %w", err)
	}

	return products, nil
}
