package model

import (
	"time"

	"github.com/google/uuid"
)

type Promotion struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	Name        string     `db:"name" json:"name"`
	Description string     `db:"description" json:"description"`
	Type        string     `db:"type" json:"type"`
	IsActive    bool       `db:"is_active" json:"is_active"`
	Priority    int        `db:"priority" json:"priority"`
	StartDate   *time.Time `db:"start_date" json:"start_date"`
	EndDate     *time.Time `db:"end_date" json:"end_date"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

type PromotionRule struct {
	ID               uuid.UUID `db:"id" json:"id"`
	PromotionID      uuid.UUID `db:"promotion_id" json:"promotion_id"`
	ConditionType    string    `db:"condition_type" json:"condition_type"`
	ConditionValue   string    `db:"condition_value" json:"condition_value"`
	ActionType       string    `db:"action_type" json:"action_type"`
	ActionValue      string    `db:"action_value" json:"action_value"`
	TargetProductSKU *string   `db:"target_product_sku" json:"target_product_sku"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

type PromotionWithRules struct {
	Promotion
	Rules []PromotionRule `json:"rules"`
}
