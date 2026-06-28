package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ifortepay/ApiBackendTest/internal/database"
	"github.com/ifortepay/ApiBackendTest/internal/model"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, executor database.DBExecutor, order model.Order) (*model.Order, error)
	CreateOrderItems(ctx context.Context, executor database.DBExecutor, items []model.OrderItem) error
	GetOrder(ctx context.Context, executor database.DBExecutor, orderID uuid.UUID) (*model.OrderWithItems, error)
	GetOrders(ctx context.Context, executor database.DBExecutor, filter model.OrderFilter) ([]model.Order, int, error)
	UpdateOrderStatus(ctx context.Context, executor database.DBExecutor, orderID uuid.UUID, status string) error
	GetOrderByIdempotencyKey(ctx context.Context, executor database.DBExecutor, key string) (*model.Order, error)
}

type orderRepository struct{}

func NewOrderRepository() OrderRepository {
	return &orderRepository{}
}

func (r *orderRepository) CreateOrder(ctx context.Context, executor database.DBExecutor, order model.Order) (*model.Order, error) {
	var createdOrder model.Order
	query := `
		INSERT INTO orders (order_number, user_id, grand_total, discount_total, status, shipping_address, idempotency_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, order_number, user_id, grand_total, discount_total, status, shipping_address, created_at, updated_at
	`

	err := executor.GetContext(ctx, &createdOrder, query,
		order.OrderNumber,
		order.UserID,
		order.GrandTotal,
		order.DiscountTotal,
		order.Status,
		order.ShippingAddress,
		order.IdempotencyKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return &createdOrder, nil
}

func (r *orderRepository) CreateOrderItems(ctx context.Context, executor database.DBExecutor, items []model.OrderItem) error {
	if len(items) == 0 {
		return nil
	}

	query := `
		INSERT INTO order_items (order_id, product_sku, quantity, unit_price, discount, final_price, promotion_applied, promotion_details)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	for _, item := range items {
		_, err := executor.ExecContext(ctx, query,
			item.OrderID,
			item.ProductSKU,
			item.Quantity,
			item.UnitPrice,
			item.Discount,
			item.FinalPrice,
			item.PromotionApplied,
			item.PromotionDetails,
		)
		if err != nil {
			return fmt.Errorf("failed to create order item: %w", err)
		}
	}

	return nil
}

func (r *orderRepository) GetOrder(ctx context.Context, executor database.DBExecutor, orderID uuid.UUID) (*model.OrderWithItems, error) {
	var order model.Order
	query := "SELECT id, order_number, user_id, grand_total, discount_total, status, shipping_address, created_at, updated_at FROM orders WHERE id = $1"

	if err := executor.GetContext(ctx, &order, query, orderID); err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	var items []model.OrderItem
	itemsQuery := "SELECT id, order_id, product_sku, quantity, unit_price, discount, final_price, promotion_applied, promotion_details, created_at, updated_at FROM order_items WHERE order_id = $1"

	if err := executor.SelectContext(ctx, &items, itemsQuery, orderID); err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}

	return &model.OrderWithItems{
		Order: order,
		Items: items,
	}, nil
}

func (r *orderRepository) GetOrders(ctx context.Context, executor database.DBExecutor, filter model.OrderFilter) ([]model.Order, int, error) {
	var orders []model.Order
	var total int

	baseQuery := "FROM orders WHERE 1=1"
	countQuery := "SELECT COUNT(*) " + baseQuery
	selectQuery := "SELECT id, order_number, user_id, grand_total, discount_total, status, shipping_address, created_at, updated_at " + baseQuery

	args := []interface{}{}
	argCount := 1

	if filter.Status != "" {
		selectQuery += fmt.Sprintf(" AND status = $%d", argCount)
		countQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, filter.Status)
		argCount++
	}

	if filter.StartDate != nil {
		selectQuery += fmt.Sprintf(" AND created_at >= $%d", argCount)
		countQuery += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, filter.StartDate)
		argCount++
	}

	if filter.EndDate != nil {
		selectQuery += fmt.Sprintf(" AND created_at <= $%d", argCount)
		countQuery += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, filter.EndDate)
		argCount++
	}

	if err := executor.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
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

	if err := executor.SelectContext(ctx, &orders, selectQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to get orders: %w", err)
	}

	return orders, total, nil
}

func (r *orderRepository) UpdateOrderStatus(ctx context.Context, executor database.DBExecutor, orderID uuid.UUID, status string) error {
	query := "UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2"

	result, err := executor.ExecContext(ctx, query, status, orderID)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}

func (r *orderRepository) GetOrderByIdempotencyKey(ctx context.Context, executor database.DBExecutor, key string) (*model.Order, error) {
	var order model.Order
	query := "SELECT id, order_number, user_id, grand_total, discount_total, status, shipping_address, created_at, updated_at FROM orders WHERE idempotency_key = $1"

	err := executor.GetContext(ctx, &order, query, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get order by idempotency key: %w", err)
	}

	return &order, nil
}
