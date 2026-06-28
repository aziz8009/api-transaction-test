package model

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID              uuid.UUID `db:"id" json:"id"`
	OrderNumber     string    `db:"order_number" json:"order_number"`
	UserID          uuid.UUID `db:"user_id" json:"user_id"`
	GrandTotal      float64   `db:"grand_total" json:"grand_total"`
	DiscountTotal   float64   `db:"discount_total" json:"discount_total"`
	Status          string    `db:"status" json:"status"`
	ShippingAddress string    `db:"shipping_address" json:"shipping_address"`
	IdempotencyKey  string    `db:"idempotency_key" json:"-"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}

type OrderItem struct {
	ID               uuid.UUID `db:"id" json:"id"`
	OrderID          uuid.UUID `db:"order_id" json:"order_id"`
	ProductSKU       string    `db:"product_sku" json:"product_sku"`
	Quantity         int       `db:"quantity" json:"quantity"`
	UnitPrice        float64   `db:"unit_price" json:"unit_price"`
	Discount         float64   `db:"discount" json:"discount"`
	FinalPrice       float64   `db:"final_price" json:"final_price"`
	PromotionApplied string    `db:"promotion_applied" json:"promotion_applied"`
	PromotionDetails string    `db:"promotion_details" json:"promotion_details"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

type OrderWithItems struct {
	Order
	Items []OrderItem `json:"items"`
}

type OrderFilter struct {
	Status    string     `query:"status"`
	StartDate *time.Time `query:"start_date"`
	EndDate   *time.Time `query:"end_date"`
	Page      int        `query:"page"`
	Limit     int        `query:"limit"`
}
