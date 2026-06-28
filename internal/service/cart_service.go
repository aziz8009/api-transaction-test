package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ifortepay/ApiBackendTest/internal/database"
	"github.com/ifortepay/ApiBackendTest/internal/model"
	"github.com/ifortepay/ApiBackendTest/internal/repository"
)

type CartService interface {
	GetActiveCart(ctx context.Context) (*model.CartWithItems, error)
	AddToCart(ctx context.Context, productSKU string, quantity int) (*model.CartWithItems, error)
	UpdateCartItem(ctx context.Context, itemID uuid.UUID, quantity int) (*model.CartWithItems, error)
	RemoveCartItem(ctx context.Context, itemID uuid.UUID) (*model.CartWithItems, error)
	GetCart(ctx context.Context, cartID uuid.UUID) (*model.CartWithItems, error)
	CalculateCartTotal(ctx context.Context, cart *model.CartWithItems) (float64, float64, float64, error)
}

type cartService struct {
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
	txManager   database.TransactionManager
	db          *database.PostgresDB
}

func NewCartService(
	cartRepo repository.CartRepository,
	productRepo repository.ProductRepository,
	txManager database.TransactionManager,
	db *database.PostgresDB,
) CartService {
	return &cartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
		txManager:   txManager,
		db:          db,
	}
}

func (s *cartService) GetActiveCart(ctx context.Context) (*model.CartWithItems, error) {
	executor := &database.SqlxDBExecutor{DB: s.db.GetDB()}
	return s.cartRepo.GetActiveCart(ctx, executor)
}

func (s *cartService) AddToCart(ctx context.Context, productSKU string, quantity int) (*model.CartWithItems, error) {
	executor := &database.SqlxDBExecutor{DB: s.db.GetDB()}

	// Get product
	product, err := s.productRepo.GetProductBySKU(ctx, executor, productSKU)
	if err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// Check stock
	if product.StockQuantity < quantity {
		return nil, fmt.Errorf("insufficient stock: available %d, requested %d", product.StockQuantity, quantity)
	}

	// Get or create active cart
	cart, err := s.cartRepo.GetActiveCart(ctx, executor)
	if err != nil {
		return nil, err
	}

	// Add item to cart
	item := model.CartItem{
		CartID:     cart.ID,
		ProductSKU: productSKU,
		Quantity:   quantity,
		UnitPrice:  product.Price,
		TotalPrice: product.Price * float64(quantity),
	}

	if err := s.cartRepo.AddItemToCart(ctx, executor, cart.ID, item); err != nil {
		return nil, err
	}

	// Get updated cart
	updatedCart, err := s.cartRepo.GetCart(ctx, executor, cart.ID)
	if err != nil {
		return nil, err
	}

	// Calculate and update totals
	totalAmount, discountAmount, finalAmount, err := s.CalculateCartTotal(ctx, updatedCart)
	if err != nil {
		return nil, err
	}

	if err := s.cartRepo.UpdateCartTotals(ctx, executor, cart.ID, totalAmount, discountAmount, finalAmount); err != nil {
		return nil, err
	}

	return s.cartRepo.GetCart(ctx, executor, cart.ID)
}

func (s *cartService) UpdateCartItem(ctx context.Context, itemID uuid.UUID, quantity int) (*model.CartWithItems, error) {
	executor := &database.SqlxDBExecutor{DB: s.db.GetDB()}

	if err := s.cartRepo.UpdateCartItem(ctx, executor, itemID, quantity); err != nil {
		return nil, err
	}

	items, err := s.cartRepo.GetCartItemByID(ctx, executor, itemID)
	if err != nil {
		return nil, err
	}

	var cartID uuid.UUID
	cartID = items.CartID

	cart, err := s.cartRepo.GetCart(ctx, executor, cartID)
	if err != nil {
		return nil, err
	}

	totalAmount, discountAmount, finalAmount, err := s.CalculateCartTotal(ctx, cart)
	if err != nil {
		return nil, err
	}

	if err := s.cartRepo.UpdateCartTotals(
		ctx,
		executor,
		cartID,
		totalAmount,
		discountAmount,
		finalAmount,
	); err != nil {
		return nil, err
	}

	return s.cartRepo.GetCart(ctx, executor, cartID)
}

func (s *cartService) RemoveCartItem(ctx context.Context, itemID uuid.UUID) (*model.CartWithItems, error) {
	executor := &database.SqlxDBExecutor{DB: s.db.GetDB()}

	// Get cart ID before removing
	items, err := s.cartRepo.GetCartItemByID(ctx, executor, itemID)
	if err != nil {
		return nil, err
	}

	var cartID uuid.UUID
	cartID = items.CartID

	if cartID == uuid.Nil {
		return nil, fmt.Errorf("cart item not found")
	}

	if err := s.cartRepo.RemoveCartItem(ctx, executor, itemID); err != nil {
		return nil, err
	}

	// Get updated cart
	cart, err := s.cartRepo.GetCart(ctx, executor, cartID)
	if err != nil {
		return nil, err
	}

	// Calculate and update totals
	totalAmount, discountAmount, finalAmount, err := s.CalculateCartTotal(ctx, cart)
	if err != nil {
		return nil, err
	}

	if err := s.cartRepo.UpdateCartTotals(ctx, executor, cartID, totalAmount, discountAmount, finalAmount); err != nil {
		return nil, err
	}

	return s.cartRepo.GetCart(ctx, executor, cartID)
}

func (s *cartService) GetCart(ctx context.Context, cartID uuid.UUID) (*model.CartWithItems, error) {
	executor := &database.SqlxDBExecutor{DB: s.db.GetDB()}
	return s.cartRepo.GetCart(ctx, executor, cartID)
}

func (s *cartService) CalculateCartTotal(ctx context.Context, cart *model.CartWithItems) (float64, float64, float64, error) {
	var totalAmount float64
	for _, item := range cart.Items {
		totalAmount += item.TotalPrice
	}

	// TODO: Apply promotions dynamically
	discountAmount := 0.0
	finalAmount := totalAmount - discountAmount

	return totalAmount, discountAmount, finalAmount, nil
}
