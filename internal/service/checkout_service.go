package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ifortepay/ApiBackendTest/internal/database"
	"github.com/ifortepay/ApiBackendTest/internal/model"
	"github.com/ifortepay/ApiBackendTest/internal/repository"
)

type CheckoutService interface {
	ProcessCheckout(ctx context.Context, cartID uuid.UUID, userID uuid.UUID, shippingAddress, paymentMethod, idempotencyKey string) (*model.OrderWithItems, error)
	ConfirmPayment(ctx context.Context, orderID uuid.UUID, paymentStatus, transactionRef string) error
	GetOrder(ctx context.Context, orderID uuid.UUID) (*model.OrderWithItems, error)
	GetOrders(ctx context.Context, filter model.OrderFilter) ([]model.Order, int, error)
	UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string) error
}

type checkoutService struct {
	cartRepo     repository.CartRepository
	productRepo  repository.ProductRepository
	orderRepo    repository.OrderRepository
	promotionSvc PromotionService
	txManager    database.TransactionManager
	db           *database.PostgresDB
}

func NewCheckoutService(
	cartRepo repository.CartRepository,
	productRepo repository.ProductRepository,
	orderRepo repository.OrderRepository,
	promotionSvc PromotionService,
	txManager database.TransactionManager,
	db *database.PostgresDB,
) CheckoutService {
	return &checkoutService{
		cartRepo:     cartRepo,
		productRepo:  productRepo,
		orderRepo:    orderRepo,
		promotionSvc: promotionSvc,
		txManager:    txManager,
		db:           db,
	}
}

func (s *checkoutService) ProcessCheckout(
	ctx context.Context,
	cartID uuid.UUID,
	userID uuid.UUID,
	shippingAddress string,
	paymentMethod string,
	idempotencyKey string,
) (*model.OrderWithItems, error) {
	// Check idempotency
	executor := &database.SqlxDBExecutor{DB: s.db.GetDB()}
	existingOrder, err := s.orderRepo.GetOrderByIdempotencyKey(ctx, executor, idempotencyKey)
	if err == nil && existingOrder != nil {
		return s.orderRepo.GetOrder(ctx, executor, existingOrder.ID)
	}

	var order *model.OrderWithItems
	err = s.txManager.DoWithExecutor(ctx, func(executor database.DBExecutor) error {
		// Get cart
		cart, err := s.cartRepo.GetCart(ctx, executor, cartID)
		if err != nil {

			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("cart id not found")
			}

			return fmt.Errorf("failed to get cart item")
		}

		if cart.Status != "active" {
			return fmt.Errorf("cart is not active")
		}

		if len(cart.Items) == 0 {
			return fmt.Errorf("cart is empty")
		}

		// Get product details and validate stock
		skus := make([]string, len(cart.Items))
		for i, item := range cart.Items {
			skus[i] = item.ProductSKU
		}

		products, err := s.productRepo.GetProductsBySKUs(ctx, executor, skus)
		if err != nil {
			return fmt.Errorf("failed to get products: %w", err)
		}

		productMap := make(map[string]model.Product)
		for _, p := range products {
			productMap[p.SKU] = p
		}

		// Validate stock and calculate totals
		var grandTotal float64
		var discountTotal float64
		orderItems := []model.OrderItem{}

		// Apply promotions
		promotionResult, err := s.promotionSvc.ApplyPromotions(ctx, cart)
		if err != nil {
			return fmt.Errorf("failed to apply promotions: %w", err)
		}

		for _, cartItem := range cart.Items {
			product, exists := productMap[cartItem.ProductSKU]
			if !exists {
				return fmt.Errorf("product not found: %s", cartItem.ProductSKU)
			}

			if product.StockQuantity < cartItem.Quantity {
				return fmt.Errorf("insufficient stock for product %s: available %d, requested %d",
					cartItem.ProductSKU, product.StockQuantity, cartItem.Quantity)
			}

			// Calculate discount
			discount := promotionResult.Discounts[cartItem.ProductSKU]
			if discount > 0 {
				discountTotal += discount
			}

			finalPrice := cartItem.TotalPrice - discount

			// Prepare order item
			orderItem := model.OrderItem{
				OrderID:          uuid.Nil, // Will be set after order creation
				ProductSKU:       cartItem.ProductSKU,
				Quantity:         cartItem.Quantity,
				UnitPrice:        cartItem.UnitPrice,
				Discount:         discount,
				FinalPrice:       finalPrice,
				PromotionApplied: promotionResult.Promotions[cartItem.ProductSKU],
			}

			if orderItem.PromotionApplied != "" {
				promoDetailJSON, _ := json.Marshal(promotionResult.PromotionDetail)
				orderItem.PromotionDetails = string(promoDetailJSON)
			}

			orderItems = append(orderItems, orderItem)
			grandTotal += finalPrice

			// Update stock
			if err := s.productRepo.UpdateProductStock(ctx, executor, cartItem.ProductSKU, cartItem.Quantity); err != nil {
				return fmt.Errorf("failed to update stock for product %s: %w", cartItem.ProductSKU, err)
			}
		}

		// Create order
		orderNumber := fmt.Sprintf("ORD-%s-%d", time.Now().Format("20060102"), time.Now().UnixNano()%10000)
		newOrder := model.Order{
			OrderNumber:     orderNumber,
			UserID:          userID,
			GrandTotal:      grandTotal,
			DiscountTotal:   discountTotal,
			Status:          "pending_payment",
			ShippingAddress: shippingAddress,
			IdempotencyKey:  idempotencyKey,
		}

		createdOrder, err := s.orderRepo.CreateOrder(ctx, executor, newOrder)
		if err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}

		// Update order items with order ID
		for i := range orderItems {
			orderItems[i].OrderID = createdOrder.ID
		}

		if err := s.orderRepo.CreateOrderItems(ctx, executor, orderItems); err != nil {
			return fmt.Errorf("failed to create order items: %w", err)
		}

		// Mark cart as completed
		if err := s.cartRepo.UpdateCartTotals(ctx, executor, cartID, cart.TotalAmount, discountTotal, grandTotal); err != nil {
			return fmt.Errorf("failed to update cart totals: %w", err)
		}

		// Get complete order
		order, err = s.orderRepo.GetOrder(ctx, executor, createdOrder.ID)
		if err != nil {
			return fmt.Errorf("failed to get order: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return order, nil
}

func (s *checkoutService) ConfirmPayment(ctx context.Context, orderID uuid.UUID, paymentStatus, transactionRef string) error {
	executor := &database.SqlxDBExecutor{DB: s.db.GetDB()}

	if paymentStatus == "success" {
		if err := s.orderRepo.UpdateOrderStatus(ctx, executor, orderID, "paid"); err != nil {
			return err
		}
	} else {
		if err := s.orderRepo.UpdateOrderStatus(ctx, executor, orderID, "payment_failed"); err != nil {
			return err
		}
	}

	return nil
}

func (s *checkoutService) GetOrder(ctx context.Context, orderID uuid.UUID) (*model.OrderWithItems, error) {
	executor := &database.SqlxDBExecutor{DB: s.db.GetDB()}
	return s.orderRepo.GetOrder(ctx, executor, orderID)
}

func (s *checkoutService) GetOrders(ctx context.Context, filter model.OrderFilter) ([]model.Order, int, error) {
	executor := &database.SqlxDBExecutor{DB: s.db.GetDB()}
	return s.orderRepo.GetOrders(ctx, executor, filter)
}

func (s *checkoutService) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string) error {
	executor := &database.SqlxDBExecutor{DB: s.db.GetDB()}
	return s.orderRepo.UpdateOrderStatus(ctx, executor, orderID, status)
}
