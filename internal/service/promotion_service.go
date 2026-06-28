package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/ifortepay/ApiBackendTest/internal/database"
	"github.com/ifortepay/ApiBackendTest/internal/model"
	"github.com/ifortepay/ApiBackendTest/internal/repository"
)

type PromotionService interface {
	GetActivePromotions(ctx context.Context) ([]model.PromotionWithRules, error)
	ApplyPromotions(ctx context.Context, cart *model.CartWithItems) (*PromotionResult, error)
	CalculateDiscounts(ctx context.Context, cart *model.CartWithItems) (map[string]float64, map[string]string, error)
}

type PromotionResult struct {
	Discounts       map[string]float64 `json:"discounts"`  // SKU -> discount amount
	Promotions      map[string]string  `json:"promotions"` // SKU -> promotion name
	FreeItems       map[string]int     `json:"free_items"` // SKU -> quantity
	TotalDiscount   float64            `json:"total_discount"`
	PromotionDetail []PromotionDetail  `json:"promotion_detail"`
}

type PromotionDetail struct {
	PromotionName string   `json:"promotion_name"`
	Description   string   `json:"description"`
	Discount      float64  `json:"discount"`
	AffectedItems []string `json:"affected_items"`
}

type promotionService struct {
	promotionRepo repository.PromotionRepository
	productRepo   repository.ProductRepository
	db            *database.PostgresDB
}

func NewPromotionService(
	promotionRepo repository.PromotionRepository,
	productRepo repository.ProductRepository,
	db *database.PostgresDB,
) PromotionService {
	return &promotionService{
		promotionRepo: promotionRepo,
		productRepo:   productRepo,
		db:            db,
	}
}

func (s *promotionService) GetActivePromotions(ctx context.Context) ([]model.PromotionWithRules, error) {
	executor := &database.SqlxDBExecutor{DB: s.db.GetDB()}
	return s.promotionRepo.GetActivePromotionsWithRules(ctx, executor)
}

func (s *promotionService) ApplyPromotions(ctx context.Context, cart *model.CartWithItems) (*PromotionResult, error) {
	result := &PromotionResult{
		Discounts:       make(map[string]float64),
		Promotions:      make(map[string]string),
		FreeItems:       make(map[string]int),
		PromotionDetail: []PromotionDetail{},
	}

	// Get active promotions
	promotions, err := s.GetActivePromotions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active promotions: %w", err)
	}

	if len(promotions) == 0 {
		return result, nil
	}

	// Group cart items by SKU
	itemMap := make(map[string]int)
	for _, item := range cart.Items {
		itemMap[item.ProductSKU] = item.Quantity
	}

	// Track applied discounts per SKU
	appliedDiscounts := make(map[string]float64)
	appliedPromotions := make(map[string]string)

	// Apply each promotion
	for _, promo := range promotions {
		detail := PromotionDetail{
			PromotionName: promo.Name,
			Description:   promo.Description,
			AffectedItems: []string{},
		}

		for _, rule := range promo.Rules {
			discount, affectedSKUs, err := s.applyPromotionRule(ctx, rule, itemMap, cart)
			if err != nil {
				continue
			}

			if discount != nil && len(discount) > 0 {
				// Apply discount
				for sku, amount := range discount {
					appliedDiscounts[sku] += amount
					appliedPromotions[sku] = promo.Name
					detail.AffectedItems = append(detail.AffectedItems, sku)
					detail.Discount += amount
				}
				result.Discounts = appliedDiscounts
				result.Promotions = appliedPromotions
			}

			// Add affected SKUs from rule
			if len(affectedSKUs) > 0 {
				for _, sku := range affectedSKUs {
					if !contains(detail.AffectedItems, sku) {
						detail.AffectedItems = append(detail.AffectedItems, sku)
					}
				}
			}
		}

		if detail.Discount > 0 {
			result.PromotionDetail = append(result.PromotionDetail, detail)
			result.TotalDiscount += detail.Discount
		}
	}

	return result, nil
}

func (s *promotionService) applyPromotionRule(
	ctx context.Context,
	rule model.PromotionRule,
	itemMap map[string]int,
	cart *model.CartWithItems,
) (map[string]float64, []string, error) {
	discounts := make(map[string]float64)
	affectedSKUs := []string{}

	switch rule.ConditionType {
	case "product_sku":
		// Check if product exists in cart
		sku := rule.ConditionValue
		quantity, exists := itemMap[sku]
		if !exists {
			return nil, nil, nil
		}

		// Find product in cart
		var product *model.CartItemDetail
		for _, item := range cart.Items {
			if item.ProductSKU == sku {
				product = &item
				break
			}
		}

		if product == nil {
			return nil, nil, nil
		}

		disc, affected, err := s.applyAction(rule, product, quantity, cart)
		if err != nil {
			return nil, nil, err
		}
		if disc != nil {
			for k, v := range disc {
				discounts[k] = v
			}
			affectedSKUs = append(affectedSKUs, affected...)
		}
		if len(discounts) > 0 {
			return discounts, affectedSKUs, nil
		}
		return nil, nil, nil

	case "min_quantity":
		// Check if any product meets minimum quantity
		minQty, _ := strconv.Atoi(rule.ConditionValue)
		for _, item := range cart.Items {
			if item.Quantity >= minQty {
				affectedSKUs = append(affectedSKUs, item.ProductSKU)
				disc, affected, err := s.applyAction(rule, &item, item.Quantity, cart)
				if err == nil && disc != nil {
					for k, v := range disc {
						discounts[k] = v
					}
					affectedSKUs = append(affectedSKUs, affected...)
				}
			}
		}
		if len(discounts) > 0 {
			return discounts, affectedSKUs, nil
		}
		return nil, nil, nil

	case "cart_total":
		// Calculate cart total
		var total float64
		for _, item := range cart.Items {
			total += item.TotalPrice
		}
		minAmount, _ := strconv.ParseFloat(rule.ConditionValue, 64)
		if total >= minAmount {
			// Apply to all items
			for _, item := range cart.Items {
				disc, affected, err := s.applyAction(rule, &item, item.Quantity, cart)
				if err == nil && disc != nil {
					for k, v := range disc {
						discounts[k] = v
					}
					affectedSKUs = append(affectedSKUs, affected...)
				}
			}
		}
		if len(discounts) > 0 {
			return discounts, affectedSKUs, nil
		}
		return nil, nil, nil

	default:
		return nil, nil, nil
	}
}

func (s *promotionService) applyAction(
	rule model.PromotionRule,
	item *model.CartItemDetail,
	quantity int,
	cart *model.CartWithItems,
) (map[string]float64, []string, error) {
	discounts := make(map[string]float64)
	affectedSKUs := []string{}

	switch rule.ActionType {
	case "free_product":
		// Free product (used in bundle)
		if rule.TargetProductSKU != nil {
			// Check if target product exists in cart
			for _, cartItem := range cart.Items {
				if cartItem.ProductSKU == *rule.TargetProductSKU {
					// Mark as free by adding to free items
					discount := cartItem.UnitPrice * float64(cartItem.Quantity)
					discounts[*rule.TargetProductSKU] = discount
					affectedSKUs = append(affectedSKUs, *rule.TargetProductSKU)
				}
			}
		}
		if len(discounts) > 0 {
			return discounts, affectedSKUs, nil
		}
		return nil, nil, nil

	case "discount_percentage":
		percentage, _ := strconv.ParseFloat(rule.ActionValue, 64)
		discount := item.UnitPrice * (percentage / 100) * float64(quantity)
		discounts[item.ProductSKU] = discount
		affectedSKUs = append(affectedSKUs, item.ProductSKU)
		return discounts, affectedSKUs, nil

	case "discount_fixed":
		fixedAmount, _ := strconv.ParseFloat(rule.ActionValue, 64)
		discounts[item.ProductSKU] = fixedAmount * float64(quantity)
		affectedSKUs = append(affectedSKUs, item.ProductSKU)
		return discounts, affectedSKUs, nil

	case "buy_x_get_y_free":
		// Format: "3:2" means buy 3 get 2 free
		parts := strings.Split(rule.ActionValue, ":")
		if len(parts) != 2 {
			return nil, nil, fmt.Errorf("invalid buy_x_get_y_free format: %s", rule.ActionValue)
		}
		buyQty, _ := strconv.Atoi(parts[0])
		freeQty, _ := strconv.Atoi(parts[1])

		if quantity >= buyQty {
			freeItems := (quantity / buyQty) * freeQty
			discount := item.UnitPrice * float64(freeItems)
			discounts[item.ProductSKU] = discount
			affectedSKUs = append(affectedSKUs, item.ProductSKU)
		}
		if len(discounts) > 0 {
			return discounts, affectedSKUs, nil
		}
		return nil, nil, nil

	case "bulk_discount":
		// Format: "3:10%" means min 3 items for 10% discount
		parts := strings.Split(rule.ActionValue, ":")
		if len(parts) != 2 {
			return nil, nil, fmt.Errorf("invalid bulk_discount format: %s", rule.ActionValue)
		}
		minQty, _ := strconv.Atoi(parts[0])
		percentage, _ := strconv.ParseFloat(strings.TrimSuffix(parts[1], "%"), 64)

		if quantity >= minQty {
			discount := item.UnitPrice * (percentage / 100) * float64(quantity)
			discounts[item.ProductSKU] = discount
			affectedSKUs = append(affectedSKUs, item.ProductSKU)
		}
		if len(discounts) > 0 {
			return discounts, affectedSKUs, nil
		}
		return nil, nil, nil

	default:
		return nil, nil, fmt.Errorf("unknown action type: %s", rule.ActionType)
	}
}

func (s *promotionService) CalculateDiscounts(ctx context.Context, cart *model.CartWithItems) (map[string]float64, map[string]string, error) {
	result, err := s.ApplyPromotions(ctx, cart)
	if err != nil {
		return nil, nil, err
	}
	return result.Discounts, result.Promotions, nil
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
