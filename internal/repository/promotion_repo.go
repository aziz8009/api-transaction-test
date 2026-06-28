package repository

import (
	"context"
	"fmt"

	"github.com/ifortepay/ApiBackendTest/internal/database"
	"github.com/ifortepay/ApiBackendTest/internal/model"
)

type PromotionRepository interface {
	GetActivePromotionsWithRules(ctx context.Context, executor database.DBExecutor) ([]model.PromotionWithRules, error)
	GetPromotionByID(ctx context.Context, executor database.DBExecutor, id string) (*model.PromotionWithRules, error)
}

type promotionRepository struct{}

func NewPromotionRepository() PromotionRepository {
	return &promotionRepository{}
}

func (r *promotionRepository) GetActivePromotionsWithRules(ctx context.Context, executor database.DBExecutor) ([]model.PromotionWithRules, error) {
	// Get active promotions
	var promotions []model.Promotion
	query := `
		SELECT id, name, description, type, is_active, priority, start_date, end_date, created_at, updated_at
		FROM promotions
		WHERE is_active = true
		AND (start_date IS NULL OR start_date <= NOW())
		AND (end_date IS NULL OR end_date >= NOW())
		ORDER BY priority DESC
	`

	if err := executor.SelectContext(ctx, &promotions, query); err != nil {
		return nil, fmt.Errorf("failed to get promotions: %w", err)
	}

	if len(promotions) == 0 {
		return []model.PromotionWithRules{}, nil
	}

	// Get rules for each promotion
	var result []model.PromotionWithRules
	for _, promo := range promotions {
		var rules []model.PromotionRule
		rulesQuery := `
			SELECT id, promotion_id, condition_type, condition_value, action_type, action_value, target_product_sku, created_at, updated_at
			FROM promotion_rules
			WHERE promotion_id = $1
		`

		if err := executor.SelectContext(ctx, &rules, rulesQuery, promo.ID); err != nil {
			continue
		}

		result = append(result, model.PromotionWithRules{
			Promotion: promo,
			Rules:     rules,
		})
	}

	return result, nil
}

func (r *promotionRepository) GetPromotionByID(ctx context.Context, executor database.DBExecutor, id string) (*model.PromotionWithRules, error) {
	var promo model.Promotion
	query := "SELECT id, name, description, type, is_active, priority, start_date, end_date, created_at, updated_at FROM promotions WHERE id = $1"

	if err := executor.GetContext(ctx, &promo, query, id); err != nil {
		return nil, fmt.Errorf("failed to get promotion: %w", err)
	}

	var rules []model.PromotionRule
	rulesQuery := "SELECT id, promotion_id, condition_type, condition_value, action_type, action_value, target_product_sku, created_at, updated_at FROM promotion_rules WHERE promotion_id = $1"

	if err := executor.SelectContext(ctx, &rules, rulesQuery, promo.ID); err != nil {
		return nil, fmt.Errorf("failed to get promotion rules: %w", err)
	}

	return &model.PromotionWithRules{
		Promotion: promo,
		Rules:     rules,
	}, nil
}
