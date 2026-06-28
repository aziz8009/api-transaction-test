package service

import (
	"context"
	"fmt"

	"github.com/ifortepay/ApiBackendTest/internal/database"
	"github.com/ifortepay/ApiBackendTest/internal/model"
	"github.com/ifortepay/ApiBackendTest/internal/repository"
)

type ProductService interface {
	GetProducts(ctx context.Context, filter model.ProductFilter) ([]model.Product, int, error)
	GetProductBySKU(ctx context.Context, sku string) (*model.Product, error)
	ValidateProductsStock(ctx context.Context, items map[string]int) error
}

type productService struct {
	productRepo repository.ProductRepository
	db          *database.PostgresDB
}

func NewProductService(productRepo repository.ProductRepository, db *database.PostgresDB) ProductService {
	return &productService{
		productRepo: productRepo,
		db:          db,
	}
}

func (s *productService) GetProducts(ctx context.Context, filter model.ProductFilter) ([]model.Product, int, error) {
	executor := &database.SqlxDBExecutor{DB: s.db.GetDB()}
	return s.productRepo.GetProducts(ctx, executor, filter)
}

func (s *productService) GetProductBySKU(ctx context.Context, sku string) (*model.Product, error) {
	executor := &database.SqlxDBExecutor{DB: s.db.GetDB()}
	return s.productRepo.GetProductBySKU(ctx, executor, sku)
}

func (s *productService) ValidateProductsStock(ctx context.Context, items map[string]int) error {
	executor := &database.SqlxDBExecutor{DB: s.db.GetDB()}

	skus := make([]string, 0, len(items))
	for sku := range items {
		skus = append(skus, sku)
	}

	products, err := s.productRepo.GetProductsBySKUs(ctx, executor, skus)
	if err != nil {
		return fmt.Errorf("failed to get products: %w", err)
	}

	productMap := make(map[string]model.Product)
	for _, p := range products {
		productMap[p.SKU] = p
	}

	for sku, quantity := range items {
		product, exists := productMap[sku]
		if !exists {
			return fmt.Errorf("product with SKU %s not found", sku)
		}
		if product.StockQuantity < quantity {
			return fmt.Errorf("insufficient stock for product %s: available %d, requested %d", sku, product.StockQuantity, quantity)
		}
	}

	return nil
}
